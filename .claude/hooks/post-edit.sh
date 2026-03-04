#!/bin/bash
# .claude/hooks/post-edit.sh
# 수정된 파일에 따라 적절한 포맷터/린터 실행

INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | grep -o '"file_path"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"file_path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')

if [ -z "$FILE_PATH" ]; then
  FILE_PATH=$(echo "$INPUT" | grep -o '"path"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
fi

if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# Go 파일
if [[ "$FILE_PATH" == *.go ]]; then
  cd backend 2>/dev/null
  gofmt -w "../$FILE_PATH" 2>/dev/null || gofmt -w "$FILE_PATH" 2>/dev/null
  echo '{"feedback": "[QA] Go 파일 자동 포맷팅 완료 (gofmt)"}'
fi

# TypeScript/React 파일
if [[ "$FILE_PATH" == *.ts || "$FILE_PATH" == *.tsx ]]; then
  cd frontend 2>/dev/null
  bunx prettier --write "../$FILE_PATH" 2>/dev/null || bunx prettier --write "$FILE_PATH" 2>/dev/null
  echo '{"feedback": "[QA] TypeScript 파일 자동 포맷팅 완료 (prettier)"}'
fi

exit 0
