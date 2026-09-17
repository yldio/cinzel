---
title: "feat: a comment travels with what it is attached to, in both directions"
type: feat
status: active
date: 2026-09-17
---

# feat: a comment travels with what it is attached to, in both directions

## The rule

A comment belongs to the thing it is written against — an attribute, a block,
or a top-level resource — and survives conversion attached to that same thing,
in its original position, in both directions, in every provider.

Nothing is dropped silently. Where a comment cannot be carried, the conversion
says so rather than losing it.

## Mapping table

Three positions, each with its own home on both sides. HCL and YAML agree on
all three, so nothing needs to be moved or normalised.

### Above a resource

```hcl
# comment
step "step-1" {}
```

```yaml
steps:
  # comment
  - run: make
```

### Above an attribute

```hcl
step "step-1" {
  # comment
  run = "make"
}
```

```yaml
steps:
  - # comment
    run: make
```

### Trailing an attribute

```hcl
step "step-1" {
  run = "make" # comment
}
```

```yaml
steps:
  - run: make # comment
```

### Trailing a block, with nothing after it

```hcl
step "step-1" {
  run = "make"
  # comment
}
```

```yaml
steps:
  - run: make
    # comment
```

An earlier read of this had the head form collapsing into the trailing form on
the way out. It does not need to: yaml.v3 keeps a `HeadComment` on a mapping
key nested inside a sequence item and re-reads it intact, verified on `run`
inside a step. Position is preserved rather than normalised.

## What works today

- `parseBodyMap` (`provider/github/parse_workflow.go:624`) reads a trailing
  comment off every attribute of every body it walks, via
  `HCLVars.TrailingComment`, and wraps the value in `annotated`.
- `jobHeadComments` (`provider/github/job_comments.go`) collects the comment
  above a `job` block, and `writeLeadingComment`
  (`provider/github/unparse_emit.go`) writes it back.
- `Step.UsesComment` carries the pin tag off `version` — hand-wired, one
  attribute, one direction.

## What is broken

Verified by running the CLI, not by reading.

1. **Steps bypass the generic attribute path.** They decode through
   `StepConfig.Parse` → `yamlwriter.Convert`, a typed path with no source
   ranges, so no step attribute except `uses` carries a comment. Comments on
   `name`, `run`, `if`, and on `env`/`with` values are all dropped on parse.
2. **Head comments exist for job blocks only.** `jobHeadComments` is general in
   shape but called once, for `job`, and keyed by job label. A comment above a
   `step`, `env`, `with`, or `runs_on` block is dropped.
3. **Head comments on attributes are not read at all.** Only the trailing form
   is. `# c` above `run = "make"` is dropped on parse.
4. **Foot comments are not read at all.** A comment at the end of a block body
   with nothing after it is dropped.
5. **Unparse writes no attribute comments.** `writeLeadingComment` is called
   for jobs and nothing else; no attribute emit writes a trailing comment. A
   YAML comment on any step or attribute is dropped on the way to HCL.
6. **The pin roundtrip does not close.** `cinzel github pin` writes
   `version = "abc123" # v4`; parse carries it to YAML; unparse drops it. So
   pin output does not survive HCL → YAML → HCL today. Closing this is commit
   2, not a later nicety: a dropped pin tag loses the only record of which
   version a SHA stands for.
7. **GitLab has none of it.** `parseGenericBodyMap`
   (`provider/gitlab/parse_pipeline.go:988`) discards `SrcRange`, and
   `pipeline_yaml.go` is a separate hand-rolled node builder that never touches
   comments.

## Step dedup

`writeJobSteps` dedupes on `stepFingerprint`, a JSON marshal of the step map.
Two jobs running the same `actions/checkout@v4` collapse to one `step` block,
so two differently-commented uses of it have one block to live on.

Ruling: a comment is part of a step's identity. Two steps that differ only by a
comment are two steps, and the dedup must break. This needs no collision policy
— no "first wins", no concatenation.

This likely falls out for free. Once every step attribute carries its comment
into the step map, `stepFingerprint` marshals that map and sees them. Whether
it does depends on whether the fingerprint runs before or after `plain`/
`plainMap` strip the `annotated` wrapper. To be confirmed when reached, not
assumed.

Blast radius: no YAML fixture has a comment under `steps:`, and the only two
HCL fixtures with step reuse are parse-direction, where the fingerprint is not
involved. No golden expected to move. To be confirmed with `mise run drift`.

## Commits

Each is independently reviewable and leaves the tree green.

1. **Read head and foot comments off a body.** Generalise `commentRunAbove`
   into a shared collector that serves attributes and blocks, not just job
   headers. Parse only; nothing emits yet.
2. **Write attribute comments on unparse.** Head, trailing, and foot, from
   `annotated` through to `hclwrite`. Closes the pin roundtrip (item 6).
3. **Head comments for every block, not just jobs.** Generalise
   `jobHeadComments` past its single caller and single block type.
4. **Steps carry comments.** The structural one. Either steps join the generic
   source-range path, or `StepConfig` carries comments through the typed
   decode. `UsesComment` collapses into whichever wins.
5. **Confirm or fix `stepFingerprint`.** Verify comments reach it. If they do,
   this is a test and a note. If they do not, fold them in.
6. **GitLab.** The same rule in its own parse and emit paths.

Order matters: 4 is the largest and depends on 1-3 existing. 5 cannot be judged
until 4 lands.

## Comment text is never rewritten

A comment is prose. Its spacing, its `#` count and its indentation are things
its author chose, so the text is carried byte for byte and never reformatted,
re-wrapped or re-prefixed.

`hclwrite` emits any of `#no space`, `##  double`, `#   wide   spacing`
verbatim and the output reparses clean, so verbatim costs nothing. The one
thing ever added is a missing `#`: a line reaching HCL without one is not a
comment but a syntax error in the generated file. yaml.v3 hands the text back
with the `#` attached and the spacing intact, so that guard is close to
unreachable — it stays because it is the only case that can emit invalid HCL.

## Out of scope

- Comment reflow, re-wrapping, or normalising `#` spacing. See above: the text
  is carried as written.
- Comments inside expressions, such as between elements of a list. HCL accepts
  them and they round-trip, but they have no YAML home and no author writes
  them today.
- `//` and `/* */` HCL comment forms. Only `#` is read now; widening is a
  separate question.

## Acceptance

- [ ] Each of the four table rows round-trips HCL → YAML → HCL unchanged, in
      both providers.
- [ ] Each row round-trips YAML → HCL → YAML unchanged, in both providers.
- [ ] `cinzel github pin` output survives a full roundtrip, tag comment
      included.
- [ ] Comment text is byte-identical across a roundtrip: spacing, `#` count
      and indentation all as written.
- [ ] Two steps differing only by a comment produce two blocks.
- [ ] No comment present in an input is absent from the output of a conversion
      that has somewhere to put it.
- [ ] `go test -count=1 ./...`, `mise run lint`, `mise run drift`,
      `mise run license-check` green.

## Sources

- `provider/github/parse_workflow.go:624` — the generic attribute comment site
- `provider/github/job_comments.go` — `jobHeadComments`, `commentRunAbove`
- `provider/github/unparse_emit.go` — `writeLeadingComment`, `writeJobSteps`
- `provider/github/unparse_workflow.go:587` — `stepFingerprint`
- `internal/hclparser/sources.go` — `TrailingComment`
- `internal/yamldoc/document.go` — `WithComment`, `WithHeadComment`
- `provider/gitlab/parse_pipeline.go:988` — `parseGenericBodyMap`
- Prior plan: `docs/plans/2026-04-13-feat-step-uses-pin-comment-propagation-plan.md`
