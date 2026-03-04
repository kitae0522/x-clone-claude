---
globs: frontend/src/**/*.{ts,tsx}
---

# React Frontend Code Rules (Auto-loaded when modifying *.ts, *.tsx files)

## Mandatory Checklist
- [ ] Is it a functional component? (class components are forbidden)
- [ ] Is the component name in PascalCase?
- [ ] Are API calls extracted into custom hooks?
- [ ] Is bun used instead of npm/yarn?
- [ ] Is server state managed with React Query?
