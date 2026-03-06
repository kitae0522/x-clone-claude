---
name: api-patterns
description: |
  Front-back API integration patterns. JWT auth, Cursor Pagination, error response format.
  Triggers on: API, endpoint, JWT, auth, pagination, fetch keywords.
---

# API Patterns -- X Clone

## Authentication Flow
- Pass JWT token in Authorization header as Bearer format
- Handle token refresh logic in axios/fetch interceptor on expiry
- Frontend token storage: httpOnly cookie or in-memory

## Cursor Pagination Rules
- Never use offset-based pagination
- Response format:
  ```json
  {
    "data": [...],
    "next_cursor": "base64_encoded_cursor",
    "has_more": true
  }
  ```
- Use React Query's `useInfiniteQuery`

## Standard API Response Format
- Success: `{ "data": ... }`
- Error: `{ "error": { "code": "NOT_FOUND", "message": "..." } }`
- In Go backend, use Fiber's `c.Status(xxx).JSON()`

## ReBAC Verification Points
- Feed listing: Filter posts from blocked/blocking users
- Profile viewing: Private accounts accessible only to followers
- Notifications: Never generate notifications in blocked relationships
