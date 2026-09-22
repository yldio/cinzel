---
title: "github unparse writes a null parse cannot read, and a blank line where it used to be"
module: "provider/github"
problem_type: logic_error
component: "provider/github"
severity: high
root_cause: "the attribute emit funnel had no nil guard, and the section separator was appended before the write that turned out to produce nothing"
symptoms:
  - "unparse writes 'defaults = null' or 'strategy = null', and parse refuses the file it just wrote"
  - "eleven more job and workflow keywords write a null parse silently drops, so HCL pass 1 and pass 2 differ"
  - "a step 'env:' or 'with:' key with no value under it is written and then refused with 'value must be set'"
  - "dropping the null leaves the blank line that separated it, so the two passes still differ"
  - "exit code 0 throughout, except where parse refuses"
tags:
  - "null"
  - "roundtrip"
  - "hclwrite"
status: "fixed"
created_date: "2026-09-20"
updated_date: "2026-09-20"
related:
  - docs/solutions/logic-errors/HANDOFF-github-internal-review.md
---

# github unparse writes a null parse cannot read

The github half of the defect fixed for gitlab in `8c2acee`, found by the probe
recorded in `HANDOFF-github-internal-review.md`. Three faults, each of which
hid the next.

## The criterion

A defect is cinzel's only if cinzel writes output its own parser rejects, or a
roundtrip loses information. "GitHub would reject this" is actionlint's job.
All three below fail the first test, not the second.

## 1. The emit funnel wrote nulls

`writeCommentedAttribute` in `provider/github/unparse_workflow.go` is the single
funnel every attribute goes through, and it had no nil guard where its gitlab
twin, `provider/gitlab/unparse_emit.go`, opens with one. A YAML key written with
no value under it became `<attr> = null`.

Parse drops such an attribute, so `timeout-minutes:`, `if:`, `name:`,
`concurrency:`, `container:`, `environment:`, `permissions:` and
`continue-on-error:` on a job, and `permissions:`, `concurrency:` and
`run-name:` on a workflow, each made the HCL from pass 1 differ from pass 2 by
a line carrying nothing.

`defaults` and `strategy` are worse, at both levels: they are blocks in the
schema, not attributes, so `defaults = null` is HCL cinzel's own parse refuses
to open, with `An argument named "defaults" is not expected here`. Unparse
wrote a file it could not read back, at exit 0.

The guard now sits in the funnel, matching gitlab. Unlike gitlab, github has no
keyword GitHub reads as "clear what this would otherwise inherit", so nothing
at job or workflow level needs an exception.

## 2. env and with values are the exception, and were refused

`env:` and `with:` are different. `V:` under them is a name GitHub defines as
empty, not a name it leaves undefined, so the null is the value and has to
survive. Unparse wrote `value = null`, which is right, and parse then refused
the whole file with `value must be set` — again, unparse writing HCL its own
parse cannot open.

Telling "written as null" from "not written at all" needs the source range, not
the value: gohcl fills an absent `hcl.Expression` field with a *synthetic* null
whose range is empty, so both resolve to `cty.NilVal`. `action.ValueWritten`
reads the range; `EnvListConfig.Parse` and `WithListConfig.Parse` keep the null
when the attribute was written and still refuse when it was not.

On the way out, `writeNameValueBlocks` calls `writeNullAttribute` rather than
going through the funnel, which would now drop it.

## 3. The separator outlived the attribute

Fixing 1 exposed this. Both writers append the blank line *before* writing the
key:

```go
if len(jobBody.Attributes()) > 0 || len(jobBody.Blocks()) > 0 {
    jobBody.AppendNewline()
}
```

With the guard in place the write that follows can produce nothing, and the
newline stays. Pass 2, whose YAML no longer carries the key, does not write it,
so the two passes still differed — by a blank line instead of by a null.

`bodySections` records how many attributes and blocks the body held when it
last separated, and skips a separator that nothing has closed. The workflow
body, each job body, the action body and the action `runs` body all use it. The
workflow body shares one instance between `writeWorkflowMetadata` and the
`jobs` list that follows, so the metadata's last section separates the list
exactly once.

## Reproducing

Build and probe through the CLI, one keyword per run — a file with several null
keywords stops at the first that errors and tells you nothing about the rest:

```
go build -o /tmp/cinzelbin .
/tmp/cinzelbin github unparse --file w.yml --output-directory h1
/tmp/cinzelbin github parse   --file h1/w.hcl --output-directory y1
/tmp/cinzelbin github unparse --file y1/w.yaml --output-directory h2
diff h1/w.hcl h2/w.hcl
```

Note the extension changes: unparse of `w.yml` parses back to `w.yaml`. A probe
script that assumes the name it wrote reports every case as unstable. An action
is named after the directory holding it, not after its own file, which is why
`TestActionHCLIsStableAcrossTwoPasses` finds the output rather than naming it.

## Tests

`provider/github/null_not_written_test.go`, modelled on
`provider/gitlab/null_not_written_test.go`. Two traps, both of which produced a
green test against broken code while the gitlab fix was being written:

1. hclwrite pads attribute names to a common width within a block, so
   `strings.Contains(hcl, "if = null")` misses `if  = null`. Normalise first
   with `strings.Join(strings.Fields(hcl), " ")`.
2. Neuter the fix and confirm the new test fails before believing it. Each of
   the three above was neutered separately and the matching test observed to
   fail. Neutering only part of a fix is not enough: sentinelling one
   `bodySections` instance left both stability tests green, because the fault
   needs the tracking gone everywhere.

## Not covered

Pass 2 gains an empty `permissions {}` block pass 1 has none of, because parse
writes `permissions: {}` into the YAML by default
(`provider/github/parse_workflow.go`, and the note in
`github-pin-comment-placement-and-empty-permissions-default.md`). That is a
deliberate least-privilege default, not part of this finding, and the tests
here set `permissions: {}` in their input so it is not what they measure.
