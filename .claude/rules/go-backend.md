---
globs: backend/**/*.go
---

# Go Backend Code Rules (Auto-loaded when modifying *.go files)

## Mandatory Checklist
- [ ] Is `ctx context.Context` the first parameter?
- [ ] Are interfaces NOT received as pointers?
- [ ] Are errors wrapped with `fmt.Errorf` instead of `panic`?
- [ ] Are sentinel errors defined at the top of the package?
- [ ] Are empty slices declared with `var` instead of `make`?
- [ ] Does map iteration NOT depend on order?
- [ ] Are there no `init()` functions?
- [ ] Are side effects (time.Now, etc.) injected externally?
