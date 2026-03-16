# Issue #80: 탈퇴 후 동일 username 재가입 시 이전 사용자 게시글 노출 버그 수정

## 1. 목적 (Why)

Phase 18(마이그레이션 018)에서 `users` 테이블의 UNIQUE 제약을 partial unique index(`WHERE deleted_at IS NULL`)로 변경하여, 탈퇴(soft-delete)된 계정의 email/username을 새로운 사용자가 재사용할 수 있게 했다. 그러나 `post_repository.go`의 handle(username) 기반 조회 쿼리들이 `u.username = $1` 조건만으로 사용자를 식별하고 있어, soft-deleted 사용자 행까지 매칭된다.

결과적으로 동일 username으로 재가입한 신규 사용자의 프로필 페이지에 이전(탈퇴) 사용자가 작성한 게시글, 답글, 좋아요한 게시글이 함께 노출되는 데이터 유출 버그가 발생한다.

## 2. 버그 재현 시나리오

1. `alice`라는 username으로 가입, 게시글 3개 작성
2. 계정 탈퇴 (users.deleted_at 설정, posts는 그대로 유지)
3. 동일 username `alice`로 신규 가입 (새로운 user ID 부여)
4. 신규 `alice`의 프로필 페이지 진입
5. **기대**: 게시글 0개 (신규 사용자이므로)
6. **실제**: 이전 `alice`의 게시글 3개 노출

## 3. 근본 원인

`LEFT JOIN users u ON p.author_id = u.id` + `WHERE u.username = $1` 쿼리 구조에서, username이 동일한 사용자 행이 2개(soft-deleted + active) 존재하면 양쪽 모두 매칭된다.

- **원글 쿼리**: `u.username = $1` -- 탈퇴 사용자의 author_id로 작성된 게시글도 매칭
- **리포스트 쿼리**: `ru.username = $1` -- 탈퇴 사용자가 리포스트한 게시글도 매칭
- **좋아요 쿼리**: `target.username = $1` -- 탈퇴 사용자가 좋아요한 게시글도 매칭

## 4. 영향 범위

### 4.1 영향받는 API 엔드포인트

| API | 메서드 | 용도 |
|-----|--------|------|
| `GET /api/users/:handle/posts` | ListPostsByHandle | 프로필 > 게시물 탭 |
| `GET /api/users/:handle/replies` | ListRepliesByHandle | 프로필 > 답글 탭 |
| `GET /api/users/:handle/likes` | ListLikedPostsByHandle | 프로필 > 좋아요 탭 |

### 4.2 영향받지 않는 영역

- `FindByID`, `FindAll` 등 ID 기반 조회: username 필터를 사용하지 않으므로 무관
- `FindRepliesByPostID` 등 post ID 기반 조회: 무관
- `FindDeletedByAuthor` 등 trash 관련: author_id(UUID) 기반이므로 무관
- 프론트엔드: 백엔드 쿼리 수정만으로 해결, UI 변경 불필요

## 5. 수정 대상 쿼리 상세

### 5.1 FindByAuthorHandle (line 477)

**원글 서브쿼리** (line 497):
```
현재: WHERE u.username = $1 AND p.parent_id IS NULL
수정: WHERE u.username = $1 AND u.deleted_at IS NULL AND p.parent_id IS NULL
```

**리포스트 서브쿼리** (line 512):
```
현재: WHERE ru.username = $1 AND p.parent_id IS NULL
수정: WHERE ru.username = $1 AND ru.deleted_at IS NULL AND p.parent_id IS NULL
```

### 5.2 FindByAuthorHandleWithUser (line 545)

**원글 서브쿼리** (line 570):
```
현재: WHERE u.username = $1 AND p.parent_id IS NULL AND p.deleted_at IS NULL
수정: WHERE u.username = $1 AND u.deleted_at IS NULL AND p.parent_id IS NULL AND p.deleted_at IS NULL
```

**리포스트 서브쿼리** (line 595):
```
현재: WHERE ru.username = $1 AND p.parent_id IS NULL AND p.deleted_at IS NULL
수정: WHERE ru.username = $1 AND ru.deleted_at IS NULL AND p.parent_id IS NULL AND p.deleted_at IS NULL
```

### 5.3 FindRepliesByAuthorHandle (line 665)

**쿼리** (line 677):
```
현재: WHERE u.username = $1 AND p.parent_id IS NOT NULL
수정: WHERE u.username = $1 AND u.deleted_at IS NULL AND p.parent_id IS NOT NULL
```

### 5.4 FindRepliesByAuthorHandleWithUser (line 689)

**쿼리** (line 704):
```
현재: WHERE u.username = $1 AND p.parent_id IS NOT NULL AND p.deleted_at IS NULL
수정: WHERE u.username = $1 AND u.deleted_at IS NULL AND p.parent_id IS NOT NULL AND p.deleted_at IS NULL
```

### 5.5 FindLikedByUserHandle (line 723)

**쿼리** (line 730-733):
```
현재: JOIN users target ON target.username = $1
수정: JOIN users target ON target.username = $1 AND target.deleted_at IS NULL
```

### 5.6 FindLikedByUserHandleWithViewer (line 745)

**쿼리** (line 755):
```
현재: JOIN users target ON target.username = $1
수정: JOIN users target ON target.username = $1 AND target.deleted_at IS NULL
```

## 6. 수정 방법 요약

총 **8곳** 수정:

| # | 메서드 | 별칭 | 추가 조건 |
|---|--------|------|-----------|
| 1 | FindByAuthorHandle | u (원글) | `AND u.deleted_at IS NULL` |
| 2 | FindByAuthorHandle | ru (리포스트) | `AND ru.deleted_at IS NULL` |
| 3 | FindByAuthorHandleWithUser | u (원글) | `AND u.deleted_at IS NULL` |
| 4 | FindByAuthorHandleWithUser | ru (리포스트) | `AND ru.deleted_at IS NULL` |
| 5 | FindRepliesByAuthorHandle | u | `AND u.deleted_at IS NULL` |
| 6 | FindRepliesByAuthorHandleWithUser | u | `AND u.deleted_at IS NULL` |
| 7 | FindLikedByUserHandle | target | `AND target.deleted_at IS NULL` (JOIN 조건에 추가) |
| 8 | FindLikedByUserHandleWithViewer | target | `AND target.deleted_at IS NULL` (JOIN 조건에 추가) |

수정 패턴은 일관적이다:
- WHERE 절의 `u.username = $1` 또는 `ru.username = $1` 옆에 `AND [alias].deleted_at IS NULL` 추가
- JOIN 절의 `target.username = $1` 옆에 `AND target.deleted_at IS NULL` 추가

## 7. 수락 기준 (Acceptance Criteria)

1. 탈퇴한 사용자와 동일 username으로 재가입한 사용자의 프로필 페이지에서 이전 사용자의 게시글이 **표시되지 않아야** 한다
2. 탈퇴한 사용자와 동일 username으로 재가입한 사용자의 프로필 페이지에서 이전 사용자의 답글이 **표시되지 않아야** 한다
3. 탈퇴한 사용자와 동일 username으로 재가입한 사용자의 프로필 페이지에서 이전 사용자가 좋아요한 게시글이 **표시되지 않아야** 한다
4. 탈퇴하지 않은 일반 사용자의 프로필 조회는 기존과 동일하게 동작해야 한다
5. `go test ./...` 전체 통과
6. `bun run check` 통과 (프론트엔드 변경 없으므로 기존 통과 상태 유지)

## 8. 테스트 시나리오

### 8.1 단위 테스트 (Repository 레벨)

기존 mock 기반 테스트는 쿼리 자체를 검증하지 않으므로, 이 버그에 대한 통합 테스트 또는 수동 검증이 필요하다.

### 8.2 수동 검증 시나리오

| # | 시나리오 | 검증 항목 | 기대 결과 |
|---|---------|-----------|-----------|
| 1 | 사용자 A 가입 > 게시글 작성 > 탈퇴 > 동일 username으로 B 가입 | B 프로필 > 게시물 탭 | 빈 목록 |
| 2 | 위와 동일 | B 프로필 > 답글 탭 | 빈 목록 |
| 3 | 사용자 A 가입 > 게시글 좋아요 > 탈퇴 > 동일 username으로 B 가입 | B 프로필 > 좋아요 탭 | 빈 목록 |
| 4 | 사용자 A 가입 > 게시글 리포스트 > 탈퇴 > 동일 username으로 B 가입 | B 프로필 > 게시물 탭 | A의 리포스트 미노출 |
| 5 | 일반 사용자 (탈퇴 이력 없음) | 프로필 > 모든 탭 | 기존과 동일 |
| 6 | 비로그인 상태에서 프로필 조회 | 게시물/답글/좋아요 탭 | 탈퇴 사용자 데이터 미노출 |

## 9. 위험 요소 및 주의사항

### 9.1 성능 영향
- `u.deleted_at IS NULL` 조건 추가는 기존 `idx_users_username_active` partial unique index와 일치하므로 인덱스 활용 가능. 성능 저하 없음.
- `target.deleted_at IS NULL`도 마찬가지로 username 검색 시 partial index가 커버.

### 9.2 기존 동작 변화
- 탈퇴 사용자의 게시글은 이미 `author_deleted = true`로 표시되어 익명화 렌더링 중이었으나, 동일 username 재가입 시 신규 사용자 프로필에 혼입되는 것이 문제였다.
- 이 수정 후에도 피드(FindAll/FindAllWithUser)에서는 탈퇴 사용자 게시글이 `author_deleted = true`로 계속 표시된다 (이는 의도된 동작).

### 9.3 추가 확인 필요 사항
- `user_repository.go`의 `FindByUsername` 메서드에 이미 `deleted_at IS NULL` 필터가 있는지 확인 필요. Service 레이어에서 사용자 존재 여부를 먼저 확인하는 경우 해당 메서드가 이미 활성 사용자만 반환한다면 이중 보호가 된다.
- 단, Repository 쿼리 자체에서 방어하는 것이 올바른 접근이다 (깊은 방어, defense in depth).

## 10. 의존성 및 제약사항

- **선행 조건**: Phase 18 (마이그레이션 018) 적용 완료 상태
- **수정 파일**: `backend/internal/repository/post_repository.go` 1개 파일만 수정
- **DB 마이그레이션**: 불필요 (쿼리 로직 수정만)
- **프론트엔드 변경**: 불필요
- **하위 호환성**: 완전 호환 (기존 API 인터페이스 변경 없음)

## 11. 구현 권장사항

1. `post_repository.go`에서 8곳의 쿼리를 위 표대로 수정
2. `go test ./...` 실행하여 기존 테스트 통과 확인
3. Docker 환경에서 수동 검증 시나리오 8.2 수행
4. PR 생성 시 이슈 #80 연결

---

**다음 단계**: Backend Agent가 `post_repository.go` 수정 수행
