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

## 최근 변경 로그
- 2026-03-06: 이슈 #18 shadcn/ui 기반 공통 UI 컴포넌트 시스템 구축
- 2026-03-05: 이슈 #28 Post Detail API nested replies 최적화 (110+→1 요청)
- 2026-03-05: 이슈 #27 대댓글 depth 2 자동 펼침 + 부모 스레드 체인 표시 구현
- 2026-03-04: 이슈 #26 대댓글(중첩 답글) + 답글 좋아요 Optimistic UI 구현
- 2026-03-04: 이슈 #6 답글(Reply) 시스템 구현 완료
- 2026-03-04: Phase 1 인증 시스템 작업 시작
