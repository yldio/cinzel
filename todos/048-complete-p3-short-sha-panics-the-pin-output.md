---
status: complete
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

- [x] A short SHA prints without panicking
- [x] Full SHAs print the same twelve characters as now

## Technical Details

Two changes, because the helper on its own makes the bug quieter rather than
gone.

`shortSHA` returns the string whole when it is twelve characters or fewer, and
both call sites use it. That is what the todo asked for and it stops the panic.

But the panic was the only thing announcing a bad response. With it gone, a
resolve returning `""` writes `version = "" # v4` into the user's file and
prints "pinned actions/checkout@v4 → " as if it had worked — a pin silently
replaced by nothing, which is worse than a crash. `ResolveTag` decodes the SHA
out of JSON with no check, so an absent field gives exactly that.

So `isCommitSHA` gates both resolve sites: 40 characters, hex only. Anything
else becomes `errShortSHA` and takes the path a failed request already takes —
a warning naming what came back, an `Error` on the result, the file untouched.

`PinFile` and `UpgradeFile` both needed it; the two resolves are separate.

## Work Log

2026-09-16

Reproduced the panic by handing `mockResolver` a six-character SHA:

    panic: runtime error: slice bounds out of range [:12] with length 6
    internal/pin/pin.go:408

Added `shortSHA`, then checked what the fixed code did with the same input
rather than stopping there. It wrote `version = ""` into the file and reported
success, which is how the second half of the fix was found.

Three tests: `TestShortSHA` and `TestIsCommitSHA` on the helpers, and
`TestAShortSHAIsRefusedNotWritten` / `TestUpgradeRefusesAShortSHA` driving
`PinFile` and `UpgradeFile` end to end — each asserting a failed result, the
reason on stdout, and the file byte-identical afterwards.

Proving the tests fail without the fix took two passes, since stashing the
whole change only broke the build, which proves nothing. Keeping the helpers
and reverting their use reproduced the original panic; keeping the length check
and dropping only the refusal reproduced the silent empty pin:

    want one failed result, got [{Action:actions/checkout Tag:v4 SHA: Error:<nil>}]

Both halves are load-bearing.

Verified with `go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check`.
