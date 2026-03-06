# Spec: Post/Reply 수정(Update) 및 삭제(Delete) 기능

- **작성일**: 2026-03-06
- **상태**: Draft
- **관련 테이블**: posts, post_media, polls, poll_options, poll_votes, likes, bookmarks

---

## 1. 기능 개요

### 1.1 What

Post와 Reply(같은 posts 테이블, parent_id 유무로 구분)에 대해 수정(Update)과 삭제(Delete) 기능을 추가한다. 작성자 본인만 자신의 게시물을 수정하거나 삭제할 수 있다.

- **수정**: content(본문)와 visibility(공개 범위)만 변경 가능. 미디어, 위치, 투표는 수정 불가.
- **삭제**: 게시물 자체를 삭제하며, DB의 ON DELETE CASCADE 설정에 의해 관련 데이터(좋아요, 북마크, 미디어 링크, 투표, 하위 답글)가 자동 삭제된다.
- **"edited" 표시**: `updated_at > created_at` 비교로 수정 여부를 UI에 표시한다.

### 1.2 Why

- **기본 CRUD 완성**: 현재 Create와 Read만 존재하며, Update/Delete가 없어 사용자가 오타 수정이나 실수 삭제를 할 수 없다.
- **X(Twitter)와의 기능 패리티**: X는 삭제를 지원하며, 2023년부터 유료 사용자에게 수정 기능도 제공한다. 본 클론에서는 모든 사용자에게 수정을 허용한다.
- **사용자 경험**: 한번 게시하면 수정할 수 없는 시스템은 사용자 이탈을 유발한다.

### 1.3 핵심 제약

- `handler -> service -> repository` 레이어 구조 준수
- `ctx context.Context`가 모든 Go 메서드의 첫 번째 파라미터
- 인터페이스 기반 DI (PostRepository, PostService 인터페이스에 메서드 추가)
- 미디어, 위치, 투표는 수정 시 변경 불가 (content, visibility만 수정)
- Reply 수정 시 visibility 변경 불가 (Reply는 항상 public)
- 삭제 시 DB CASCADE에 의존하되, Reply 삭제 시 부모의 reply_count는 명시적으로 감소

---

## 2. API 엔드포인트 설계

### 2.1 게시물/답글 수정

```
PUT /api/posts/:id
```

- **인증**: Required (AuthRequired 미들웨어)
- **권한**: 작성자 본인만 허용

**Request Body**:
```json
{
  "content": "수정된 본문 내용",
  "visibility": "public"
}
```

| 필드 | 타입 | 필수 | 검증 규칙 |
|------|------|------|-----------|
| content | string | N (기존 값 유지) | 1~500자 (rune count) |
| visibility | string | N (기존 값 유지) | "public", "follower", "private" 중 하나 |

**주의**: Reply(parent_id가 있는 게시물)를 수정할 때 visibility 필드는 무시된다. Reply는 항상 "public"으로 고정.

**Success Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "authorId": "uuid",
    "parentId": null,
    "content": "수정된 본문 내용",
    "visibility": "public",
    "author": {
      "username": "user1",
      "displayName": "User One",
      "profileImageUrl": "..."
    },
    "likeCount": 5,
    "replyCount": 3,
    "viewCount": 142,
    "isLiked": true,
    "isBookmarked": false,
    "media": [...],
    "location": {...},
    "poll": {...},
    "topReplies": [],
    "createdAt": "2026-03-06T10:00:00Z",
    "updatedAt": "2026-03-06T12:30:00Z"
  }
}
```

**Error Responses**:

| 상태코드 | 조건 | 응답 |
|----------|------|------|
| 400 | 잘못된 post ID, 빈 content, 500자 초과, 잘못된 visibility | `{"error": {"message": "..."}}` |
| 401 | 미인증 | `{"error": {"message": "not authenticated"}}` |
| 403 | 작성자가 아닌 사용자 | `{"error": {"message": "you can only edit your own post"}}` |
| 404 | 존재하지 않는 post | `{"error": {"message": "post not found"}}` |

### 2.2 게시물/답글 삭제

```
DELETE /api/posts/:id
```

- **인증**: Required (AuthRequired 미들웨어)
- **권한**: 작성자 본인만 허용

**Request Body**: 없음

**Success Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "message": "post deleted successfully"
  }
}
```

**Error Responses**:

| 상태코드 | 조건 | 응답 |
|----------|------|------|
| 400 | 잘못된 post ID | `{"error": {"message": "invalid post ID"}}` |
| 401 | 미인증 | `{"error": {"message": "not authenticated"}}` |
| 403 | 작성자가 아닌 사용자 | `{"error": {"message": "you can only delete your own post"}}` |
| 404 | 존재하지 않는 post | `{"error": {"message": "post not found"}}` |

---

## 3. DTO 설계

### 3.1 새로 추가할 DTO

**UpdatePostRequest** (`backend/internal/dto/post_dto.go`):

```go
type UpdatePostRequest struct {
    Content    *string `json:"content"    validate:"omitempty,min=1,max=500"`
    Visibility *string `json:"visibility" validate:"omitempty,oneof=public follower private"`
}
```

- 포인터 타입을 사용하여 "필드가 전송되지 않음(nil)" vs "빈 문자열"을 구분한다.
- content만 보내면 visibility는 기존 값 유지, visibility만 보내면 content는 기존 값 유지.

**DeletePostResponse** (`backend/internal/dto/post_dto.go`):

```go
type DeletePostResponse struct {
    Message string `json:"message"`
}
```

### 3.2 기존 DTO 변경 없음

- `PostDetailResponse`에 이미 `updatedAt` 필드가 존재하므로 수정 결과를 그대로 반환한다.
- 프론트엔드 `PostDetail` 인터페이스에도 이미 `updatedAt: string`이 존재한다.

---

## 4. 비즈니스 규칙

### 4.1 권한 검사 (ReBAC)

1. **소유권 확인**: Service 레이어에서 `post.AuthorID == requesterID`를 검사한다.
2. 소유자가 아니면 `apperror.Forbidden("you can only edit your own post")` (수정) 또는 `apperror.Forbidden("you can only delete your own post")` (삭제) 반환.
3. **apperror에 Forbidden 추가 필요**: 현재 `apperror` 패키지에 `Forbidden` (403) 헬퍼가 없으므로 추가해야 한다.

```go
// apperror/apperror.go 에 추가
func Forbidden(msg string, args ...interface{}) *AppError {
    return &AppError{Code: 403, Message: fmt.Sprintf(msg, args...)}
}
```

### 4.2 수정 가능 필드

| 필드 | Post 수정 | Reply 수정 | 이유 |
|------|-----------|-----------|------|
| content | O | O | 본문은 수정 가능 |
| visibility | O | X (무시) | Reply는 항상 public |
| media | X | X | 미디어 교체는 복잡도가 높아 스코프 밖 |
| location | X | X | 위치는 작성 시점의 정보이므로 수정 불가 |
| poll | X | X | 투표 진행 중 옵션 변경은 공정성 문제 |

### 4.3 수정 시 updated_at 갱신

- `UPDATE posts SET content = $1, visibility = $2, updated_at = NOW() WHERE id = $3`
- DB의 `updated_at` 컬럼은 이미 존재하며, `CREATE` 시 `created_at`과 동일 값으로 설정된다.
- 수정이 발생하면 `updated_at > created_at`이 되어 "edited" 표시의 근거가 된다.

### 4.4 삭제 시 연쇄 처리

**DB CASCADE가 자동 처리하는 항목** (ON DELETE CASCADE 설정 확인됨):
- `likes` (post_id FK)
- `bookmarks` (post_id FK)
- `post_media` (post_id FK)
- `polls` (post_id FK) -> `poll_options` (poll_id FK) -> `poll_votes` (poll_id FK)
- 하위 `posts` (parent_id FK) -- 답글들도 연쇄 삭제됨

**Service 레이어에서 명시적으로 처리해야 하는 항목**:
- **Reply 삭제 시 부모의 reply_count 감소**: 삭제 대상이 Reply(parent_id가 있음)인 경우, 부모 post의 `reply_count`를 감소시켜야 한다.
- 이 작업은 트랜잭션 내에서 수행되어야 한다 (삭제 + reply_count 감소가 원자적이어야 함).

**주의 -- 재귀 삭제 시 reply_count 처리**:
- Post A에 Reply B, Reply B에 Reply C가 있는 경우: Post A를 삭제하면 CASCADE로 B와 C 모두 삭제되므로 부모 reply_count 조정이 불필요하다.
- Reply B를 직접 삭제하면: Post A의 reply_count만 1 감소시킨다. Reply C는 CASCADE로 자동 삭제되며, Reply B의 reply_count 조정은 불필요하다 (B 자체가 삭제되므로).

### 4.5 "edited" 표시 판별

프론트엔드에서 `post.updatedAt !== post.createdAt` 비교로 수정 여부를 판별한다. 시간 정밀도 문제를 방지하기 위해, 밀리초 단위까지 비교하는 대신 두 문자열의 직접 비교를 사용한다 (서버에서 동일 포맷으로 반환하므로 안전).

---

## 5. 백엔드 구현 상세

### 5.1 Repository 변경

**PostRepository 인터페이스에 추가** (`backend/internal/repository/post_repository.go`):

```go
Update(ctx context.Context, id uuid.UUID, content string, visibility model.Visibility) error
Delete(ctx context.Context, id uuid.UUID) error
DeleteReply(ctx context.Context, id uuid.UUID, parentID uuid.UUID) error
```

**Update 구현**:
```go
func (r *postRepository) Update(ctx context.Context, id uuid.UUID, content string, visibility model.Visibility) error {
    query := `UPDATE posts SET content = $1, visibility = $2, updated_at = NOW() WHERE id = $3`
    result, err := r.pool.Exec(ctx, query, content, string(visibility), id)
    if err != nil {
        return fmt.Errorf("failed to update post: %w", err)
    }
    if result.RowsAffected() == 0 {
        return pgx.ErrNoRows
    }
    return nil
}
```

**Delete 구현** (일반 Post 삭제 -- parent_id가 없는 경우):
```go
func (r *postRepository) Delete(ctx context.Context, id uuid.UUID) error {
    result, err := r.pool.Exec(ctx, `DELETE FROM posts WHERE id = $1`, id)
    if err != nil {
        return fmt.Errorf("failed to delete post: %w", err)
    }
    if result.RowsAffected() == 0 {
        return pgx.ErrNoRows
    }
    return nil
}
```

**DeleteReply 구현** (Reply 삭제 -- parent_id가 있는 경우, 트랜잭션):
```go
func (r *postRepository) DeleteReply(ctx context.Context, id uuid.UUID, parentID uuid.UUID) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    result, err := tx.Exec(ctx, `DELETE FROM posts WHERE id = $1`, id)
    if err != nil {
        return fmt.Errorf("failed to delete reply: %w", err)
    }
    if result.RowsAffected() == 0 {
        return pgx.ErrNoRows
    }

    _, err = tx.Exec(ctx,
        `UPDATE posts SET reply_count = GREATEST(reply_count - 1, 0) WHERE id = $1`,
        parentID,
    )
    if err != nil {
        return fmt.Errorf("failed to decrement reply_count: %w", err)
    }

    return tx.Commit(ctx)
}
```

**설계 결정**: `GREATEST(reply_count - 1, 0)` 사용으로 reply_count가 음수가 되는 것을 방지한다. 동시성 이슈로 카운트가 불일치할 가능성이 있으므로 방어적으로 처리한다.

### 5.2 Service 변경

**PostService 인터페이스에 추가** (`backend/internal/service/post_service.go`):

```go
UpdatePost(ctx context.Context, postID, requesterID uuid.UUID, req dto.UpdatePostRequest) (*dto.PostDetailResponse, error)
DeletePost(ctx context.Context, postID, requesterID uuid.UUID) error
```

**UpdatePost 구현 로직**:
1. `postRepo.FindByID`로 게시물 조회 (존재 여부 + 작성자 확인)
2. `post.AuthorID != requesterID`이면 `Forbidden` 반환
3. `req.Content`가 nil이면 기존 content 유지, 아니면 새 값으로 교체 (rune count 검증)
4. `req.Visibility`가 nil이면 기존 visibility 유지; Reply인 경우 visibility 변경 무시
5. content 또는 visibility가 변경된 경우에만 `postRepo.Update` 호출 (불필요한 DB 쓰기 방지)
6. 수정 후 `postRepo.FindByIDWithUser`로 최신 데이터 조회하여 반환 (enrichWithPollAndMedia 포함)

**DeletePost 구현 로직**:
1. `postRepo.FindByID`로 게시물 조회 (존재 여부 + 작성자 확인)
2. `post.AuthorID != requesterID`이면 `Forbidden` 반환
3. Reply인 경우 (`post.ParentID != nil`): `postRepo.DeleteReply(ctx, postID, *post.ParentID)` 호출
4. 일반 Post인 경우: `postRepo.Delete(ctx, postID)` 호출

### 5.3 Handler 변경

**PostHandler에 추가** (`backend/internal/handler/post_handler.go`):

```go
func (h *PostHandler) UpdatePost(c *fiber.Ctx) error
func (h *PostHandler) DeletePost(c *fiber.Ctx) error
```

**UpdatePost 핸들러 로직**:
1. `c.Locals("userID")`에서 인증된 사용자 ID 추출
2. `c.Params("id")`에서 post ID 파싱
3. `parseAndValidate`로 `UpdatePostRequest` 파싱 및 검증
4. `postService.UpdatePost` 호출
5. 200 OK + `PostDetailResponse` 반환

**DeletePost 핸들러 로직**:
1. `c.Locals("userID")`에서 인증된 사용자 ID 추출
2. `c.Params("id")`에서 post ID 파싱
3. `postService.DeletePost` 호출
4. 200 OK + `DeletePostResponse{Message: "post deleted successfully"}` 반환

### 5.4 Router 변경

**router.go에 추가**:
```go
posts.Put("/:id", middleware.AuthRequired(jwtSecret), p.PostHandler.UpdatePost)
posts.Delete("/:id", middleware.AuthRequired(jwtSecret), p.PostHandler.DeletePost)
```

위치: `posts.Get("/:id", ...)` 바로 아래에 추가.

### 5.5 apperror 변경

**apperror.go에 추가**:
```go
func Forbidden(msg string, args ...interface{}) *AppError {
    return &AppError{Code: 403, Message: fmt.Sprintf(msg, args...)}
}
```

### 5.6 Handler의 respondError 함수 확인

현재 `respondError`가 403 상태 코드를 올바르게 처리하는지 확인 필요. `AppError.Code`를 HTTP 상태 코드로 사용하므로 별도 변경 없이 동작해야 한다.

---

## 6. 프론트엔드 UI/UX 설계

### 6.1 수정/삭제 드롭다운 메뉴

**적용 위치**: `PostCard.tsx`, `ReplyCard.tsx`, `PostDetailPage.tsx`

**디자인**:
- 게시물 우측 상단에 `MoreHorizontal` (lucide-react) 아이콘 버튼 (세 점 메뉴)
- 작성자 본인의 게시물에만 표시
- 클릭 시 `DropdownMenu` (shadcn/ui) 표시
  - "Edit" (Pencil 아이콘) -- 수정 모드 진입
  - "Delete" (Trash2 아이콘, text-destructive 색상) -- 삭제 확인 다이얼로그 표시

### 6.2 수정 모드 (Inline Edit)

**PostCard / ReplyCard에서의 수정**:
- "Edit" 클릭 시 content 영역이 `Textarea`로 전환된다 (inline editing).
- 기존 content가 Textarea에 pre-fill된다.
- Textarea 아래에 "Cancel" 버튼과 "Save" 버튼이 표시된다.
- Post인 경우 `VisibilitySelector`도 함께 표시하여 변경 가능하게 한다.
- "Save" 클릭 시 `PUT /api/posts/:id` 호출.
- 성공 시 수정 모드 해제, React Query `invalidateQueries` 호출.
- 실패 시 toast 에러 메시지 표시.

**글자 수 제한**: 수정 시에도 500자 제한 유지. 원형 프로그레스 바로 표시 (ComposeForm과 동일).

### 6.3 삭제 확인 다이얼로그

- "Delete" 클릭 시 `AlertDialog` (shadcn/ui) 표시.
- 제목: "Delete post?" (또는 Reply인 경우 "Delete reply?")
- 본문: "This action cannot be undone. This will permanently delete your post and all associated data including replies, likes, and bookmarks."
- 버튼: "Cancel" (outline) + "Delete" (destructive)
- "Delete" 확인 시 `DELETE /api/posts/:id` 호출.
- 성공 시:
  - PostCard에서 삭제한 경우: 피드에서 해당 카드 제거 (React Query invalidation)
  - PostDetailPage에서 삭제한 경우: 홈으로 navigate
  - Reply를 삭제한 경우: 부모 게시물의 상세 페이지 갱신

### 6.4 "edited" 표시

- PostCard, ReplyCard, PostDetailPage에서 작성 시간 옆에 "(edited)" 텍스트 표시.
- 조건: `post.updatedAt !== post.createdAt`
- 스타일: `text-muted-foreground text-xs` (작성 시간과 동일 톤)
- 예시: `2h ago (edited)` 또는 `Mar 6, 2026 (edited)`

### 6.5 Hooks 추가

**useUpdatePost** (`frontend/src/hooks/usePosts.ts`에 추가):
```typescript
export function useUpdatePost(postId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { content?: string; visibility?: string }) =>
      api.put<APIResponse<PostDetail>>(`/posts/${postId}`, data).then(r => r.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["posts"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
    },
  });
}
```

**useDeletePost** (`frontend/src/hooks/usePosts.ts`에 추가):
```typescript
export function useDeletePost(postId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () =>
      api.delete<APIResponse<{ message: string }>>(`/posts/${postId}`).then(r => r.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["posts"] });
    },
  });
}
```

### 6.6 필요한 신규 shadcn/ui 컴포넌트

- `AlertDialog`: 삭제 확인용 (이미 설치되어 있을 수 있으므로 확인 필요)
- `DropdownMenu`: 세 점 메뉴용 (이미 설치되어 있을 수 있으므로 확인 필요)

---

## 7. 수락 기준 (Acceptance Criteria)

### AC-1: apperror Forbidden 추가
- [ ] `apperror.Forbidden("message")` 호출 시 `{Code: 403, Message: "message"}`를 반환한다
- [ ] `respondError`에서 403 상태 코드로 올바르게 응답한다

### AC-2: Repository -- Update
- [ ] `PostRepository` 인터페이스에 `Update(ctx, id, content, visibility) error` 메서드가 존재한다
- [ ] `Update`는 content와 visibility를 갱신하고 `updated_at = NOW()`를 설정한다
- [ ] 존재하지 않는 ID로 호출 시 `pgx.ErrNoRows`를 반환한다

### AC-3: Repository -- Delete / DeleteReply
- [ ] `PostRepository` 인터페이스에 `Delete(ctx, id) error`와 `DeleteReply(ctx, id, parentID) error`가 존재한다
- [ ] `Delete`는 해당 post를 삭제한다 (CASCADE로 연관 데이터 자동 삭제)
- [ ] `DeleteReply`는 트랜잭션 내에서 reply 삭제 + 부모 post의 reply_count 감소를 수행한다
- [ ] `DeleteReply`에서 `reply_count`가 0 미만으로 내려가지 않는다 (`GREATEST(reply_count - 1, 0)`)
- [ ] 존재하지 않는 ID로 호출 시 `pgx.ErrNoRows`를 반환한다

### AC-4: Service -- UpdatePost
- [ ] `PostService` 인터페이스에 `UpdatePost(ctx, postID, requesterID, req) (*PostDetailResponse, error)`가 존재한다
- [ ] 작성자가 아닌 사용자가 호출하면 403 Forbidden을 반환한다
- [ ] content가 nil이면 기존 content를 유지한다
- [ ] visibility가 nil이면 기존 visibility를 유지한다
- [ ] Reply 수정 시 visibility 변경을 무시하고 "public"을 유지한다
- [ ] content가 빈 문자열("")이면 400 Bad Request를 반환한다
- [ ] content가 500자(rune count)를 초과하면 400 Bad Request를 반환한다
- [ ] 수정 후 updated_at이 갱신된 PostDetailResponse를 반환한다 (enrichWithPollAndMedia 포함)

### AC-5: Service -- DeletePost
- [ ] `PostService` 인터페이스에 `DeletePost(ctx, postID, requesterID) error`가 존재한다
- [ ] 작성자가 아닌 사용자가 호출하면 403 Forbidden을 반환한다
- [ ] Reply 삭제 시 부모의 reply_count가 1 감소한다
- [ ] 일반 Post 삭제 시 CASCADE로 모든 연관 데이터가 삭제된다

### AC-6: Handler + Router
- [ ] `PUT /api/posts/:id` 엔드포인트가 AuthRequired 미들웨어와 함께 등록된다
- [ ] `DELETE /api/posts/:id` 엔드포인트가 AuthRequired 미들웨어와 함께 등록된다
- [ ] UpdatePost 핸들러가 200 OK + PostDetailResponse를 반환한다
- [ ] DeletePost 핸들러가 200 OK + `{"message": "post deleted successfully"}`를 반환한다

### AC-7: 프론트엔드 -- 드롭다운 메뉴
- [ ] PostCard, ReplyCard에 작성자 본인의 게시물에만 MoreHorizontal 버튼이 표시된다
- [ ] 드롭다운에 "Edit"과 "Delete" 항목이 존재한다

### AC-8: 프론트엔드 -- 수정 UI
- [ ] "Edit" 클릭 시 content가 Textarea로 전환된다 (기존 값 pre-fill)
- [ ] "Save" 클릭 시 PUT API를 호출하고 성공 시 수정 모드가 해제된다
- [ ] "Cancel" 클릭 시 수정 모드가 해제된다
- [ ] 500자 제한이 적용된다

### AC-9: 프론트엔드 -- 삭제 UI
- [ ] "Delete" 클릭 시 확인 AlertDialog가 표시된다
- [ ] 확인 시 DELETE API를 호출한다
- [ ] PostCard에서 삭제 시 피드가 갱신된다
- [ ] PostDetailPage에서 삭제 시 홈으로 이동한다

### AC-10: 프론트엔드 -- "edited" 표시
- [ ] `updatedAt !== createdAt`인 게시물에 "(edited)" 텍스트가 표시된다
- [ ] PostCard, ReplyCard, PostDetailPage 모두에 적용된다

---

## 8. 엣지 케이스 목록

| # | 케이스 | 예상 동작 |
|---|--------|-----------|
| 1 | 비로그인 사용자가 수정/삭제 시도 | 401 Unauthorized |
| 2 | 다른 사용자의 게시물을 수정/삭제 시도 | 403 Forbidden |
| 3 | 존재하지 않는 post ID로 수정/삭제 시도 | 404 Not Found |
| 4 | content를 빈 문자열로 수정 | 400 Bad Request ("content must not be empty") |
| 5 | content를 500자 초과로 수정 | 400 Bad Request ("content must not exceed 500 characters") |
| 6 | content 없이 visibility만 수정 | 정상 처리, content 유지 + visibility만 변경 |
| 7 | visibility 없이 content만 수정 | 정상 처리, visibility 유지 + content만 변경 |
| 8 | Reply의 visibility를 변경 시도 | visibility 변경 무시, "public" 유지 |
| 9 | 수정 요청에 아무 필드도 없음 | 변경 없이 현재 상태 반환 (불필요한 DB 쓰기 방지) |
| 10 | 미디어가 있는 게시물 수정 | content/visibility만 변경, 미디어는 유지 |
| 11 | 투표가 있는 게시물 수정 | content만 변경, 투표는 유지 |
| 12 | 답글이 달린 게시물 삭제 | CASCADE로 모든 하위 답글 삭제 |
| 13 | 부모 게시물이 삭제된 상태에서 답글 삭제 시도 | CASCADE로 이미 삭제됨, 404 반환 |
| 14 | Reply 삭제 시 부모의 reply_count가 이미 0 | GREATEST(0-1, 0) = 0, 음수 방지 |
| 15 | 동시에 같은 게시물에 수정 + 삭제 요청 | 선행 처리된 요청이 성공, 후행 요청은 404 (삭제 선행) 또는 정상 처리 (수정 선행) |
| 16 | 수정 후 content가 null/undefined (잘못된 요청) | DTO 검증에서 차단 |
| 17 | PostDetailPage에서 현재 보고 있는 게시물이 삭제됨 | 삭제 후 navigate("/") |
| 18 | PostDetailPage에서 답글을 삭제 | 부모 게시물의 상세 페이지가 갱신 (reply_count 감소 반영) |
| 19 | 수정한 게시물에 "(edited)" 표시 | updatedAt > createdAt이므로 표시됨 |
| 20 | 한 번도 수정하지 않은 게시물 | updatedAt === createdAt이므로 "(edited)" 미표시 |
| 21 | content에 미디어만 있고 text가 없는 게시물 수정 | content가 빈 문자열이 되면 거부; 미디어가 있어도 content는 최소 1자 필요 (현재 Update는 content+visibility만 변경하므로 미디어 존재 여부와 무관) |

---

## 9. 의존성 및 제약사항

### 9.1 의존성

- `apperror.Forbidden` 헬퍼 함수 추가 필요 (신규)
- `shadcn/ui`의 `DropdownMenu`, `AlertDialog` 컴포넌트 (설치 여부 확인 필요)
- `lucide-react`의 `MoreHorizontal`, `Pencil`, `Trash2` 아이콘 (이미 설치된 패키지)
- 기존 `PostRepository`, `PostService`, `PostHandler` 인터페이스 확장

### 9.2 제약사항

- DB 마이그레이션 불필요: `updated_at` 컬럼이 이미 존재하며, 새 테이블이나 컬럼 추가가 없다.
- 미디어/위치/투표 수정은 현재 스코프 밖 (향후 별도 이슈로 분리 가능).
- 수정 이력(edit history) 추적은 현재 스코프 밖 (X의 유료 기능과 동일하게, 추후 별도 테이블로 확장 가능).

### 9.3 테스트 요구사항

**Service 테스트 (table-driven, interface mock)**:
- 작성자 본인이 수정 성공
- 다른 사용자가 수정 시도 시 Forbidden
- 존재하지 않는 게시물 수정 시 NotFound
- content만 수정 (visibility 유지)
- visibility만 수정 (content 유지)
- Reply의 visibility 변경 무시
- 빈 content로 수정 시 BadRequest
- 500자 초과 content 수정 시 BadRequest
- 작성자 본인이 삭제 성공
- 다른 사용자가 삭제 시도 시 Forbidden
- Reply 삭제 시 부모 reply_count 감소

---

## 10. 구현 순서 (Phase별)

### Phase A: 백엔드 기반 (Backend Agent)

1. `apperror.Forbidden` 추가
2. `PostRepository` 인터페이스에 `Update`, `Delete`, `DeleteReply` 추가 및 구현
3. `PostService` 인터페이스에 `UpdatePost`, `DeletePost` 추가 및 구현
4. `PostHandler`에 `UpdatePost`, `DeletePost` 핸들러 추가
5. `router.go`에 `PUT /:id`, `DELETE /:id` 라우트 등록
6. Service 레이어 테스트 작성

### Phase B: 프론트엔드 기능 (Frontend Agent)

1. `useUpdatePost`, `useDeletePost` 훅 추가 (`usePosts.ts`)
2. `PostCard`에 MoreHorizontal 드롭다운 메뉴 추가 (Edit/Delete)
3. `PostCard` inline edit 모드 구현
4. `PostCard` 삭제 확인 AlertDialog 구현
5. `PostCard`에 "(edited)" 표시 추가

### Phase C: 프론트엔드 확장 (Frontend Agent)

1. `ReplyCard`에 동일한 드롭다운/수정/삭제 UI 추가
2. `PostDetailPage`에 수정/삭제 UI 추가
3. 삭제 후 네비게이션 처리 (PostDetailPage -> Home)
4. shadcn/ui 컴포넌트 설치 (DropdownMenu, AlertDialog - 미설치 시)

### Phase D: QA 및 마무리

1. E2E 시나리오 테스트
2. 코드 리뷰
3. PR 생성

---

## 11. 남은 의문점

1. **미디어가 있고 content가 없는 게시물의 수정**: 현재 Create 시에는 미디어만 있으면 content가 빈 문자열이어도 허용된다. 수정 시에는 content만 변경 가능한데, 미디어만 있는 게시물의 content를 빈 문자열로 두는 것이 허용되어야 하는지? -> 현재 제안: 미디어 존재 여부를 확인하여, 미디어가 있으면 content 빈 문자열 허용. 미디어가 없으면 content 최소 1자 필요.

2. **수정 이력**: X는 수정 이력을 보여주는 기능이 있다. 본 구현에서는 "edited" 표시만 하고 이력은 저장하지 않는다. 추후 `post_edits(post_id, old_content, edited_at)` 테이블로 확장 가능.

3. **삭제 후 알림 처리**: 게시물 삭제 시 해당 게시물에 대한 알림(좋아요/답글 알림)도 삭제해야 하는지? -> 알림 시스템 구현 전이므로 현재 스코프 밖. 알림 테이블 설계 시 `ON DELETE CASCADE`를 적용하면 자동 처리 가능.

4. **Rate Limiting**: 수정 요청에 대한 rate limiting이 필요한지? -> 현재 스코프 밖. 전체 API rate limiting 구현 시 함께 처리.

---

## 12. 다음 단계 권장

- **Backend Agent**: Phase A 구현 (apperror -> repository -> service -> handler -> router -> 테스트)
- **Frontend Agent**: Phase B, C 구현 (hooks -> PostCard -> ReplyCard -> PostDetailPage)
- 구현 완료 후 **Review Agent**가 코드 리뷰 수행
