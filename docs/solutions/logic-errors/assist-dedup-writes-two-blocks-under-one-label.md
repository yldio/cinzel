---
title: "Assist wrote two blocks under one label, in the directory it reads them from"
module: "internal/command"
problem_type: "logic_error"
component: "internal/command"
severity: "high"
root_cause: "a generated block colliding with the context kept the label the context already held"
symptoms:
  - "parse refuses the assist session: two step blocks share a label: 'checkout' is declared at ... and again at ..."
  - "The note branch and the reuse branch of the same comparison are each correct in the mode the other fails in"
  - "Dropping the block instead converts cleanly and silently binds the job to the context's version"
tags:
  - "assist"
  - "naming"
  - "dedup"
created_date: "2026-09-22"
updated_date: "2026-09-22"
---

## Problem Description

`deduplicateWithExisting` compares each generated block against `cinzel/*.hcl`
and takes one of two branches. A block identical to one already there becomes a
`// reuses:` comment and is dropped. A block with the same signature and a
different body was kept with a `// note:` saying where the other one is.

Both are right about the same thing and disagree about what to do with it. The
reuse branch drops its block because the context still holds one under that
label, so the reference resolves. The note branch kept its block under that same
label, so the label was declared twice.

That only matters because of where assist writes. Output goes to
`cinzel/assist/{timestamp}/assist.hcl`, inside the directory the comparison read
from, and `// reuses:` only means anything if the two are read together. Read
together is exactly when two blocks share a label:

```
two step blocks share a label: 'checkout' is declared at
cinzel/assist/20260922-120000/assist.hcl:2,1-16 and again at
cinzel/steps.hcl:1,1-16
```

Each branch is correct in one reading mode and wrong in the other:

| | parsed alone | parsed with the context |
| --- | --- | --- |
| reuse branch | step not found | ok |
| note branch | ok | duplicate label |

The `// reuses:` comment is what settles which mode is the real one. It is
written on the assumption the context is there, so the note branch is the one
that was wrong.

## Why not simply drop the block

Dropping the differing block, the way the identical one is dropped, parses at
exit 0. The job then binds to the context's block, which is the block that
differs — the generator asked for `actions/checkout@v4` and the workflow comes
out on `v3`. That trades a refusal for a wrong workflow, which is worse: the
refusal names the file and the line, and the silent version names nothing.

## The fix

The generated block takes a label nothing else holds, through
`naming.UniqueIdentifierInSet`, the same helper `provider/github` already uses
when its own unparse hits a collision. The references follow it: a reference is
a traversal, `step.checkout`, sitting in a `steps` list on a job and a `jobs`
list on a workflow, so the rewrite walks every attribute in every body to any
depth. A rename that moves the block and not the references is the silent case
again, in a new place — the same shape as
`matrix-axis-name-renamed-away-from-its-reference.md`, where an axis was renamed
and `${{ matrix.X }}` was left pointing at the old name.

The free label has to clear this run's own blocks as well as the context's. The
first version of the fix collected taken labels from the context and from what
had been kept so far, which let a rename land on the label of a generated block
further down the same file: `checkout` became `checkout_2` and there already was
a `checkout_2`. The clash came back with the names swapped and the job's step
list named one block twice. The set is now seeded from every generated block
before any rename happens, and each rename claims its new label as it takes it.

## Verification

Five reverts, each one compiling, each failing differently:

1. Keep the block under its original label — the original duplicate-label error.
2. Rename the block, leave the references — parses, job uses the context's `v3`.
3. Drop the block instead of renaming it — parses, job uses the context's `v3`.
4. Rename identical blocks too — two of the block that should have been reused.
5. Collect taken labels from kept blocks only — `checkout_2` declared twice.

The tests assert through a real `Parse` of the context directory rather than on
the merged text, because the text cannot tell a working rename from one that
lost its references. The version in the emitted YAML is what separates those
two: both parse, and only one is the workflow that was generated.
