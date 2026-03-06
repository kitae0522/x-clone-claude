# Backend Infrastructure Improvements Report

> Date: 2026-03-06
> Branch: feat/issue-9-bookmark
> Status: Implemented + Reviewed

## Summary

Backend에 3가지 인프라 개선을 도입했습니다:

| # | 항목 | 라이브러리 | 상태 |
|---|------|-----------|------|
| 1 | Request Validator | go-playground/validator/v10 | Done |
| 2 | Structured Logging | log/slog (표준 라이브러리) | Done |
| 3 | Dependency Injection | uber-go/fx | Done |
| 4 | ~~Query Builder~~ | ~~squirrel + sqlx~~ | Cancelled (raw SQL 유지) |

## Changes

### 1. Request Validator
- **New**: `pkg/validator/validator.go` - validate 래퍼 + 에러 메시지 포맷
- **New**: `internal/handler/parse.go` - parseAndValidate 공통 헬퍼
- **Modified**: DTO에 validate 태그 추가 (auth_dto, post_dto, user_dto)
- **Modified**: 핸들러에서 parseAndValidate 사용 (auth, post, user)

### 2. Structured Logging (slog)
- **New**: `pkg/logger/logger.go` - 환경별 slog 핸들러 팩토리
- **New**: `internal/middleware/request_logger.go` - 요청 로깅 미들웨어 (request_id, method, path, status, latency, user_id)
- **Modified**: `pkg/config/config.go` - Env 필드 추가, production에서 JWT_SECRET 필수
- **Modified**: `internal/handler/response.go` - 5xx 에러 시 slog.ErrorContext 로깅

### 3. Dependency Injection (fx)
- **New**: `internal/{repository,service,handler}/module.go` - fx.Module 정의
- **New**: `internal/router/router.go` - 라우팅 로직 분리 (fx.In 파라미터)
- **Modified**: `main.go` - fx 기반 리팩토링 (89줄 -> 85줄, 수동 와이어링 제거)
- **Modified**: `internal/service/auth_service.go` - NewAuthService가 *config.Config 수용

## Code Review Fixes
- **Bug fix**: parseAndValidate가 검증 실패 시 nil 반환하던 버그 수정 -> errResponseSent 반환
- **Improvement**: router.Params에 fx.In 적용하여 DI 보일러플레이트 제거
- **Security**: config.Load가 production에서 JWT_SECRET 미설정 시 에러 반환

## Test Results
- New tests: 23 cases (validator 20 + logger 3)
- Existing tests: All passing
- Total: 0 failures

## Decision Log
- Query Builder(squirrel + sqlx) 도입 취소 -> raw SQL 유지가 유지보수에 유리하다는 판단
- slog 선택 (zap 대신) -> Go 1.24 표준, context 네이티브 지원, 외부 의존성 0
- fx 선택 (wire 대신) -> 런타임 DI지만 라이프사이클 관리 내장, wire 유지보수 중단

## Added Dependencies
| Package | Version | Purpose |
|---------|---------|---------|
| github.com/go-playground/validator/v10 | v10.30.1 | Request validation |
| go.uber.org/fx | v1.24.0 | Dependency injection |
| go.uber.org/dig | v1.19.0 | fx dependency |
