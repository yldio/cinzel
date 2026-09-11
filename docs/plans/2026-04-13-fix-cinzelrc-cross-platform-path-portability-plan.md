---
title: "fix: cinzelrc path portability across OS"
type: fix
status: active
date: 2026-04-13
---

# fix: cinzelrc path portability across OS

## Problem Statement

`.cinzelrc.yaml` is committed to git and shared across developers. Path values (`file`, `directory`, `output-directory`) are stored verbatim — no normalization, no validation. Two failure modes:

1. **Absolute paths** — a dev on macOS writes `/Users/alice/project/cinzel/main.hcl`; breaks on every other machine and OS.
2. **Windows path separators** — `cinzel\main.hcl` is valid on Windows but not POSIX. Go's `filepath.FromSlash` handles this, but it is not applied on load.

The contract should be: **all paths in `.cinzelrc.yaml` must be relative and use forward slashes.** Relative paths with forward slashes work on all platforms Go supports.

## Proposed Solution

Two changes, both in `internal/command/config.go:loadProviderCommandConfig`:

### 1. Reject absolute paths at load time

When a path value is set for `file`, `directory`, or `output-directory`, check `filepath.IsAbs`. If true, return an error with a message that names the key and explains the fix:

```
.cinzelrc.yaml.github.parse.file must be a relative path (got "/Users/alice/...")
```

Also reject any path beginning with `~` — `filepath.IsAbs` does not catch these on any platform. Use `strings.HasPrefix(value, "~")` and return the same error pattern.

### 2. Normalize separators on load

Apply `filepath.FromSlash` to every path value before storing. This converts forward-slash paths to the OS-native separator, so a git-tracked config with `./cinzel/main.hcl` works on Windows without modification.

```go
// internal/command/config.go
config.file = filepath.FromSlash(valueNode.Value)
```

## Acceptance Criteria

- [ ] Absolute path in any path field → clear error naming the field, before any provider code runs
- [ ] Forward-slash relative paths (`./cinzel/main.hcl`) work on Windows after normalization
- [ ] `~` in paths → error with message directing user to use a relative path
- [ ] Existing tests in `internal/command/config_test.go` pass; new cases added for:
  - Absolute path rejection (each path field: `file`, `directory`, `output-directory`)
  - `~` path rejection
  - Separator normalization (forward slash input → OS native)

## Files

- `internal/command/config.go` — main change: `loadProviderCommandConfig`
- `internal/command/config_test.go` — new test cases
- `README.md` — clarify path contract in config section

## Out of Scope

- Walk-up search for `.cinzelrc.yaml` (separate feature)
- `~` tilde expansion (rejected with error instead)
- Changing the user-level AI config (`internal/ai/config.go`) — uses `os.UserConfigDir()`, not shared in git

## Sources

- Config loader: `internal/command/config.go:67–175`
- Path fields stored at lines 135, 140, 145
- Related plan: `docs/plans/2026-03-09-feat-cinzelrc-provider-config-precedence-plan.md`
- Related solution: `docs/solutions/logic-errors/config-input-precedence-ignored-for-parse.md`
