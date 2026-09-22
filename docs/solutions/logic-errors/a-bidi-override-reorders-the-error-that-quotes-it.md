---
title: "A bidirectional override rewrote the error message quoting it"
module: "internal/cinzelerror"
problem_type: "logic_error"
component: "internal/cinzelerror"
severity: "medium"
root_cause: "the terminal guard escaped control characters only, and a bidi override is not one"
symptoms:
  - "A job name ending in U+202E renders the rest of the error backwards, so the file named at fault is not the file shown"
  - "The same forgery an ANSI cursor move already could not do, by another route"
tags:
  - "terminal"
  - "unicode"
  - "errors"
created_date: "2026-09-22"
updated_date: "2026-09-22"
---

## Problem Description

`SafeForTerminal` exists because errors quote names taken from the input, and a
name is free to carry an ANSI escape. Written to a terminal as-is the escape is
acted on rather than shown, so a crafted name can erase the line naming the file
at fault and put its own there. The function escapes every C0 and C1 control to
stop that.

A bidirectional override is not a control character, and it does the same thing.
U+202E reverses everything after it for the rest of the line, so a name ending
in one renders the text that follows backwards:

```
error in job 'build<U+202E> lmth.tsefinam ni': boom
```

reads on a terminal as `error in job 'build in manifest.html': boom`. The name
is not what it appears to be and neither is the file. The guard passed it
through untouched because it only asked whether the rune was a control.

Twelve characters do this: the five embeddings and overrides with the pop that
closes them, the four isolates, and the three marks that set the direction of
neutral characters beside them.

## Why not escape every format character

The zero-width joiner is category Cf too, and an emoji sequence is built out of
it. `unicode-emoji-zwj-escape-roundtrip.md` is a whole fix for carrying those
through unchanged; escaping the joiner here would print a workflow named
`👮‍♂️ Lint` as its pieces. The named set is the point, not a convenience.

A name that leans on a bidi mark to set its own direction now reads with the
mark shown rather than applied. Worse to look at, and it still says what the
name is.

## Verification

Two reverts, each compiling:

1. The bidi cases removed: 12 of the 12 subtests fail on
   `U+…  reached the terminal`.
2. Every Cf character escaped, which also makes those 12 pass: 2 emoji subtests
   fail on `want it unchanged`.

The second revert is why the emoji control exists. Widening the switch to all
of category Cf looks correct and breaks the roundtrip the other note fixed.
