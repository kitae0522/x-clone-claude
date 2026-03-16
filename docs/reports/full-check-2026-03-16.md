# Full Quality Check Report — 2026-03-16

**Branch**: `feat/issue-56-settings-password-delete`
**Scope**: Issue #65 — 탈퇴 사용자 게시글 작성자 익명 처리

---

## 1. Build & Test

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go vet ./...` | PASS |
| `go test ./...` | PASS (4/4 packages) |
| `bun run check` (tsc + eslint) | PASS |

---

## 2. Code Review (변경 사항)

### SQL Column/Scan Alignment

모든 repository 메서드의 SELECT 컬럼 수와 Scan 인자 수가 일치합니다.

| Method | SELECT cols | Scan args | Match |
|--------|------------|-----------|-------|
| `FindByID` | 17 | 17 | OK |
| `FindAll` | 20 | 20 | OK |
| `FindByIDWithUser` | 20 | 20 | OK |
| `FindAllWithUser` | 23 | 23 | OK |
| `FindRepliesByPostID` | 17 | 17 | OK |
| `FindRepliesByPostIDWithUser` | 20 | 20 | OK |
| `FindAuthorReplyByPostID` | 17 | 17 | OK |
| `FindAuthorReplyByPostIDWithUser` | 20 | 20 | OK |
| `FindByAuthorHandle` | 20 | 20 | OK |
| `FindByAuthorHandleWithUser` | 23 | 23 | OK |
| `FindRepliesByAuthorHandle` | 22 | 22 | OK |
| `FindRepliesByAuthorHandleWithUser` | 25 | 25 | OK |
| `FindLikedByUserHandle` | 17 | 17 | OK |
| `FindLikedByUserHandleWithViewer` | 20 | 20 | OK |
| Bookmark `ListByUserID` | 15 | 15 | OK |

### Findings

| # | Severity | File | Issue |
|---|----------|------|-------|
| 1 | **CRITICAL** | `model/post.go:47-49` | `LocationLat/Lng/Name` 필드가 embedded `Post`와 `PostWithAuthor`에 중복 선언 — Go field shadowing 발생. 기존 코드부터 존재하는 문제이며 현재 scan이 outer field를 사용하므로 런타임 오류는 없으나 `p.Post.LocationLat` 접근 시 nil 반환 |
| 2 | **WARNING** | `bookmark_repository.go:76` | Bookmark 쿼리에 `view_count`, `repost_count` 미포함 — 프론트에서 0으로 표시됨 (기존 문제) |
| 3 | **WARNING** | `bookmark_repository.go:76` | Bookmark 쿼리에 `location_lat/lng/name` 미포함 — 위치 정보 누락 (기존 문제) |
| 4 | **WARNING** | `dto/post_dto.go:194` | Parent author 익명화가 `username == nil \|\| ""` 휴리스틱 사용. 명시적 `ParentAuthorDeleted` bool이 더 안전 |
| 5 | **WARNING** | `user_service.go` | Soft-deleted 사용자 JWT가 만료까지 유효 (쿠키만 제거) |
| 6 | **WARNING** | `PostCard.tsx:247` | "replying to @deleted" 링크에 isDeleted 가드 미적용 |
| 7 | **WARNING** | `ParentPostCard.tsx:20` | Avatar 클릭에 isDeleted 가드 미적용 |
| 8 | **WARNING** | `PostDetailPage.tsx:59,385` | 탈퇴 사용자 username="deleted"로 인해 다른 탈퇴 사용자 답글에도 OP 뱃지 표시 가능 |
| 9 | INFO | `user_repository.go:118` | Soft delete 시 follows/likes/bookmarks 정리 안 됨 — count 불일치 가능 |
| 10 | INFO | `ParentPostCard.tsx:34` | 비삭제 작성자에도 ProfileHoverCard 미적용 (기존 문제) |
| 11 | INFO | `dto/post_dto.go:154` | 탈퇴 사용자 게시글에 `authorId` UUID 노출 (프라이버시) |
| 12 | INFO | `dto/user_dto.go:7` | 비밀번호 복잡성 검증 없음 (길이만 체크) |

### 이번 PR에서 수정 권장 (merge 전)

- **#4**: `ParentAuthorDeleted` 명시적 boolean 추가 또는 현재 휴리스틱 허용 여부 결정
- **#6**: PostCard "replying to" 영역 isDeleted 스타일 가드
- **#7**: ParentPostCard avatar 클릭 가드
- **#8**: OP 뱃지 비교를 `authorId`로 변경하거나 탈퇴 시 OP 뱃지 숨김

### 기존 코드 문제 (별도 이슈 권장)

- **#1**: `PostWithAuthor` Location 필드 shadowing 제거
- **#2, #3**: Bookmark 쿼리 view_count/repost_count/location 추가
- **#5**: JWT 무효화 전략 (token blocklist 또는 tokenVersion)
- **#9**: Soft delete cascade 처리

---

## 3. DB Migration 정합성

### 3-1. Migration Up/Down Pairs

총 18개 마이그레이션, **모두 up/down 쌍 완전**.

| Migration | Up | Down |
|-----------|-----|------|
| 001 ~ 018 | 18/18 | 18/18 |

### 3-2. FK Index Coverage

| Table.Column | References | Index | Status |
|-------------|------------|-------|--------|
| `posts.author_id` | `users(id)` | `idx_posts_author_id` | OK |
| `follows.follower_id` | `users(id)` | `idx_follows_follower_id` | OK |
| `follows.following_id` | `users(id)` | `idx_follows_following_id` | OK |
| **`likes.user_id`** | **`users(id)`** | **NONE** | **MISSING** |
| `likes.post_id` | `posts(id)` | `idx_likes_post_id` | OK |
| `posts.parent_id` | `posts(id)` | `idx_posts_parent_id` | OK |
| `bookmarks.user_id` | `users(id)` | composite PK | OK |
| `bookmarks.post_id` | `posts(id)` | composite PK | OK |
| `post_media.post_id` | `posts(id)` | `idx_post_media_post_id` | OK |
| `post_media.uploader_id` | `users(id)` | `idx_post_media_uploader_id` | OK |
| `polls.post_id` | `posts(id)` | `idx_polls_post_id` | OK |
| `poll_options.poll_id` | `polls(id)` | `idx_poll_options_poll_id` | OK |
| `poll_votes.poll_id` | `polls(id)` | `idx_poll_votes_poll_id` | OK |
| `poll_votes.user_id` | `users(id)` | `idx_poll_votes_user_id` | OK |
| `reposts.user_id` | `users(id)` | `idx_reposts_user_created` | OK |
| `reposts.post_id` | `posts(id)` | `idx_reposts_post_id` | OK |

**누락 인덱스 1건**: `likes.user_id` — 사용자별 좋아요 조회 성능에 영향

### 3-3. NOT NULL Default 검사

모든 NOT NULL 컬럼이 DEFAULT 값을 가지거나, INSERT 시 명시적으로 설정되는 필수 필드입니다. 문제 없음.

---

## 4. E2E 테스트 (Playwright MCP)

> **SKIPPED** — Playwright MCP 서버가 현재 환경에 설정되어 있지 않습니다.

수동 검증 항목:
- [ ] 로그인 플로우
- [ ] 게시물 CRUD 플로우
- [ ] 좋아요/리포스트 인터랙션
- [ ] 실시간 알림 수신
- [ ] **탈퇴 사용자 게시글이 "탈퇴한 사용자"로 표시되는지**
- [ ] **탈퇴 사용자 프로필 클릭 시 네비게이션 미동작**

---

## 5. DB 정합성 (PostgreSQL MCP)

> **SKIPPED** — PostgreSQL MCP 서버가 현재 환경에 설정되어 있지 않습니다.

마이그레이션 파일 기반 정적 분석 결과는 Section 3 참조.

---

## Summary

| Category | Status |
|----------|--------|
| Go Build | PASS |
| Go Vet | PASS |
| Go Test | PASS |
| Frontend Check | PASS |
| SQL Column/Scan Alignment | PASS (all 15 methods) |
| Migration Pairs | PASS (18/18) |
| FK Index Coverage | 1 MISSING (`likes.user_id`) |
| NOT NULL Defaults | PASS |
| Code Review | 4 WARNING (이번 PR), 4 WARNING (기존), 4 INFO |
| E2E (Playwright) | SKIPPED (MCP 미설정) |
| DB Live (PostgreSQL) | SKIPPED (MCP 미설정) |
