# Action Dropdown Unification Report

**Date**: 2026-03-07
**Issue**: #53
**Branch**: `feat/issue-53-action-dropdown-unification`

## Summary

PostCard, ReplyCard, PostDetailPage 3개 컴포넌트의 액션 버튼을 드롭다운 메뉴로 통합하여 UI 일관성을 확보했다.

## Changes

| File | Lines Changed | Description |
|------|--------------|-------------|
| `PostCard.tsx` | -70 / +30 | 팔로우 버튼 제거, 드롭다운 통합 (북마크/공유, 본인→수정/삭제), 하단 Bookmark/Share 제거 |
| `ReplyCard.tsx` | +61 / -2 | 모든 로그인 사용자 드롭다운 추가, useBookmark/ShareModal 도입 |
| `PostDetailPage.tsx` | -67 / +50 | 팔로우 버튼 제거, 드롭다운 통합, 하단 Bookmark/Share 제거 |
| `docs/TODO.md` | +12 | Phase 15 추가 |

## Key Decisions

1. **팔로우 버튼**: PostCard + PostDetailPage에서 완전 제거. 프로필 페이지에서만 팔로우 가능.
2. **드롭다운 구조**: 본인 글 → 수정/삭제 + 구분선 + 북마크/공유, 타인 글 → 북마크/공유만
3. **ShareModal + DropdownMenu 포커스 충돌**: `onSelect(e.preventDefault())` + `setTimeout`으로 해결
4. **비로그인 사용자**: 드롭다운 미표시

## Quality Check

- `bun run check` (tsc + eslint): PASS
- 백엔드 변경: 없음
