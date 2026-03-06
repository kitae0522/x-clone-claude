# Pipeline Report: Visibility + Compose Page

**Date**: 2026-03-06
**Branch**: feat/issue-37-compose-customization

---

## Summary

3가지 기능을 구현: (1) 게시물 공개 범위 설정, (2) 별도 글쓰기 페이지, (3) 답글 시 부모글 컨텍스트 표시.

## Changes

### Backend (5 files modified, 1 file created)
- `model/post.go`: `VisibilityFriends` -> `VisibilityFollower`
- `dto/post_dto.go`: validate 태그 `friends` -> `follower`
- `service/post_service.go`: `checkVisibilityAccess` 메서드 추가, `FollowRepository` 의존성
- `repository/post_repository.go`: 8개 피드 쿼리에 visibility WHERE 필터 추가
- `service/post_service_test.go`: 9개 visibility 접근 제어 테스트 케이스
- `migrations/012_rename_friends_to_follower.up.sql`: DB 마이그레이션

### Frontend (9 files modified/created)
- **New**: `ComposePage.tsx`, `VisibilitySelector.tsx`, `VisibilityBadge.tsx`
- **Modified**: `App.tsx`, `HomePage.tsx`, `PostCard.tsx`, `PostDetailPage.tsx`, `Sidebar.tsx`, `MobileNav.tsx`, `api.ts`

## Test Results

| Suite | Cases | Pass | Fail |
|-------|-------|------|------|
| Go Visibility Access | 9 | 9 | 0 |
| Go Full Suite | All | All | 0 |
| TypeScript Check | - | Pass | - |

## Code Review Summary

| Level | Count | Status |
|-------|-------|--------|
| Critical | 0 applicable | - |
| Warning | 2 | Fixed (MobileNav dead ternary, PostDetailPage ReplyForm removed) |
| Info | 3 | Acknowledged |

## Security Verification
- SQL injection: PASS (parameterized queries)
- Unauthenticated visibility: PASS (public only)
- Unauthorized access: PASS (404 not 403)
