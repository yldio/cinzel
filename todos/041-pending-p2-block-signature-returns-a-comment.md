---
status: pending
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

- [ ] A commented block still matches an identical existing block
- [ ] Uncommented blocks behave as before
