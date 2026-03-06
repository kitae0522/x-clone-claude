# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A full-stack X (Twitter) clone monorepo.
**Core flow**: User Auth (JWT) -> Feed (Cursor Pagination) -> Interactions (Likes/Reposts) -> Real-time Notifications (WebSocket).

## Build & Development Commands

```bash
# Frontend (Requires `bun` - NEVER use npm/yarn)
cd frontend && bun run dev     # Start Vite dev server
cd frontend && bun run check   # Typecheck & Lint

# Backend
cd backend && go run main.go   # Start Fiber server
cd backend && go test ./...    # Run tests
```

## Architecture Quick Reference

| Area | Key Rules |
|------|-----------|
| Frontend | React 19 + TypeScript + Tailwind + shadcn/ui, React Query required |
| Backend | handler -> service -> repository layers, Go interface communication |
| Auth | JWT Bearer, httpOnly cookie |
| Pagination | Cursor-based only (offset forbidden) |

## Navigation
- Go code convention -> `.claude/rules/go-backend.md`
- React code convention -> `.claude/rules/react-frontend.md`
- API patterns -> `.claude/skills/api-patterns/SKILL.md`
- WebSocket patterns -> `.claude/skills/websocket/SKILL.md`
- Testing -> `.claude/skills/testing/SKILL.md`
- Git convention -> `.claude/skills/git-convention/SKILL.md`
- DB patterns -> `.claude/skills/database/SKILL.md`

## AI Directives (CRITICAL)

1. **Plan Mode**: Always output a numbered plan and wait for user approval before writing code.
2. **Token Efficiency**: Explicitly remind the user to run `/clear` after successfully completing a task.
3. **ReBAC**: Service layer must explicitly verify user relationships (followers, blocks) before returning resources.
4. **Working Memory**: Read docs/TODO.md before starting any task. Update it after completing.
