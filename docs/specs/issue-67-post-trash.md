# Spec: 삭제된 게시글 접근 제어 + 휴지통 API

- **작성일**: 2026-03-16
- **상태**: Draft
- **이슈**: #67
- **관련**: #56 (비밀번호 변경 및 계정 탈퇴), Phase 12 (Post/Reply Soft Delete)
- **관련 테이블**: posts, post_media, polls, poll_options, poll_votes, likes, bookmarks

---

## 1. 개요

### 1.1 What (기능 설명)

두 가지 기능을 구현한다:

**A. 삭제된 게시글 접근 제어**
- 현재 모든 조회 쿼리에 `WHERE p.deleted_at IS NULL` 필터가 적용되어 삭제된 게시글은 `pgx.ErrNoRows`로 처리된다. 이는 "post not found"(404)라는 일반적인 메시지를 반환한다.
- 삭제된 게시글에 직접 URL 접근 시, 단순 "not found"가 아닌 "이 게시글은 삭제되었습니다" 메시지를 명확히 반환하도록 개선한다.
- 이를 위해 Repository에 삭제 여부만 확인하는 메서드를 추가하고, Service에서 "not found"와 "deleted"를 구분한다.

**B. 휴지통(Trash) API**
- 사용자가 자신이 soft delete한 게시글 목록을 조회할 수 있다.
- 삭제된 게시글을 복원(`deleted_at = NULL`)할 수 있다.
- 삭제된 게시글을 영구 삭제(hard delete)할 수 있다.
- 자동 영구 삭제 정책: 삭제 후 30일이 지난 게시글은 복원 불가 표시 (MVP에서는 UI 경고만, 자동 purge는 후속 작업).

### 1.2 Why (구현 이유)

- **사용자 경험**: "게시글을 찾을 수 없습니다"는 URL 오타인지 삭제된 것인지 구분이 안 된다. 삭제 여부를 명확히 알려주면 사용자 혼란을 줄인다.
- **데이터 복구**: 실수로 삭제한 게시글을 복원할 수 있는 안전망을 제공한다. 이는 Phase 12에서 soft delete를 채택한 이유("향후 복구 기능 확장 용이")를 실현하는 것이다.
- **X(Twitter) 패리티**: X는 삭제 취소 기능은 없지만, 사용자 실수 방지를 위한 UX 개선은 경쟁력 있는 차별점이다.

### 1.3 핵심 제약

- `handler -> service -> repository` 레이어 구조 준수
- `ctx context.Context`가 모든 Go 메서드의 첫 번째 파라미터
- 인터페이스 기반 DI (PostRepository, PostService 인터페이스에 메서드 추가)
- Cursor-based pagination (offset 금지) -- 단, 현재 코드는 limit/offset 사용 중이므로 기존 패턴과 일관성 유지
- 휴지통은 본인 게시글만 조회/복원/영구삭제 가능 (ReBAC: 소유자 검증)
- Reply(parent_id가 있는 게시글)도 휴지통에 포함

---

## 2. API 엔드포인트 상세

### 2.1 삭제된 게시글 접근 시 응답 개선

기존 엔드포인트 `GET /api/posts/:id`의 동작 변경:

| 상태 | 현재 응답 | 변경 후 응답 |
|------|----------|-------------|
| 게시글 존재 | 200 + 게시글 데이터 | (변경 없음) |
| 게시글 없음 (존재하지 않는 ID) | 404 "post not found" | (변경 없음) |
| 게시글 삭제됨 | 404 "post not found" | **410 "this post has been deleted"** |

**HTTP 410 Gone**: 리소스가 의도적으로 더 이상 사용할 수 없음을 나타내는 표준 상태 코드. 404와 달리 리소스가 존재했었음을 명시한다.

응답 형식:
```json
{
  "success": false,
  "error": "this post has been deleted"
}
```

`apperror` 패키지에 `Gone` 함수 추가 필요:
```go
func Gone(msg string, args ...interface{}) *AppError {
    return &AppError{Code: 410, Message: fmt.Sprintf(msg, args...)}
}
```

### 2.2 휴지통 목록 조회

```
GET /api/users/trash
```

- **인증**: Required (AuthRequired 미들웨어)
- **권한**: 본인의 삭제된 게시글만 조회

**Query Parameters**:

| 파라미터 | 타입 | 필수 | 설명 |
|---------|------|------|------|
| limit | int | N | 한 번에 가져올 개수 (기본값: 20, 최대: 50) |
| cursor | string | N | 다음 페이지 커서 (deleted_at 기준 ISO8601 타임스탬프) |

**성공 응답 (200)**:
```json
{
  "success": true,
  "data": {
    "posts": [
      {
        "id": "uuid",
        "content": "삭제된 게시글 내용",
        "visibility": "public",
        "parentId": null,
        "author": {
          "username": "user1",
          "displayName": "User One",
          "profileImageUrl": "..."
        },
        "likeCount": 5,
        "replyCount": 2,
        "viewCount": 100,
        "repostCount": 0,
        "location": null,
        "media": [],
        "poll": null,
        "createdAt": "2026-03-10T12:00:00Z",
        "deletedAt": "2026-03-15T08:30:00Z",
        "canRestore": true
      }
    ],
    "nextCursor": "2026-03-14T10:00:00Z",
    "hasMore": true
  }
}
```

`canRestore` 필드: 삭제 후 30일 이내이면 `true`, 이후이면 `false`.

**빈 결과 (200)**:
```json
{
  "success": true,
  "data": {
    "posts": [],
    "nextCursor": null,
    "hasMore": false
  }
}
```

### 2.3 삭제된 게시글 복원

```
PUT /api/posts/:id/restore
```

- **인증**: Required (AuthRequired 미들웨어)
- **권한**: 게시글 작성자 본인만 복원 가능

**Request Body**: 없음

**성공 응답 (200)**:
```json
{
  "success": true,
  "data": {
    "message": "post restored successfully",
    "post": {
      "id": "uuid",
      "content": "복원된 게시글 내용",
      ...
    }
  }
}
```

**에러 응답**:

| 상황 | 코드 | 메시지 |
|------|------|--------|
| 인증 안 됨 | 401 | "not authenticated" |
| 게시글 없음 | 404 | "post not found" |
| 삭제되지 않은 게시글 | 400 | "post is not deleted" |
| 본인 게시글 아님 | 403 | "you can only restore your own post" |
| 30일 초과 | 400 | "post cannot be restored after 30 days" |
| Reply의 부모가 삭제됨 | 400 | "cannot restore reply: parent post is deleted" |

### 2.4 영구 삭제

```
DELETE /api/posts/:id/permanent
```

- **인증**: Required (AuthRequired 미들웨어)
- **권한**: 게시글 작성자 본인만 영구 삭제 가능

**Request Body**: 없음

**성공 응답 (200)**:
```json
{
  "success": true,
  "data": {
    "message": "post permanently deleted"
  }
}
```

**에러 응답**:

| 상황 | 코드 | 메시지 |
|------|------|--------|
| 인증 안 됨 | 401 | "not authenticated" |
| 게시글 없음 | 404 | "post not found" |
| 삭제되지 않은 게시글 | 400 | "post is not in trash" |
| 본인 게시글 아님 | 403 | "you can only permanently delete your own post" |

---

## 3. 백엔드 구현 상세

### 3.1 Repository 레이어

#### 3.1.1 PostRepository 인터페이스 추가 메서드

```go
type PostRepository interface {
    // ... 기존 메서드 ...

    // 삭제 여부 확인 (deleted_at 필터 없이 조회)
    ExistsIncludingDeleted(ctx context.Context, id uuid.UUID) (exists bool, isDeleted bool, err error)

    // 삭제된 게시글 조회 (본인 것만, cursor pagination)
    FindDeletedByAuthor(ctx context.Context, authorID uuid.UUID, limit int, cursor *time.Time) ([]model.PostWithAuthor, error)

    // 삭제된 게시글 복원
    Restore(ctx context.Context, id uuid.UUID) error

    // 삭제된 Reply 복원 (reply_count 증가 포함)
    RestoreReply(ctx context.Context, id uuid.UUID, parentID uuid.UUID) error

    // 영구 삭제 (hard delete)
    HardDelete(ctx context.Context, id uuid.UUID) error
}
```

#### 3.1.2 ExistsIncludingDeleted 구현

```sql
SELECT
    EXISTS(SELECT 1 FROM posts WHERE id = $1) AS exists,
    EXISTS(SELECT 1 FROM posts WHERE id = $1 AND deleted_at IS NOT NULL) AS is_deleted
```

목적: 기존 `FindByID`는 `deleted_at IS NULL` 필터가 있어 삭제된 게시글을 구분 불가. 이 메서드로 "존재하지 않음" vs "삭제됨"을 구분한다.

#### 3.1.3 FindDeletedByAuthor 구현

```sql
SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility,
       p.like_count, p.reply_count, p.view_count, p.repost_count,
       p.created_at, p.updated_at, p.deleted_at,
       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
       (u.deleted_at IS NOT NULL OR u.id IS NULL),
       p.location_lat, p.location_lng, p.location_name
FROM posts p
LEFT JOIN users u ON p.author_id = u.id
WHERE p.author_id = $1
  AND p.deleted_at IS NOT NULL
  AND ($2::TIMESTAMPTZ IS NULL OR p.deleted_at < $2)
ORDER BY p.deleted_at DESC
LIMIT $3
```

Cursor: `deleted_at` 기준 내림차순. 최근 삭제된 것이 먼저 표시.

#### 3.1.4 Restore 구현

```sql
UPDATE posts SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL
```

#### 3.1.5 RestoreReply 구현

트랜잭션으로 처리:
1. `UPDATE posts SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL`
2. `UPDATE posts SET reply_count = reply_count + 1 WHERE id = $2` (부모의 reply_count 복원)

SoftDeleteReply의 역연산.

#### 3.1.6 HardDelete 구현

```sql
DELETE FROM posts WHERE id = $1 AND deleted_at IS NOT NULL
```

`deleted_at IS NOT NULL` 조건: 휴지통에 있는(이미 soft delete된) 게시글만 영구 삭제 가능. 활성 게시글을 직접 hard delete하는 것을 방지.

관련 데이터(likes, bookmarks, post_media, polls 등)는 DB의 `ON DELETE CASCADE`로 자동 정리.

#### 3.1.7 FindByIDIncludingDeleted (내부용)

복원/영구삭제 시 권한 검증을 위해 deleted_at 필터 없이 게시글 정보를 조회하는 내부 메서드.

```go
FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*model.PostWithAuthor, error)
```

```sql
SELECT p.id, p.author_id, p.parent_id, p.content, p.visibility,
       p.like_count, p.reply_count, p.view_count, p.repost_count,
       p.created_at, p.updated_at, p.deleted_at,
       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''),
       (u.deleted_at IS NOT NULL OR u.id IS NULL),
       p.location_lat, p.location_lng, p.location_name
FROM posts p
LEFT JOIN users u ON p.author_id = u.id
WHERE p.id = $1
```

### 3.2 Service 레이어

#### 3.2.1 PostService 인터페이스 추가 메서드

```go
type PostService interface {
    // ... 기존 메서드 ...

    // 휴지통 목록 조회
    ListTrash(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) (*dto.TrashListResponse, error)

    // 삭제된 게시글 복원
    RestorePost(ctx context.Context, postID, requesterID uuid.UUID) (*dto.PostDetailResponse, error)

    // 영구 삭제
    PermanentDeletePost(ctx context.Context, postID, requesterID uuid.UUID) error
}
```

#### 3.2.2 GetPostByID 변경

기존 로직에서 `pgx.ErrNoRows` 처리 부분을 수정:

```go
func (s *postService) GetPostByID(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*dto.PostDetailResponse, error) {
    // ... 기존 조회 로직 ...

    if err != nil {
        if err == pgx.ErrNoRows {
            // 삭제된 게시글인지 확인
            exists, isDeleted, checkErr := s.postRepo.ExistsIncludingDeleted(ctx, id)
            if checkErr == nil && exists && isDeleted {
                return nil, apperror.Gone("this post has been deleted")
            }
            return nil, apperror.NotFound("post not found")
        }
        return nil, apperror.Internal("failed to retrieve post")
    }
    // ... 나머지 로직 ...
}
```

#### 3.2.3 ListTrash 구현

```go
func (s *postService) ListTrash(ctx context.Context, userID uuid.UUID, limit int, cursor *time.Time) (*dto.TrashListResponse, error) {
    if limit <= 0 || limit > 50 {
        limit = 20
    }

    // limit+1로 조회하여 hasMore 판단
    posts, err := s.postRepo.FindDeletedByAuthor(ctx, userID, limit+1, cursor)
    if err != nil {
        return nil, apperror.Internal("failed to retrieve trash")
    }

    hasMore := len(posts) > limit
    if hasMore {
        posts = posts[:limit]
    }

    // DTO 변환 (deletedAt, canRestore 포함)
    items := make([]dto.TrashPostResponse, len(posts))
    now := time.Now()
    for i, p := range posts {
        items[i] = dto.ToTrashPostResponse(p, now)
    }

    var nextCursor *string
    if hasMore && len(posts) > 0 {
        last := posts[len(posts)-1]
        if last.DeletedAt != nil {
            c := last.DeletedAt.Format(time.RFC3339)
            nextCursor = &c
        }
    }

    return &dto.TrashListResponse{
        Posts:      items,
        NextCursor: nextCursor,
        HasMore:    hasMore,
    }, nil
}
```

#### 3.2.4 RestorePost 구현

비즈니스 로직 순서:
1. `FindByIDIncludingDeleted`로 게시글 조회
2. 게시글 존재 확인 (없으면 404)
3. `deleted_at` 확인 (NULL이면 400 "post is not deleted")
4. 소유자 검증 (본인 아니면 403)
5. 30일 초과 확인 (초과 시 400)
6. Reply인 경우 부모 게시글 삭제 여부 확인 (부모가 삭제됨이면 400)
7. Reply이면 `RestoreReply`, 아니면 `Restore` 호출
8. 복원된 게시글 조회하여 반환

#### 3.2.5 PermanentDeletePost 구현

비즈니스 로직 순서:
1. `FindByIDIncludingDeleted`로 게시글 조회
2. 게시글 존재 확인 (없으면 404)
3. `deleted_at` 확인 (NULL이면 400 "post is not in trash")
4. 소유자 검증 (본인 아니면 403)
5. `HardDelete` 호출

### 3.3 Handler 레이어

#### 3.3.1 새 핸들러 메서드 (PostHandler에 추가)

```go
// ListTrash - GET /api/users/trash
func (h *PostHandler) ListTrash(c *fiber.Ctx) error

// RestorePost - PUT /api/posts/:id/restore
func (h *PostHandler) RestorePost(c *fiber.Ctx) error

// PermanentDeletePost - DELETE /api/posts/:id/permanent
func (h *PostHandler) PermanentDeletePost(c *fiber.Ctx) error
```

#### 3.3.2 라우터 등록 (router.go)

```go
// 기존 users 그룹에 추가 (/:handle 위에 배치)
users.Get("/trash", middleware.AuthRequired(jwtSecret), p.PostHandler.ListTrash)

// 기존 posts 그룹에 추가
posts.Put("/:id/restore", middleware.AuthRequired(jwtSecret), p.PostHandler.RestorePost)
posts.Delete("/:id/permanent", middleware.AuthRequired(jwtSecret), p.PostHandler.PermanentDeletePost)
```

라우트 순서 주의: `/users/trash`를 `/users/:handle` 위에 등록해야 "trash"가 handle로 해석되지 않는다. 현재 `/users/bookmarks`, `/users/password`, `/users/account`, `/users/profile`이 이미 이 패턴으로 등록되어 있으므로 동일하게 배치.

### 3.4 DTO

#### 3.4.1 새 DTO 타입

```go
// TrashPostResponse - 휴지통 게시글 응답
type TrashPostResponse struct {
    ID         string            `json:"id"`
    AuthorID   string            `json:"authorId"`
    ParentID   *string           `json:"parentId"`
    Content    string            `json:"content"`
    Visibility string            `json:"visibility"`
    Author     PostAuthor        `json:"author"`
    LikeCount  int               `json:"likeCount"`
    ReplyCount int               `json:"replyCount"`
    ViewCount  int               `json:"viewCount"`
    RepostCount int              `json:"repostCount"`
    Location   *LocationResponse `json:"location,omitempty"`
    Media      []MediaResponse   `json:"media,omitempty"`
    Poll       *PollResponse     `json:"poll,omitempty"`
    CreatedAt  string            `json:"createdAt"`
    DeletedAt  string            `json:"deletedAt"`
    CanRestore bool              `json:"canRestore"`
}

// TrashListResponse - 휴지통 목록 응답
type TrashListResponse struct {
    Posts      []TrashPostResponse `json:"posts"`
    NextCursor *string            `json:"nextCursor"`
    HasMore    bool               `json:"hasMore"`
}

// RestorePostResponse - 복원 응답
type RestorePostResponse struct {
    Message string             `json:"message"`
    Post    PostDetailResponse `json:"post"`
}

// PermanentDeleteResponse - 영구 삭제 응답
type PermanentDeleteResponse struct {
    Message string `json:"message"`
}
```

#### 3.4.2 ToTrashPostResponse 변환 함수

```go
const trashRetentionDays = 30

func ToTrashPostResponse(p model.PostWithAuthor, now time.Time) TrashPostResponse {
    canRestore := true
    deletedAtStr := ""
    if p.DeletedAt != nil {
        deletedAtStr = p.DeletedAt.Format("2006-01-02T15:04:05Z")
        if now.Sub(*p.DeletedAt) > time.Duration(trashRetentionDays) * 24 * time.Hour {
            canRestore = false
        }
    }

    // PostDetailResponse와 유사한 변환 로직
    // ...
}
```

### 3.5 apperror 추가

```go
func Gone(msg string, args ...interface{}) *AppError {
    return &AppError{Code: 410, Message: fmt.Sprintf(msg, args...)}
}
```

### 3.6 DB 마이그레이션

별도 마이그레이션 불필요. `posts.deleted_at` 컬럼은 이미 존재한다. 인덱스 추가 권장:

```sql
-- 019_trash_index.up.sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_author_deleted
ON posts (author_id, deleted_at DESC)
WHERE deleted_at IS NOT NULL;
```

```sql
-- 019_trash_index.down.sql
DROP INDEX IF EXISTS idx_posts_author_deleted;
```

이 부분 인덱스는 `FindDeletedByAuthor` 쿼리를 최적화한다.

---

## 4. 프론트엔드 구현 상세

### 4.1 새 타입 (types.ts 또는 api.ts)

```typescript
interface TrashPost {
  id: string;
  authorId: string;
  parentId: string | null;
  content: string;
  visibility: string;
  author: PostAuthor;
  likeCount: number;
  replyCount: number;
  viewCount: number;
  repostCount: number;
  location?: Location;
  media?: MediaItem[];
  poll?: Poll;
  createdAt: string;
  deletedAt: string;
  canRestore: boolean;
}

interface TrashListResponse {
  posts: TrashPost[];
  nextCursor: string | null;
  hasMore: boolean;
}
```

### 4.2 Custom Hooks

```typescript
// hooks/useTrash.ts

// 휴지통 목록 조회 (useInfiniteQuery)
export function useTrash() {
  return useInfiniteQuery({
    queryKey: ["trash"],
    queryFn: ({ pageParam }) => fetchTrash(pageParam),
    getNextPageParam: (lastPage) => lastPage.hasMore ? lastPage.nextCursor : undefined,
    initialPageParam: undefined as string | undefined,
  });
}

// 게시글 복원 (useMutation)
export function useRestorePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (postId: string) => restorePost(postId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["trash"] });
      queryClient.invalidateQueries({ queryKey: ["posts"] });
      toast.success("게시글이 복원되었습니다");
    },
  });
}

// 영구 삭제 (useMutation)
export function usePermanentDelete() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (postId: string) => permanentDeletePost(postId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["trash"] });
      toast.success("게시글이 영구 삭제되었습니다");
    },
  });
}
```

### 4.3 TrashPage 컴포넌트

경로: `/trash` (SettingsPage와 유사한 독립 페이지)

구성:
- 헤더: "휴지통" 제목 + ArrowLeft 뒤로가기
- 안내 문구: "삭제된 게시글은 30일 후 자동으로 영구 삭제됩니다"
- 게시글 목록: 각 항목에 내용 미리보기, 삭제 일시, 복원/영구삭제 버튼
- 무한 스크롤 (IntersectionObserver)
- 빈 상태: "휴지통이 비어있습니다"

각 휴지통 항목 UI:
```
+------------------------------------------+
| [Avatar] @username                       |
| 게시글 내용 미리보기 (최대 2줄)            |
| 삭제일: 2026-03-15 08:30                 |
| [복원] [영구 삭제]                        |
+------------------------------------------+
```

- 복원 불가(30일 초과) 게시글: 복원 버튼 비활성화 + "복원 기간이 만료되었습니다" 텍스트
- 영구 삭제: AlertDialog로 확인 ("이 작업은 되돌릴 수 없습니다")

### 4.4 삭제된 게시글 접근 시 UI

PostDetailPage에서 410 응답 수신 시:
- 전체 화면 메시지: "이 게시글은 삭제되었습니다"
- 홈으로 돌아가기 버튼
- 기존 404 처리와 분리

### 4.5 네비게이션

- SettingsPage에 "휴지통" 링크 추가 (Trash2 아이콘 사용)
- 또는 Sidebar 프로필 영역 근처에 휴지통 메뉴 추가
- App.tsx에 `/trash` 라우트 등록 (AuthRequired)

### 4.6 Optimistic UI

- 복원: 목록에서 즉시 제거 + 성공 시 toast
- 영구 삭제: 목록에서 즉시 제거 + 성공 시 toast
- 실패 시 목록 복원 + 에러 toast

---

## 5. 에러 처리

### 5.1 백엔드 에러 코드 매핑

| 상황 | HTTP 코드 | apperror 함수 | 메시지 |
|------|-----------|--------------|--------|
| 삭제된 게시글 직접 접근 | 410 | `Gone` | "this post has been deleted" |
| 게시글 없음 | 404 | `NotFound` | "post not found" |
| 삭제되지 않은 게시글 복원 시도 | 400 | `BadRequest` | "post is not deleted" |
| 삭제되지 않은 게시글 영구삭제 시도 | 400 | `BadRequest` | "post is not in trash" |
| 타인 게시글 복원 시도 | 403 | `Forbidden` | "you can only restore your own post" |
| 타인 게시글 영구삭제 시도 | 403 | `Forbidden` | "you can only permanently delete your own post" |
| 30일 초과 복원 시도 | 400 | `BadRequest` | "post cannot be restored after 30 days" |
| 부모 삭제된 Reply 복원 | 400 | `BadRequest` | "cannot restore reply: parent post is deleted" |
| 인증 안 됨 | 401 | `Unauthorized` | "not authenticated" |

### 5.2 프론트엔드 에러 처리

- 410 응답: PostDetailPage에서 별도 "삭제된 게시글" 화면 렌더링
- 403 응답: toast.error("권한이 없습니다")
- 400 응답: toast.error(서버 메시지 그대로 표시)
- 네트워크 에러: toast.error("네트워크 오류가 발생했습니다")

---

## 6. 보안 고려사항

### 6.1 접근 제어 (ReBAC)

- **휴지통 조회**: JWT에서 추출한 `userID`로만 조회. 쿼리에 `WHERE author_id = $1` 강제.
- **복원/영구삭제**: Service 레이어에서 `post.AuthorID == requesterID` 검증 필수.
- **삭제된 게시글 내용 노출 방지**: 410 응답 시 게시글 내용을 포함하지 않는다. 메시지만 반환.

### 6.2 Rate Limiting (후속 작업)

- 복원/영구삭제 API에 rate limit 적용 권장 (남용 방지)
- MVP에서는 미적용, 후속 이슈로 등록

### 6.3 영구 삭제 안전장치

- `HardDelete`는 반드시 `deleted_at IS NOT NULL` 조건을 포함하여 활성 게시글이 실수로 삭제되는 것을 방지
- 프론트엔드에서 AlertDialog 이중 확인
- 영구 삭제 전 poll/media 데이터는 DB CASCADE로 자동 정리

### 6.4 데이터 무결성

- Reply 복원 시 부모 게시글 존재 + 미삭제 확인 필수
- Reply 복원 시 부모의 `reply_count` 증가 (트랜잭션)
- 부모 게시글 영구 삭제 시 하위 Reply도 CASCADE로 함께 삭제

---

## 7. 테스트 계획

### 7.1 Go Service 테스트 (Table-driven)

```go
func TestListTrash(t *testing.T) {
    tests := []struct {
        name        string
        userID      uuid.UUID
        mockPosts   []model.PostWithAuthor
        wantLen     int
        wantHasMore bool
        wantErr     bool
    }{
        {"empty trash", ...},
        {"with posts", ...},
        {"pagination hasMore true", ...},
        {"pagination hasMore false", ...},
    }
}

func TestRestorePost(t *testing.T) {
    tests := []struct {
        name       string
        postID     uuid.UUID
        requester  uuid.UUID
        mockPost   *model.PostWithAuthor
        wantErr    bool
        wantErrMsg string
    }{
        {"success - restore own post", ...},
        {"success - restore own reply with live parent", ...},
        {"fail - post not found", ...},
        {"fail - post not deleted", ...},
        {"fail - not owner", ...},
        {"fail - past 30 days", ...},
        {"fail - reply parent deleted", ...},
    }
}

func TestPermanentDeletePost(t *testing.T) {
    tests := []struct {
        name       string
        postID     uuid.UUID
        requester  uuid.UUID
        mockPost   *model.PostWithAuthor
        wantErr    bool
        wantErrMsg string
    }{
        {"success - permanent delete own post", ...},
        {"fail - post not found", ...},
        {"fail - post not in trash", ...},
        {"fail - not owner", ...},
    }
}

func TestGetPostByID_DeletedPost(t *testing.T) {
    tests := []struct {
        name       string
        postID     uuid.UUID
        mockExists bool
        mockDeleted bool
        wantCode   int
        wantMsg    string
    }{
        {"deleted post returns 410", ...},
        {"non-existent post returns 404", ...},
    }
}
```

### 7.2 Mock 인터페이스 업데이트

기존 테스트 파일의 mock 구조체에 새 메서드 stub 추가 필요:
- `mockPostRepo` in `post_service_test.go`
- `mockPostRepoForPoll` in `poll_service_test.go`
- `mockPostRepoForBookmark` in `bookmark_service_test.go`
- `mockPostRepoForFollow` in `follow_service_test.go` (만약 PostRepository 사용 시)

### 7.3 React 테스트

- TrashPage 렌더링 테스트 (빈 상태, 데이터 있는 상태)
- 복원 버튼 클릭 시 API 호출 확인 (MSW)
- 영구 삭제 AlertDialog 확인 절차 테스트
- 복원 불가(canRestore=false) 상태에서 복원 버튼 비활성화 확인
- PostDetailPage 410 응답 시 "삭제된 게시글" 화면 표시 확인

---

## 8. 수락 기준 (Acceptance Criteria)

1. [ ] `GET /api/posts/:id`에서 삭제된 게시글 접근 시 HTTP 410 + "this post has been deleted" 메시지 반환
2. [ ] `GET /api/posts/:id`에서 존재하지 않는 게시글은 기존대로 HTTP 404 반환
3. [ ] `GET /api/users/trash`에서 본인이 삭제한 게시글 목록을 cursor pagination으로 조회 가능
4. [ ] 휴지통 목록에 `deletedAt`, `canRestore` 필드 포함
5. [ ] `PUT /api/posts/:id/restore`로 삭제된 게시글 복원 가능 (deleted_at = NULL)
6. [ ] Reply 복원 시 부모의 reply_count 증가
7. [ ] Reply 복원 시 부모가 삭제 상태이면 400 에러
8. [ ] 삭제 후 30일 초과 게시글은 복원 불가 (400 에러)
9. [ ] `DELETE /api/posts/:id/permanent`로 영구 삭제 가능 (DB에서 완전 제거)
10. [ ] 영구 삭제는 이미 soft delete된 게시글만 대상 (활성 게시글 보호)
11. [ ] 모든 휴지통 API는 본인 게시글만 조작 가능 (403 검증)
12. [ ] 프론트엔드 TrashPage에서 휴지통 목록 조회, 복원, 영구 삭제 UI 제공
13. [ ] PostDetailPage에서 410 응답 시 "이 게시글은 삭제되었습니다" 화면 표시
14. [ ] Go service 테스트 (table-driven) 전체 통과
15. [ ] `bun run check` (프론트엔드 typecheck + lint) 통과

---

## 9. 엣지 케이스 목록

1. **부모 게시글과 Reply가 동시에 휴지통에 있는 경우**: 부모를 먼저 복원해야 Reply 복원 가능
2. **부모 게시글 영구 삭제 시**: CASCADE로 하위 Reply도 함께 삭제됨 (휴지통의 Reply 포함)
3. **동일 게시글 복원 중복 요청**: 첫 번째만 성공, 이후는 400 "post is not deleted"
4. **동일 게시글 영구 삭제 중복 요청**: 첫 번째만 성공, 이후는 404 "post not found"
5. **탈퇴한 사용자의 게시글**: 사용자가 탈퇴(soft delete)한 경우 JWT가 무효화되므로 휴지통 접근 자체가 불가
6. **게시글에 연결된 미디어/투표**: 복원 시 미디어/투표 데이터가 그대로 유지됨 (soft delete는 posts 테이블만 대상)
7. **cursor 값이 유효하지 않은 ISO8601**: 400 에러 반환
8. **삭제 직후(deleted_at이 매우 최근) 복원**: 정상 동작, canRestore = true

---

## 10. 의존성 및 제약사항

### 의존성
- Phase 12 (Post/Reply Soft Delete) 완료 -- 이미 완료됨
- `posts.deleted_at` 컬럼 존재 -- 이미 존재
- `apperror` 패키지 -- `Gone` 함수 추가 필요
- `ON DELETE CASCADE` 설정 -- 이미 적용됨

### 제약사항
- 자동 purge(30일 초과 게시글 자동 영구 삭제)는 이 스펙 범위 밖. 후속 작업으로 cron job 또는 pg_cron 도입 검토.
- 현재 코드의 페이지네이션이 limit/offset 기반이므로, 휴지통 API의 cursor pagination은 `deleted_at` 타임스탬프 기반으로 새롭게 구현. 기존 API의 cursor 전환은 별도 이슈.

---

## 11. 구현 순서 권장

1. **Phase A**: `apperror.Gone` 추가 + `ExistsIncludingDeleted` + `GetPostByID` 410 응답 개선
2. **Phase B**: Repository 메서드 추가 (FindByIDIncludingDeleted, FindDeletedByAuthor, Restore, RestoreReply, HardDelete) + DB 인덱스 마이그레이션
3. **Phase C**: Service 메서드 추가 (ListTrash, RestorePost, PermanentDeletePost) + DTO
4. **Phase D**: Handler 메서드 + 라우터 등록
5. **Phase E**: Go 테스트 작성 + 전체 테스트 통과 확인
6. **Phase F**: 프론트엔드 hooks + TrashPage + PostDetailPage 410 처리
7. **Phase G**: 프론트엔드 typecheck/lint 통과 + 코드 리뷰
