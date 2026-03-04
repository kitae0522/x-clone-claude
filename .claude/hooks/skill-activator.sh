#!/bin/bash
# .claude/hooks/skill-activator.sh

INPUT=$(cat)
PROMPT=$(echo "$INPUT" | grep -o '"prompt"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"prompt"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')

[ -z "$PROMPT" ] && exit 0

SUGGESTIONS=""

# go-backend-patterns
if echo "$PROMPT" | grep -qiE "(handler|service|repository|Go|Fiber|미들웨어|라우터|백엔드)"; then
  SUGGESTIONS="${SUGGESTIONS}\n  -> go-backend-patterns [critical]"
fi

# react-frontend-patterns
if echo "$PROMPT" | grep -qiE "(컴포넌트|화면|UI|페이지|hook|React|프론트)"; then
  SUGGESTIONS="${SUGGESTIONS}\n  -> react-frontend-patterns [critical]"
fi

# api-patterns
if echo "$PROMPT" | grep -qiE "(API|엔드포인트|JWT|인증|페이지네이션|fetch|요청)"; then
  SUGGESTIONS="${SUGGESTIONS}\n  -> api-patterns [high]"
fi

# websocket-patterns
if echo "$PROMPT" | grep -qiE "(WebSocket|실시간|알림|notification|소켓|푸시)"; then
  SUGGESTIONS="${SUGGESTIONS}\n  -> websocket-patterns [high]"
fi

# testing-patterns
if echo "$PROMPT" | grep -qiE "(테스트|test|TDD|mock|커버리지)"; then
  SUGGESTIONS="${SUGGESTIONS}\n  -> testing-patterns [high]"
fi

if [ -n "$SUGGESTIONS" ]; then
  FEEDBACK="[SKILL ACTIVATION] 추천 스킬:${SUGGESTIONS}\n해당 스킬의 SKILL.md를 읽고 규칙을 따르세요."
  echo "{\"feedback\": \"$FEEDBACK\"}"
fi

exit 0
