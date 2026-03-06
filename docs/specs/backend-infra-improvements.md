# 백엔드 인프라 개선 스펙

> 작성일: 2026-03-06
> 상태: In Progress (Query Builder 제외 — raw SQL 유지 결정)
> 관련 문서: `docs/PLAN.md`, `docs/CONTEXT.md`

---

## 목차
1. [개요](#1-개요)
2. [Structured Logging](#2-structured-logging)
3. [Request Validator](#3-request-validator)
4. [Dependency Injection](#4-dependency-injection)
5. [DB Query Builder](#5-db-query-builder)
6. [도입 순서 및 의존 관계](#6-도입-순서-및-의존-관계)
7. [열린 질문](#7-열린-질문)

---

## 1. 개요

### 현재 상태 (As-Is)
| 영역 | 현재 방식 | 문제점 |
|------|----------|--------|
| 로깅 | `log.Fatalf`, `log.Fatal` (표준 라이브러리) | 로그 레벨 없음, JSON 포맷 불가, 요청별 컨텍스트(request ID, user ID) 추적 불가 |
| 요청 검증 | `c.BodyParser(&req)` 후 수동 검증 없음 | 빈 content, 잘못된 visibility 값 등이 서비스 레이어까지 도달 |
| DI | `main.go`에서 수동 와이어링 (약 20줄) | 의존성 추가 시 main.go 비대화, 테스트 시 수동 목 구성 필요 |
| DB 쿼리 | raw SQL 문자열 직접 작성 | 유사 쿼리 복붙 (PostRepository에 16개 메서드, SELECT 컬럼 목록 반복), 조건부 WHERE/JOIN 구성 어려움 |

### 설계 원칙
- 기존 handler -> service -> repository 레이어 구조를 유지한다
- 한 번에 전부 바꾸지 않고 점진적으로 도입한다
- Go 생태계에서 사실상 표준(de facto standard)인 라이브러리를 선택한다
- 과도한 추상화를 지양한다

---

## 2. Structured Logging

### 기능 설명 (What)
표준 `log` 패키지를 구조화된 로깅 라이브러리로 교체한다. 모든 로그에 레벨(debug/info/warn/error), 타임스탬프, 요청 컨텍스트(request ID, user ID, method, path)를 포함한다.

### 구현 이유 (Why)
- 프로덕션 환경에서 로그를 JSON으로 수집/검색하려면 구조화된 로그가 필수
- 현재 `log.Fatalf`만 사용하여 에러가 발생하면 프로세스가 즉시 종료됨
- 요청 추적(tracing)을 위한 request ID 전파가 불가능

### 라이브러리 선택: `log/slog` (Go 1.21+ 표준 라이브러리)

**선택 이유:**
- Go 1.21부터 표준 라이브러리에 포함 (외부 의존성 불필요)
- 프로젝트가 Go 1.24를 사용하므로 바로 사용 가능
- `slog.Logger`는 `context.Context` 기반으로 값 전파 지원
- JSON/Text 핸들러 내장, 커스텀 핸들러 확장 가능
- zerolog/zap 대비 성능은 약간 낮지만 표준이라는 이점이 압도적

**대안 비교:**
| 라이브러리 | 장점 | 단점 | 결론 |
|-----------|------|------|------|
| `log/slog` | 표준, 의존성 0, context 네이티브 | 성능이 zap보다 약간 낮음 | **채택** |
| `uber-go/zap` | 최고 성능 | 외부 의존성, 설정 복잡 | 현 규모에서 불필요 |
| `rs/zerolog` | 빠름, zero-alloc | 외부 의존성, API가 빌더 패턴 | 표준 slog로 충분 |

### 수락 기준 (Acceptance Criteria)
- [ ] `log.Fatalf` / `log.Fatal` 호출이 프로젝트에서 0건이다
- [ ] 모든 HTTP 요청에 대해 method, path, status, latency가 JSON 형식으로 기록된다
- [ ] 각 요청에 고유한 `request_id`가 부여되고 로그에 포함된다
- [ ] 인증된 요청에는 `user_id`가 로그에 포함된다
- [ ] 에러 발생 시 `slog.Error` 레벨로 에러 메시지와 스택 정보가 기록된다
- [ ] 개발 환경에서는 Text 핸들러, 프로덕션에서는 JSON 핸들러를 사용한다

### 구현 방안

#### 2-1. 로거 초기화 (`pkg/logger/logger.go` 신규)
```go
package logger

import (
    "log/slog"
    "os"
)

func New(env string) *slog.Logger {
    var handler slog.Handler
    opts := &slog.HandlerOptions{Level: slog.LevelDebug}

    if env == "production" {
        handler = slog.NewJSONHandler(os.Stdout, opts)
    } else {
        handler = slog.NewTextHandler(os.Stdout, opts)
    }

    return slog.New(handler)
}
```

#### 2-2. 요청 로깅 미들웨어 (`internal/middleware/request_logger.go` 신규)
```go
func RequestLogger(logger *slog.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        requestID := uuid.New().String()
        c.Locals("requestID", requestID)

        start := time.Now()
        err := c.Next()
        latency := time.Since(start)

        attrs := []slog.Attr{
            slog.String("request_id", requestID),
            slog.String("method", c.Method()),
            slog.String("path", c.Path()),
            slog.Int("status", c.Response().StatusCode()),
            slog.Duration("latency", latency),
            slog.String("ip", c.IP()),
        }

        if userID, ok := c.Locals("userID").(string); ok {
            attrs = append(attrs, slog.String("user_id", userID))
        }

        level := slog.LevelInfo
        if c.Response().StatusCode() >= 500 {
            level = slog.LevelError
        } else if c.Response().StatusCode() >= 400 {
            level = slog.LevelWarn
        }

        logger.LogAttrs(c.Context(), level, "HTTP request", attrs...)
        return err
    }
}
```

#### 2-3. 서비스/리포지토리 레이어에서의 사용
`context.Context`를 통해 logger를 전파하거나, 구조체 필드로 `*slog.Logger`를 주입한다.
DI 도입 후에는 DI 컨테이너를 통해 자동 주입한다.

### 변경 파일 목록
| 파일 | 변경 유형 |
|------|----------|
| `pkg/logger/logger.go` | **신규** - 로거 팩토리 |
| `internal/middleware/request_logger.go` | **신규** - 요청 로깅 미들웨어 |
| `pkg/config/config.go` | **수정** - `Env` 필드 추가 (development/production) |
| `main.go` | **수정** - `log.Fatalf` 제거, `slog.Logger` 초기화 및 미들웨어 등록 |
| `internal/handler/response.go` | **수정** - 5xx 에러 시 로깅 추가 |

### 엣지 케이스
- `log.Fatal` 호출 위치(DB 연결 실패, 마이그레이션 실패)는 앱 시작 전이므로 `slog`로 교체 후에도 `os.Exit(1)`로 종료해야 함
- 미들웨어에서 패닉 발생 시 로깅 누락 가능 -> Fiber의 Recover 미들웨어와 순서 조율 필요

---

## 3. Request Validator

### 기능 설명 (What)
DTO 구조체에 검증 태그를 추가하고, 핸들러에서 `BodyParser` 직후 자동으로 구조체 필드를 검증한다.

### 구현 이유 (Why)
- 현재 `CreatePostRequest`의 `Content` 필드가 빈 문자열이어도 서비스까지 도달함
- `RegisterRequest`의 email 형식, password 길이 등을 수동으로 검증하지 않음
- 검증 로직이 핸들러마다 중복되거나 누락될 위험이 있음

### 라이브러리 선택: `go-playground/validator/v10`

**선택 이유:**
- Go 생태계에서 가장 널리 사용되는 검증 라이브러리 (GitHub stars 17k+)
- 구조체 태그 기반으로 선언적 검증
- 커스텀 검증 함수 등록 가능
- Fiber 공식 문서에서도 권장

**대안 비교:**
| 라이브러리 | 장점 | 단점 | 결론 |
|-----------|------|------|------|
| `go-playground/validator` | 사실상 표준, 풍부한 태그 | 에러 메시지 커스터마이징에 약간의 보일러플레이트 | **채택** |
| `go-ozzo/ozzo-validation` | 코드 기반 검증, 유연함 | 태그 기반 대비 직관성 떨어짐 | 패스 |
| 수동 검증 | 의존성 없음 | 누락/중복 위험, 코드 증가 | 현재 상태 |

### 수락 기준 (Acceptance Criteria)
- [ ] 모든 요청 DTO(`*Request` 구조체)에 `validate` 태그가 적용되어 있다
- [ ] 검증 실패 시 400 응답에 실패한 필드명과 사유가 포함된다
- [ ] 빈 content로 게시물 생성 시 400 에러가 반환된다
- [ ] email 형식이 아닌 값으로 회원가입 시 400 에러가 반환된다
- [ ] 검증 로직이 핸들러 공통 헬퍼에 1곳에만 존재한다

### 구현 방안

#### 3-1. 검증 헬퍼 (`pkg/validator/validator.go` 신규)
```go
package validator

import (
    "fmt"
    "strings"

    "github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func Validate(s interface{}) []ValidationError {
    var errors []ValidationError
    err := validate.Struct(s)
    if err != nil {
        for _, e := range err.(validator.ValidationErrors) {
            errors = append(errors, ValidationError{
                Field:   toSnakeCase(e.Field()),
                Message: msgForTag(e),
            })
        }
    }
    return errors
}

func msgForTag(fe validator.FieldError) string {
    switch fe.Tag() {
    case "required":
        return "this field is required"
    case "email":
        return "must be a valid email address"
    case "min":
        return fmt.Sprintf("must be at least %s characters", fe.Param())
    case "max":
        return fmt.Sprintf("must be at most %s characters", fe.Param())
    case "oneof":
        return fmt.Sprintf("must be one of: %s", fe.Param())
    default:
        return fmt.Sprintf("failed on '%s' validation", fe.Tag())
    }
}
```

#### 3-2. DTO 태그 적용 예시
```go
// auth_dto.go
type RegisterRequest struct {
    Email    string `json:"email"    validate:"required,email"`
    Username string `json:"username" validate:"required,min=3,max=30,alphanum"`
    Password string `json:"password" validate:"required,min=8,max=128"`
}

type LoginRequest struct {
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

// post_dto.go
type CreatePostRequest struct {
    Content    string `json:"content"    validate:"required,min=1,max=280"`
    Visibility string `json:"visibility" validate:"omitempty,oneof=public followers_only"`
}

type CreateReplyRequest struct {
    Content string `json:"content" validate:"required,min=1,max=280"`
}

// user_dto.go
type UpdateProfileRequest struct {
    DisplayName     *string `json:"displayName"     validate:"omitempty,max=50"`
    Bio             *string `json:"bio"             validate:"omitempty,max=160"`
    ProfileImageURL *string `json:"profileImageUrl" validate:"omitempty,url"`
    HeaderImageURL  *string `json:"headerImageUrl"  validate:"omitempty,url"`
}
```

#### 3-3. 핸들러 공통 파서 (`internal/handler/parse.go` 신규)
```go
func parseAndValidate(c *fiber.Ctx, out interface{}) error {
    if err := c.BodyParser(out); err != nil {
        return respondError(c, apperror.BadRequest("invalid request body"))
    }
    if errors := validator.Validate(out); len(errors) > 0 {
        return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse{
            Success: false,
            Error:   &"validation failed",
            Data:    errors, // 필드별 에러 상세
        })
    }
    return nil
}
```

### 변경 파일 목록
| 파일 | 변경 유형 |
|------|----------|
| `pkg/validator/validator.go` | **신규** - 검증 래퍼 |
| `internal/handler/parse.go` | **신규** - 파싱+검증 공통 헬퍼 |
| `internal/dto/auth_dto.go` | **수정** - validate 태그 추가 |
| `internal/dto/post_dto.go` | **수정** - validate 태그 추가 |
| `internal/dto/user_dto.go` | **수정** - validate 태그 추가 (UpdateProfileRequest) |
| `internal/handler/auth_handler.go` | **수정** - `parseAndValidate` 사용 |
| `internal/handler/post_handler.go` | **수정** - `parseAndValidate` 사용 |
| `internal/handler/user_handler.go` | **수정** - `parseAndValidate` 사용 |
| `go.mod` | **수정** - `github.com/go-playground/validator/v10` 추가 |

### 엣지 케이스
- `Visibility`가 빈 문자열일 때: `omitempty`로 처리하되, 서비스에서 기본값 `public` 적용
- `Content`가 공백만 있을 때: `min=1`은 공백을 허용하므로 커스텀 태그 `notblank`이 필요할 수 있음
- 중첩 구조체 검증: 현재 DTO에는 없지만 추후 필요 시 `dive` 태그 사용
- URL 필드: `url` 태그는 scheme을 요구하므로 상대경로를 쓰면 실패

---

## 4. Dependency Injection

### 기능 설명 (What)
`main.go`의 수동 의존성 와이어링을 DI 컨테이너로 교체한다. Repository -> Service -> Handler의 생성 순서와 의존 관계를 선언적으로 관리한다.

### 구현 이유 (Why)
- 현재 `main.go`에서 6개 Repository, 6개 Service, 6개 Handler를 수동 생성 (약 25줄)
- 기능 추가 시마다 main.go에 와이어링 코드가 계속 증가
- 의존성 순서를 개발자가 직접 관리해야 하므로 실수 가능
- 테스트에서 목(mock) 교체가 수동적

### 라이브러리 선택: `uber-go/fx`

**선택 이유:**
- Uber에서 개발한 프로덕션 검증된 DI 프레임워크
- 런타임 DI로 리플렉션 기반이지만, 앱 시작 시 한 번만 실행
- 의존성 그래프를 자동으로 해석하고 순서를 결정
- 라이프사이클 훅(OnStart/OnStop)으로 graceful shutdown 지원
- 로깅, 에러 핸들링 등 인프라 관심사와 잘 어울림

**대안 비교:**
| 라이브러리 | 장점 | 단점 | 결론 |
|-----------|------|------|------|
| `uber-go/fx` | 프로덕션 검증, 라이프사이클 관리, 자동 그래프 해석 | 런타임 DI, 약간의 학습 곡선 | **채택** |
| `google/wire` | 컴파일 타임 DI, 타입 안전 | 코드 생성 필요, 유지보수 중단 상태 | 유지보수 중단 리스크 |
| `samber/do` | 경량, 제네릭 기반 | 상대적으로 작은 커뮤니티 | 대안 후보 |
| 수동 와이어링 유지 | 의존성 없음, 명시적 | 확장 시 main.go 비대화 | 현재 상태 |

### 수락 기준 (Acceptance Criteria)
- [ ] `main.go`의 수동 와이어링 코드가 `fx.Provide` 호출로 대체되었다
- [ ] `main.go`가 30줄 이하로 유지된다
- [ ] Repository, Service, Handler 각각의 생성자가 `fx.Provide`로 등록되어 있다
- [ ] 앱 시작 시 의존성 그래프가 올바르게 해석된다 (순환 의존 없음)
- [ ] `pool.Close()`가 `fx.Lifecycle`의 `OnStop` 훅으로 관리된다
- [ ] 기존 인터페이스 기반 구조가 그대로 유지된다

### 구현 방안

#### 4-1. 모듈 구조
```
backend/
  internal/
    app/
      app.go          # fx.App 구성 (최상위)
      modules.go      # 모듈 그룹 정의
    handler/
      module.go       # handler 모듈 (fx.Provide 등록)
    service/
      module.go       # service 모듈
    repository/
      module.go       # repository 모듈
```

#### 4-2. Repository 모듈 예시 (`internal/repository/module.go` 신규)
```go
package repository

import "go.uber.org/fx"

var Module = fx.Module("repository",
    fx.Provide(
        NewPostRepository,
        NewUserRepository,
        NewLikeRepository,
        NewFollowRepository,
        NewBookmarkRepository,
    ),
)
```

#### 4-3. Service 모듈 (`internal/service/module.go` 신규)
```go
package service

import "go.uber.org/fx"

var Module = fx.Module("service",
    fx.Provide(
        NewPostService,
        NewAuthService,
        NewLikeService,
        NewFollowService,
        NewBookmarkService,
        NewUserService,
    ),
)
```

#### 4-4. Handler 모듈 (`internal/handler/module.go` 신규)
```go
package handler

import "go.uber.org/fx"

var Module = fx.Module("handler",
    fx.Provide(
        NewPostHandler,
        NewAuthHandler,
        NewLikeHandler,
        NewFollowHandler,
        NewBookmarkHandler,
        NewUserHandler,
    ),
)
```

#### 4-5. 라우터 분리 (`internal/router/router.go` 신규)
```go
package router

func Setup(
    app *fiber.App,
    cfg *config.Config,
    postHandler *handler.PostHandler,
    authHandler *handler.AuthHandler,
    likeHandler *handler.LikeHandler,
    followHandler *handler.FollowHandler,
    bookmarkHandler *handler.BookmarkHandler,
    userHandler *handler.UserHandler,
) {
    // 현재 main.go의 라우팅 코드를 이동
}
```

#### 4-6. 리팩토링된 main.go
```go
func main() {
    fx.New(
        config.Module,
        database.Module,
        logger.Module,
        repository.Module,
        service.Module,
        handler.Module,
        router.Module,
    ).Run()
}
```

### 변경 파일 목록
| 파일 | 변경 유형 |
|------|----------|
| `internal/repository/module.go` | **신규** |
| `internal/service/module.go` | **신규** |
| `internal/handler/module.go` | **신규** |
| `internal/router/router.go` | **신규** - 라우팅 로직 분리 |
| `pkg/config/config.go` | **수정** - fx.Module 래퍼 추가 |
| `pkg/database/database.go` | **수정** - fx.Module 래퍼 + Lifecycle 훅 |
| `main.go` | **수정** - 대폭 단순화 |
| `go.mod` | **수정** - `go.uber.org/fx` 추가 |

### 엣지 케이스
- `AuthService`는 `userRepo`, `jwtSecret`, `jwtExpiryHours` 3개를 받으므로 스칼라 값(`string`, `int`)의 DI 주입 방식 결정 필요 -> `config.Config` 구조체를 통째로 주입하고 서비스 내에서 필요한 값 추출
- 인터페이스 반환 생성자(`NewPostRepository`가 `PostRepository` 인터페이스 반환)와 fx의 타입 해석이 호환됨 -> 문제 없음
- `UserHandler`가 `UserService`와 `PostService` 둘 다 의존 -> fx가 자동 해석

### 의존성 및 제약사항
- Structured Logging(항목 2)을 먼저 도입하면 fx의 내부 로거를 slog와 연결할 수 있어 디버깅이 용이
- fx 도입 시 기존 테스트에는 영향 없음 (인터페이스 기반 목은 그대로 사용)

---

## 5. DB Query Builder

### 기능 설명 (What)
반복되는 raw SQL 문자열 조합을 쿼리 빌더로 교체한다. 특히 SELECT 컬럼 목록, 조건부 WHERE, JOIN 구성의 중복을 제거한다.

### 구현 이유 (Why)
- `PostRepository`에 16개 메서드가 있으며, SELECT 컬럼 12개가 거의 모든 메서드에서 반복됨
- `WithUser` 변형 메서드는 `EXISTS(SELECT 1 FROM likes ...)` 서브쿼리만 추가되는데, 이를 위해 메서드 전체를 복사함
- 향후 검색, 필터링, 정렬 등 동적 쿼리 요구사항에 대응 어려움
- SQL 인젝션 위험은 pgx의 파라미터 바인딩으로 방지되고 있으나, 문자열 조합이 복잡해지면 실수 여지 증가

### 라이브러리 선택: `Masterminds/squirrel`

**선택 이유:**
- Go에서 가장 널리 사용되는 SQL 빌더 (GitHub stars 7k+)
- `database/sql` 호환 인터페이스 (`ToSql()` 메서드가 query string + args 반환)
- pgx와 조합 가능 (squirrel로 쿼리 생성 -> pgx로 실행)
- PostgreSQL의 `$1, $2` 플레이스홀더 지원 (`squirrel.Dollar`)
- ORM이 아니므로 기존 raw SQL 마이그레이션이 점진적으로 가능

**대안 비교:**
| 라이브러리 | 장점 | 단점 | 결론 |
|-----------|------|------|------|
| `Masterminds/squirrel` | 가벼움, 직관적 API, PostgreSQL 지원 | SELECT만 강하고 복잡한 서브쿼리는 여전히 수동 | **채택** |
| `doug-martin/goqu` | 더 풍부한 표현식, 다양한 dialect | API가 복잡, 학습 곡선 높음 | 대안 후보 |
| `jmoiron/sqlx` | 구조체 스캔 자동화 | 쿼리 빌더가 아닌 확장 라이브러리 | 보완재로 고려 |
| GORM | 풀 ORM, 마이그레이션 포함 | 과도한 추상화, 성능 오버헤드, 기존 pgx와 충돌 | 부적합 |
| raw SQL 유지 | 의존성 없음, 완전한 제어 | 현재 중복 문제 해결 불가 | 현재 상태 |

### 수락 기준 (Acceptance Criteria)
- [ ] `PostRepository`의 SELECT 컬럼 목록이 상수/함수 1곳에서만 정의된다
- [ ] `WithUser` / `WithoutUser` 변형이 조건부 컬럼 추가로 처리된다 (메서드 복사 없이)
- [ ] 기존 모든 쿼리가 동일한 결과를 반환한다 (회귀 테스트 통과)
- [ ] squirrel 사용 시 항상 `squirrel.Dollar` 포맷을 사용한다 (PostgreSQL)
- [ ] 복잡한 서브쿼리(EXISTS, CTE 등)는 `squirrel.Expr`로 처리한다

### 구현 방안

#### 5-1. 공통 컬럼 정의 (`internal/repository/query_helpers.go` 신규)
```go
package repository

import sq "github.com/Masterminds/squirrel"

// PostgreSQL 플레이스홀더 사용
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// 게시물 기본 컬럼
var postColumns = []string{
    "p.id", "p.author_id", "p.parent_id", "p.content", "p.visibility",
    "p.like_count", "p.reply_count", "p.created_at", "p.updated_at",
    "u.username", "u.display_name", "u.profile_image_url",
}

// 로그인 사용자 추가 컬럼 (is_liked, is_bookmarked)
func withUserColumns(userID uuid.UUID) []string {
    return []string{
        fmt.Sprintf("EXISTS(SELECT 1 FROM likes l WHERE l.user_id = '%s' AND l.post_id = p.id) AS is_liked", userID),
        fmt.Sprintf("EXISTS(SELECT 1 FROM bookmarks b WHERE b.user_id = '%s' AND b.post_id = p.id) AS is_bookmarked", userID),
    }
}

// 기본 게시물 SELECT 빌더
func basePostQuery() sq.SelectBuilder {
    return psql.Select(postColumns...).
        From("posts p").
        Join("users u ON p.author_id = u.id")
}
```

#### 5-2. 리팩토링 예시 - FindAll / FindAllWithUser 통합
```go
func (r *postRepository) FindAll(ctx context.Context, limit, offset int) ([]model.PostWithAuthor, error) {
    return r.findAllPosts(ctx, limit, offset, nil)
}

func (r *postRepository) FindAllWithUser(ctx context.Context, limit, offset int, userID uuid.UUID) ([]model.PostWithAuthor, error) {
    return r.findAllPosts(ctx, limit, offset, &userID)
}

func (r *postRepository) findAllPosts(ctx context.Context, limit, offset int, userID *uuid.UUID) ([]model.PostWithAuthor, error) {
    q := basePostQuery().
        Where(sq.Eq{"p.parent_id": nil}).
        OrderBy("p.created_at DESC").
        Limit(uint64(limit)).
        Offset(uint64(offset))

    if userID != nil {
        q = q.Columns(withUserColumns(*userID)...)
    }

    query, args, err := q.ToSql()
    if err != nil {
        return nil, fmt.Errorf("build query: %w", err)
    }

    rows, err := r.pool.Query(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    return r.scanPostRows(rows, userID != nil)
}
```

#### 5-3. 점진적 마이그레이션 전략
1. **Phase A**: `query_helpers.go`에 공통 빌더 함수 추가, `FindAll`/`FindAllWithUser`만 먼저 전환
2. **Phase B**: `FindByID` 계열 전환
3. **Phase C**: `FindByAuthorHandle` 계열 전환
4. **Phase D**: `FindReplies` 계열 전환
5. 각 Phase마다 기존 테스트 통과 확인

### 변경 파일 목록
| 파일 | 변경 유형 |
|------|----------|
| `internal/repository/query_helpers.go` | **신규** - 공통 쿼리 빌더 |
| `internal/repository/post_repository.go` | **수정** - squirrel 기반으로 리팩토링 |
| `internal/repository/bookmark_repository.go` | **수정** - squirrel 적용 |
| `internal/repository/like_repository.go` | **수정** - squirrel 적용 (해당 시) |
| `go.mod` | **수정** - `github.com/Masterminds/squirrel` 추가 |

### 엣지 케이스
- `EXISTS` 서브쿼리에 사용자 UUID를 직접 문자열로 삽입하면 SQL 인젝션 위험 -> `squirrel.Expr("EXISTS(SELECT 1 FROM likes l WHERE l.user_id = ? AND l.post_id = p.id)", userID)` 형태로 파라미터 바인딩 사용
- `parent_id IS NULL` 조건: squirrel에서 `sq.Eq{"p.parent_id": nil}`은 `p.parent_id IS NULL`로 변환됨 -> 정상 동작
- `ORDER BY` 방향이 쿼리마다 다름 (ASC/DESC) -> 빌더 함수에 파라미터로 전달
- CTE(WITH 절)가 필요한 경우 squirrel이 직접 지원하지 않음 -> `squirrel.Expr` 또는 raw prefix 사용

### 의존성 및 제약사항
- squirrel은 쿼리 문자열만 생성하므로 pgx 실행 레이어는 그대로 유지
- 기존 `scanPostRows` 헬퍼는 그대로 재사용
- 커서 기반 페이지네이션 전환 시 WHERE 조건 추가가 squirrel로 훨씬 자연스러워짐

---

## 6. 도입 순서 및 의존 관계

```
[1] Structured Logging ─────┐
                             ├──> [3] Dependency Injection
[2] Request Validator ──────┘         │
                                      v
                              [4] DB Query Builder
```

### 권장 순서

| 순서 | 항목 | 이유 | 예상 작업량 |
|------|------|------|-----------|
| 1차 | **Request Validator** | 독립적, 변경 범위 작음, 즉시 효과 | 소 (1-2시간) |
| 2차 | **Structured Logging** | 독립적, 이후 DI에서 로거 주입에 활용 | 소 (2-3시간) |
| 3차 | **Dependency Injection** | 1, 2차에서 추가된 로거/검증기를 DI로 관리 | 중 (3-4시간) |
| 4차 | **DB Query Builder** | 가장 큰 리팩토링, 회귀 테스트 필요 | 대 (4-6시간) |

### 이유
- Validator는 DTO 태그만 추가하면 되므로 가장 빠르고 안전
- Logging은 DI 없이도 수동 주입 가능하지만, DI 이후에 도입하면 더 깔끔 -> 그래도 DI 전에 도입하는 게 디버깅에 유리
- DI는 로거와 검증기가 준비된 후 한 번에 정리하는 것이 효율적
- Query Builder는 기존 쿼리 결과가 바뀌면 안 되므로 충분한 테스트 후 마지막에 적용

---

## 7. 열린 질문

| # | 질문 | 영향 범위 | 의사결정 필요 시점 |
|---|------|----------|------------------|
| Q1 | fx 대신 수동 와이어링을 유지하되 `internal/app/` 패키지로 분리만 하는 것이 더 나은가? 현재 규모(6개 서비스)에서 fx 도입이 과도할 수 있음 | DI 전체 | DI 구현 전 |
| Q2 | squirrel의 `withUserColumns`에서 UUID를 Expr 파라미터로 전달할 때, pgx의 `$N` 넘버링과 충돌하지 않는지 검증 필요 | Query Builder | Query Builder 구현 전 |
| Q3 | `Content` 필드의 공백만 있는 경우를 어떻게 처리할 것인가? (커스텀 `notblank` 태그 vs 서비스 레이어에서 `strings.TrimSpace`) | Validator | Validator 구현 시 |
| Q4 | 현재 offset 기반 페이지네이션이 일부 남아있는데(`LIMIT $1 OFFSET $2`), cursor 기반으로의 전환을 Query Builder 도입과 함께 진행할 것인가? | Query Builder + API | Query Builder 구현 시 |
| Q5 | 로그에 request body를 포함할 것인가? (디버깅 편의 vs 개인정보 보호) | Logging | Logging 구현 시 |

---

## 부록: 추가되는 외부 의존성 요약

| 라이브러리 | 버전 | 용도 | 라이선스 |
|-----------|------|------|---------|
| `github.com/go-playground/validator/v10` | v10.x | 구조체 태그 기반 요청 검증 | MIT |
| `go.uber.org/fx` | v1.x | 의존성 주입 컨테이너 | MIT |
| `github.com/Masterminds/squirrel` | v1.x | SQL 쿼리 빌더 | MIT |
| _(log/slog)_ | 표준 라이브러리 | 구조화된 로깅 | - |
