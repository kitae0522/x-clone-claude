#!/bin/bash
# .claude/hooks/safety-check.sh
# 위험한 셸 명령 차단

INPUT=$(cat)
CMD=$(echo "$INPUT" | grep -o '"command"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"command"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')

# npm/yarn 사용 차단 (bun만 허용)
if echo "$CMD" | grep -qE "^(npm |yarn )"; then
  echo '{"decision": "block", "reason": "이 프로젝트는 bun만 사용합니다. npm/yarn 사용 금지!"}'
  exit 2
fi

# 위험 명령 차단
if echo "$CMD" | grep -qE "(rm -rf /|DROP TABLE|DELETE FROM.*WHERE 1|sudo)"; then
  echo '{"decision": "block", "reason": "위험 명령이 차단되었습니다."}'
  exit 2
fi

exit 0
