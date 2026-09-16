---
status: complete
priority: p2
issue_id: "041"
tags: [code-review, correctness, assist]
dependencies: []
---

# A block with a leading comment gets the comment as its signature

## Problem Statement

`blockSignature` takes the first line of the block. When the block carries a
leading `//` comment, that comment is returned instead of `step "checkout"`, so
the signature never matches and `deduplicateWithExisting` cannot emit the
`// reuses:` reference for it.

## Findings

`internal/command/assist.go:372`.

## Recommended Action

Skip leading comment and blank lines, then take the first line that opens a
block.

## Acceptance Criteria

- [x] A commented block still matches an identical existing block
- [x] Uncommented blocks behave as before

## Technical Details

`blockSignature` (`internal/command/assist.go`) walks the block's lines and
returns the first that is neither blank nor a comment, trimmed of its opening
brace. It used to take line one outright.

Both comment spellings are skipped, `//` and `#`. A block that is nothing but
comments returns the empty string, which `deduplicateWithExisting` already
treats as "no signature" when indexing the context directory.

This matters most for pinned actions: `cinzel github pin` writes an
`// action tag` line above each step, so in practice a large share of the
context blocks carried one and none of them could ever be matched.

## Work Log

### 2026-09-16

Reproduced by calling `blockSignature` over a block with an `// action tag`
comment: it returned `// action tag: v4` rather than `step "checkout"`. After
the fix it returns the block header.

Tests: `TestBlockSignature` gained three cases (a `//` comment, a `#` comment
with a blank line after it, and a block that is only a comment), and a new
`TestDeduplicateMatchesACommentedBlock` writes a commented block into the
context directory and asserts the generated copy collapses to
`// reuses: step "checkout" from steps.hcl`. All four were confirmed to fail
with the fix removed.

`go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check` all pass.
