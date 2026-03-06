---
name: testing-patterns
description: |
  Go test and React test writing guide.
  Triggers on: test, TDD, mock, unit, integration keywords.
---

# Testing Patterns -- X Clone

## Go Backend Tests
- Use table-driven test pattern
- Interface-based mock injection (gomock or manual mocks)
- Inject side effects like `time.Now()` for testability
- Never assert on map iteration order (prevents flaky tests)
- File naming: `*_test.go`, co-located in the same package

## React Frontend Tests
- Use React Testing Library + Vitest
- Test user behavior, not implementation details
- API call mocking: MSW (Mock Service Worker) recommended
- Custom hook testing: use `renderHook`

## Coverage Targets
- Backend service layer: 80%+
- Frontend custom hooks: 70%+
- Handlers/components: critical paths only
- **Overall goal: maintain 80%+ coverage on core business logic**
