---
status: complete
priority: p1
issue_id: "033"
tags: [code-review, correctness, gitlab, unparse]
dependencies: []
---

# A list of artifacts writes more blocks than parse accepts

## Problem Statement

`cache` and `artifacts` share an unparse branch that emits one block per list
entry. `parseJobBlock` allows at most one `artifacts` block, so a two-entry
list writes HCL that this tool cannot read back.

## Findings

`provider/gitlab/unparse_pipeline.go:707-745` writes the blocks;
`provider/gitlab/parse_pipeline.go:512` refuses them:
`job can include at most one artifacts block`.

## Recommended Action

In that branch, when the key is `artifacts` and the list holds more than one
entry, fail at write time with a message saying artifacts takes a single
object.

## Technical Details

- `provider/gitlab/errors.go`: `errArtifactsNotAList`, carrying the count so
  the message says how many were found
- The check sits after the single-value normalisation, so the object form and
  a one-entry list both pass through unchanged and `cache` is untouched
- Refusing rather than merging is right on the input's own terms: GitLab reads
  a single `artifacts` object per job, so a two-entry list is not valid
  pipeline input either, and merging would have to guess which `paths` wins

Worth knowing, not a defect: a one-entry `artifacts` list comes back as the
object form, since a single block has no list to be. GitLab reads the two the
same way, and the test pins that the output parses.

## Acceptance Criteria

- [x] A multi-entry artifacts list is refused with a clear error
- [x] A single-entry list and the object form still work
- [x] `cache` lists are unaffected

## Work Log

- 2026-09-16: Fixed as recommended. Every non-error case in the test parses
  its own output back, which is the property the list broke.
