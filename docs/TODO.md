# X Clone — 작업 체크리스트

## Phase 1: 인증 시스템 (진행중)
- [x] JWT 토큰 발급/검증 로직
- [x] 회원가입 API (handler → service → repository)
- [ ] 로그인 API + 리프레시 토큰
- [ ] 프론트엔드 로그인/회원가입 페이지

## Phase 2: 피드 시스템 (대기)
- [ ] 게시물 CRUD API
- [ ] Cursor Pagination 구현
- [ ] 피드 타임라인 UI

## Phase 3: 인터랙션 (진행중)
- [ ] 좋아요/리포스트 API
- [x] 답글(Reply) 시스템 (#6)
- [x] Post Detail API nested replies 최적화 (#28)
- [ ] ReBAC 검증 로직 (차단 관계 필터링)

## Phase 4: 실시간 알림 (대기)
- [ ] WebSocket 연결 관리
- [ ] 알림 이벤트 처리
- [ ] 프론트엔드 알림 UI

## 차단된 항목
- (없음)

## Phase 5: UI 컴포넌트 시스템 (완료)
- [x] shadcn/ui 컴포넌트 설치 (Input, Textarea, Avatar, Card, Dialog, Sonner, Form, Label)
- [x] Avatar size variant 확장 + UserAvatar 래퍼 컴포넌트
- [x] Button follow variant 추가
- [x] Toast(Sonner) 설정
- [x] 기존 컴포넌트 마이그레이션 (Avatar 6곳, Dialog 2곳, Input/Textarea, Button)
- [x] Toast 알림 연동 (ComposeForm, ReplyForm, EditProfileModal)
- [x] ComponentShowcasePage (/dev/components)
- [x] orphan .module.css 파일 정리 (#18)

## Phase 6: UI 레이아웃 고도화 (완료)
- [x] MainLayout 3단 레이아웃 (사이드바 + 피드 + 위젯 패널)
- [x] Sidebar 컴포넌트 (반응형: 데스크톱/태블릿/모바일)
- [x] MobileNav 하단 네비게이션 바
- [x] RightPanel 검색 + 트렌드 위젯
- [x] PostCard UI 고도화 (상대시간, Repost/Share 버튼, X 스타일 레이아웃)
- [x] ComposeForm 고도화 (원형 글자수 프로그레스, 아바타)
- [x] ProfilePage 고도화 (탭 네비게이션, ArrowLeft 아이콘, CalendarDays)
- [x] PostDetailPage 고도화 (액션 버튼 4종, 통계 표시)
- [x] LoginPage/RegisterPage shadcn/ui 마이그레이션
- [x] formatRelativeTime 유틸리티 함수

## Phase 7: Profile 탭 콘텐츠 (완료)
- [x] PostRepository handle 기반 조회 메서드 6개 추가 (#31)
- [x] PostService 3개 조회 메서드 (ListPostsByHandle, ListRepliesByHandle, ListLikedPostsByHandle)
- [x] UserHandler PostService 의존성 + 핸들러 3개 (GetUserPosts, GetUserReplies, GetUserLikes)
- [x] main.go 라우트 3개 등록 (OptionalAuth)
- [x] useUserPosts.ts hook (useUserPosts, useUserReplies, useUserLikes)
- [x] ProfilePage.tsx 탭 콘텐츠 실제 데이터 연동 (PostCard 렌더링, 로딩/빈 상태)

## Phase 8: 북마크 시스템 (진행중)
- [x] bookmarks 테이블 마이그레이션 (#9)
- [x] BookmarkRepository (Bookmark, Unbookmark, IsBookmarked, ListByUserID)
- [x] BookmarkService (Bookmark, Unbookmark, ListBookmarks - cursor 기반 페이지네이션)
- [x] BookmarkHandler (POST/DELETE /posts/:id/bookmark, GET /users/bookmarks)
- [x] 라우팅 등록 (main.go, /bookmarks를 /:handle보다 먼저 등록)
- [x] Frontend useBookmark hook (optimistic UI toggle)
- [x] PostCard 북마크 버튼 추가
- [x] ProfilePage 북마크 탭 (본인만 표시)
- [ ] 테스트 작성
- [ ] 코드 리뷰
- [ ] PR 생성 및 이슈 연결

## Phase 9: 백엔드 인프라 개선 (완료)
- [x] Request Validator (go-playground/validator/v10) - DTO 태그 + 공통 헬퍼
- [x] Structured Logging (log/slog) - 로거 팩토리 + 요청 로깅 미들웨어
- [x] Dependency Injection (uber-go/fx) - 모듈 분리 + 라우터 분리 + main.go 리팩토링
- [x] 테스트 작성 (validator 20 + logger 3 = 23 cases)
- [x] 코드 리뷰 (버그 1건 수정, 개선 2건 적용)
- [ ] PR 생성 및 이슈 연결

## Phase 10: 글쓰기 커스터마이징 (진행중)
- [x] 스펙 문서 작성 (docs/specs/compose-customization-spec.md)
- [x] Phase A: 마크다운 지원
  - [x] DB 마이그레이션 (content VARCHAR→TEXT)
  - [x] 글자 수 제한 280→500 (백엔드 + 프론트엔드)
  - [x] react-markdown + remark-gfm + rehype-sanitize 설치
  - [x] MarkdownRenderer 컴포넌트
  - [x] ComposeForm 라이브 프리뷰
  - [x] PostCard/PostDetailPage/ReplyCard/ParentPostCard 마크다운 렌더링
  - [x] highlight.js 코드 하이라이팅
- [x] Phase B: 미디어 업로드 백엔드
  - [x] post_media 테이블 마이그레이션
  - [x] MediaStorage 인터페이스 + LocalStorage 구현
  - [x] MediaRepository, MediaService, MediaHandler
  - [x] 정적 파일 서빙 설정
- [x] Phase C: 미디어 업로드 프론트엔드
  - [x] useMediaUpload 훅
  - [x] MediaPreview, MediaGrid 컴포넌트
  - [x] ComposeForm 미디어 통합
  - [x] PostCard/PostDetailPage 미디어 표시
- [x] Phase D: GPS 위치 태그
  - [x] posts 테이블 위치 컬럼 마이그레이션
  - [x] 백엔드 위치 필드 (model, dto, repository)
  - [x] useGeolocation 훅 (Nominatim 역지오코딩)
  - [x] ComposeForm 위치 태그 UI
  - [x] PostCard/PostDetailPage 위치 표시
- [x] Phase E: 투표 백엔드
  - [x] polls/poll_options/poll_votes 테이블 마이그레이션
  - [x] PollRepository, PollService, PollHandler
  - [x] PostService에 투표 생성 통합
- [x] Phase F: 투표 프론트엔드
  - [x] usePoll 훅
  - [x] PollCreator, PollDisplay 컴포넌트
  - [x] ComposeForm 투표 통합
  - [x] PostCard/PostDetailPage 투표 표시
- [x] 테스트 작성 (Go service 테스트 통과)
- [x] 코드 리뷰 (Critical 4건 수정, Warning 수정)
- [ ] PR 생성
- [x] Phase G: 공개 범위(Visibility) 설정
  - [x] friends → follower 리네이밍 (model, dto, DB 마이그레이션)
  - [x] 백엔드 Repository visibility 필터링 (피드 쿼리 8개+ 수정)
  - [x] Service 레이어 checkVisibilityAccess (ReBAC 접근 제어)
  - [x] PostService에 FollowRepository 의존성 추가
  - [x] VisibilitySelector 컴포넌트 (Globe/Users/Lock 드롭다운)
  - [x] VisibilityBadge 컴포넌트 (PostCard/PostDetailPage)
- [x] Phase H: 글쓰기 페이지 분리
  - [x] ComposePage (/compose) 신규 생성
  - [x] HomePage 인라인 ComposeForm 제거
  - [x] Sidebar 글쓰기 버튼 → /compose 이동
  - [x] MobileNav 홈-글쓰기(하이라이트)-프로필 순서 변경
  - [x] 답글 버튼 → /compose?replyTo={id} (부모글 컨텍스트 표시)
  - [x] PostDetailPage ReplyForm 제거 (compose 페이지로 통합)
- [x] Visibility 테스트 (9개 table-driven 케이스 통과)
- [x] 코드 리뷰 (보안 체크 통과, MobileNav/PostDetailPage 수정)

## Phase 11: 조회수(View Count) 시스템 (완료)
- [x] 스펙 문서 작성 (docs/specs/view-count-spec.md)
- [x] DB 마이그레이션 (posts.view_count 컬럼)
- [x] 백엔드 Model/DTO/Repository/Service 구현
- [x] IncrementViewCount (비본인 상세 조회 시만)
- [x] 프론트엔드 PostCard/PostDetailPage/ReplyCard Eye 아이콘
- [x] formatCompactNumber 유틸 (1.2K, 3.4M)
- [x] 테스트 작성 (4 서브케이스 통과)
- [x] 코드 리뷰 (Critical 2건 수정)

## Phase 12: Post/Reply 수정 및 삭제 (진행중)
- [x] 스펙 문서 작성 (docs/specs/post-edit-delete-spec.md)
- [x] DB 마이그레이션 (posts.deleted_at 컬럼)
- [x] apperror.Forbidden (403) 추가
- [x] 백엔드 Model (DeletedAt 필드)
- [x] 백엔드 DTO (UpdatePostRequest, DeletePostResponse)
- [x] Repository: Update, SoftDelete, SoftDeleteReply + deleted_at IS NULL 필터 (14개 쿼리)
- [x] Repository: PollRepository.DeleteByPostID, MediaRepository.UnlinkByPostID
- [x] Service: UpdatePost (content/visibility/media/location/poll 수정)
- [x] Service: DeletePost (soft delete, Reply는 본인+Post 작성자 삭제 가능)
- [x] Handler: UpdatePost (PUT /api/posts/:id), DeletePost (DELETE /api/posts/:id)
- [x] Router: PUT/DELETE /:id 라우트 등록
- [x] 프론트엔드 UpdatePostRequest 타입
- [x] useUpdatePost, useDeletePost 훅
- [x] PostCard: 드롭다운 메뉴 + 인라인 수정 + 삭제 확인 + (edited) 표시
- [x] ReplyCard: 동일
- [x] PostDetailPage: 동일 + 삭제 후 홈 이동
- [x] shadcn/ui DropdownMenu, AlertDialog 설치
- [x] 기존 테스트 mock 업데이트 + 전체 테스트 통과
- [x] 코드 리뷰
- [ ] PR 생성

## Phase 13: Seed Data & Schema 최신화 (완료)
- [x] 스펙 문서 작성 (docs/specs/seed-data-spec.md)
- [x] backend/docs/schema.md 최신화 (실제 DB 스키마 반영)
- [x] 015_seed_data.up.sql (5명 사용자, 15 게시물, 11 팔로우, 13 좋아요, 3 북마크)
- [x] 015_seed_data.down.sql (롤백)
- [x] docker compose down -v && up 검증 완료

## 최근 변경 로그
- 2026-03-06: Seed Data 마이그레이션 + schema.md 최신화
- 2026-03-06: Post/Reply 수정 및 삭제 구현 - soft delete, 인라인 편집, 권한 검증
- 2026-03-06: 조회수(View Count) 시스템 구현 - 모든 post/reply에 조회수 표시, 큰 숫자 포맷팅
- 2026-03-06: 공개 범위 설정 + 글쓰기 페이지 분리 + 답글 컨텍스트 구현
- 2026-03-06: 글쓰기 커스터마이징 - 마크다운/미디어/위치/투표 구현
- 2026-03-06: 백엔드 인프라 개선 - Validator, Logging(slog), DI(fx) 도입
- 2026-03-06: 이슈 #31 Profile 페이지 탭 콘텐츠 구현 (게시물/답글/좋아요)
- 2026-03-06: 이슈 #19 주요 UI 레이아웃 및 스타일링 고도화 (3단 레이아웃, 반응형)
- 2026-03-06: 이슈 #18 shadcn/ui 기반 공통 UI 컴포넌트 시스템 구축
- 2026-03-05: 이슈 #28 Post Detail API nested replies 최적화 (110+→1 요청)
- 2026-03-05: 이슈 #27 대댓글 depth 2 자동 펼침 + 부모 스레드 체인 표시 구현
- 2026-03-04: 이슈 #26 대댓글(중첩 답글) + 답글 좋아요 Optimistic UI 구현
- 2026-03-04: 이슈 #6 답글(Reply) 시스템 구현 완료
- 2026-03-04: Phase 1 인증 시스템 작업 시작
