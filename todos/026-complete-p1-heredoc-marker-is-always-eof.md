---
status: complete
priority: p1
issue_id: "026"
tags: [code-review, correctness, github, unparse]
dependencies: []
---

# A script containing EOF writes HCL that will not reparse

## Problem Statement

`setAsHeredoc` opens with `<<-EOF` and closes with `EOF` regardless of what the
script contains. A multi-line `run` with a bare `EOF` line closes the heredoc
early, and the rest of the script is read as HCL.

## Findings

`provider/github/step/stepdecode.go:266,281`. Unparsing a workflow whose `run`
contains a `cat <<EOF ... EOF` block produced HCL that failed to reparse.
Verified end to end.

## Recommended Action

Choose a marker not present in the body: try `EOF`, then `EOF_1`, `EOF_2` until
no line of the content equals it. Keep the `<<-` form.

## Technical Details

- `provider/github/step/stepdecode.go`, `setAsHeredoc` and the new
  `freeHeredocMarker`
- The comparison is against the raw lines, not the escaped ones:
  `escapeTemplateMarkers` only doubles `${` and `%{`, so a bare `EOF` reaches
  the output as written and is what would close the heredoc
- Lines are trimmed before comparing. The `<<-` form matches a closing marker
  after stripping indentation, so an indented `  EOF` closes it just as a bare
  one does
- A marker appearing inside a longer line is not a clash and does not move the
  marker, so the common case keeps `EOF`

## Acceptance Criteria

- [x] A `run` containing `EOF` on its own line roundtrips
- [x] Nested case (`EOF` and `EOF_1` both present) picks a free marker
- [x] Existing golden output is unchanged for scripts with no marker clash

## Work Log

- 2026-09-16: Reproduced through the CLI: the reparse failed with "A block
  definition must have block content delimited". Fixed by walking `EOF`,
  `EOF_1`, ... until no trimmed line matches. No golden file moved, since no
  fixture had a clash.
