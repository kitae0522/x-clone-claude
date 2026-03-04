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

## 2026-03-04 — 답글(Reply) 자기참조 구조 채택
**상황**: 답글을 별도 테이블로 분리할지, posts 테이블에 parent_id 자기참조로 처리할지
**결정**: posts 테이블에 `parent_id` (nullable, self-referencing FK) 추가
**이유**:
- X(트위터)의 답글은 사실상 부모를 가진 포스트이므로 동일 테이블이 자연스러움
- 답글에도 좋아요/답글 달기 등 동일 인터랙션 적용 가능
- 별도 테이블 대비 JOIN 복잡도 감소
**트레이드오프**: N-depth 지원 가능하지만, UI 복잡도 관리를 위해 1-depth만 렌더링. 피드 쿼리에 `WHERE parent_id IS NULL` 필터 필요
