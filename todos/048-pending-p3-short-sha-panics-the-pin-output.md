---
status: pending
priority: p3
issue_id: "048"
tags: [code-review, pin, robustness]
dependencies: []
---

# sha[:12] assumes the API returned a full SHA

## Problem Statement

Both progress lines slice the resolved SHA to twelve characters with no length
check. A short or empty value from the API panics instead of reporting a bad
response.

## Findings

`internal/pin/pin.go:408` and `internal/pin/upgrade.go:113`.

## Recommended Action

One helper that returns the whole string when it is shorter than twelve.

## Acceptance Criteria

- [ ] A short SHA prints without panicking
- [ ] Full SHAs print the same twelve characters as now
