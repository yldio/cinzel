---
status: complete
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

- [x] Packages removed, or their intended use recorded
- [x] `go build ./...` and the full test run stay clean

## Technical Details

Deleted. 413 lines across 8 files.

The question the todo raises — whether either is the intended home for the I/O
now in the providers — is answered by the history rather than by intent. Both
packages lost their last importer in `a60fdb4`, "feat: refactor to introduce
providers", which is where the work moved out rather than a point where it was
meant to move back. What replaced them is live and doing the job:
`provider/github/io_helpers.go`, `provider/gitlab/gitlab.go` and
`internal/fsutil`.

`filewriter.Writer.Do` was a wrapper around `os.Create` plus a write, which is
`os.WriteFile`. `filereader.Reader` was a generic extension-filtered directory
walk whose equivalent already exists per provider.

## Work Log

2026-09-16

Confirmed the todo's claim independently: a repo-wide grep for both package
names outside their own directories returns nothing in Go source, and nothing
outside a CHANGELOG line recording an old refactor.

Traced when they went dead with `git log -S` on the import path, which named
`a60fdb4` — the providers refactor. That settles the "is this the intended
home" question against keeping them.

Checked what took over before deleting, so the removal is not leaving a gap:
`filepath.Ext` and `os.ReadDir` discovery live in the two providers' own I/O
helpers.

`go build ./...`, `go test -count=1 ./...`, `mise run lint`, `mise run drift`
and `mise run license-check` all clean afterwards. `go mod tidy` moved nothing,
so neither package was holding a dependency on its own.
