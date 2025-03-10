---
title: "Non-deterministic YAML output from unsorted map iteration in parse_action.go"
module: "GitHubProvider"
problem_type: "logic_error"
component: "parse_action"
severity: "high"
root_cause: "logic_error"
symptoms:
  - "Golden tests pass intermittently"
  - "Same input produces different YAML key ordering"
  - "CI flakes on action parse tests"
tags:
  - "determinism"
  - "map-iteration"
  - "golden-tests"
  - "flaky-tests"
created_date: "2026-03-08"
updated_date: "2026-03-08"
---

## Problem Description

`parse_action.go` iterated HCL body attributes using `for name, attr := range sb.Attributes`, which in Go produces non-deterministic ordering. This caused YAML output key order to vary between runs.

## Root Cause

Go map iteration order is intentionally randomized. Any code that iterates a map and writes ordered output (YAML, HCL, JSON) will produce non-deterministic results.

## Solution Implemented

Replaced direct map iteration with `maputil.SortedKeys()`:

```go
// BEFORE
for name, attr := range sb.Attributes {

// AFTER
for _, name := range maputil.SortedKeys(sb.Attributes) {
    attr := sb.Attributes[name]
```

Applied consistently across `parseActionBody`, `parseActionRunsBlock`, and `parseActionBlockAttrs`.

## Prevention Guidance

- **Rule**: Never use `for k, v := range someMap` when the output order matters.
- Search for direct map iteration in any new parse/unparse code: `grep -n "range.*\.Attributes\|range.*Map\|range.*map\[" provider/github/*.go`
- The existing codebase uses `maputil.SortedKeys()` (cross-package) and `sortedKeys()` (local to github package) — use whichever is in scope.
- Golden tests will eventually catch this, but the failures are intermittent and hard to reproduce locally.
