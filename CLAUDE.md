# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 프로젝트 개요

트위터 클론 모노레포 프로젝트. 프론트엔드와 백엔드가 하나의 저장소에 존재합니다.

## 프로젝트 구조

```
├── frontend/          # React + TypeScript (Vite)
├── backend/           # Go + Fiber v2
├── docker-compose.yml # Docker Compose 개발 환경
├── INSTRUCTIONS.md    # 코딩 컨벤션 및 규칙
├── CLAUDE.md          # Claude Code 가이드
└── .claudeignore      # Claude Code 무시 파일
```

## 빌드 & 실행 명령어

### 프론트엔드
```bash
cd frontend
bun install          # 의존성 설치
bun run dev          # 개발 서버 (Vite)
bun run build        # 프로덕션 빌드
bun run preview      # 빌드 결과 미리보기
```

### 백엔드
```bash
cd backend
go build ./...       # 빌드
go run main.go       # 실행
go test ./...        # 테스트
```

### Docker Compose (개발 환경)
```bash
docker compose up --build    # 전체 서비스 빌드 및 실행
docker compose up -d         # 백그라운드 실행
docker compose down          # 서비스 중지
```

## 아키텍처

- **프론트엔드**: Vite + React 19 + TypeScript. 서버 상태는 React Query, API 호출은 커스텀 hooks로 분리.
- **백엔드**: Go + Fiber v2. handler → service → repository 레이어드 아키텍처. 인터페이스 기반 의존성 주입.
- **데이터베이스**: PostgreSQL. 마이그레이션은 `backend/migrations/`에서 관리.

## 규칙

- 코드 작성 전 반드시 INSTRUCTIONS.md의 컨벤션을 따를 것
- 계획 모드에서 먼저 설계 후, 사용자 승인을 받고 구현
- Conventional Commits 사용
