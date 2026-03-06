---
description: GitHub Issue를 읽고 기획 -> 구현 -> 테스트 -> PR 생성까지 자동 실행
---

# Issue 기반 개발: $ARGUMENTS

## Step 1: Issue 분석
GitHub Issue $ARGUMENTS의 내용을 읽어줘.
- 제목, 설명, 라벨, 코멘트 확인
- 관련 이슈 참조 확인

## Step 2: 기획 (pm-spec 서브에이전트)
pm-spec 서브에이전트를 사용하여 Issue 내용 기반으로 스펙 작성.
- docs/specs/ 에 저장
- 결과를 나에게 요약하고 승인을 기다려라

## Step 3: 브랜치 생성
Issue 번호 기반으로 feature 브랜치 생성:
- `feat/issue-{번호}-{간단한설명}` 형식

## Step 4: 구현
승인된 스펙 기반으로 코드 구현.
- docs/TODO.md 업데이트하면서 진행

## Step 5: 테스트 (test-runner 서브에이전트)
test-runner 서브에이전트로 테스트 작성 + 실행.

## Step 6: 코드 리뷰 (code-reviewer 서브에이전트)
code-reviewer 서브에이전트로 셀프 리뷰.

## Step 7: PR 생성
GitHub MCP를 사용하여 PR 생성:
- 제목: Conventional Commits 형식
- 본문: 변경 사항, 이유, 테스트 결과
- 원본 Issue 자동 링크 (`Closes #번호`)
- 리뷰 보고서를 PR 코멘트로 첨부

## Step 8: Issue 업데이트
원본 Issue에 "PR #번호로 구현 완료" 코멘트 작성.
