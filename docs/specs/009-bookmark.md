# Spec 009: 사용자별 포스트 북마크(저장하기) 시스템

- **Issue**: #9
- **작성일**: 2026-03-06
- **상태**: Draft

---

## 1. 개요

### What
사용자가 나중에 다시 보기 위해 특정 포스트를 개인 서랍에 저장해 두는 북마크 기능.
X(Twitter)의 Bookmarks 기능과 동일하게, 사용자 본인만 볼 수 있는 비공개 저장 목록을 제공한다.

### Why
- 좋아요와 달리 다른 사용자에게 노출되지 않는 개인적인 저장 수단 필요
- X에서 핵심 인터랙션 기능 중 하나로, 사용자 경험 완성도를 위해 필수
- 기존 likes 패턴(매핑 테이블 + toggle API)을 재사용하여 빠르게 구현 가능

### 핵심 제약
- 북마크 목록은 **본인만 조회** 가능 (타인의 북마크 목록 접근 불가)
- 북마크 수(bookmark_count)는 포스트에 **노출하지 않음** (likes와의 차이점)
- 따라서 posts 테이블에 bookmark_count 컬럼을 추가하지 않음

---

## 2. API 설계

### 2.1 북마크 추가

```
POST /api/posts/:id/bookmark
```

- **인증**: AuthRequired
- **Request Body**: 없음
- **성공 응답** (200):
```json
{
  "success": true,
  "data": {
    "bookmarked": true
  }
}
```
- **에러 응답**:
  - 401: 미인증
  - 404: 포스트 없음
  - 409: 이미 북마크됨

### 2.2 북마크 취소

```
DELETE /api/posts/:id/bookmark
```

- **인증**: AuthRequired
- **Request Body**: 없음
- **성공 응답** (200):
```json
{
  "success": true,
  "data": {
    "bookmarked": false
  }
}
```
- **에러 응답**:
  - 401: 미인증
  - 404: 포스트 없음
  - 409: 북마크하지 않은 포스트

### 2.3 내 북마크 목록 조회

```
GET /api/users/bookmarks?cursor=<created_at>&limit=20
```

- **인증**: AuthRequired (본인만 접근 가능)
- **Query Parameters**:
  - `cursor` (optional): 마지막으로 받은 bookmark의 `created_at` ISO 문자열
  - `limit` (optional, default=20, max=50): 페이지 크기
- **성공 응답** (200):
```json
{
  "success": true,
  "data": {
    "posts": [PostDetailResponse],
    "nextCursor": "2026-03-06T12:00:00Z",
    "hasMore": true
  }
}
```

> **설계 결정**: 북마크 목록은 북마크한 시간(bookmarks.created_at) 기준 최신순 정렬.
> PostDetailResponse 형태를 재사용하여 isLiked, likeCount 등 기존 정보도 함께 반환한다.
> isBookmarked 필드는 목록 내에서 항상 true이지만, PostCard의 일관성을 위해 포함한다.

---

## 3. DB 스키마

### 3.1 마이그레이션 (007_create_bookmarks.up.sql)

```sql
CREATE TABLE bookmarks (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id    UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);

-- 사용자별 북마크 목록 조회 (최신순) 최적화 인덱스
CREATE INDEX idx_bookmarks_user_created ON bookmarks(user_id, created_at DESC);
```

### 3.2 다운 마이그레이션 (007_create_bookmarks.down.sql)

```sql
DROP TABLE IF EXISTS bookmarks;
```

### 3.3 설계 결정 사항
- **bookmark_count 컬럼 없음**: likes와 달리 포스트에 북마크 수를 노출하지 않으므로 posts 테이블에 카운터 컬럼을 추가하지 않는다. 이에 따라 Bookmark/Unbookmark 시 트랜잭션이 필요 없다 (단일 INSERT/DELETE).
- **인덱스 전략**: `(user_id, created_at DESC)` 복합 인덱스로 사용자별 최신순 목록 조회를 커버링 인덱스로 처리.

---

## 4. Backend 구현 계획

### 4.1 Model

```go
// backend/internal/model/bookmark.go
package model

type Bookmark struct {
    UserID    uuid.UUID
    PostID    uuid.UUID
    CreatedAt time.Time
}
```

### 4.2 DTO

```go
// backend/internal/dto/bookmark_dto.go
package dto

type BookmarkStatusResponse struct {
    Bookmarked bool `json:"bookmarked"`
}

type BookmarkListResponse struct {
    Posts      []PostDetailResponse `json:"posts"`
    NextCursor string              `json:"nextCursor,omitempty"`
    HasMore    bool                `json:"hasMore"`
}
```

### 4.3 Repository Interface

```go
// backend/internal/repository/bookmark_repository.go
type BookmarkRepository interface {
    Bookmark(ctx context.Context, userID, postID uuid.UUID) error
    Unbookmark(ctx context.Context, userID, postID uuid.UUID) error
    IsBookmarked(ctx context.Context, userID, postID uuid.UUID) (bool, error)
    ListByUserID(ctx context.Context, userID uuid.UUID, cursor time.Time, limit int) ([]model.PostWithAuthor, bool, error)
}
```

**구현 노트**:
- `Bookmark`: `INSERT INTO bookmarks ... ON CONFLICT DO NOTHING` (트랜잭션 불필요, 카운터 없음)
- `Unbookmark`: `DELETE FROM bookmarks WHERE user_id = $1 AND post_id = $2`
- `IsBookmarked`: `SELECT EXISTS(...)` 패턴 (likes와 동일)
- `ListByUserID`: bookmarks JOIN posts JOIN users (author 정보 포함), cursor 기반 페이지네이션
  - JOIN 구조: `bookmarks b JOIN posts p ON b.post_id = p.id JOIN users u ON p.author_id = u.id`
  - 정렬: `ORDER BY b.created_at DESC`
  - isLiked 서브쿼리: `EXISTS(SELECT 1 FROM likes WHERE user_id = $userID AND post_id = p.id)`
  - isBookmarked: 목록 자체가 북마크 목록이므로 항상 true로 세팅

### 4.4 Service Interface

```go
// backend/internal/service/bookmark_service.go
type BookmarkService interface {
    Bookmark(ctx context.Context, userID, postID uuid.UUID) (*dto.BookmarkStatusResponse, error)
    Unbookmark(ctx context.Context, userID, postID uuid.UUID) (*dto.BookmarkStatusResponse, error)
    ListBookmarks(ctx context.Context, userID uuid.UUID, cursor string, limit int) (*dto.BookmarkListResponse, error)
}
```

**비즈니스 로직**:
- Bookmark/Unbookmark: 포스트 존재 확인 -> 중복/미존재 확인 -> 실행 (likes 서비스 패턴 재사용)
- ListBookmarks: cursor 파싱 -> repository 호출 -> DTO 변환

### 4.5 Handler

```go
// backend/internal/handler/bookmark_handler.go
type BookmarkHandler struct {
    bookmarkService service.BookmarkService
}

func NewBookmarkHandler(bs service.BookmarkService) *BookmarkHandler
func (h *BookmarkHandler) Bookmark(c *fiber.Ctx) error      // POST /api/posts/:id/bookmark
func (h *BookmarkHandler) Unbookmark(c *fiber.Ctx) error    // DELETE /api/posts/:id/bookmark
func (h *BookmarkHandler) ListBookmarks(c *fiber.Ctx) error // GET /api/users/bookmarks
```

### 4.6 라우트 등록 (main.go)

```go
// 기존 posts 그룹에 추가
posts.Post("/:id/bookmark", middleware.AuthRequired(cfg.JWTSecret), bookmarkHandler.Bookmark)
posts.Delete("/:id/bookmark", middleware.AuthRequired(cfg.JWTSecret), bookmarkHandler.Unbookmark)

// 기존 users 그룹에 추가
users.Get("/bookmarks", middleware.AuthRequired(cfg.JWTSecret), bookmarkHandler.ListBookmarks)
```

> **주의**: `users.Get("/bookmarks", ...)` 라우트는 `users.Get("/:handle", ...)` 보다 **앞에** 등록해야 한다.
> Fiber는 라우트를 등록 순서대로 매칭하므로, `/bookmarks`가 `:handle` 파라미터로 해석되는 것을 방지해야 한다.

### 4.7 PostDetailResponse에 isBookmarked 필드 추가

기존 `PostDetailResponse`에 `isBookmarked` 필드를 추가하여, 피드/포스트 상세 등 모든 포스트 조회 시 로그인 유저의 북마크 여부를 함께 반환한다.

```go
// dto/post_dto.go 변경
type PostDetailResponse struct {
    // ... 기존 필드
    IsBookmarked bool `json:"isBookmarked"`
}
```

이에 따라 `model.PostWithAuthor`에도 `IsBookmarked bool` 필드 추가가 필요하며, 기존 포스트 조회 쿼리에 bookmarks 테이블 LEFT JOIN 또는 EXISTS 서브쿼리를 추가해야 한다.

---

## 5. Frontend 구현 계획

### 5.1 타입 추가

```typescript
// types/api.ts
export interface BookmarkStatusResponse {
  bookmarked: boolean;
}

// PostDetail 인터페이스에 필드 추가
export interface PostDetail {
  // ... 기존 필드
  isBookmarked: boolean;
}

export interface BookmarkListResponse {
  posts: PostDetail[];
  nextCursor: string;
  hasMore: boolean;
}
```

### 5.2 Custom Hooks

#### useBookmark.ts (토글 hook)
```typescript
// hooks/useBookmark.ts
export function useBookmark(postId: string, isBookmarked: boolean)
```
- useLike.ts 패턴과 동일한 Optimistic UI 적용
- mutationFn: isBookmarked ? DELETE : POST
- onMutate: 캐시에서 isBookmarked 토글
- onError: 이전 캐시 복원
- onSettled: 관련 쿼리 invalidate (["posts"], ["post", postId], ["bookmarks"])

#### useBookmarks.ts (목록 조회 hook)
```typescript
// hooks/useBookmarks.ts
export function useBookmarks()
```
- useInfiniteQuery 사용 (cursor 기반 페이지네이션)
- queryKey: ["bookmarks"]
- queryFn: `GET /api/users/bookmarks?cursor=...&limit=20`
- getNextPageParam: response.hasMore ? response.nextCursor : undefined

### 5.3 PostCard 변경

`PostCard.tsx`의 Action Buttons 영역에 북마크 토글 버튼 추가:
- 기존 Share 버튼 옆 또는 대체 위치에 Bookmark 아이콘 배치
- lucide-react의 `Bookmark` 아이콘 사용 (북마크됨 상태: `fill` 적용)
- 색상: 기본 `text-muted-foreground`, 활성화 시 `text-primary fill-primary`
- hover: `hover:bg-primary/10`
- onClick: `useBookmark` hook 호출

```tsx
import { Bookmark } from "lucide-react";

// Action Buttons 내 Share 버튼 자리를 Bookmark 버튼으로 교체
<button
  onClick={handleBookmarkClick}
  className="group flex cursor-pointer items-center gap-1.5 rounded-full border-none bg-transparent p-2 transition-colors hover:bg-primary/10"
>
  <Bookmark
    size={18}
    className={cn(
      "transition-colors group-hover:text-primary",
      post.isBookmarked
        ? "fill-primary text-primary"
        : "text-muted-foreground",
    )}
  />
</button>
```

### 5.4 ProfilePage 변경

프로필 페이지에 "북마크" 탭 추가. 단, **본인 프로필에서만** 노출.

```typescript
type Tab = "posts" | "replies" | "likes" | "bookmarks";

// 탭 목록 동적 구성
const tabs: { key: Tab; label: string }[] = [
  { key: "posts", label: "게시물" },
  { key: "replies", label: "답글" },
  { key: "likes", label: "마음에 들어요" },
  ...(isOwner ? [{ key: "bookmarks" as Tab, label: "북마크" }] : []),
];
```

TabContent에서 bookmarks 탭일 때 useBookmarks() hook으로 데이터를 가져와 PostCard 목록을 렌더링한다.

### 5.5 Sidebar 네비게이션 (선택사항)

Sidebar에 "북마크" 메뉴 항목 추가 가능. 클릭 시 `/bookmarks` 전용 페이지로 이동하거나, 본인 프로필의 북마크 탭으로 이동. (이슈 범위 외일 수 있으나, X 원본 UX 참고)

---

## 6. 보안 고려사항 (북마크 비공개 보장)

### 6.1 API 레벨 보호
- 북마크 추가/삭제: `AuthRequired` 미들웨어로 인증된 사용자만 접근
- 북마크 목록 조회: `AuthRequired` + JWT에서 추출한 userID로만 조회 (파라미터로 다른 유저 ID 받지 않음)
- 라우트가 `GET /api/users/bookmarks` (handle 파라미터 없음)이므로, 구조적으로 타인의 북마크에 접근할 경로가 없음

### 6.2 데이터 레벨 보호
- Repository의 `ListByUserID`는 항상 JWT에서 추출한 userID를 WHERE 조건에 사용
- 다른 유저의 userID를 주입할 수 있는 API 경로가 존재하지 않음

### 6.3 응답 레벨 보호
- PostDetailResponse의 `isBookmarked` 필드는 현재 로그인 유저 기준으로만 계산
- 비로그인 상태(OptionalAuth)에서는 isBookmarked가 항상 false

### 6.4 테스트 시나리오
1. 인증되지 않은 사용자가 `GET /api/users/bookmarks` 접근 시 401 반환
2. 사용자 A의 JWT로 북마크 목록 조회 시 사용자 A의 북마크만 반환
3. 포스트 조회 시 `isBookmarked`가 다른 유저에게는 false로 반환

---

## 7. 수락 기준 (Acceptance Criteria)

### Backend
- [ ] bookmarks 테이블 마이그레이션 적용 (007)
- [ ] `POST /api/posts/:id/bookmark` 호출 시 북마크 추가, 중복 시 409
- [ ] `DELETE /api/posts/:id/bookmark` 호출 시 북마크 제거, 미존재 시 409
- [ ] 존재하지 않는 포스트에 대한 북마크 시 404
- [ ] `GET /api/users/bookmarks` 호출 시 본인의 북마크 목록 cursor 기반 페이지네이션 반환
- [ ] 미인증 상태에서 북마크 관련 API 호출 시 모두 401
- [ ] 기존 포스트 조회 API(피드, 상세)에 `isBookmarked` 필드 포함

### Frontend
- [ ] PostCard에 북마크 토글 버튼 표시 (Bookmark 아이콘)
- [ ] 북마크 상태에 따라 아이콘 fill 토글 (Optimistic UI)
- [ ] 본인 프로필 페이지에 "북마크" 탭 표시 (타인 프로필에서는 미표시)
- [ ] 북마크 탭에서 북마크한 포스트 목록 표시

### Security
- [ ] 타인의 북마크 목록에 접근할 수 있는 API 경로가 없음을 확인
- [ ] 비로그인 시 isBookmarked가 항상 false

---

## 8. 엣지 케이스

| # | 시나리오 | 기대 동작 |
|---|---------|----------|
| 1 | 삭제된 포스트를 북마크 시도 | 404 Not Found |
| 2 | 이미 북마크한 포스트를 다시 북마크 | 409 Conflict ("already bookmarked") |
| 3 | 북마크하지 않은 포스트를 북마크 취소 | 409 Conflict ("not bookmarked yet") |
| 4 | 북마크한 포스트가 나중에 삭제됨 | ON DELETE CASCADE로 bookmarks 행 자동 삭제 |
| 5 | 북마크한 사용자가 계정 삭제 | ON DELETE CASCADE로 bookmarks 행 자동 삭제 |
| 6 | 비공개(private) 포스트를 북마크 | 현재 단계에서는 허용 (향후 ReBAC으로 제한 가능) |
| 7 | 빈 북마크 목록 조회 | `{ posts: [], nextCursor: "", hasMore: false }` |
| 8 | 동시에 같은 포스트를 북마크/취소 (race condition) | ON CONFLICT DO NOTHING으로 DB 레벨에서 안전 |
| 9 | 비로그인 상태에서 PostCard 북마크 버튼 클릭 | 아무 동작 안 함 (useLike와 동일 패턴) |

---

## 9. 의존성 및 제약사항

### 의존성
- likes 패턴 (repository/service/handler 구조) -- 참조 모델
- PostWithAuthor 모델 및 기존 포스트 조회 쿼리 -- isBookmarked 필드 추가 필요
- lucide-react의 Bookmark 아이콘 -- 이미 node_modules에 포함 확인됨

### 제약사항
- `/api/users/bookmarks` 라우트는 `/api/users/:handle` 보다 앞에 등록해야 Fiber 라우팅 충돌 방지
- 기존 포스트 조회 쿼리 변경 시 성능 영향 최소화 필요 (EXISTS 서브쿼리 vs LEFT JOIN 선택)

---

## 10. 변경 대상 파일 목록

### 신규 파일
| 파일 | 설명 |
|------|------|
| `backend/migrations/007_create_bookmarks.up.sql` | bookmarks 테이블 생성 |
| `backend/migrations/007_create_bookmarks.down.sql` | bookmarks 테이블 삭제 |
| `backend/internal/model/bookmark.go` | Bookmark 모델 |
| `backend/internal/dto/bookmark_dto.go` | BookmarkStatusResponse, BookmarkListResponse DTO |
| `backend/internal/repository/bookmark_repository.go` | BookmarkRepository 인터페이스 및 구현 |
| `backend/internal/service/bookmark_service.go` | BookmarkService 인터페이스 및 구현 |
| `backend/internal/handler/bookmark_handler.go` | BookmarkHandler (3개 핸들러) |
| `frontend/src/hooks/useBookmark.ts` | 북마크 토글 mutation hook |
| `frontend/src/hooks/useBookmarks.ts` | 북마크 목록 조회 infinite query hook |

### 수정 파일
| 파일 | 변경 내용 |
|------|----------|
| `backend/main.go` | bookmarkRepo/Service/Handler 초기화 + 라우트 3개 등록 |
| `backend/internal/model/post.go` | PostWithAuthor에 IsBookmarked 필드 추가 |
| `backend/internal/dto/post_dto.go` | PostDetailResponse에 IsBookmarked 필드 + ToPostDetailResponse 변환 |
| `backend/internal/repository/post_repository.go` | 포스트 조회 쿼리에 bookmarks EXISTS 서브쿼리 추가 |
| `frontend/src/types/api.ts` | PostDetail에 isBookmarked, BookmarkStatusResponse 타입 추가 |
| `frontend/src/components/PostCard.tsx` | 북마크 토글 버튼 추가 (Share 버튼 위치) |
| `frontend/src/pages/ProfilePage.tsx` | Tab 타입에 "bookmarks" 추가, 본인일 때만 탭 표시 |

---

## 11. 구현 순서 권장

1. **DB 마이그레이션** (007_create_bookmarks)
2. **Backend**: model -> dto -> repository -> service -> handler -> main.go 라우트
3. **Backend 기존 쿼리 수정**: PostWithAuthor에 isBookmarked 추가
4. **Frontend**: types -> hooks -> PostCard 버튼 -> ProfilePage 탭
5. **테스트**: 보안 테스트 (타인 접근 불가) + CRUD 테스트
