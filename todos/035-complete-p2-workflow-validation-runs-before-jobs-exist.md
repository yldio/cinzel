---
status: complete
priority: p2
issue_id: "035"
tags: [code-review, validation, github, parse]
dependencies: []
---

# Validation walks a workflow that has no jobs yet

## Problem Statement

`validateParsedWorkflow(workflow)` is called before `workflow.Body["jobs"]` is
populated, so every check that walks jobs and steps — the expression syntax
check among them — sees an empty workflow. Those checks are dead on the parse
path.

## Findings

`provider/github/parse_workflow.go:115` validates; `:135` assigns the jobs.

## Recommended Action

Move the call to after the jobs assignment. Expect currently-passing fixtures
to start failing; that is the point, and each failure needs checking.

## Technical Details

- The call moves below the `if len(workflow.JobRefs) > 0` block, after
  `workflow.Body["jobs"]` is written
- Ordering against the other checks is unchanged in effect: the filename,
  boundary and claim checks still run first, so a workflow with a bad filename
  still fails on the filename
- **No fixture started failing.** The handoff expected some to. The whole
  suite passes untouched, and the repo's own `cinzel/*.hcl` regenerates
  `.github/workflows` byte-identical. So the newly-live checks find nothing
  wrong in what this repo already has, which is the good outcome, not a
  silenced one
- Verified the check really was dead: the same input exits 0 with the call in
  its old position and fails with it moved

## Acceptance Criteria

- [x] A bad `${{` inside a step is caught on the parse path
- [x] Existing fixtures either pass or are corrected, not silenced

## Work Log

- 2026-09-16: Moved as recommended. The new test drives Parse rather than
  calling the validator, since the validator was always right and only its
  position was wrong.
