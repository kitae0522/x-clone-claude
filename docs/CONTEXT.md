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
