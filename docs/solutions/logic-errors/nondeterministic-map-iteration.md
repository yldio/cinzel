---
title: "Non-deterministic output from map iteration: unsorted order, and mutation while ranging"
module: "GitHubProvider"
problem_type: "logic_error"
component: "parse_action, gitlab/parse_pipeline"
severity: "high"
root_cause: "logic_error"
symptoms:
  - "Golden tests pass intermittently"
  - "Same input produces different YAML key ordering"
  - "CI flakes on action parse tests"
  - "The same input moves a comment onto a different key from one run to the next"
tags:
  - "determinism"
  - "map-iteration"
  - "map-mutation"
  - "golden-tests"
  - "flaky-tests"
created_date: "2026-03-08"
updated_date: "2026-09-18"
---

> The functions named below were replaced by typed decode. `parse_action.go`
> no longer ranges over `sb.Attributes` at all: it reads a struct
> (`parseActionConfig`, `parseActionRunsConfig`), and a struct has field order.
> `sortedKeys` is still the answer wherever a map does reach output — see
> `provider/github/unparse_workflow.go:732`. The rule below stands; the call
> sites it names do not.

## Problem Description

`parse_action.go` iterated HCL body attributes using `for name, attr := range sb.Attributes`, which in Go produces non-deterministic ordering. This caused YAML output key order to vary between runs.

## Root Cause

Go map iteration order is intentionally randomized. Any code that iterates a map and writes ordered output (YAML, HCL, JSON) will produce non-deterministic results.

## Solution Implemented

Replaced direct map iteration with deterministic key-order helpers:

```go
// BEFORE
for name, attr := range sb.Attributes {

// AFTER
for _, name := range sortedKeys(sb.Attributes) {
    attr := sb.Attributes[name]
```

Applied consistently across `parseActionBody`, `parseActionRunsBlock`, and `parseActionBlockAttrs`.

## The second instance: mutating the map being ranged over

`relabel` (`provider/gitlab/parse_pipeline.go`) hit the same root cause from
the other side. It renamed comment tree labels in place:

```go
// BEFORE
for label, child := range mapping.children {
	key, renamed := keys[label]

	if !renamed || key == label {
		continue
	}

	delete(mapping.children, label)
	mapping.children[key] = child
}
```

The spec leaves it undefined whether a key added during a range is visited by
that same range. With a rename chain — `A` → `B` and `B` → `C`, which is what a
pipeline gets when two jobs swap names — the `B` inserted by the first
iteration was sometimes visited and renamed again to `C`, and sometimes not.
One run produced two children, the next produced one, from the same input. The
comment written above `A` landed on a different key each time.

The fix collects the renames first and applies them after the range ends, so
nothing is inserted into the map while it is being walked
(`TestRelabelIsDeterministic` runs it 200 times).

Sorting the keys would not have helped here: the order was not the problem, the
mutation was.

## Prevention Guidance

- **Rule**: Never use `for k, v := range someMap` when the output order matters.
- **Rule**: Never insert into a map inside a `range` over that same map. Collect
  what to change, end the range, then apply it. Deleting alone is defined and
  safe; inserting is not.
- Search for direct map iteration in any new parse/unparse code: `grep -n "range.*\.Attributes\|range.*Map\|range.*map\[" provider/github/*.go`
- Use deterministic key-order helpers that match local package conventions (`sortedKeys()` or equivalent) to avoid introducing new dependencies.
- Golden tests will eventually catch this, but the failures are intermittent and hard to reproduce locally.
