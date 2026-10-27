---
title: "A comment above env, with, outputs or secrets is reattributed to one key"
module: "provider/github"
problem_type: logic_error
component: "provider/github, provider/github/step"
severity: medium
root_cause: "a YAML mapping becoming one HCL block per key, leaving the comment above the mapping with no block to sit on, and the writer putting it above the first one"
symptoms:
  - "a comment above 'env:' comes back inside the mapping, above whichever key sorts first"
  - "a comment above a mapping and one on its first key merge into a single run, indistinguishable"
  - "the same comment on a step's 'env' or 'with' is dropped outright, with no warning"
  - "exit code 0 throughout, and the second roundtrip is stable, so nothing reports it"
tags:
  - "comments"
  - "roundtrip"
  - "fan-out"
status: "fixed"
created_date: "2026-10-26"
updated_date: "2026-10-26"
related:
  - docs/solutions/logic-errors/github-env-block-comments-read-by-halves.md
  - docs/solutions/logic-errors/hcl-inline-comment-propagation-to-yaml.md
---

# A comment above a fan-out mapping has no block to sit on

`env`, `with`, `outputs` and `secrets` are mappings in YAML and one block per
key in HCL:

```yaml
# about the environment
env:
  # about ALPHA only
  ALPHA: one
  ZEBRA: two
```

```hcl
env {
  name  = "ALPHA"
  value = "one"
}
env {
  name  = "ZEBRA"
  value = "two"
}
```

Four comment positions are already handled, and the note above `env:` is a
fifth. It is about the set. The set stops existing at the conversion, so there
is no block for it.

## What each path did

`writeNameValueBlocks` (`provider/github/unparse_workflow.go`) wrote it above
the first block. Parse then read it as that key's own head, so a note about the
whole environment came back as a note about whichever key sorted first. With
both comments written, the two merged into one run and no reader could separate
them again.

The step path did not write it at all. `stepComments` collected it into
`Attrs["env"]`, and nothing in `provider/github/step` ever writes the head of a
mapping attribute, so it was gone with no warning. Same trigger, opposite
failure.

Both ended at exit 0, and the second roundtrip was a fixed point, so neither
showed up in a diff.

## The shape of the fix

There is no correct block, so the comment is not written. `headForMapping`
checks the key against `fanOutKeys`, reports the drop, and returns an empty
head.

The report names the file, the line the mapping key was written on, and the
path to it:

```
warning: dropping the comment above .github/workflows/ci.yml:21: jobs.build.steps[0].with: ...
```

A `commentSite` carries the file and path together through every writer from
`unparseYAMLFile` down to a step, rather than two parameters threaded through
the same dozen signatures. The line comes from `nodeComment.srcLine`, which
`collectComments` reads off the yaml.v3 key node it already walks; `empty()`
ignores it, so a key with a line and no comment is still empty.

Dropping it loudly beats relocating it quietly: the author has one comment to
move, rather than a claim they never wrote sitting on a key they never meant.

`permissions`, `defaults` and `concurrency` stay one block and are untouched.
So are the comments on the mapping's own keys and the one closing it — each of
those has a block to live in, which is why the earlier fix recorded in
`github-env-block-comments-read-by-halves.md` could keep them.

## The empty site

`stepFingerprint` builds a step's comments only to measure them, twice per
step. A `commentSite` with no file renders empty and silences the report, so
the conversion says it once rather than three times.

The first attempt threaded a bare path string and built the key's path with
`path + "." + key`, which turned the empty path into `".env"` — non-empty, so
it warned after all: nine warnings for five comments. `commentSite.String`
keys on the file instead, which cannot be grown into something by accident.
