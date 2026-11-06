---
title: "pin --dry-run exited 0 whether or not the HCL was pinned"
module: "internal/command"
problem_type: logic_error
component: "internal/command"
severity: medium
root_cause: "dry-run reported what a run would do and returned nil either way, so there was no flag that answered whether the work was already done"
symptoms:
  - "cinzel github pin --dry-run exits 0 on a workflow still following a tag"
  - "a CI step gating on it passes whatever the HCL holds"
  - "the dry-run report names each action in the past tense, over a file it did not write"
tags:
  - pin
  - upgrade
  - exit code
  - ci
status: fixed
created_date: "2026-11-06"
updated_date: "2026-11-06"
---

# pin --dry-run exited 0 whether or not the HCL was pinned

## What happened

`--dry-run` worked: both commands resolved every action, printed what they
found and wrote nothing. Verified against the real API, with the files compared
byte for byte before and after.

What it could not do is fail. The exit code was 0 on a directory where every
action followed a moving tag and 0 on one already pinned, so a CI step running
`cinzel github pin --dry-run` passed either way. The report was on screen; the
only part another program reads said nothing.

This is the same shape as
`pin-and-upgrade-reported-a-failure-at-exit-zero.md`, one step further along.
That note made a *failed* resolve exit non-zero. A successful one — an action
that resolved, in a run told not to write — stayed at exit 0, which is the
case a pinning check is actually about.

## Why

There was no flag for the question. `--dry-run` is a mode of the work: it
answers "what would this do", and a preview that aborted on finding something
to preview would be useless. The missing thing was the assertion, which `fmt`
and `diff` both already had — `fmt --check` fails on a file it would rewrite,
`diff` fails on YAML that drifted.

## The fix

`--check` on both commands. It implies `--dry-run`, so it writes nothing: a
check that pinned what it was asked to look for would pass on its own side
effect, and the second run of a two-run step would find the file clean because
the first run cleaned it.

```go
dryRun := c.Bool("dry-run") || check
```

The summaries already count what happened, so the finding is a count that was
there all along read the other way round. An action that had to be resolved was
not pinned when the run started, and nothing was written, so it still is not:

```go
if check && pinned > 0 {
	return fmt.Errorf("%w: %d of %d", errActionsNotPinned, pinned, len(results))
}
```

A failed resolve still outranks it. A run that could not reach an action does
not know whether it is pinned, so reporting it as unpinned would be a guess.

`--dry-run` alone is untouched and still exits 0.

Two reporting changes came with it, both about a run describing work it did not
do. The summary read "1 pinned" over a file left alone, and each action was
named in the past tense — "pinned actions/checkout@v4". Under `--check` those
per-action lines are the whole of what the reader opens the failed step to see,
and every one of them claimed the fix was already in. They now read "would pin"
and "would upgrade" under any dry run, `--check` or not.

## Prevention

A preview and an assertion are different commands wearing the same flag. When a
`--dry-run` starts getting recommended for CI, that is the signal the assertion
is missing: ask what the step is meant to go red on, and whether the exit code
says it.

The wording follows from the same place. If a flag means "do not do this", then
every line the run prints about having done it is wrong, including the ones
that were there before the flag existed.

## Tests

`internal/command/pin_check_test.go`. Resolving a tag is an API request, so the
CLI-level tests use what answers offline: an action on a full SHA is already
pinned and skipped without a request, a local action names no repository and
fails without one.

The decision itself is in the summary, so the finding is tested there, driving
`printPinSummary` and `printUpgradeSummary` with a result that carries a SHA
and no error.

`TestPinCheckWritesNothingToAFileItWouldPin` needs a run that genuinely
resolves, since a run that resolved nothing leaves the file alone too and
proves nothing. Pin reads a 24h cache before the API, so the test seeds it. It
asserts on the report as well as the file: if the cache key ever stops
matching, the run reports nothing pinned and the test fails rather than passing
on a resolution it never made. Upgrade has no offline equivalent — it asks for
the latest release through an uncached resolver — so the no-write case covers
pin only.
