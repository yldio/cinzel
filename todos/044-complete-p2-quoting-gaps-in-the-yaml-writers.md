---
status: complete
priority: p2
issue_id: "044"
tags: [code-review, output, github, gitlab]
dependencies: []
---

# Three gaps between the quoting rule and the quoting code

## Problem Statement

The project rule is: quote when needed, with double quotes, never relying on
single quotes. Three cases miss it.

`Yes`, `Off`, `y` and `N` in their capitalized forms are not recognised as
booleans and go out unquoted.

A string with a leading or trailing space is left to yaml.v3, which wraps it in
single quotes.

GitLab mapping keys never go through the check at all, so a key like `#x` or
`a: b` comes out single-quoted.

## Findings

- `internal/yamldoc/encode.go`, `needsQuoting`
- `provider/gitlab/pipeline_yaml.go`, `stringNeedsQuoting`, and the key node
  built at `:155`

## Recommended Action

Match the YAML 1.1 boolean set case-insensitively, quote strings whose ends are
whitespace, and run mapping keys through the same check as values.

## Acceptance Criteria

- [x] `Yes` / `Off` / `y` / `N` are emitted double-quoted
- [x] `" padded "` is emitted double-quoted
- [x] No single quotes appear in either provider's output

## Technical Details

Both writers keep their own copy of the predicate, so all three fixes landed
twice: `needsQuoting` in `internal/yamldoc/encode.go` and
`stringNeedsQuoting` in `provider/gitlab/pipeline_yaml.go`.

The literal chain `v == "true" || v == "yes" || ...` became a lookup in a
`plainWords` map on `strings.ToLower(v)`, which brings in the single-letter
`y` and `n` alongside the words. A `strings.TrimSpace(v) != v` check covers the
padded strings: yaml.v3 did quote those, but picked single quotes.

Keys did not take the same predicate. Running them through it would have
quoted every GitLab job name holding a colon, and
`TestJobNamesThatAreNotIdentifiersSurviveRoundtrip` caught that immediately —
`test:unit` came back as `"test:unit"`. A colon only ends a key when a space
follows it, so keys got `keyNeedsQuoting`, which shares the bool, null and
whitespace rules but replaces the blanket `:` and `#` scan with the two shapes
that actually break a key: `": "` or a trailing colon, and `" #"` or a leading
`#`.

Values kept the blanket scan. `TestJobScalarKeywordsParse` pins
`container: "node:18"`, so quoting a value with a colon is deliberate and was
left alone.

## Work Log

2026-09-16

Reproduced all three by encoding directly: `Yes`, `Off`, `y` and `N` came out
bare, `" padded "` came out single-quoted, and `#x` and `a: b` came out
single-quoted as keys.

First attempt narrowed the colon rule for values as well as keys, which broke
`TestJobScalarKeywordsParse` and `TestJobScalarKeywordsStillAcceptBlocks`.
Those tests are the existing decision on values, so the narrowing was pulled
back to keys only.

Reverted both writers with `git stash push` and ran the new tests against the
old code: ten assertions failed across the two packages, covering every case
listed above.

Verified with `go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check`. No golden file moved.
