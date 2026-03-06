---
name: websocket-patterns
description: |
  WebSocket-based real-time notification system patterns.
  Triggers on: WebSocket, real-time, notification, socket keywords.
---

# WebSocket Patterns -- X Clone

## Architecture
- Use Go Fiber WebSocket handler
- Connection management: connection pool pattern
- Auth: Validate JWT token during WebSocket handshake

## Event Types
- `notification:new` -- New notification (like, repost, follow)
- `notification:read` -- Mark as read
- `feed:update` -- Real-time feed update (optional)

## Frontend Integration
- Manage connection via custom hook `useWebSocket()`
- Reconnection logic: exponential backoff
- Auto-invalidate React Query cache on events
