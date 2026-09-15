---
status: complete
priority: p2
issue_id: "015"
tags: [code-review, correctness]
dependencies: ["014"]
---

# A renamed workflow in a subdirectory is never removed

## Problem Statement

`PruneStaleGeneratedYAML` read the top level of the output directory and
skipped every directory in it. A `filename` may name a subdirectory, so the
file a rename leaves behind sits where the prune cannot see it, and stays
there for good.

## Findings

Probed at 1066b8f. A workflow with `filename = "sub/kept"` was parsed,
renamed to `sub/renamed`, and parsed again:

```
out/sub/kept.yaml
out/sub/renamed.yaml
```

Both carry the cinzel marker. The same rename at the top level prunes
correctly, which is what makes the subdirectory the difference rather than
the rename.

Two neighbouring shapes were probed at the same time:

- Symlinked directories. `filepath.WalkDir` does not follow them, so
  widening the walk does not reach outside the output directory through one.
- An emitted `action.yml`. It carries no marker at all: only workflows go
  through `PrependGeneratedMarker`. A stale action directory therefore
  survives the wider walk too, because the prune only removes what it owns.
  Recorded here, not fixed: marking actions changes what every existing
  action output looks like, which is its own intent.

## Recommended Action

Walk the whole tree with `filepath.WalkDir`. The ownership check already
protects hand-written files, and the existing absolute-path check still
holds every removal inside the output directory.

The caller has to record every file it writes, actions included. A wider
walk reaches live action files that the old walk never saw, and anything
missing from the current set reads as stale. Actions are unmarked today so
nothing is at risk yet; the entry is what keeps it that way once they are
marked.

## Technical Details

- `internal/fsutil/generated_markers.go`, `PruneStaleGeneratedYAML`
- `provider/github/github.go`, `Parse`

The steps-only early return keeps its single-entry current set. It already
claimed the whole output directory, and the wider walk extends that claim to
subdirectories. Writing steps into a directory that also holds generated
workflows was already destructive; it is now destructive one level deeper.

## Acceptance Criteria

- [x] A stale generated file in a subdirectory is removed
- [x] A hand-written file, and one owned by another provider, are left alone
- [x] A missing output directory is not an error

## Work Log

- 2026-09-15: Created and closed. Found while checking what the filename
  subdirectory allowed by 013 reaches.
