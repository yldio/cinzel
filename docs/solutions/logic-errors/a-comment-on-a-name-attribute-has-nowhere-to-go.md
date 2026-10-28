---
title: "A comment was dropped wherever its attribute moved, and diff had nothing to report"
module: "provider/github, provider/gitlab"
problem_type: logic_error
component: "provider/github/parse_workflow.go"
severity: medium
root_cause: "a name/value block becomes one key with one line, and only the value's comment was read for it"
symptoms:
  - "a comment beside `name = \"...\"` in an env, with, output or secret block does not reach the YAML"
  - "the same for a GitLab `variable` block that collapses to a scalar"
  - "`cinzel github diff` reports no difference after the comment is added"
  - "exit code 0"
  - "`strategy.matrix.variable entries require a non-empty name` for a comment beside a matrix variable's name"
  - "`strategy.matrix.variable must be an object or list` for a comment above a matrix `variable {`"
  - "`schedule cron must be a string or list` for a comment beside `cron =`"
  - "a comment on a list-valued attribute comes out beside the next key down"
  - "a comment on a list-valued attribute does not survive YAML to HCL"
tags:
  - "comments"
  - "parse"
  - "annotated"
  - "yaml"
status: fixed
created_date: "2026-10-28"
updated_date: "2026-10-28"
---

# A comment on a name attribute has nowhere to go

## The symptom

A note written beside the name of a name/value block did not reach the YAML:

```hcl
step "drift" {
  name = "Check Generated Workflows" # fails if any committed YAML has drifted from its HCL
  run  = "mise run drift"
}
```

That case worked. This one did not:

```hcl
with {
  name  = "repositories" # only these two
  value = "cinzel,homebrew-cinzel"
}
```

Exit code 0, nothing on stderr, and the generated file came out byte for byte
what it was before. `cinzel github diff` therefore reported no difference
either, so the one command that exists to say "the YAML is not what the HCL
produces" agreed that nothing had happened. The author has no way of telling
a comment that was carried from one that was thrown away: both look like a
clean run.

Three neighbouring shapes did worse than drop the comment. They stopped the
parse and blamed the value:

```
error in job 'a': strategy.matrix.variable entries require a non-empty name
error in job 'a': strategy.matrix.variable must be an object or list
error in workflow 'w': workflow.w.on.schedule: schedule cron must be a string or list
```

all three from HCL that is valid, in reply to a comment being added to it.

And one put the comment somewhere it was never written:

```yaml
on:
  push:
    tags:
      - v*
jobs: # list comment on tags
```

## The cause

The general rule, which the reported symptom is one case of: **where an HCL
attribute and the YAML key it becomes are the same line, the comment was
carried. Where they move, nothing carried it.**

The attribute reported was `name`, but it was never about `name`. It is about
every attribute whose comment has no line of its own at the other end, and
there are three ways that happens.

**Several HCL lines become one YAML key.** An `env`, `with`, `output`, `secret`
or GitLab `variable` block is a block whose name and value become a single
key. A `uses` block's two attributes become one `uses:` string. The readers
took the value's comment and nothing else, so a note beside `name =` or
`action =` was dropped — not for being unreadable, but for having nowhere
obvious to sit.

**The key is not where the attribute was.** `jobs` and `steps` are reference
lists with no value of their own. `runs-on` collapses to the single `runners`
attribute inside its block. A matrix axis becomes a two-attribute `variable`
block. In each the writer on one side or the other took a head comment and no
trailing one, or took no comments at all: `strategy`, `matrix`, an action's
`runs.using` and its `branding` keys were written through `writeAttributeAny`,
which exists to say "no comments here".

**A wrapper read as an absence.** A comment travels through the parse inside
`annotated`, which wraps the value it was written on. A reader that takes a
value apart with a type assertion sees the wrapper instead of the type it
expected. Where that failure is read as "the field is not there" the comment is
dropped; where it reaches a `default:` or a `!ok` that returns, the comment
stops the conversion outright. `NormalizeStrategyMatrix` reads
`entry["name"].(string)` and reports the failure as a missing name; its type
switch has no case for a wrapper and reports that as "must be an object or
list"; `DenormalizeScheduleEvent` reads `cronVals.([]any)` and the caller
reports the leftover wrapper as a bad cron. This one is already named in
`patterns/critical-patterns.md`.

The misplacement is separate, in the encoder. yaml.v3 renders a line comment
where the node's own line ends. For a scalar that is the key's line; for a
filled block collection it is after the last entry, which is the line the
*next* key is on. A note on `tags = [...]` was emitted against `jobs:`, where
the next reader — cinzel included — reads it as a note on a key its author
never wrote.

### Both directions, or it is not carried

A comment is only carried if it survives the conversion and the one back. The
parse side was fixed first and the unparse side then dropped seven of them, so
the file converted and converted back had lost the note either way. Anything
changed on one side has to be checked on the other.

## The fix

Unwrap before the assertion, and give each comment the one place it belongs.

`namedBlockComments` and `Comments.setNested` now read the name as well as the
value. The key has one line and the value has first claim on it, since the
value is the attribute the key carries and a pin tag or a unit lands there; a
comment beside the name stacks above the key instead of taking a line already
spoken for. `liftCollapsedVariables` does the same for GitLab.

`unwrapMatrixVariableNames` and `unwrapMatrixVariableBlocks` take the wrapper
off before the normalizer reads it and put the comments back on the axis each
block becomes. `denormalizeSchedule` does the same for `cron`, moving the
comment to the `schedule` key: one cron line becomes several entries and none
of them is the line the comment was on.

`moveLineToKey` (`internal/yamldoc/encode.go`) moves a line comment off a
filled collection and onto its key, which is the line it was written on. An
empty collection renders as `[]` or `{}` on the key's line and keeps its own.

Moving it there exposed the other half: `collectComments`
(`provider/github/yaml_comments.go`) read `value.LineComment` only, so a
comment now emitted on the key was not read back and the roundtrip lost it. It
reads both, the way the GitLab reader already did.

## What is still not carried

Verified by probing every attribute and block opener in a workflow, an action
and a pipeline, in both positions, through parse and unparse, against the
released binary and this one: every roundtrip loss closed, 0 regressions. What
is left is the same before and after.

**A trailing comment on a block opener** — `job "a" { # note`, and the same
for `env {`, `with {`, `runs {`, `artifacts {`, every block type in both
providers. The key a block becomes has one line, and the head comment above
the block already occupies the one position it has. A comment written on its
own line above the block does reach the YAML, which is the position to use.

**A comment on `filename`**, on a workflow or an action. It names the file to
write and becomes no key in it, so there is nowhere in the output to put it.
It stays in the HCL, which is the only place it means anything.

A GitLab variable's name comment comes back on its `value` rather than its
`name`. The block collapses to one key, and the value is the line that key is
written on, so this is the comment arriving where it belongs rather than a
loss.

**A comment above a `workflow` block.** The workflow becomes the document,
whose head is the generated-marker block. Unparse puts a document head comment
above `name` instead, so a roundtrip moves it rather than keeping it.

The first two are properties of the output format rather than things left
undone: there is no key to carry them. The third is a real gap, and fixing it
means deciding how a document head comment and the generated markers share the
top of the file.

## Prevention

The shape to grep for is a type assertion on a value that came out of the parse
map, where failing the assertion is read as the field being absent:

```go
name, ok := entry["name"].(string)
if !ok { /* treated as "no name" */ }
```

Any of those breaks the moment a comment is written on that line, and the
breakage is reported in the vocabulary of the value rather than of the comment.
`plain` is the boundary where the wrapper comes off; anything reading parsed
content rather than emitting it goes through `plain` first. See
[[hcl-inline-comment-propagation-to-yaml]], whose "Prevention" section
describes the same failure in `emittedStepID`, and
[[a-reference-list-loses-the-comment-above-it]] for the unparse-side version
of the `jobs`/`steps` problem.
