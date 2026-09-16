---
status: pending
priority: p3
issue_id: "047"
tags: [code-review, cleanup]
dependencies: []
---

# Two packages with no callers

## Problem Statement

`internal/filereader` and `internal/filewriter` have no users outside their own
directories. Verified with a repo-wide grep excluding each package's own path.

## Recommended Action

Delete both packages and their tests. If either is meant as the intended home
for the I/O now living in the providers, say so in a doc comment and open a
separate issue for the move.

## Acceptance Criteria

- [ ] Packages removed, or their intended use recorded
- [ ] `go build ./...` and the full test run stay clean
