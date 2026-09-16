---
status: complete
priority: p1
issue_id: "030"
tags: [code-review, correctness, github, unparse]
dependencies: []
---

# Two steps can be given the same label

## Problem Statement

`stepIdentifier` appends `_N` from a counter without checking that the result
is itself free. A document containing `build` twice plus a literal `build_1`
produces two `step "build_1"` blocks.

## Findings

`provider/github/unparse_workflow.go:523-563`. Reparsing the output fails:
`error in step 'build_1': already defined`.

## Recommended Action

Use `naming.UniqueIdentifierInSet`, which already loops until the candidate is
absent from the set.

## Technical Details

- `used` changes from `map[string]int` (a per-name counter) to
  `map[string]struct{}` (the set `UniqueIdentifierInSet` takes). The type is
  threaded through `writeWorkflowJobs`, `writeJobBody`, `writeJobKey` and
  `writeJobSteps`; `unparse_action.go` builds its own set the same way
- The suffix a plain repeat gets moves from `_1` to `_2`, since
  `UniqueIdentifierInSet` starts at 2. No golden holds a suffixed step label,
  so nothing moved: the labels emitted from `provider/github/testdata` are
  byte-identical before and after
- The positional fallback (`step_%d`) goes through the same claim, so a
  document mixing a literal `step_1` with an unnamed step no longer collides

## Acceptance Criteria

- [x] The collision case above produces three distinct labels
- [x] Existing golden labels are unchanged where no collision exists

## Work Log

- 2026-09-16: Fixed as recommended. Verified through the CLI that the
  three-step collision now emits `build`, `build_1`, `build_2` and parses
  back, and diffed every step label in the golden tree to confirm no churn.
