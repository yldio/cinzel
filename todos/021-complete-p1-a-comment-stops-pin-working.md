---
status: complete
priority: p1
issue_id: "021"
tags: [code-review, correctness, supply-chain]
dependencies: ["019"]
---

# A comment stops pin working on the whole file

## Problem Statement

`findActionRefs` matched the text of every `action = "..."` and
`version = "..."` in the file and paired them by position. Anything that read
that way counted, a comment included. One line such as

    // TODO: was on version = "v3" before the bump

left one more version than actions, so the file was refused with
`mismatched action/version count: 1 actions, 2 versions` and nothing in it
was pinned.

`PinDirectory` reports a failed file as a warning and carries on, so the run
exits 0 and the summary reads `0 pinned, 0 already pinned, 0 failed`. The
action stays on a mutable tag and nothing says so. `upgrade` behaves the same
way, through the same function.

## Findings

All three comment syntaxes trigger it: `//`, `#`, and `/* */`. Reproduced
end to end:

    warning: steps.hcl: failed to parse action refs in hcl/steps.hcl:
      mismatched action/version count: 1 actions, 2 versions
    Pin summary: 0 pinned, 0 already pinned, 0 failed

Two shapes probed and found not to be defects:

- Attribute order inside a uses block. Reordering `version` before `action`
  still paired correctly, since the counts stay equal and positions still
  interleave.
- `with { version = "1.21" }`. Not reachable: `WithConfig` in
  `provider/github/action/with.go` declares `name` and `value` as required
  attributes, so a bare `version` there is invalid HCL and the parser
  rejects it long before pin runs.

## Fix

`findActionRefs` parses with `hclsyntax.ParseConfig`, already used for the
same purpose in `internal/ai/strip.go`, and walks the blocks looking for
`uses`. Comments are not part of the parsed body, so they cannot be counted.
The parser also gives the byte range of each attribute, which is what the
rewrite from 019 needs, so the offsets come from the parse rather than from
a second regex.

A uses block holding only one of the two attributes is skipped rather than
paired with a neighbour's half. An action whose value is an expression rather
than a literal string is left alone, since rewriting it would destroy the
expression.

## Verification

`internal/pin/find_refs_test.go`. Proven to fail with only `findActionRefs`
reverted to the regex:

- `TestCommentMentioningAVersionIsIgnored` — all three comment forms
- `TestAnActionBesideACommentIsStillPinned` — end to end through PinFile
- `TestIncompleteUsesBlockIsSkipped` — halves paired across blocks

`TestRefsComeBackInSourceOrder` is a regression guard, not a gate, and is
labelled as such in the file. It passes with the sort removed in every shape
tried, because hclsyntax keeps `Blocks` in source order and the walk follows
it. The sort stays because the back-to-front splice depends on the ordering
and should state that requirement rather than rest on a property of
hclsyntax that nothing here records.

Against the repository's own `cinzel/steps.hcl` the parser finds 12 refs,
matching the 12 `version` attributes, 12 `action` attributes and 12 uses
blocks the old regex counted. Behaviour on valid input is unchanged.
