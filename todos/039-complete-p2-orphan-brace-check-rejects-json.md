---
status: complete
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

- [x] A script with `{"a": {"b": 1}}` and a later `${{ github.sha }}` parses
- [x] A genuinely unmatched `}}` is still reported

## Technical Details

`validateExpressionSyntax` had two passes: one walking the `${{ ... }}` pairs,
and a second looking for an orphaned closer. The second compared the first `}}`
in the string against the first `${{` and refused the string whenever the closer
came first, which is what a shell script ending nested JSON looks like.

Both passes are now one left-to-right scan. At each position the string either
opens an expression — checked for its closer and a non-empty body, then skipped
past — closes one, or is ordinary text. Ordinary text keeps a count of single
braces, so the `}}` ending `{"a": {"b": 1}}` has two open braces to account for
it and is not an orphan. A `}}` with fewer than two open braces behind it is
reported, and the existing carve-out stands: a string holding no `${{` at all is
left alone, so `echo }}` in a script that never interpolates still passes.

## Work Log

### 2026-09-16

Reproduced through the CLI: a step whose `run` held `{"a": {"b": 1}}` followed
by `$${{ github.sha }}` was refused with `orphaned '}}' without matching '${{'`.
After the fix it parses and the YAML carries the script unchanged; a genuine
`echo }} && echo $${{ github.sha }}` is still refused.

`provider/github/expression_test.go` gained six cases — JSON before and after an
expression, JSON with no expression at all, two genuine orphans and a lone
closer in a plain string. The two JSON-with-expression cases were confirmed to
fail with the fix removed.

`go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check` all pass.
