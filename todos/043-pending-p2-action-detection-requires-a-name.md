---
status: pending
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

- [ ] An action with no `name` reports the missing field, not a type error
- [ ] Step-only documents are still detected
