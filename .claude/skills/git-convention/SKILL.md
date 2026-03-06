---
name: git-convention
description: |
  Git commit message, branch strategy, and PR rules.
  Triggers on: commit, branch, PR, merge keywords.
---

# Git Convention -- X Clone

## Conventional Commits
Commit messages MUST follow this format:

```
<type>(<scope>): <subject>
```

### Allowed Types
| Type | Description |
|------|-------------|
| feat | New feature |
| fix | Bug fix |
| docs | Documentation changes |
| style | Code formatting (no functional changes) |
| refactor | Refactoring (not a feature or bug fix) |
| test | Adding/modifying tests |
| chore | Build, config, dependency changes |
| perf | Performance improvement |
| ci | CI/CD configuration changes |

### Rules
- Subject must be under 50 characters, written in imperative mood
- Scope is optional (e.g., `feat(auth):`, `fix(feed):`)
- Body and footer are optional

## Branch Strategy
- `master` -- Production branch (direct push forbidden)
- `feat/<issue-description>` -- Feature development
- `fix/<issue-description>` -- Bug fixes
- `chore/<description>` -- Other tasks

## PR Rules
- New issue -> New branch -> New PR (one branch per issue)
- PR title under 70 characters
- Description must include Summary and Test plan
- At least 1 review recommended before merge

## Forbidden
- Never commit `package-lock.json` or `yarn.lock` (bun-only project)
- Never commit `.env` or credential files
- Never `--force-push` to master/main
