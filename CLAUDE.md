# CLAUDE.md

## Project Overview
A full-stack X (Twitter) clone monorepo.
**Core flow**: User Auth (JWT) → Feed (Cursor Pagination) → Interactions (Likes/Reposts) → Real-time Notifications (WebSocket).

## Build & Development Commands
# Frontend (Requires `bun` - NEVER use npm/yarn)
cd frontend && bun run dev        # Vite dev server
cd frontend && bun run check      # Typecheck & Lint

# Backend
cd backend && go run main.go      # Fiber server
cd backend && go test ./...       # Run tests

## Architecture Quick Reference
- **Frontend**: React 19 + TypeScript + Vite + Tailwind + shadcn/ui
- **Backend**: Go Fiber, Layered (handler → service → repository)
- **Auth**: JWT
- **Real-time**: WebSocket

## Navigation
- Go code convention → `.claude/rules/go-backend.md`
- React code convention → `.claude/rules/react-frontend.md`
- API pattern convention → `.claude/skills/api-patterns/SKILL.md`
- WebSocket pattern convention → `.claude/skills/websocket/SKILL.md`
- Testing convention → `.claude/skills/testing/SKILL.md`

## AI Directives (CRITICAL)
1. **Plan Mode**: Always output a numbered plan and wait for approval before writing code.
2. **Token Efficiency**: Remind user to run `/clear` after completing a task.
3. **ReBAC**: Service layer must verify user relationships (followers, blocks) before returning resources.
4. **Working Memory**: Read docs/TODO.md before starting any task. Update it after completing.
