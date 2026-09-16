---
status: complete
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

- [x] A summary-only diagnostic prints the summary
- [x] No output contains `%!w(<nil>)`

## Technical Details

`ProcessHCLDiags` falls back to `diag.Summary` when `diag.Detail` is empty, and
returns `ErrOpenIssue` on its own rather than joining an empty slice. Both
halves were needed: the fallback covers a summary-only diagnostic, and the
early return covers `hcl.Diagnostics{}` and any diagnostic with neither field
set, where `errors.Join` still returns nil.

A second, quieter defect went with it. `errors.Join` on a slice whose only
entries came from details meant a run of diagnostics where some carried a
detail and some did not reported only the ones that did. Collecting the summary
puts the rest back.

## Work Log

### 2026-09-16

This one could not be driven from the CLI: no `hcl.Diagnostic` constructed in
hcl v2.24.0 omits `Detail`, so every diagnostic the parser produces today
carries one. Reproduced by calling `ProcessHCLDiags` directly with a
summary-only diagnostic and with an empty set — both printed
`%!w(<nil>), if you think this is incorrect...`. After the fix the first prints
`Unsupported block type, ...` and the second the issue line alone.

That the trigger is not reachable through the parser today is the reason this
is worth pinning rather than leaving: the sentinel that would surface it is a
diagnostic cinzel or a future hcl builds itself, and nothing about the call
sites would announce the regression.

`internal/cinzelerror/process_diags_test.go` covers four cases — a summary
alone, a detail preferred over a summary, one of each joined, and an empty set —
and asserts no message carries a formatting verb. Three of the four were
confirmed to fail with the fix removed.

`go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check` all pass.
