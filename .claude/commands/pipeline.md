---
description: 기획 → 구현 → 테스트 → 리뷰 전체 파이프라인을 실행합니다.
---

# 전체 개발 파이프라인: $ARGUMENTS

다음 단계를 순서대로 진행해줘:

## Step 1: 기획 (pm-spec 서브에이전트 사용)
pm-spec 서브에이전트를 사용하여 "$ARGUMENTS"에 대한 스펙을 작성.
- docs/specs/에 스펙 저장
- 스펙 결과를 나에게 요약 보고

## Step 2: 구현 (사용자 승인 후)
스펙 문서를 기반으로 코드 구현:
- 백엔드: handler → service → repository 순서
- 프론트엔드: 커스텀 훅 → 컴포넌트
- docs/TODO.md 업데이트하면서 진행

## Step 3: 테스트 (test-runner 서브에이전트 사용)
test-runner 서브에이전트를 사용하여:
- Go: 테이블 드리븐 테스트 작성 + 실행
- React: 커스텀 훅 + 컴포넌트 테스트 작성 + 실행

## Step 4: 코드 리뷰 (code-reviewer 서브에이전트 사용)
code-reviewer 서브에이전트를 사용하여:
- Go 체크리스트 12항목 검증
- React 체크리스트 5항목 검증
- 리뷰 보고서 작성

## Step 5: 최종 보고
- 전체 과정 요약 보고서를 docs/reports/에 저장
- docs/TODO.md 체크 표시 업데이트
- docs/CONTEXT.md에 의사결정 기록
- "/clear 실행을 권장합니다" 알림
