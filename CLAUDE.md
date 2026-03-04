# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A full-stack X (Twitter) clone monorepo. 
**Core flow**: User Auth (JWT) → Feed (Cursor Pagination) → Interactions (Likes/Reposts) → Real-time Notifications (WebSocket).

## Build & Development Commands

```bash
# Frontend (Requires `bun` - NEVER use npm/yarn)
cd frontend
bun run dev          # Start Vite dev server
bun run check        # Typecheck & Lint

# Backend
cd backend
go run main.go       # Start Fiber server
go test ./...        # Run tests

```

## Architecture & Conventions

### Frontend (`frontend/`)

* React 19, TypeScript, Vite, Tailwind CSS, shadcn/ui.
* Server state via React Query. API calls MUST be separated into custom hooks.
* Use Functional Components (PascalCase) only.

### Backend (`backend/`)

* **Layered Architecture**: `handler` -> `service` -> `repository`. All communication MUST use Go interfaces.
* **Interfaces (Uber)**: NEVER use pointers to interfaces (e.g., avoid `*MyInterface`). Use the interface directly.
* **Context First**: `ctx context.Context` MUST always be the first parameter in all function signatures.
* **Parameter Order**: `ctx` -> clients/interfaces -> heavy types (slices/maps) -> light types (strings, bools).
* **Error Handling**: NEVER use `panic`. Use `fmt.Errorf("failed to do X: %w", err)` for error wrapping. Define Sentinel Errors (e.g., `ErrNotFound`) at the top of the package.
* **Dependency Injection**: Inject side effects (e.g., `time.Now()`, `rand.Int()`) from outside to ensure pure testability. Avoid `init()` functions (Uber).
* **Concurrency**: Embed `sync.Mutex` by value (not pointer) inside structs, and always pass the struct by pointer.
* **Slice Initialization**: Use `var s []string` for empty slices, NOT `s := make([]string, 0)`. If length is known, explicitly allocate using `make([]T, len, cap)`.
* **Avoid Map Loops**: Do NOT loop over maps if deterministic order is required (prevents flaky tests).
* **Naming Rules**: Use `get` for singular, `list` for plural. Do NOT use ambiguous terms like `info` or `details`.

## AI Directives (CRITICAL)

1. **Plan Mode**: Always output a numbered plan and wait for user approval before writing code.
2. **Token Efficiency**: Explicitly remind the user to run `/clear` after successfully completing a task.
3. **ReBAC**: Service layer must explicitly verify user relationships (followers, blocks) before returning resources.