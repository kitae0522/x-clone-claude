---
description: 전체 품질 검사 + E2E 테스트까지 포함한 풀 검사
---

# 풀 품질 검사

기존 quality-check 항목을 모두 실행하고, 추가로:

## E2E 테스트 (Playwright MCP)
1. 로그인 플로우 테스트
2. 게시물 CRUD 플로우 테스트
3. 좋아요/리포스트 인터랙션 테스트
4. 실시간 알림 수신 테스트

## DB 정합성 (PostgreSQL MCP)
5. 모든 FK에 인덱스가 있는지 확인
6. NOT NULL 컬럼에 default 값이 설정되어 있는지
7. 마이그레이션 up/down 쌍이 모두 존재하는지

결과를 docs/reports/full-check-{오늘날짜}.md에 저장.
