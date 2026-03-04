---
description: 프로젝트 전체 품질 검사를 실행하고 보고서를 생성합니다.
---

# 품질 검사 실행

다음 검사를 순서대로 실행하고, 각 결과를 보고서로 작성해줘:

## Backend (cd backend)
1. **빌드 검사**: `go build ./...`
2. **테스트**: `go test ./...`
3. **정적 분석**: `go vet ./...`
4. **포맷 검사**: `gofmt -l .` (포맷 안 된 파일 목록)

## Frontend (cd frontend)
5. **타입 체크 & 린트**: `bun run check`
6. **빌드 검사**: `bun run build`

## 코드 품질 (전체)
7. **TODO/FIXME 주석 수집**: 코드 내 TODO, FIXME, HACK 주석 목록
8. **미사용 import 탐지** (Go 파일)

## 보고서 형식
각 항목별:
- 상태: ✅ 통과 / ❌ 실패
- 발견된 문제 수
- 주요 문제 (최대 5개)
- 권장 수정 방안

결과를 `docs/reports/quality-{오늘날짜}.md`에 저장해줘.
