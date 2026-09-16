---
status: pending
priority: p2
issue_id: "042"
tags: [code-review, errors]
dependencies: []
---

# An HCL error with no detail renders as %!w(<nil>)

## Problem Statement

`ProcessHCLDiags` collects `diag.Detail` only. When every diagnostic has an
empty detail, `errors.Join` returns nil and the user is shown
`%!w(<nil>), if you think this is incorrect…`.

## Findings

`internal/cinzelerror/error.go`.

## Recommended Action

Fall back to `diag.Summary` when the detail is empty, and skip the join
entirely when nothing was collected.

## Acceptance Criteria

- [ ] A summary-only diagnostic prints the summary
- [ ] No output contains `%!w(<nil>)`
