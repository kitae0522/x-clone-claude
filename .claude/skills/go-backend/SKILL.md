---
name: go-backend-patterns
description: |
  Go backend development patterns and Fiber framework rules.
  Triggers on: handler, service, repository, Go, Fiber, middleware, router keywords.
---

# Go Backend Patterns — X Clone

## Layered Architecture Rules
- Follow handler → service → repository order strictly
- Use Go interfaces for all inter-layer communication
- Never use interface pointers (`*MyInterface` is forbidden)

## Function Signature Rules
- First parameter: `ctx context.Context` (always)
- Parameter order: ctx → clients/interfaces → heavy types (slices/maps) → light types (strings, bools)

## Error Handling
- Never use `panic`
- Wrap errors: `fmt.Errorf("failed to do X: %w", err)`
- Define sentinel errors at package top: `var ErrNotFound = errors.New("not found")`

## Dependency Injection
- Inject side effects externally: `time.Now()`, `rand.Int()`, etc.
- Never use `init()` functions (per Uber Go Style Guide)

## Concurrency
- Embed `sync.Mutex` as a value in structs (not as a pointer)
- Always pass that struct by pointer

## Slices/Maps
- Empty slices: `var s []string` (do not use make)
- Known length: `make([]T, len, cap)` with explicit allocation
- Never iterate over maps when deterministic order is required (prevents flaky tests)

## Naming
- Singular lookup: `get`, plural lookup: `list`
- Avoid vague names like `info`, `details`

## ReBAC (Relationship-Based Access Control)
- Service layer must verify user relationships (followers, blocks) before returning resources
