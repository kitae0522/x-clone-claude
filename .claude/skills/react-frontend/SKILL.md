---
name: react-frontend-patterns
description: |
  React 19 frontend development patterns. Tailwind, shadcn/ui, React Query rules.
  Triggers on: component, screen, UI, page, hook, React, frontend keywords.
---

# React Frontend Patterns — X Clone

## Tech Stack
- React 19 + TypeScript + Vite
- Styling: Tailwind CSS + shadcn/ui
- Server state: React Query (TanStack Query)
- Package manager: bun (NEVER use npm/yarn)

## Mandatory Rules
- Use functional components only (PascalCase naming)
- All API calls must be extracted into custom hooks
  - e.g., `useGetFeed()`, `useLikePost()`
  - Never call fetch/axios directly inside components
- React Query queryKey must use consistent array format
  - e.g., `['posts', 'feed', { cursor }]`

## Directory Rules
- `src/components/` — Reusable UI components
- `src/features/` — Domain-specific feature modules
- `src/hooks/` — Custom hooks (including API calls)
- `src/lib/` — Utilities, types, constants

## shadcn/ui Usage Rules
- Check for existing shadcn components before custom styling
- Use `cn()` utility for conditional class composition
