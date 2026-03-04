---
globs: backend/**/*_test.go, frontend/src/**/*.test.{ts,tsx}
---

# Test File Rules

## Go Tests
- Table-driven pattern is mandatory
- Never assert on map iteration order
- Mocks must be interface-based

## React Tests
- Test user behavior (getByRole, getByText)
- Never test implementation details (direct state/props access)
- Use MSW for API mocking
