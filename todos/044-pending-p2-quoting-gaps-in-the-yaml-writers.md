---
status: pending
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

- [ ] `Yes` / `Off` / `y` / `N` are emitted double-quoted
- [ ] `" padded "` is emitted double-quoted
- [ ] No single quotes appear in either provider's output
