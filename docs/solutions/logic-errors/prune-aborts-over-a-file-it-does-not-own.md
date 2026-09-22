---
title: "The prune ended the run over a file it was never allowed to delete"
module: "internal/fsutil"
problem_type: "logic_error"
component: "internal/fsutil"
severity: "high"
root_cause: "the ownership check reported a read failure as an error instead of as not-owned"
symptoms:
  - "parse exits 1 having already written its YAML, naming the issue tracker over a file the author wrote"
  - "bufio.Scanner: token too long, from a hand-written file in the output directory with a line over 64KB"
  - "open ...: permission denied, from a file in the output directory the process cannot read"
tags:
  - "prune"
  - "ownership"
  - "markers"
created_date: "2026-09-22"
updated_date: "2026-09-22"
---

## Problem Description

`PruneStaleGeneratedYAML` deletes files. `HasGeneratedMarker` is what stands
between it and the author's own work: a file carrying cinzel's two marker lines
may be removed, anything else may not. So every answer the check cannot give
has to fall on the same side, and one of them did not.

A file it could not read at all came back as an error. The error travelled up
through the walk, out of the prune, and ended the run. Two shapes reach it:

- a line longer than 64KB, which `bufio.Scanner` refuses with `token too long`
- a file the process cannot open, which `os.Open` refuses

Neither is a file cinzel wrote. Its markers are lines 1 and 2 of everything it
writes, so the scan returns `true` before a long line further down is ever
reached, and it does not write files it cannot read back. The read failure is
the evidence that the file belongs to someone else, and it was being treated as
a reason to stop.

The failure lands late. `parse` writes its YAML first and prunes afterwards, so
the run has already done its work when it exits 1, and the message points at
cinzel's issue tracker over a file the author put there.

## Why the missing-file case was the tell

`os.Open` on a path that is not there already returned `false, nil`. That case
was reasoned about and the rest was not: the same question, "may this be
deleted", has the same answer for every file the check cannot open, and only
one of them was answered.

## The Fix

Any file that cannot be opened, and any file that cannot be scanned through,
counts as not-owned. `scanner.Err()` is no longer consulted, because a file that
could not be read through is one whose markers were not found, which is what the
function already returns in that case.

`HasGeneratedMarker` keeps its `(bool, error)` signature. It is exported and
called from `provider/github/action_marker_test.go`, and dropping the error
would be a public interface change this does not need.

## Why not skip the file and keep the error

Because there is nothing left for the error to report. The caller's only
question is whether to delete, and "no" is a complete answer. An error that
says "no, and also stop" makes the prune's failure mode worse than doing
nothing: the file was going to be kept either way.

## Verification

Reverted separately, each still compiling, each failing:

1. `scanner.Err()` consulted again: the 100KB-line file ends the run with
   `bufio.Scanner: token too long`
2. the `os.Open` error returned again: the mode-`0000` file ends the run with
   `permission denied`

The test asserts three things per case, not one. The unreadable file survives,
this run's own output survives, and `stale.yaml` is gone. Without the third, a
prune that deletes nothing at all would pass.
