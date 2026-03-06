# Pipeline Report: Compose Customization

**Date**: 2026-03-06
**Branch**: feat/backend-infra-improvements
**Status**: COMPLETE

---

## Feature Summary

글쓰기 환경 커스터마이징 4가지 기능을 구현했습니다.

| Feature | Status | Description |
|---------|--------|-------------|
| Markdown | DONE | react-markdown + rehype-sanitize, 라이브 프리뷰, 코드 하이라이팅 |
| Media Upload | DONE | 이미지 4장/동영상 1개/GIF 1개, 로컬 스토리지, 업로드 진행률 |
| GPS Location | DONE | Geolocation API + Nominatim 역지오코딩, 프라이버시 라운딩 |
| Polls | DONE | 2-4 옵션, 투표 기간, 중복 투표 방지, Optimistic UI |

---

## Step 1: Spec
- `docs/specs/compose-customization-spec.md` 작성 완료
- 4가지 기능의 API 설계, DB 스키마, 컴포넌트 구조, 보안 고려사항 포함

## Step 2: Implementation

### Backend (Go + Fiber)
**New Files (14):**
- 4 migrations (008-011): content TEXT, post_media, location columns, polls
- storage/storage.go, local_storage.go: MediaStorage 인터페이스
- model/media.go, poll.go: 데이터 모델
- dto/media_dto.go: MediaResponse DTO
- repository/media_repository.go, poll_repository.go
- service/media_service.go, poll_service.go
- handler/media_handler.go, poll_handler.go

**Modified Files (7):**
- model/post.go: 위치 필드 추가
- dto/post_dto.go: Location/Poll/Media 필드, 500자 제한
- repository/post_repository.go: 위치 컬럼 SELECT/INSERT
- service/post_service.go: poll/media/location 통합
- handler/module.go, service/module.go, repository/module.go: DI 등록
- router/router.go: /api/media/upload, /api/posts/:id/vote 라우트
- main.go: LocalStorage provider

### Frontend (React + TypeScript)
**New Files (9):**
- components/MarkdownRenderer.tsx
- components/MediaGrid.tsx, MediaPreview.tsx
- components/PollDisplay.tsx, PollCreator.tsx
- hooks/useMediaUpload.ts, useGeolocation.ts, usePoll.ts
- highlight.js CSS import

**Modified Files (7):**
- ComposeForm.tsx: 마크다운 프리뷰 + 미디어 + 위치 + 투표 통합
- PostCard.tsx, PostDetailPage.tsx: 마크다운/미디어/위치/투표 표시
- ParentPostCard.tsx, ReplyCard.tsx, ReplyForm.tsx: 마크다운 렌더링, 500자
- types/api.ts: MediaItem, LocationData, PollData 타입
- hooks/usePosts.ts: createPost에 mediaIds/location/poll 지원

## Step 3: Tests
- Go 테스트: service 패키지 전체 PASS (post_service + media_service 등)
- TypeScript: tsc -b 에러 없음

## Step 4: Code Review
- `docs/reports/compose-customization-review.md` 작성 완료
- Critical 4건 모두 수정 완료:
  - BE-01: DB 테이블명 불일치 (media → post_media)
  - BE-02: 미디어-게시물 연결 누락 (mediaRepo 주입 + LinkToPost 호출)
  - BE-03: XHR 업로드 인증 누락 (withCredentials 추가)
  - BE-04: 경로 순회 취약점 (path traversal 방지 추가)
- Warning FE-01 수정: toast 무한 루프 → useEffect로 이동

---

## Build Verification
- `go build ./...` PASS
- `go test ./...` PASS (all packages)
- `npx tsc -b` PASS (zero errors)

---

## Architecture Decision Records
- 글자수 제한: 280 → 500 (마크다운 기호 포함 계산)
- DB content: VARCHAR(280) → TEXT
- 마크다운 렌더링: 프론트엔드에서 수행, 서버는 원본 저장
- 미디어 스토리지: 로컬 파일시스템 (MediaStorage 인터페이스로 S3 전환 가능)
- 역지오코딩: Nominatim (OpenStreetMap) 무료 API, 프론트엔드에서 호출
- 투표 + 미디어: 상호 배타 (동시 사용 불가)
