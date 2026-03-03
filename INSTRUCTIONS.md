# INSTRUCTIONS.md

이 문서는 트위터 클론 프로젝트의 기술 스택과 코딩 컨벤션을 정의합니다.

## 기술 스택

| 영역 | 기술 |
|------|------|
| 프론트엔드 | TypeScript, React 19, Vite, React Query (TanStack Query) |
| 백엔드 | Go, Fiber v2 |
| 데이터베이스 | PostgreSQL |
| 캐시 | Redis |
| 패키지 매니저 | bun (프론트엔드) |

## 프론트엔드 컨벤션

### 컴포넌트
- 함수형 컴포넌트만 사용 (클래스 컴포넌트 금지)
- 컴포넌트 파일명은 PascalCase (`TweetCard.tsx`)
- 한 파일에 하나의 export 컴포넌트만 작성

### Import / Export
- barrel export (`index.ts`에서 re-export) 금지
- 절대경로 import 사용 (`@/components/TweetCard`)
- import 순서: 외부 라이브러리 → 내부 모듈 → 상대경로

### 상태 관리 & 데이터 패칭
- 서버 상태: React Query (TanStack Query)
- 클라이언트 상태: React Context 또는 Zustand (필요 시)
- API 호출은 별도 hooks 파일로 분리 (`useTweets.ts`, `useAuth.ts`)

### 스타일링
- CSS Modules 또는 Tailwind CSS
- 인라인 스타일 금지

### 타입
- `any` 타입 사용 금지
- API 응답 타입은 `types/` 디렉토리에 정의
- Props 타입은 컴포넌트 파일 내에 정의

## 백엔드 컨벤션

### 아키텍처 (레이어드)
```
handler (HTTP 레이어) → service (비즈니스 로직) → repository (DB 접근)
```
- handler: 요청 파싱, 응답 반환, 입력 검증
- service: 비즈니스 로직, 트랜잭션 관리
- repository: SQL 쿼리, DB 접근만 담당
- 각 레이어는 인터페이스를 통해 의존성 주입

### 디렉토리 구조
```
backend/
├── main.go
├── internal/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── model/
│   ├── dto/
│   └── middleware/
├── pkg/
│   ├── config/
│   └── database/
└── migrations/
```

### 에러 핸들링
- 커스텀 에러 타입 사용 (`AppError`)
- handler에서 일관된 JSON 에러 응답 반환
- 에러 로깅은 서비스 레이어에서 수행
- panic 사용 금지 — 모든 에러는 명시적으로 반환

### 네이밍 규칙
- 파일명: snake_case (`tweet_handler.go`)
- 구조체/인터페이스: PascalCase (`TweetService`)
- 변수/함수: camelCase (exported는 PascalCase)
- 상수: PascalCase 또는 ALL_CAPS

### API 응답 형식
```json
{
  "success": true,
  "data": {},
  "error": null
}
```

## 데이터베이스 컨벤션

### 네이밍 규칙
- 테이블명: snake_case, 복수형 (`tweets`, `users`)
- 컬럼명: snake_case (`created_at`, `user_id`)
- 인덱스: `idx_{테이블}_{컬럼}` (`idx_tweets_user_id`)
- FK: `fk_{테이블}_{참조테이블}` (`fk_tweets_users`)

### 마이그레이션
- `backend/migrations/` 디렉토리에 순번 기반 관리
- 파일명: `{번호}_{설명}.up.sql`, `{번호}_{설명}.down.sql`
- 모든 마이그레이션은 up/down 쌍으로 작성

### 공통 컬럼
- 모든 테이블에 `id` (UUID), `created_at`, `updated_at` 포함
- soft delete가 필요한 테이블에 `deleted_at` 추가

## Git 컨벤션

### Conventional Commits
```
<type>(<scope>): <description>

[optional body]
```

타입:
- `feat`: 새 기능
- `fix`: 버그 수정
- `refactor`: 리팩토링
- `docs`: 문서 변경
- `test`: 테스트 추가/수정
- `chore`: 빌드, 설정 변경

스코프: `frontend`, `backend`, `db`, `infra`
