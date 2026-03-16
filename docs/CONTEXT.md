# X Clone — 의사결정 기록

## 2026-03-04 — Go Fiber 선택
**상황**: Gin vs Fiber vs Echo 중 선택 필요
**결정**: Fiber 채택
**이유**:
- Express.js와 유사한 API로 학습 곡선이 낮음
- fasthttp 기반으로 높은 성능
- WebSocket 네이티브 지원
**트레이드오프**: Gin 대비 커뮤니티 규모가 작음

## 2026-03-04 — Cursor Pagination 채택
**상황**: Offset vs Cursor Pagination
**결정**: Cursor Pagination
**이유**:
- 피드처럼 실시간 데이터가 추가되는 환경에서 offset은 중복/누락 발생
- 대규모 데이터에서 성능 우위 (O(1) vs O(n))
**트레이드오프**: 특정 페이지로 직접 이동 불가

## 2026-03-05 — Threads 스타일 author thread continuation 채택
**상황**: Post Detail 페이지 진입 시 depth별 reply를 개별 요청하여 110+건 API 호출 발생
**결정**: `GET /api/posts/:id` 응답에 depth 1 전체 replies + 작성자 자기 답글 chain(author thread) 포함
**방식**: 각 reply에 대해 동일 작성자가 이어 쓴 답글만 재귀적으로 chain (Meta Threads 방식)
**제한**: `maxAuthorThreadDepth = 10` (무한 chain 방지)
**트레이드오프**: engagement 기반 정렬 포기 → 작성자의 대화 맥락 보존 우선. reply 클릭으로 detail 페이지 이동하여 다른 사람의 nested reply 확인

## 2026-03-04 — 답글(Reply) 자기참조 구조 채택
**상황**: 답글을 별도 테이블로 분리할지, posts 테이블에 parent_id 자기참조로 처리할지
**결정**: posts 테이블에 `parent_id` (nullable, self-referencing FK) 추가
**이유**:
- X(트위터)의 답글은 사실상 부모를 가진 포스트이므로 동일 테이블이 자연스러움
- 답글에도 좋아요/답글 달기 등 동일 인터랙션 적용 가능
- 별도 테이블 대비 JOIN 복잡도 감소
**트레이드오프**: N-depth 지원 가능하지만, UI 복잡도 관리를 위해 1-depth만 렌더링. 피드 쿼리에 `WHERE parent_id IS NULL` 필터 필요

## 2026-03-06 — 조회수(View Count) 저장 방식
**상황**: 조회수를 별도 테이블(post_views)로 관리할지, posts 테이블에 직접 저장할지
**결정**: posts.view_count 컬럼 직접 저장 (like_count/reply_count 패턴과 동일)
**이유**:
- 기존 카운트 필드와 일관성 유지
- 별도 테이블은 집계 쿼리 비용 증가, MVP 수준에서 과도한 설계
- 비본인 + 상세 조회 시에만 증가 (피드 스크롤 시 미증가)
**트레이드오프**: 중복 조회도 카운트 (X도 동일). 고트래픽 시 Redis 배치 처리로 확장 가능

## 2026-03-06 — Profile 탭 조회: handle 기반 Repository 메서드
**상황**: Profile 페이지에서 사용자 게시물/답글/좋아요를 조회할 때 handle(username)로 API를 호출
**결정**: PostRepository에 handle 기반 조회 메서드 추가 (users JOIN으로 username → author_id 해석)
**이유**:
- 프론트엔드에서 handle만 알고 있으므로 별도 user ID 조회 단계 없이 한 번의 쿼리로 해결
- 좋아요 목록은 likes 테이블 + users(target) + posts + users(author) 4-way JOIN
**트레이드오프**: Repository 메서드 수 증가 (6개), 하지만 각각 단일 쿼리로 N+1 없음

## 2026-03-06 — Post/Reply Soft Delete 패턴
**상황**: 삭제된 글을 DB에서 완전히 제거(hard delete)할지, 논리적으로만 삭제(soft delete)할지
**결정**: `deleted_at TIMESTAMPTZ` 컬럼 기반 soft delete
**이유**:
- 데이터 보존 및 감사 추적 가능
- 향후 복구 기능 확장 용이
- 삭제된 글 통계 분석 가능
**구현**: 모든 SELECT 쿼리(14개+)에 `WHERE deleted_at IS NULL` 조건, 부분 인덱스로 성능 보완
**트레이드오프**: 쿼리 복잡도 증가, 주기적 데이터 정리(purge) 필요

## 2026-03-06 — Reply 삭제 권한: Post 작성자 확장
**상황**: Reply 삭제를 작성자 본인만 허용할지, 부모 Post 작성자에게도 허용할지
**결정**: Reply 작성자 + 부모 Post 작성자 모두 삭제 가능
**이유**: X/Twitter와 동일한 패턴, Post 작성자가 자기 글의 답글을 관리할 수 있어야 함
**구현**: Service 레이어에서 `parentPost.AuthorID == userID` 추가 검증 (ReBAC)

## 2026-03-06 — 부분 업데이트(Partial Update) 포인터 타입
**상황**: UpdatePost 시 변경된 필드만 업데이트하되, 명시적 제거(null)와 미변경을 구분해야 함
**결정**: `*string` 포인터 타입 + `ClearLocation`/`ClearPoll` boolean 플래그
**이유**: Go JSON unmarshaling에서 `null` → `nil` 포인터, 필드 부재 → 포인터 그대로 nil 구분 가능
**트레이드오프**: DTO 복잡도 증가, 하지만 정확한 의도 표현 가능

## 2026-03-07 — Profile Modal 이미지 업로드 방식 (#51)
**상황**: 프로필 수정 모달에서 이미지 URL 직접 입력 방식의 UX 불편, Dialog 닫기/저장 버튼 겹침
**결정**: 기존 media upload API 재활용하여 파일 선택 → 업로드 → URL 자동 설정
**이유**:
- 백엔드 변경 없이 프론트엔드만 수정으로 해결 가능
- 기존 `/api/media/upload` 엔드포인트와 완전 호환
- X/Twitter와 동일한 인라인 이미지 선택 UX
**트레이드오프**: 프로필 이미지도 post_media 테이블에 저장되어 용도 구분 없음 (현재 문제 아님)

## 2026-03-07 — 액션 드롭다운 통합 (#53)
**상황**: PostCard에서 본인→드롭다운, 타인→팔로우 버튼으로 시각적 불일치. 하단 액션 바에 6개 버튼 과다.
**결정**: 팔로우 버튼을 프로필 페이지에만 한정, 북마크/공유를 MoreHorizontal 드롭다운으로 편입
**이유**:
- 모든 카드에 동일한 MoreHorizontal 아이콘 → UI 통일감
- 하단 액션 바 간소화 (Reply/Repost/Like/ViewCount 4개로 집중)
- ReplyCard에도 북마크/공유 접근 가능
**트레이드오프**: 북마크/공유 접근성 1탭 → 2탭으로 증가, 하지만 핵심 인터랙션에 집중

## 2026-03-16 — Handle 기반 쿼리 soft-deleted 사용자 필터링 (#80)
**상황**: Phase 18에서 partial unique index로 username 재사용을 허용했으나, post_repository.go의 handle 기반 조회 쿼리가 soft-deleted 사용자도 매칭하여 이전 사용자 게시글이 신규 사용자 프로필에 노출됨
**결정**: 8곳의 handle 기반 쿼리에 `deleted_at IS NULL` 조건 추가
**이유**:
- WHERE 절의 `u.username = $1`에 `AND u.deleted_at IS NULL` 추가하면 기존 `idx_users_username_active` partial index와 정확히 일치하여 성능 영향 없음
- Service 레이어에서 별도 사용자 조회 없이 Repository 쿼리에서 직접 방어 (defense in depth)
**트레이드오프**: 없음 (쿼리 수정만, API 인터페이스/DB 스키마 변경 없음)
