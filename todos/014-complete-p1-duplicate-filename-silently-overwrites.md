---
status: complete
priority: p1
issue_id: "014"
tags: [code-review, correctness]
dependencies: ["013"]
---

# Two definitions sharing a filename overwrite one another

## Problem Statement

`filename` is not required to be unique. Two workflows naming the same file
produce one file, holding whichever came last. The other is gone and nothing
reports it.

## Findings

Probed at a786403 through the built CLI.

```
workflow "first"  { filename = "shared" ... }
workflow "second" { filename = "shared" ... }

$ cinzel github parse --file in/w.hcl --output-directory out
RC=0
out/shared.yaml     <- name: Second. "first" was never written.
```

The same holds for two actions, and across two HCL files in one directory,
where the file read second wins. Which one that is depends on walk order, so
the surviving definition is not something the author chose.

Filenames differing only in case collide on macOS and Windows but not on
Linux, so the same input produces different output depending on where it
ran.

Two shapes were probed and are not affected:

- A workflow and an action may share a name. They are written to
  `<name>.yaml` and `<name>/action.yml`, which are different paths.
- Duplicate block labels are already refused for steps, and two workflows
  with the same label but different filenames write both files, which is the
  existing behaviour.

## Recommended Action

Record each filename as it is claimed and refuse the second. Compare on the
cleaned, case-folded name so the answer does not change with the filesystem.
Name both definitions in the error, since the point is to say which two to
go and fix.

## Known gap

A workflow with `--yml` and `filename = "x/action"` writes `x/action.yml`,
which is exactly where an action named `x` writes. The two are claimed in
separate maps, so this is not caught. It needs the output extension plumbed
into the parser, which the parser does not otherwise see. Left alone rather
than widened for one shape that requires a flag and a deliberate name.

## Technical Details

- `provider/github/io_helpers.go`, `claimFilename`
- called from `parseHCLToWorkflows` and `parseHCLActions`

## Acceptance Criteria

- [x] Two workflows, and two actions, sharing a filename are refused
- [x] A case-only difference is refused
- [x] The same file reached by different paths is refused
- [x] Distinct filenames still work, and a workflow and an action may share a
      name

## Work Log

- 2026-09-15: Created and closed. Found while probing input-side file
  boundaries after #63. The cross-category collision is recorded above rather
  than fixed.
