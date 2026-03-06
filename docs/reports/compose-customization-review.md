# Code Review: Compose Customization Feature

**Date**: 2026-03-06
**Reviewer**: Claude Opus 4.6
**Branch**: feat/backend-infra-improvements
**Scope**: Media upload, Poll, Location, Markdown rendering

---

## Summary

| Severity | Count |
|----------|-------|
| Critical | 4     |
| Warning  | 8     |
| Info     | 5     |

---

## Critical Issues

### [BE-01] DB Table Name Mismatch: migration creates `post_media`, repository queries `media`

- **Severity**: CRITICAL
- **Location**: `backend/migrations/009_create_post_media.up.sql:1` vs `backend/internal/repository/media_repository.go:29-30`
- **Problem**: The migration creates a table called `post_media`, but all repository SQL queries reference a table called `media`. This will cause runtime SQL errors on every media operation.
- **Rationale**: Migration DDL and application queries must target the same table name.
- **Fix**: Either rename the migration table to `media`, or update all SQL queries in `media_repository.go` to use `post_media`.

```sql
-- Option A: Change migration to use `media`
CREATE TABLE media (
    ...
);
```

### [BE-02] Media not linked to posts after upload -- mediaIds are accepted but never processed

- **Severity**: CRITICAL
- **Location**: `backend/internal/service/post_service.go:42-112`
- **Problem**: `CreatePostRequest` accepts `mediaIds`, and the frontend sends them, but `postService.CreatePost` never calls `mediaRepo.LinkToPost()`. The `postService` struct does not even have a `mediaRepo` field. Uploaded media will remain orphaned (with `post_id = NULL`) and will never appear on posts.
- **Rationale**: The entire media-post association feature is non-functional.
- **Fix**: Inject `MediaRepository` into `postService` and call `LinkToPost` after post creation:

```go
type postService struct {
    postRepo  repository.PostRepository
    pollRepo  repository.PollRepository
    mediaRepo repository.MediaRepository  // add this
}

// In CreatePost, after s.postRepo.Create(ctx, post):
if len(req.MediaIds) > 0 {
    var mediaIDs []uuid.UUID
    for _, idStr := range req.MediaIds {
        id, err := uuid.Parse(idStr)
        if err != nil {
            return nil, apperror.BadRequest("invalid media ID: %s", idStr)
        }
        mediaIDs = append(mediaIDs, id)
    }
    if err := s.mediaRepo.LinkToPost(ctx, mediaIDs, post.ID); err != nil {
        return nil, apperror.Internal("failed to link media to post")
    }
}
```

### [BE-03] XHR upload request lacks authentication credentials

- **Severity**: CRITICAL
- **Location**: `frontend/src/hooks/useMediaUpload.ts:78-79`
- **Problem**: The `XMLHttpRequest` for media upload does not set `withCredentials = true` and does not include any Authorization header. The backend requires JWT auth via `middleware.AuthRequired`. This means every upload request will be rejected with 401 Unauthorized.
- **Rationale**: The rest of the app likely uses `fetch` with cookies or headers configured via a shared client. The XHR here bypasses that.
- **Fix**: Add credentials to the XHR:

```typescript
xhr.open("POST", "/api/media/upload");
xhr.withCredentials = true;  // if using httpOnly cookies
// OR: xhr.setRequestHeader("Authorization", `Bearer ${getToken()}`);
xhr.send(formData);
```

### [BE-04] Path traversal vulnerability in local storage Delete

- **Severity**: CRITICAL
- **Location**: `backend/internal/storage/local_storage.go:51-60`
- **Problem**: The `Delete` method accepts a URL string, strips the leading slash, and passes it directly to `os.Remove()` without validating that the resolved path stays within `basePath`. A malicious URL like `/../../etc/important-file` would delete files outside the upload directory.
- **Rationale**: Security -- file system operations must validate path boundaries.
- **Fix**: Validate the resolved path is within basePath:

```go
func (s *localStorage) Delete(ctx context.Context, url string) error {
    path := strings.TrimPrefix(url, "/")
    absPath, err := filepath.Abs(path)
    if err != nil {
        return fmt.Errorf("failed to resolve path: %w", err)
    }
    absBase, _ := filepath.Abs(s.basePath)
    if !strings.HasPrefix(absPath, absBase+string(os.PathSeparator)) {
        return fmt.Errorf("path traversal attempt blocked")
    }
    if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to delete file: %w", err)
    }
    return nil
}
```

---

## Warning Issues

### [BE-05] `time.Now()` used directly in multiple service functions -- not injected

- **Severity**: WARNING
- **Location**: `backend/internal/service/poll_service.go:52,96` and `backend/internal/service/post_service.go:86`
- **Problem**: `time.Now()` is called directly inside `Vote()`, `buildPollResponse()`, and `CreatePost()`. This violates the project rule requiring side effects like `time.Now()` to be injected externally. It makes these functions untestable for time-dependent logic (e.g., poll expiration).
- **Rationale**: Go backend rule -- "Are side effects (time.Now, etc.) injected externally?"
- **Fix**: Inject a `func() time.Time` (or a clock interface) into the service structs.

### [BE-06] Error comparison using `==` instead of `errors.Is()`

- **Severity**: WARNING
- **Location**: `backend/internal/service/poll_service.go:34`, `backend/internal/service/post_service.go:125,235,276,305`, `backend/internal/repository/poll_repository.go:73,154`
- **Problem**: `err == pgx.ErrNoRows` is used throughout instead of `errors.Is(err, pgx.ErrNoRows)`. If the error is wrapped at any point, the comparison will fail silently and produce incorrect behavior (e.g., returning 500 instead of 404).
- **Rationale**: Go standard practice since Go 1.13; consistent with `bookmark_service.go` which already uses `errors.Is()`.
- **Fix**: Replace all `err == pgx.ErrNoRows` with `errors.Is(err, pgx.ErrNoRows)`.

### [BE-07] Sentinel errors not defined at package top for poll_service

- **Severity**: WARNING
- **Location**: `backend/internal/service/poll_service.go` (entire file)
- **Problem**: The poll service uses inline error creation via `apperror.BadRequest(...)` and `apperror.NotFound(...)` throughout, but does not define sentinel errors at the package level. The media service correctly defines `ErrUnsupportedMediaType` and `ErrFileTooLarge` at the top, but the poll service does not follow the same pattern.
- **Rationale**: Go backend rule -- "Are sentinel errors defined at the top of the package?"
- **Fix**: Define sentinel errors at the top of `poll_service.go`:

```go
var (
    ErrPollNotFound    = apperror.NotFound("poll not found")
    ErrPollExpired     = apperror.BadRequest("poll has expired")
    ErrAlreadyVoted    = apperror.Conflict("already voted on this poll")
    ErrOwnPollVote     = apperror.BadRequest("cannot vote on your own poll")
)
```

### [BE-08] Missing media ownership validation during LinkToPost

- **Severity**: WARNING
- **Location**: `backend/internal/repository/media_repository.go:95-113`
- **Problem**: `LinkToPost` does not verify that the media items actually belong to the requesting user (`uploader_id`). A user could attach another user's uploaded media to their own post by guessing/knowing the media UUID.
- **Rationale**: ReBAC rule -- service layer must verify user relationships before modifying resources.
- **Fix**: Add a `WHERE uploader_id = $4` condition to the UPDATE query, or validate ownership in the service layer before calling `LinkToPost`.

### [BE-09] Upload method uses `time.Now()` directly for directory naming

- **Severity**: WARNING
- **Location**: `backend/internal/storage/local_storage.go:23`
- **Problem**: `time.Now()` is called directly in the Upload method for generating the directory path. This violates the "externally inject side effects" rule and makes the storage behavior non-deterministic in tests.
- **Rationale**: Go backend rule -- side effects should be injected.
- **Fix**: Accept a `nowFunc func() time.Time` in the `localStorage` struct.

### [FE-01] `toast.error(locationError)` called directly in render body -- causes infinite re-render loop

- **Severity**: WARNING
- **Location**: `frontend/src/components/ComposeForm.tsx:149-151`
- **Problem**: The code `if (locationError) { toast.error(locationError); }` is placed directly in the component render body (not inside a `useEffect`). Every re-render triggered by the toast will see the error still set, triggering another toast, creating an infinite loop.
- **Rationale**: Side effects in React must be placed in `useEffect` or event handlers.
- **Fix**:

```tsx
useEffect(() => {
  if (locationError) {
    toast.error(locationError);
  }
}, [locationError]);
```

### [FE-02] Winning option calculation uses pre-optimistic values

- **Severity**: WARNING
- **Location**: `frontend/src/components/PollDisplay.tsx:57-58`
- **Problem**: The `isWinning` calculation uses `option.voteCount` (the original values) in `Math.max(...)` but compares with the potentially optimistic `voteCount`. This can lead to incorrect highlighting when the user just voted.
- **Rationale**: Logic inconsistency between optimistic and non-optimistic values.
- **Fix**: Compute max from the same set of adjusted vote counts used for display:

```tsx
const adjustedCounts = poll.options.map((o, i) =>
  optimisticVotedIndex === i ? o.voteCount + 1 : o.voteCount
);
const maxCount = Math.max(...adjustedCounts);
// then: isWinning = showResults && adjustedCounts[index] === maxCount;
```

### [FE-03] File upload validation race condition

- **Severity**: WARNING
- **Location**: `frontend/src/hooks/useMediaUpload.ts:100-131`
- **Problem**: The `addFiles` function checks `mediaItems.length` for the max count validation, but then calls `Promise.all(files.map(uploadFile))` which adds items asynchronously. If `addFiles` is called rapidly twice with 3 images each (total 6), both calls could pass the `mediaItems.length >= 4` check since neither has updated state yet, resulting in more than 4 images being uploaded.
- **Rationale**: State-based validation before async operations can be stale.
- **Fix**: Use a ref to track the total pending + completed count, or serialize uploads.

---

## Info Issues

### [BE-10] PostWithAuthor has duplicate location fields

- **Severity**: INFO
- **Location**: `backend/internal/model/post.go:32-48`
- **Problem**: `PostWithAuthor` embeds `Post` (which already has `LocationLat`, `LocationLng`, `LocationName`) and then re-declares these same three fields. This means the embedded fields are shadowed, which could lead to confusion about which fields are populated.
- **Rationale**: Code clarity and maintainability.
- **Fix**: Remove the duplicate `LocationLat`, `LocationLng`, `LocationName` fields from `PostWithAuthor` since they are inherited from the embedded `Post`.

### [BE-11] Post creation and poll creation are not in a single transaction

- **Severity**: INFO
- **Location**: `backend/internal/service/post_service.go:81-103`
- **Problem**: `s.postRepo.Create(ctx, post)` and `s.pollRepo.CreatePoll(ctx, poll, options)` are separate operations. If poll creation fails after the post is committed, you end up with an orphaned post without its intended poll.
- **Rationale**: Data consistency.
- **Fix**: Either use a single transaction wrapping both operations, or implement a compensation/cleanup mechanism.

### [BE-12] Content-Type from multipart header is trusted without verification

- **Severity**: INFO
- **Location**: `backend/internal/handler/media_handler.go:41`
- **Problem**: The MIME type comes from the HTTP multipart header (`Content-Type`), which is set by the client and can be spoofed. A user could upload a malicious file with a fake `image/jpeg` Content-Type.
- **Rationale**: Defense in depth for file upload security.
- **Fix**: Use magic bytes detection (e.g., `http.DetectContentType` or the `gabriel-vasile/mimetype` library) to verify the actual file content type.

### [FE-04] ParentPostCard does not render media, polls, or location

- **Severity**: INFO
- **Location**: `frontend/src/components/ParentPostCard.tsx:1-49`
- **Problem**: When viewing a reply's detail page, the parent post chain uses `ParentPostCard` which only renders text content (via MarkdownRenderer). If the parent post has media, a poll, or a location, those are not displayed.
- **Rationale**: Feature completeness / consistency.
- **Fix**: Add `MediaGrid`, `PollDisplay`, and location display to `ParentPostCard`, consistent with `PostCard`.

### [FE-05] ReplyCard does not render media, polls, or location

- **Severity**: INFO
- **Location**: `frontend/src/components/ReplyCard.tsx:113-115`
- **Problem**: Similar to FE-04, replies can have media/polls/location (the `PostDetail` type supports them) but `ReplyCard` only renders text via `MarkdownRenderer`.
- **Rationale**: Feature completeness.
- **Fix**: Add media/poll/location rendering to `ReplyCard`.

---

## Checklist Results

### Go Backend Checklist

| # | Rule | Status | Notes |
|---|------|--------|-------|
| 1 | `ctx context.Context` first parameter | PASS | All service/repository functions follow this |
| 2 | No pointer interfaces (`*I`) | PASS | All interfaces are passed by value |
| 3 | Error wrapping with `fmt.Errorf` | PASS | No `panic()` found anywhere |
| 4 | Sentinel errors at package top | PARTIAL | `media_service.go` passes; `poll_service.go` fails (BE-07) |
| 5 | Empty slices use `var` not `make` | PASS | `var media []model.Media`, `var options []model.PollOption`, etc. |
| 6 | No map order dependence | PASS | Maps used for lookup only, no order-dependent iteration |
| 7 | No `init()` functions | PASS | No init functions found |
| 8 | Side effects injected | FAIL | `time.Now()` used directly (BE-05, BE-09) |
| 9 | File upload security | PARTIAL | MIME type allowed-list exists, UUID filenames used, size limits enforced; but Content-Type is client-trusted (BE-12), path traversal in Delete (BE-04) |
| 10 | handler->service->repo layers | PASS | Clean layer separation throughout |

### React Frontend Checklist

| # | Rule | Status | Notes |
|---|------|--------|-------|
| 1 | Functional components | PASS | All components are function components |
| 2 | PascalCase naming | PASS | MarkdownRenderer, MediaGrid, PollDisplay, etc. |
| 3 | API calls in custom hooks | PASS | useMediaUpload, useVote, useGeolocation |
| 4 | React Query for server state | PASS | useVote uses useMutation + queryClient invalidation |
| 5 | XSS: rehype-sanitize | PASS | Applied with custom schema in MarkdownRenderer |

---

## Architecture Notes

- The overall handler -> service -> repository layering is well maintained.
- DI via uber/fx is properly configured with module files for each layer.
- Migration files are well-structured with proper indexes and foreign keys.
- The poll voting system has good protection against double-voting (UNIQUE constraint + application-level check).
- The frontend component decomposition (ComposeForm, PollCreator, MediaPreview, PollDisplay, MediaGrid) is clean and follows single responsibility.

---

## Priority Fix Order

1. **BE-01** (table name mismatch) -- nothing works without this
2. **BE-02** (media not linked to posts) -- core feature broken
3. **BE-03** (XHR missing auth) -- uploads will 401
4. **BE-04** (path traversal) -- security vulnerability
5. **FE-01** (infinite toast loop) -- UX-breaking bug
6. **BE-06** (errors.Is) -- silent incorrect error handling
7. **BE-08** (media ownership) -- security concern
8. Remaining warnings and info items
