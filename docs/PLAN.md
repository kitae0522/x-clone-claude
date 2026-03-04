# X Clone — 아키텍처 계획서

## 시스템 구조
```
[React 19 SPA] ←→ [Go Fiber API] ←→ [PostgreSQL]
      ↕                    ↕
[React Query]       [WebSocket Hub]
```

## 백엔드 레이어
1. **Handler**: HTTP 요청 파싱, 응답 반환
2. **Service**: 비즈니스 로직, ReBAC 검증
3. **Repository**: DB 쿼리 (인터페이스 기반)

## 인증 흐름
1. 로그인 → JWT access + refresh 토큰 발급
2. 요청마다 Authorization 헤더에 access 토큰
3. 만료 시 refresh 토큰으로 자동 갱신

## API 응답 규격
- 성공: `{ "data": ... }`
- 에러: `{ "error": { "code": "...", "message": "..." } }`
- 페이지네이션: `{ "data": [...], "next_cursor": "...", "has_more": true }`
