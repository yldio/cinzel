---
status: complete
priority: p2
issue_id: "043"
tags: [code-review, github, unparse]
dependencies: []
---

# An action file with no name is read as a bare step

## Problem Statement

`isActionDocument` requires both `name` and `runs`. `name` is often absent in a
local action, so the file falls through to the step-only path and fails with an
opaque "not a valid type".

## Findings

`provider/github/unparse_action.go:26-33`.

## Recommended Action

Treat `runs` present with `on` and `jobs` absent as an action, and let
`validateActionDocument` report the missing `name` in its own words.

## Acceptance Criteria

- [x] An action with no `name` reports the missing field, not a type error
- [x] Step-only documents are still detected

## Technical Details

`isActionDocument` (`provider/github/unparse_action.go`) no longer looks for a
`name`. `runs` present with `on` and `jobs` absent is the whole test.

Nothing about the requirement changed: `validateActionDocument` still refuses a
document with no `name`, and it runs first thing in `actionToHCL`. Detection and
validation were doing the same check, and having detection do it meant a
nameless action was routed to the step-only path, where it failed on its shape
rather than its content.

A step-only document is a map of step ids and has no `runs` key, so it still
takes the step path. The two are not ambiguous.

## Work Log

### 2026-09-16

Reproduced through the CLI: an `action.yml` with `description` and `runs` but no
`name` failed with `not a valid type`. After the fix the same file reports
`action must define 'name'`.

`provider/github/action_detection_test.go` drives Unparse over four documents —
a nameless action, an action with no `runs.using`, a complete action, and a
step-only map. The first was confirmed to fail with the fix removed; the others
pin that the routing did not move.

`go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check` all pass.
