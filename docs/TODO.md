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
- [ ] ReBAC 검증 로직 (차단 관계 필터링)

## Phase 4: 실시간 알림 (대기)
- [ ] WebSocket 연결 관리
- [ ] 알림 이벤트 처리
- [ ] 프론트엔드 알림 UI

## 차단된 항목
- (없음)

## 최근 변경 로그
- 2026-03-04: 이슈 #6 답글(Reply) 시스템 구현 완료
- 2026-03-04: Phase 1 인증 시스템 작업 시작
