#!/bin/bash
# .claude/hooks/git-commit-guard.sh
# Validates git commit messages follow Conventional Commits format
# and blocks forbidden files (package-lock.json, yarn.lock).

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('command',''))" 2>/dev/null)

# Only check git commit commands
echo "$COMMAND" | grep -qE "git commit" || exit 0

# Check for forbidden lock files in staged changes
STAGED=$(git diff --cached --name-only 2>/dev/null)
if echo "$STAGED" | grep -qE "(package-lock\.json|yarn\.lock)"; then
  echo '{"decision": "block", "reason": "package-lock.json / yarn.lock이 스테이징에 포함됨. bun 전용 프로젝트입니다."}'
  exit 0
fi

# Extract commit message from -m flag
COMMIT_MSG=$(echo "$COMMAND" | sed -n 's/.*-m[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)

# If using heredoc style (-m "$(cat <<'EOF' ...), try to extract from that
if [ -z "$COMMIT_MSG" ]; then
  COMMIT_MSG=$(echo "$COMMAND" | sed -n "s/.*-m[[:space:]]*['\"]\\$(.*)/\1/p" | head -1)
fi

# If no inline message found, allow (might be using editor or heredoc)
[ -z "$COMMIT_MSG" ] && exit 0

# Validate Conventional Commits format
VALID_TYPES="feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert"
if ! echo "$COMMIT_MSG" | grep -qE "^($VALID_TYPES)(\(.+\))?!?:\s.+"; then
  echo "{\"decision\": \"block\", \"reason\": \"Conventional Commits 형식 위반. 형식: <type>(<scope>): <subject>\"}"
  exit 0
fi

exit 0
