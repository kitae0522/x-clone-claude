# Issue #68: 탈퇴 시 좋아요 Soft Delete 처리

## 개요

### 문제 (What)
사용자가 탈퇴(soft delete)해도 해당 사용자의 `likes` 레코드가 그대로 남아 있어, 게시글의 `like_count`가 실제 활성 사용자의 좋아요 수보다 부풀려진 상태로 유지된다. 또한 프로필 탭의 "좋아요한 게시글" 목록에서 탈퇴한 사용자의 좋아요가 조회 결과에 영향을 준다.

### 해결 방안 (Why)
기존 `users.deleted_at`, `posts.deleted_at` 패턴과 동일하게 `likes` 테이블에도 `deleted_at` 컬럼을 추가하여 soft delete를 적용한다. 사용자 탈퇴 시 해당 사용자의 모든 좋아요를 soft delete 처리하고, 관련 게시글의 `like_count`를 감소시킨다.

### 영향 범위
- DB 마이그레이션: `likes.deleted_at` 컬럼 추가
- `LikeRepository`: 전체 쿼리에 `deleted_at IS NULL` 필터 추가 + soft delete 메서드
- `UserService.DeleteAccount`: 좋아요 soft delete 로직 추가
- `PostRepository`: `is_liked` 서브쿼리에 `deleted_at IS NULL` 필터 추가
- `BookmarkRepository`: `is_liked` 서브쿼리에 `deleted_at IS NULL` 필터 추가
- `like_count` 일괄 감소 처리

### 관련 이슈
- Related to #56 (계정 탈퇴)
- Related to #67 (삭제된 게시글 접근 제어)

---

## 1. DB 마이그레이션

### 파일: `backend/migrations/020_add_deleted_at_to_likes.up.sql`

```sql
-- likes 테이블에 soft delete 컬럼 추가
ALTER TABLE likes ADD COLUMN deleted_at TIMESTAMPTZ;

-- 활성 좋아요만 조회하기 위한 부분 인덱스
CREATE INDEX idx_likes_active ON likes(user_id, post_id) WHERE deleted_at IS NULL;

-- soft delete된 좋아요 조회용 인덱스 (사용자별 일괄 처리 시)
CREATE INDEX idx_likes_user_deleted ON likes(user_id) WHERE deleted_at IS NOT NULL;

-- 기존 PK (user_id, post_id)는 유지한다.
-- soft delete 후 동일 사용자가 다시 좋아요하는 시나리오는 "탈퇴 사용자"이므로 발생하지 않는다.
-- 만약 향후 좋아요 취소/재좋아요를 soft delete로 전환한다면 PK를 재설계해야 하나,
-- 이번 스코프에서는 탈퇴 시에만 soft delete를 사용하므로 PK 변경 불필요.
```

### 파일: `backend/migrations/020_add_deleted_at_to_likes.down.sql`

```sql
DROP INDEX IF EXISTS idx_likes_user_deleted;
DROP INDEX IF EXISTS idx_likes_active;
ALTER TABLE likes DROP COLUMN IF EXISTS deleted_at;
```

---

## 2. 백엔드 변경사항

### 2-1. Model 수정

**파일: `backend/internal/model/like.go`**

```go
type Like struct {
    UserID    uuid.UUID
    PostID    uuid.UUID
    CreatedAt time.Time
    DeletedAt *time.Time  // 추가
}
```

### 2-2. LikeRepository 변경

**파일: `backend/internal/repository/like_repository.go`**

인터페이스에 메서드 추가:

```go
type LikeRepository interface {
    Like(ctx context.Context, userID, postID uuid.UUID) error
    Unlike(ctx context.Context, userID, postID uuid.UUID) error
    IsLiked(ctx context.Context, userID, postID uuid.UUID) (bool, error)
    SoftDeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error)  // 추가: 탈퇴 시 일괄 soft delete
}
```

#### 기존 쿼리 수정

1. **`IsLiked`**: `deleted_at IS NULL` 필터 추가
   ```sql
   SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND post_id = $2 AND deleted_at IS NULL)
   ```

2. **`Like` (INSERT)**: `ON CONFLICT DO NOTHING`은 PK 충돌 시 동작하므로, soft delete된 좋아요가 있는 경우를 처리해야 한다. 그러나 탈퇴 사용자가 다시 좋아요하는 것은 불가능하므로(탈퇴 사용자는 로그인 불가), 현재 INSERT 로직은 변경 불필요.

3. **`Unlike` (DELETE -> UPDATE)**: 현재 `DELETE FROM likes`를 사용하고 있다. 일반 unlike는 기존대로 hard delete(DELETE)를 유지한다. soft delete는 탈퇴 시에만 사용한다.
   - **결정**: Unlike는 기존 hard delete 유지. 이유: 일반 unlike는 사용자가 의도적으로 좋아요를 취소하는 행위이므로, 데이터 보존 필요 없음. soft delete는 탈퇴라는 시스템 이벤트에서만 사용.

#### 신규 메서드

```go
// SoftDeleteByUserID: 탈퇴 시 해당 사용자의 모든 활성 좋아요를 soft delete
// 반환값: soft delete된 행 수 (like_count 감소에 사용)
func (r *likeRepository) SoftDeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
    // 트랜잭션 내에서:
    // 1. soft delete 대상 게시글 ID 목록 조회
    // 2. likes.deleted_at = NOW() 일괄 업데이트
    // 3. 각 게시글의 like_count 감소
}
```

**상세 구현 SQL**:

```sql
-- 트랜잭션 시작

-- Step 1: soft delete 대상 좋아요의 post_id 목록과 각 post별 좋아요 수 집계
WITH deleted_likes AS (
    UPDATE likes
    SET deleted_at = NOW()
    WHERE user_id = $1 AND deleted_at IS NULL
    RETURNING post_id
)
UPDATE posts p
SET like_count = GREATEST(like_count - sub.cnt, 0)
FROM (
    SELECT post_id, COUNT(*) AS cnt
    FROM deleted_likes
    GROUP BY post_id
) sub
WHERE p.id = sub.post_id;

-- 트랜잭션 커밋
```

**설계 포인트**:
- CTE(WITH)를 사용하여 단일 쿼리로 soft delete + like_count 감소를 원자적으로 처리
- `GREATEST(like_count - cnt, 0)`으로 음수 방지
- 이미 soft delete된 좋아요는 `deleted_at IS NULL` 조건으로 제외

### 2-3. PostRepository 변경

**파일: `backend/internal/repository/post_repository.go`**

`is_liked` 서브쿼리가 포함된 모든 SELECT 쿼리에 `deleted_at IS NULL` 필터 추가.

현재 영향 받는 쿼리 위치 (줄 번호 기준):
- L161: `FindByID` (with viewer)
- L201, L225: `ListFeed` 관련
- L378, L409: profile handle 조회
- L562, L585: replies handle 조회
- L694, L750: liked posts handle 조회

수정 패턴:
```sql
-- 변경 전
EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $X AND l.post_id = p.id) AS is_liked

-- 변경 후
EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $X AND l.post_id = p.id AND l.deleted_at IS NULL) AS is_liked
```

총 수정 대상: **약 10개** 쿼리

추가로 `FindLikedByUserHandle` / `FindLikedByUserHandleWithViewer` 쿼리에서 `FROM likes lk` JOIN 조건에도 필터 추가:
```sql
-- 변경 전
WHERE lk.user_id = target.id

-- 변경 후
WHERE lk.user_id = target.id AND lk.deleted_at IS NULL
```

### 2-4. BookmarkRepository 변경

**파일: `backend/internal/repository/bookmark_repository.go`**

L79의 `is_liked` 서브쿼리에 동일한 `deleted_at IS NULL` 필터 추가:
```sql
EXISTS(SELECT 1 FROM likes l WHERE l.user_id = $1 AND l.post_id = p.id AND l.deleted_at IS NULL) AS is_liked
```

### 2-5. UserService.DeleteAccount 변경

**파일: `backend/internal/service/user_service.go`**

`userService` 구조체에 `LikeRepository` 의존성 추가:

```go
type userService struct {
    userRepo   repository.UserRepository
    followRepo repository.FollowRepository
    likeRepo   repository.LikeRepository  // 추가
}

func NewUserService(
    userRepo repository.UserRepository,
    followRepo repository.FollowRepository,
    likeRepo repository.LikeRepository,  // 추가
) UserService {
    return &userService{userRepo: userRepo, followRepo: followRepo, likeRepo: likeRepo}
}
```

`DeleteAccount` 메서드 수정:

```go
func (s *userService) DeleteAccount(ctx context.Context, userID uuid.UUID, req dto.DeleteAccountRequest) error {
    // ... 기존 비밀번호 검증 ...

    // 좋아요 soft delete (like_count 감소 포함)
    if _, err := s.likeRepo.SoftDeleteByUserID(ctx, userID); err != nil {
        return apperror.Internal("failed to soft delete likes")
    }

    // 사용자 soft delete
    if err := s.userRepo.SoftDelete(ctx, userID); err != nil {
        return apperror.Internal("failed to delete account")
    }

    return nil
}
```

**순서가 중요한 이유**: 좋아요 soft delete를 사용자 soft delete보다 먼저 실행해야 한다. 사용자가 먼저 soft delete되면 `FindByID`에서 조회 불가능해지므로, 이후 로직이 영향받을 수 있다. 다만 현재 `SoftDeleteByUserID`는 user_id로 직접 likes 테이블에 접근하므로 순서 무관하지만, 안전성을 위해 좋아요 먼저 처리한다.

### 2-6. DI 모듈 수정

**파일: `backend/internal/module/` (해당 모듈 파일)**

`NewUserService` 호출부에 `LikeRepository` 인자를 추가해야 한다. uber-go/fx 기반이므로 자동 주입될 것이나, 함수 시그니처 변경에 따른 업데이트 필요.

---

## 3. 프론트엔드 변경사항

프론트엔드는 변경 불필요.

이유:
- 좋아요/좋아요 취소 API 응답 형식 변경 없음
- `like_count`는 백엔드에서 정확하게 반영된 값을 반환
- `is_liked` 플래그도 백엔드에서 `deleted_at IS NULL` 필터를 적용하므로 정확

---

## 4. 수락 기준 (Acceptance Criteria)

1. **AC-1**: `likes` 테이블에 `deleted_at TIMESTAMPTZ` nullable 컬럼이 추가된다.
2. **AC-2**: 사용자 탈퇴 시 해당 사용자의 모든 좋아요 레코드에 `deleted_at`이 설정된다.
3. **AC-3**: 사용자 탈퇴 시 soft delete된 좋아요의 수만큼 해당 게시글들의 `like_count`가 감소한다.
4. **AC-4**: `like_count`는 0 미만이 되지 않는다 (`GREATEST(..., 0)`).
5. **AC-5**: `LikeRepository.IsLiked`는 `deleted_at IS NULL`인 좋아요만 조회한다.
6. **AC-6**: 모든 `is_liked` 서브쿼리(PostRepository 10곳, BookmarkRepository 1곳)에 `deleted_at IS NULL` 필터가 적용된다.
7. **AC-7**: 프로필 "좋아요한 게시글" 탭에서 탈퇴한 사용자의 좋아요가 제외된다 (`FindLikedByUserHandle` 쿼리의 `lk.deleted_at IS NULL`).
8. **AC-8**: 일반 Unlike(사용자가 직접 좋아요 취소)는 기존 hard delete 방식을 유지한다.
9. **AC-9**: 마이그레이션 롤백(down)이 정상 동작한다.
10. **AC-10**: 기존 테스트가 모두 통과한다 (mock 인터페이스 업데이트 포함).

---

## 5. 엣지 케이스

| # | 시나리오 | 기대 동작 |
|---|---------|----------|
| E-1 | 탈퇴 사용자가 좋아요한 게시글이 이미 삭제(soft delete)된 경우 | `like_count` 감소 대상에서 자연스럽게 제외 (게시글은 이미 조회 안됨). 하지만 `likes` soft delete와 `posts.like_count` 감소는 여전히 실행됨. 삭제된 게시글이 복원되면 정확한 카운트 반영. |
| E-2 | 동일 사용자가 탈퇴 후 재가입하여 같은 게시글에 좋아요 시도 | 기존 좋아요는 `deleted_at`이 설정되어 있고, PK(user_id, post_id) 충돌 발생. 새 계정은 새 UUID를 받으므로 PK 충돌 없음. 즉, 문제 없음. |
| E-3 | 좋아요가 0인 게시글에 대해 like_count 감소 시도 | `GREATEST(like_count - cnt, 0)` 으로 음수 방지. |
| E-4 | 탈퇴 처리 중 DB 에러 발생 (좋아요 soft delete 성공, 사용자 soft delete 실패) | 각각 별도 트랜잭션이므로 좋아요만 soft delete된 상태가 될 수 있음. 하지만 좋아요가 soft delete되어도 사용자 자체는 활성 상태이므로, 다시 좋아요를 취소/추가하는 데 문제 없음. 이상적으로는 전체를 하나의 트랜잭션으로 감싸야 하지만, 현재 아키텍처(Repository별 트랜잭션)에서는 과도한 변경. 향후 개선 과제로 등록. |
| E-5 | 탈퇴 사용자의 좋아요가 수천 건인 경우 | CTE 기반 단일 쿼리로 처리하므로 N+1 없음. 대량 UPDATE이므로 잠금 경합 가능성 있으나, MVP 수준에서는 허용. |
| E-6 | `is_liked` 서브쿼리에서 탈퇴 사용자 본인이 viewer인 경우 | 탈퇴 사용자는 로그인 불가이므로 viewer가 될 수 없음. 문제 없음. |

---

## 6. 의존성 및 제약사항

### 의존성
- Phase 17 완료 필수 (UserService.DeleteAccount 존재)
- Phase 12 완료 필수 (posts.deleted_at soft delete 패턴)

### 제약사항
- 마이그레이션 번호 020 사용 (019까지 존재 확인)
- `LikeRepository` 인터페이스 변경으로 모든 mock 업데이트 필요
- `NewUserService` 시그니처 변경으로 DI 모듈 및 테스트 mock 초기화 업데이트 필요

### 향후 개선 과제 (Out of Scope)
- 탈퇴 프로세스 전체를 단일 트랜잭션으로 묶기 (좋아요 soft delete + 사용자 soft delete)
- 북마크/리포스트도 동일한 soft delete 패턴 적용 검토
- 주기적 데이터 정리(purge): `deleted_at`이 30일 이상 된 likes 레코드 hard delete

---

## 7. 테스트 계획

### 7-1. Repository 테스트 (Go)

| # | 테스트 케이스 | 검증 |
|---|-------------|------|
| T-1 | `SoftDeleteByUserID` - 사용자의 좋아요 2건 soft delete | `deleted_at` 설정, 영향 행 수 = 2 |
| T-2 | `SoftDeleteByUserID` - 좋아요 없는 사용자 | 에러 없이 0 반환 |
| T-3 | `SoftDeleteByUserID` - like_count 감소 확인 | 관련 게시글의 like_count가 정확히 감소 |
| T-4 | `IsLiked` - soft delete된 좋아요는 false 반환 | `deleted_at`이 설정된 레코드 무시 |

### 7-2. Service 테스트 (Go, Table-Driven)

| # | 테스트 케이스 | 검증 |
|---|-------------|------|
| T-5 | `DeleteAccount` - 좋아요 soft delete 후 사용자 soft delete | `SoftDeleteByUserID` 호출 확인, `SoftDelete` 호출 확인 |
| T-6 | `DeleteAccount` - 좋아요 soft delete 실패 시 에러 반환 | 500 에러, 사용자 soft delete 미실행 |
| T-7 | 기존 `Like`/`Unlike` 동작 변경 없음 | 기존 테스트 통과 |

### 7-3. Mock 업데이트

- `LikeRepository` mock에 `SoftDeleteByUserID` 메서드 추가
- `NewUserService` 호출부에 `likeRepo` mock 인자 추가
- 기존 `user_service_test.go`의 `NewUserService` 호출 모두 업데이트

### 7-4. 통합 검증 (수동)

1. Seed data 기반으로 사용자 탈퇴 실행
2. 해당 사용자가 좋아요한 게시글의 `like_count` 감소 확인
3. 프로필 탭 "좋아요" 목록에서 탈퇴 사용자 좋아요 미표시 확인
4. 마이그레이션 up/down 반복 검증

---

## 8. 구현 순서 권장

1. DB 마이그레이션 작성 및 적용
2. `model/like.go` - `DeletedAt` 필드 추가
3. `LikeRepository` - `SoftDeleteByUserID` 메서드 구현
4. `LikeRepository` - `IsLiked` 쿼리에 `deleted_at IS NULL` 추가
5. `PostRepository` - 모든 `is_liked` 서브쿼리에 `deleted_at IS NULL` 추가 (10곳)
6. `BookmarkRepository` - `is_liked` 서브쿼리에 `deleted_at IS NULL` 추가 (1곳)
7. `PostRepository` - `FindLikedByUserHandle*` 쿼리에 `lk.deleted_at IS NULL` 추가 (2곳)
8. `UserService` - `LikeRepository` 의존성 추가 + `DeleteAccount` 수정
9. DI 모듈 업데이트
10. 테스트 mock 업데이트 + 새 테스트 작성
11. `schema.md` 업데이트
12. 전체 빌드 및 테스트 통과 확인
