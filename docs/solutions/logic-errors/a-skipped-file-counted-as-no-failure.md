---
title: "A file pin could not read was counted as no failure at all"
module: "internal/pin"
problem_type: logic_error
component: "internal/pin, internal/command"
severity: medium
root_cause: "a file-level failure was warned about and dropped, and the summary counts results"
symptoms:
  - "pin reports 0 failed on a directory holding a file it could not parse"
  - "upgrade does the same, and its --parse gate reads the same results"
  - "exit code 0"
tags:
  - pin
  - upgrade
  - reporting
status: fixed
created_date: "2026-09-25"
updated_date: "2026-09-25"
---

# A file pin could not read was counted as no failure at all

## What happened

A `cinzel` directory holding two files, one already pinned and one that does not
parse:

```
warning: broken.hcl: failed to parse action refs in cinzel/broken.hcl: ...

Pin summary: 0 pinned, 1 already pinned, 0 failed
```

Exit 0, and the summary line says nothing failed. The warning above it is the
only sign that a file was skipped, and a run over a directory of any size puts
that warning far from the line the user reads as the result. `upgrade` printed
the same shape.

## Why

`PinDirectory` and `UpgradeDirectory` warn about a file whose contents cannot be
read and `continue`, which leaves no result behind for it. Both summaries count
results, so a whole file going missing counts as nothing:

```go
results, err := PinFile(ctx, path, resolver, w, dryRun)
if err != nil {
    reportf(w, "warning: %s: %v\n", entry.Name(), err)

    continue
}
```

The per-action failures inside a file are counted, because those are appended as
results carrying an error. A file-level failure was the one kind with nowhere to
be counted, and it is the larger one: every action in that file is unpinned, not
just one.

`findActionRefs`' own doc comment records the shape from the other side, where
one comment line reading `version = "v3"` made a file unparseable: "because
PinDirectory reports a failed file as a warning, the run still exited 0."

## The fix

The failure becomes a result, named by the path it happened to:

```go
allResults = append(allResults, PinResult{Action: path, Error: err})
```

Both summaries then count it, because both count a result carrying an error.
Nothing else reads these results in a way this changes: `upgrade`'s `--parse`
gate looks for a result with `Error == nil`, so a file that failed cannot
trigger a regeneration, which is what it did before.

The exit code is left alone. A partial run is still a run, and the question of
whether pin should exit non-zero on a skipped file is a separate decision from
whether it should say so. That decision was taken later, in
`pin-and-upgrade-reported-a-failure-at-exit-zero.md`: both commands now exit
non-zero on a counted failure, which includes the file-level one added here.

## Prevention

A summary built by counting a collection reports on the collection, not on the
work. Whenever a failure path drops an item instead of marking it, the two stop
matching, and the count is the thing people read.

The warning being printed is what hides this: the information is on screen, so
the code looks like it reports the failure. It reports it in the place a user
scrolls past, and denies it in the place they trust.

## Tests

`internal/pin/unreadable_file_counted_test.go` covers both directory walkers. It
asserts that the results name the unreadable file, rather than counting failures:
the upgrade stub fails every action it is asked about, so a count would have
passed for the wrong reason.
