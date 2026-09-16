---
status: complete
priority: p2
issue_id: "038"
tags: [code-review, correctness, github, unparse]
dependencies: []
---

# Two steps that differ only by id are merged into one

## Problem Statement

`stepFingerprint` deletes `id` before hashing, so two steps with identical
content but different explicit ids share a fingerprint and the registry reuses
one block for both. The second id is lost, and any `steps.<id>.outputs`
reference to it breaks.

## Findings

`provider/github/unparse_workflow.go:600-612`; the reuse happens in
`writeJobSteps` in `provider/github/unparse_emit.go` via `stepRegistry[fp]`.

## Recommended Action

Keep `id` in the fingerprint when it is explicitly present in the source, and
strip it only when it was generated.

## Technical Details

- `stepFingerprint` now marshals the step map as it stands. The "strip only
  when generated" half of the recommendation needs no code: the map is the
  source document, and a generated id is assigned later by `stepIdentifier`
  onto the `step.Step`, so it was never in the map to strip
- `json.Marshal` sorts map keys, so the fingerprint stays canonical
- Dedup across jobs is unaffected. Two id-less steps with the same content
  still share a block, and so do two steps carrying the *same* explicit id,
  which is the case the registry exists for
- No golden moved

## Acceptance Criteria

- [x] Two same-content steps with different explicit ids stay separate
- [x] Deduplication still fires for steps with no id
- [x] A step with an id and one without are not treated as the same step
- [x] The same explicit id in two jobs still collapses to one block

## Work Log

- 2026-09-16: Fixed. Confirmed each new case fails with the old fingerprint
  and that the dedup cases pass either way.
