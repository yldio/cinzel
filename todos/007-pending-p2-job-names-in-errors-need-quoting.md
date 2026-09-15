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

**The defect is real but sits in a different file.** Terminal injection was
confirmed against 3a8ee45 — a job name carrying ANSI escapes reaches stderr
with the raw `0x1b` byte intact:

```
printf 'on: push\njobs:\n  "\e[31mPWNED\e[0m": not-an-object\n'
```

The escape arrives through the validator's message, not through the three
`unparse_emit.go` lines this issue names. `validateWorkflowYAMLDoc` runs
first, at `unparse_workflow.go:128`, so a malformed job never reaches
`buildWorkflowJobIndex`; changing those three to `%q` fixes nothing an input
can actually trigger.

Two things follow. The fix belongs in `provider/github/validate.go`, on the
messages that interpolate a YAML key. And the HCL writer is already correct —
a job name with escapes emits as `id = "\u001b[31mPWNED\u001b[0m"`, so only
the error path leaks.

The `unparse_emit.go` messages should still move to `%q` for consistency, but
as tidying, not as the security fix.

## Technical Details

- Affected file: `provider/github/unparse_emit.go`, function `buildWorkflowJobIndex`

## Acceptance Criteria

- [x] Injection reproduced and traced to its real source
- [ ] Validator messages that interpolate a YAML key quote it safely
- [ ] A test asserts no raw `0x1b` reaches the error output
- [ ] `unparse_emit.go` messages moved to `%q` for consistency

## Work Log

- 2026-03-31: Finding created during code review
- 2026-09-15: Reproduced against 3a8ee45. Confirmed real, but the named call
  site is unreachable — validation rejects the input first. Fix relocated to
  validate.go. HCL writer verified already safe.
