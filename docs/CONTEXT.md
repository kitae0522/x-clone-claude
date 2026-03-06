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
