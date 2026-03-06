# Spec: 게시물 공개 범위(Visibility) 설정

- **작성일**: 2026-03-06
- **상태**: Draft

---

## 1. 개요

### 1.1 What

게시물 작성 시 공개 범위를 선택할 수 있는 기능을 추가한다. 세 가지 공개 범위를 지원한다.

1. **Public (공개)**: 모든 사용자에게 표시
2. **Friends (팔로워 전용)**: 작성자를 팔로우하는 사용자 + 작성자 본인에게만 표시
3. **Private (비공개)**: 작성자 본인에게만 표시

현재 백엔드 모델/DTO/DB에는 visibility 필드가 이미 존재하지만, 프론트엔드 UI에서 선택할 수 없고(하드코딩 `"public"`), 피드 쿼리에서 visibility 기반 필터링이 전혀 없어 **보안 취약점**이 존재한다.

### 1.2 Why

- **보안 문제**: 현재 `friends`나 `private`로 저장된 게시물도 피드에서 모든 사용자에게 노출됨. DB에 visibility 값이 저장되지만 조회 시 무시되고 있음
- **기능 완성도**: X(Twitter)의 "팔로워 전용 공개"와 유사한 핵심 프라이버시 기능
- **기존 인프라 활용**: DB 컬럼, 모델, DTO가 모두 준비되어 있어 구현 비용이 낮음

### 1.3 핵심 제약

- DB 마이그레이션 불필요 (visibility 컬럼 이미 존재)
- 백엔드 모델/DTO 변경 불필요 (이미 정의됨)
- `follows` 테이블이 이미 존재하며, `FollowRepository.IsFollowing()` 메서드 사용 가능
- 답글(reply)의 visibility는 항상 `public` 고정 (현재 CreateReply에서 하드코딩)

---

## 2. 설계 결정 사항

### 2.1 "Friends" 범위의 의미

**결정**: `friends` visibility는 **작성자를 팔로우하는 사용자(팔로워)**에게만 표시한다. 맞팔(상호 팔로우)이 아닌 단방향 팔로우 기준이다.

**이유**:
- X(Twitter)의 "비공개 계정" 모델과 유사하게, 팔로워에게 공개하는 방식이 자연스러움
- 맞팔 기준으로 하면 팔로우를 끊는 순간 양방향 모두 접근 불가가 되어 예측 불가능한 UX
- `follows` 테이블의 `(follower_id, following_id)` 구조에서 `following_id = author_id`인 follower를 찾으면 됨

**트레이드오프**: "친구"라는 표현이 맞팔을 암시할 수 있으나, UI에서 "팔로워 전용"으로 표시하여 혼동 방지

### 2.2 Visibility 필터링 위치

**결정**: **Repository 레이어의 SQL WHERE 절**에서 필터링하고, Service 레이어에서 viewerID 기반으로 적절한 Repository 메서드를 호출한다.

**이유**:
- 애플리케이션 레벨 필터링은 불필요한 데이터 전송 발생
- SQL에서 follows JOIN으로 한 번에 처리하면 추가 쿼리 없이 해결 가능
- CLAUDE.md의 ReBAC 지시사항에 따라 Service 레이어에서 관계 검증 책임

**트레이드오프**: Repository 쿼리가 복잡해지지만, N+1 문제 없이 단일 쿼리로 해결

### 2.3 비인증 사용자 처리

**결정**: 비인증 사용자(viewerID == nil)에게는 `public` 게시물만 표시한다.

**이유**:
- 팔로우 관계를 확인할 수 없으므로 `friends` 게시물 접근 불가
- 안전한 기본값으로 정보 노출 최소화

### 2.4 답글(Reply)의 Visibility

**결정**: 답글은 항상 `public`으로 유지한다. ComposeForm의 visibility 선택기는 답글 작성 시 표시하지 않는다.

**이유**:
- 답글에 visibility를 적용하면 대화 흐름이 끊어지는 UX 문제 발생 (일부 답글만 보이는 상황)
- X(Twitter)에서도 답글의 공개 범위는 원본 게시물의 설정을 따름
- 현재 `CreateReply`에서 이미 `VisibilityPublic` 하드코딩됨

---

## 3. Phase별 상세 구현 계획

### Phase A: 프론트엔드 - ComposeForm 공개 범위 선택 UI

#### 3.1 변경 파일

| 파일 | 변경 내용 |
|------|----------|
| `frontend/src/components/ComposeForm.tsx` | visibility state 추가, 선택 드롭다운 UI, mutate 호출 시 동적 visibility 전달 |
| `frontend/src/components/VisibilitySelector.tsx` | **신규** - 공개 범위 선택 컴포넌트 |

#### 3.2 VisibilitySelector 컴포넌트 설계

```
Props:
  value: 'public' | 'friends' | 'private'
  onChange: (value: 'public' | 'friends' | 'private') => void
  disabled?: boolean
```

UI 구성:
- 현재 선택된 공개 범위를 아이콘 + 텍스트로 표시하는 버튼
- 클릭 시 드롭다운 메뉴 표시 (shadcn/ui의 DropdownMenu 사용)
- 옵션 3개:
  - Globe 아이콘 + "전체 공개" + 설명 "모든 사용자가 볼 수 있습니다"
  - Users 아이콘 + "팔로워 전용" + 설명 "나를 팔로우하는 사용자만 볼 수 있습니다"
  - Lock 아이콘 + "나만 보기" + 설명 "나만 볼 수 있습니다"
- 선택 시 드롭다운 닫힘, 버튼 텍스트 업데이트
- 위치: 텍스트 입력 영역과 하단 툴바 사이 (Textarea 아래, 미디어 프리뷰 위)

#### 3.3 ComposeForm 변경사항

1. `useState`로 `visibility` state 추가 (기본값: `"public"`)
2. line 82의 하드코딩 `visibility: "public"`을 `visibility` state로 교체
3. `VisibilitySelector` 컴포넌트를 Textarea 아래에 배치
4. 게시 성공 시 `visibility`를 `"public"`으로 리셋
5. 답글 작성 폼(ReplyForm)에서는 `VisibilitySelector`를 표시하지 않음

---

### Phase B: 백엔드 - Visibility 기반 피드 필터링 (보안)

#### 3.4 변경 파일

| 파일 | 변경 내용 |
|------|----------|
| `backend/internal/repository/post_repository.go` | `FindAll`, `FindAllWithUser`, `FindByID`, `FindByIDWithUser`, handle 기반 조회 메서드들에 visibility 필터 WHERE 조건 추가 |
| `backend/internal/service/post_service.go` | `GetPostByID`에서 visibility 기반 접근 권한 검증 추가 |

#### 3.5 Repository 쿼리 변경

**핵심 원칙**: 기존 쿼리의 WHERE 절에 visibility 필터 조건을 추가한다.

**(a) 비인증 사용자용 쿼리 (`FindAll`, `FindByID` 등)**

기존:
```sql
WHERE p.parent_id IS NULL
```

변경:
```sql
WHERE p.parent_id IS NULL
  AND p.visibility = 'public'
```

비인증 사용자에게는 public 게시물만 노출한다.

**(b) 인증 사용자용 쿼리 (`FindAllWithUser`, `FindByIDWithUser` 등)**

기존:
```sql
WHERE p.parent_id IS NULL
```

변경:
```sql
WHERE p.parent_id IS NULL
  AND (
    p.visibility = 'public'
    OR (p.visibility = 'friends' AND (
      p.author_id = $viewerID
      OR EXISTS (
        SELECT 1 FROM follows f
        WHERE f.follower_id = $viewerID AND f.following_id = p.author_id
      )
    ))
    OR (p.visibility = 'private' AND p.author_id = $viewerID)
  )
```

설명:
- `public`: 모든 인증 사용자에게 표시
- `friends`: 작성자 본인이거나 작성자를 팔로우하는 사용자에게 표시
- `private`: 작성자 본인에게만 표시

**(c) 적용 대상 메서드 목록**

| 메서드 | 필터 타입 | 비고 |
|--------|----------|------|
| `FindAll` | public only | 비인증 |
| `FindAllWithUser` | full visibility filter | 인증 |
| `FindByID` | public only | 비인증 |
| `FindByIDWithUser` | full visibility filter | 인증 |
| `FindByAuthorHandle` | public only | 비인증 |
| `FindByAuthorHandleWithUser` | full visibility filter | 인증 |
| `FindRepliesByAuthorHandle` | public only (답글의 부모 게시물 기준) | 비인증 |
| `FindRepliesByAuthorHandleWithUser` | full visibility filter | 인증 |
| `FindLikedByUserHandle` | public only | 비인증 |
| `FindLikedByUserHandleWithViewer` | full visibility filter | 인증 |
| `FindRepliesByPostID` | 변경 없음 | 답글 자체는 항상 public |
| `FindRepliesByPostIDWithUser` | 변경 없음 | 답글 자체는 항상 public |
| `FindAuthorReplyByPostID` | 변경 없음 | 내부 thread chain용 |
| `FindAuthorReplyByPostIDWithUser` | 변경 없음 | 내부 thread chain용 |

**(d) 프로필 페이지 자기 게시물 조회 특수 처리**

프로필 페이지에서 자기 게시물을 볼 때는 모든 visibility의 게시물이 표시되어야 한다. `FindByAuthorHandleWithUser`에서 viewer가 작성자 본인인 경우 이미 위의 visibility 필터 조건으로 자동 처리된다 (`p.author_id = $viewerID` 조건이 friends와 private 모두 통과).

#### 3.6 Service 레이어 변경

**`GetPostByID` 접근 제어 강화**

Repository에서 이미 visibility 필터를 적용하지만, `GetPostByID`는 단일 게시물 직접 접근이므로 Service 레이어에서 추가 검증한다.

```
로직:
1. Repository에서 게시물 조회 (visibility 필터 없이 FindByID 또는 FindByIDWithUser)
2. 조회 결과의 visibility 확인:
   - public: 통과
   - friends: viewer가 nil이면 NotFound 반환.
             viewer가 작성자 본인이면 통과.
             viewer가 작성자의 팔로워인지 FollowRepository.IsFollowing() 확인. 아니면 NotFound 반환.
   - private: viewer가 nil이면 NotFound 반환.
             viewer가 작성자 본인이면 통과.
             아니면 NotFound 반환.
3. 권한 없는 경우 "post not found" (NotFound)를 반환 (Forbidden이 아닌 NotFound로 게시물 존재 여부 자체를 숨김)
```

**의존성 추가**: `postService`에 `FollowRepository` 의존성 추가 필요.

```go
type postService struct {
    postRepo   repository.PostRepository
    pollRepo   repository.PollRepository
    mediaRepo  repository.MediaRepository
    followRepo repository.FollowRepository  // 추가
}
```

`NewPostService` 생성자에도 `FollowRepository` 파라미터 추가. DI(fx) 모듈에서 자동 주입.

---

### Phase C: 프론트엔드 - PostCard/PostDetailPage 공개 범위 표시

#### 3.7 변경 파일

| 파일 | 변경 내용 |
|------|----------|
| `frontend/src/components/PostCard.tsx` | visibility 아이콘 표시 (public이 아닌 경우) |
| `frontend/src/components/VisibilityBadge.tsx` | **신규** - visibility 아이콘 표시 컴포넌트 |

#### 3.8 VisibilityBadge 컴포넌트 설계

```
Props:
  visibility: 'public' | 'friends' | 'private'
```

동작:
- `public`인 경우: 아무것도 렌더링하지 않음 (대부분의 게시물이 public이므로 노이즈 방지)
- `friends`인 경우: Users 아이콘 (12px) + "팔로워 전용" 텍스트 (text-muted-foreground, 13px)
- `private`인 경우: Lock 아이콘 (12px) + "나만 보기" 텍스트 (text-muted-foreground, 13px)

위치: PostCard에서 작성자 이름 행 우측, 시간 옆에 표시. 또는 Location과 동일한 패턴으로 콘텐츠 위에 별도 행으로 표시.

**권장 위치**: 작성자 행의 시간 뒤, `·` 구분자 후 아이콘 표시. PostCard의 author row 내에 배치하여 자연스럽게 메타데이터로 인식.

```
@username · 5분 전 · [Lock] 나만 보기
```

PostDetailPage에서도 동일한 VisibilityBadge를 사용한다. PostDetailPage의 상세 정보 영역에 배치.

---

## 4. API 변경사항

### 4.1 기존 API 엔드포인트 (변경 없음)

API 엔드포인트 자체의 변경은 없다. 요청/응답 형식도 동일하다.

| 메서드 | 경로 | 변경 |
|--------|------|------|
| `POST /api/posts` | 없음 (이미 visibility 필드 지원) |
| `GET /api/posts` | 없음 (응답에 이미 visibility 포함) |
| `GET /api/posts/:id` | 없음 (응답에 이미 visibility 포함) |
| `GET /api/users/:handle/posts` | 없음 |
| `GET /api/users/:handle/replies` | 없음 |
| `GET /api/users/:handle/likes` | 없음 |

### 4.2 동작 변경 (Breaking Change 아님)

- 기존에 모든 visibility의 게시물이 반환되던 것이, 이제 viewer의 권한에 따라 필터링됨
- `public`으로만 사용하던 클라이언트는 영향 없음
- `friends`/`private` 게시물이 권한 없는 사용자에게 404 반환 (이전에는 200으로 반환됨 -- 보안 수정)

---

## 5. DB 변경사항

**없음.** visibility 컬럼과 follows 테이블 모두 이미 존재한다.

---

## 6. 수락 기준 (Acceptance Criteria)

### Phase A: 프론트엔드 UI

- [ ] AC-A1: ComposeForm에 공개 범위 선택 드롭다운이 표시된다
- [ ] AC-A2: Globe/Users/Lock 아이콘으로 세 가지 옵션이 구분된다
- [ ] AC-A3: 기본 선택값은 "전체 공개(public)"이다
- [ ] AC-A4: 선택한 visibility 값이 POST /api/posts 요청에 포함된다
- [ ] AC-A5: 게시 성공 후 visibility가 "public"으로 리셋된다
- [ ] AC-A6: 답글 작성 폼(ReplyForm)에서는 공개 범위 선택기가 표시되지 않는다

### Phase B: 백엔드 필터링

- [ ] AC-B1: 비인증 사용자의 피드에는 `public` 게시물만 표시된다
- [ ] AC-B2: 인증 사용자의 피드에서 `friends` 게시물은 작성자의 팔로워에게만 표시된다
- [ ] AC-B3: `private` 게시물은 작성자 본인에게만 표시된다
- [ ] AC-B4: 권한 없는 사용자가 `friends`/`private` 게시물에 직접 접근(GET /posts/:id)하면 404 반환
- [ ] AC-B5: 작성자 본인은 자신의 모든 visibility 게시물을 프로필에서 볼 수 있다
- [ ] AC-B6: `friends` 게시물의 작성자 본인도 해당 게시물을 볼 수 있다 (팔로워가 아닌 본인)

### Phase C: PostCard 표시

- [ ] AC-C1: `public` 게시물에는 visibility 아이콘이 표시되지 않는다
- [ ] AC-C2: `friends` 게시물에는 Users 아이콘 + "팔로워 전용" 텍스트가 표시된다
- [ ] AC-C3: `private` 게시물에는 Lock 아이콘 + "나만 보기" 텍스트가 표시된다
- [ ] AC-C4: PostDetailPage에서도 동일한 visibility 표시가 적용된다

---

## 7. 엣지 케이스

| # | 시나리오 | 기대 동작 |
|---|---------|----------|
| E1 | 사용자 A가 `friends` 게시물 작성 후, 팔로워 B가 언팔로우 | B의 다음 피드 조회부터 해당 게시물 미표시. 이미 로드된 캐시는 React Query refetch 시 갱신 |
| E2 | `friends` 게시물의 직접 URL 접근 (비팔로워) | 404 Not Found 반환. "게시물을 찾을 수 없습니다" UI 표시 |
| E3 | `private` 게시물에 답글 달기 시도 | 현실적으로 본인만 볼 수 있으므로 본인만 답글 가능. 답글 자체는 public이지만, 부모 게시물이 안 보이면 맥락 없는 답글이 됨 -- 별도 차단 로직은 v1에서 미적용, 향후 검토 |
| E4 | 작성자가 visibility를 변경하려는 경우 | v1에서는 게시물 수정 기능 미지원. 향후 수정 API 추가 시 고려 |
| E5 | `friends` 게시물이 좋아요/북마크/리포스트에 포함 | 좋아요/북마크 목록 조회 시에도 동일한 visibility 필터 적용 필요 (`FindLikedByUserHandle` 등) |
| E6 | 비인증 사용자가 프로필 페이지에서 게시물 조회 | `public` 게시물만 표시 |
| E7 | 작성자 계정 삭제 시 `friends`/`private` 게시물 | CASCADE DELETE로 게시물 자체가 삭제되므로 문제 없음 |
| E8 | `friends` 게시물 작성 직후, 팔로워가 0명인 경우 | 작성자 본인에게만 표시. 이후 팔로워가 생기면 해당 팔로워에게도 표시 |

---

## 8. 의존성 및 제약사항

### 8.1 의존성

| 항목 | 상태 | 비고 |
|------|------|------|
| `follows` 테이블 | 존재 | migration 004 |
| `FollowRepository.IsFollowing()` | 존재 | 단일 게시물 접근 시 서비스 레이어에서 사용 |
| `visibility` DB 컬럼 | 존재 | posts 테이블, default 'public' |
| `model.Visibility` 타입 | 존재 | public/friends/private |
| shadcn/ui DropdownMenu | 설치 필요 여부 확인 | 없으면 `bun add` 또는 `bunx shadcn-ui add dropdown-menu` |
| lucide-react Globe/Users/Lock 아이콘 | 이미 사용 중 | lucide-react 패키지 |

### 8.2 제약사항

- PostService에 FollowRepository 의존성을 추가하면 기존 테스트의 모킹이 변경될 수 있음
- Repository 쿼리 변경 시 EXISTS 서브쿼리가 추가되어 피드 쿼리 성능에 미미한 영향 가능 (follows 테이블에 이미 인덱스 존재)
- DI(fx) 모듈 설정에서 PostService 생성자 시그니처 변경 반영 필요

---

## 9. 테스트 계획

### 9.1 백엔드 Service 레이어 테스트 (Table-driven)

| 테스트 케이스 | 입력 | 기대 결과 |
|-------------|------|----------|
| public 게시물 - 비인증 사용자 조회 | viewerID=nil | 정상 반환 |
| public 게시물 - 인증 사용자 조회 | viewerID=X | 정상 반환 |
| friends 게시물 - 비인증 사용자 조회 | viewerID=nil | NotFound |
| friends 게시물 - 팔로워 조회 | viewerID=follower | 정상 반환 |
| friends 게시물 - 비팔로워 조회 | viewerID=stranger | NotFound |
| friends 게시물 - 작성자 본인 조회 | viewerID=author | 정상 반환 |
| private 게시물 - 비인증 사용자 조회 | viewerID=nil | NotFound |
| private 게시물 - 다른 사용자 조회 | viewerID=other | NotFound |
| private 게시물 - 작성자 본인 조회 | viewerID=author | 정상 반환 |
| 피드 조회 - mixed visibility | viewerID=X | public + 본인의 friends/private + 팔로우한 작성자의 friends만 반환 |

### 9.2 프론트엔드 테스트

| 테스트 케이스 | 방법 |
|-------------|------|
| VisibilitySelector 렌더링 및 옵션 선택 | getByRole('button'), click, getByText |
| ComposeForm에서 visibility 값 전달 | MSW 모킹 + 제출 요청 body 검증 |
| VisibilityBadge - public일 때 미표시 | queryByText('팔로워 전용') === null |
| VisibilityBadge - friends/private 표시 | getByText('팔로워 전용') / getByText('나만 보기') |

---

## 10. 구현 순서 권장

1. **Phase B 먼저 (백엔드 보안 수정)** -- 현재 보안 취약점이므로 최우선
2. **Phase A (프론트엔드 UI)** -- 사용자가 visibility를 선택할 수 있도록
3. **Phase C (PostCard 표시)** -- visibility 정보를 사용자에게 시각적으로 전달

---

## 11. 변경 대상 파일 요약

### 백엔드

| 파일 | 변경 유형 |
|------|----------|
| `backend/internal/repository/post_repository.go` | 수정 - 12개+ 쿼리 메서드에 visibility WHERE 조건 추가 |
| `backend/internal/service/post_service.go` | 수정 - GetPostByID 접근 제어 + FollowRepository 의존성 추가 |
| `backend/internal/service/post_service_test.go` | 수정/신규 - visibility 관련 테스트 케이스 추가 |
| DI 모듈 파일 (fx provider) | 수정 - PostService 생성자 시그니처 변경 반영 |

### 프론트엔드

| 파일 | 변경 유형 |
|------|----------|
| `frontend/src/components/ComposeForm.tsx` | 수정 - visibility state + VisibilitySelector 통합 |
| `frontend/src/components/VisibilitySelector.tsx` | 신규 |
| `frontend/src/components/VisibilityBadge.tsx` | 신규 |
| `frontend/src/components/PostCard.tsx` | 수정 - VisibilityBadge 추가 |
| PostDetailPage (해당 파일) | 수정 - VisibilityBadge 추가 |
