---
status: complete
priority: p2
issue_id: "018"
tags: [code-review, correctness]
dependencies: ["015"]
---

# A renamed action is never removed

## Problem Statement

Only workflows were written with the generated marker. An action.yml carried
nothing, so the prune could not tell an action cinzel wrote from one written
by hand, and owned neither. Renaming an action left the old directory there
with a live action.yml in it for good.

## Findings

Recorded while fixing 015, confirmed at b236f63. One action renamed from
"old" to "new", parsed twice into the same output directory:

```
out/old/action.yml
out/new/action.yml
```

Both are live actions as far as anything reading the directory is concerned,
and the first is one no HCL describes. The equivalent workflow rename prunes
correctly, and has since 015 reached subdirectories.

The first line of the emitted file was `name: A`, with no marker above it.

## Recommended Action

Write the marker on an action, the same as on a workflow. The prune already
walks the whole tree since 015, and Parse already records every action it
writes, so the marker is the only piece missing.

## Technical Details

- `provider/github/github.go`, `Parse`, the actions loop

Marking changes what every emitted action.yml looks like, so the other
direction was checked first: unparse drops the comments and the action keeps
its name. The golden tests compare semantically, so none of them move.

The empty directory left by a pruned action stays behind. The prune removes
files, not directories. It holds nothing and GitHub ignores it, so it is
recorded rather than fixed.

## Acceptance Criteria

- [x] An emitted action carries the marker
- [x] A renamed action is pruned
- [x] An action written by hand is left alone
- [x] A marked action still unparses back

## Work Log

- 2026-09-15: Created and closed. Recorded as a known gap in 015, where
  marking actions was left out to keep that change to one intent.
