---
status: pending
priority: p2
issue_id: "007"
tags: [code-review, security, quality]
dependencies: []
---

# Job names in error messages use `%s` — terminal injection via ANSI escape sequences

## Problem Statement

Three error messages in `buildWorkflowJobIndex` interpolate raw YAML job names using `%s`:

```go
fmt.Errorf("job '%s' listed in order but not found in jobs map", jobName)
fmt.Errorf("job '%s' must be an object", jobName)
fmt.Errorf("job '%s' is defined but was not included in the job order", jobName)
```

A job name containing ANSI escape sequences (e.g., `\x1b[31m`) or other control characters will be passed verbatim to the terminal when `cinzel` prints the error. This enables terminal injection for any user running `cinzel unparse` against a YAML file with a crafted job name.

## Findings

- `provider/github/unparse_emit.go:33,39,57`
- Job names come from YAML input and are not sanitized before error formatting

## Proposed Solutions

### Option A — Use `%q` instead of `%s` in all three format strings (Recommended)

```go
fmt.Errorf("job %q listed in order but not found in jobs map", jobName)
fmt.Errorf("job %q must be an object", jobName)
fmt.Errorf("job %q is defined but was not included in the job order", jobName)
```

`%q` uses Go's `strconv.Quote` semantics: control characters and escape sequences are hex-escaped, producing safe printable output.

- **Pros:** one-line fix; `%q` is already used in other error messages in this codebase
- **Cons:** changes error message format (quotes shift from `'` to `"`) — check test fixtures for exact string matches

### Option B — `strconv.Quote(jobName)` in format string

Equivalent to `%q` but more explicit.

## Recommended Action

_To be filled during triage._

## Technical Details

- Affected file: `provider/github/unparse_emit.go`, function `buildWorkflowJobIndex`

## Acceptance Criteria

- [ ] All three error messages use `%q` or equivalent safe quoting
- [ ] Tests that assert on the exact error message text are updated

## Work Log

- 2026-03-31: Finding created during code review
