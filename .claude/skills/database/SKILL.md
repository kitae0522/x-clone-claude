---
name: database-patterns
description: |
  PostgreSQL database patterns. Schema design, migration, query optimization.
  Triggers on: DB, database, table, schema, query, migration, index keywords.
---

# Database Patterns -- X Clone

## Schema Rules
- Table names: plural snake_case (users, posts, likes)
- PK: `id BIGSERIAL PRIMARY KEY`
- Timestamps: `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Soft delete: `deleted_at TIMESTAMPTZ` (nullable)
- FK: always create an index

## Cursor Pagination Index
- Feed query: `CREATE INDEX idx_posts_created_at ON posts(created_at DESC, id DESC)`
- Cursor: compose WHERE clause with `(created_at, id)` compound condition

## Migration Rules
- Filename: `{number}_{description}.up.sql` / `{number}_{description}.down.sql`
- Always write up/down pairs
- Verify schema via MCP after applying

## MCP Usage Checklist
- [ ] Query actual schema before writing Repository code
- [ ] Run EXPLAIN on new queries to check performance
- [ ] Verify schema consistency after migration
