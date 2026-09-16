---
status: pending
priority: p2
issue_id: "039"
tags: [code-review, validation, github, expression]
dependencies: []
---

# The unmatched-"}}" check fails on scripts containing JSON

## Problem Statement

The check counts `}}` against the first `${{`, so a `run` script with nested
JSON braces followed later by a real `${{ ... }}` is read as an orphaned
closer and the workflow is refused.

## Findings

`provider/github/expression.go:47-73`.

## Recommended Action

Scan left to right tracking the open `${{` depth, and report a closer only when
the depth is zero.

## Acceptance Criteria

- [ ] A script with `{"a": {"b": 1}}` and a later `${{ github.sha }}` parses
- [ ] A genuinely unmatched `}}` is still reported
