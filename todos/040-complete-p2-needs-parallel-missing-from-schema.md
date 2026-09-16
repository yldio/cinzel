---
status: complete
priority: p2
issue_id: "040"
tags: [code-review, schema, gitlab]
dependencies: []
---

# needs: parallel is written but cannot be read back

## Problem Statement

GitLab supports `needs: [{job: …, parallel: {matrix: …}}]`. Unparse writes
`parallel = {…}` inside a `need` block, but the schema has no such field, so
parse fails with `An argument named "parallel" is not expected here`.

## Findings

`provider/gitlab/config.go:42-49`, `hclNeedBlock`; the attribute list is at
`provider/gitlab/parse_pipeline.go:630`.

## Recommended Action

Add `Parallel hcl.Expression` with `hcl:"parallel,optional"` and the matching
entry in the attribute list.

## Acceptance Criteria

- [x] A parallel-matrix need roundtrips

## Technical Details

Two lines. `hclNeedBlock` (`provider/gitlab/config.go`) gained
`Parallel hcl.Expression` with `hcl:"parallel,optional"`, and `parseNeedBlocks`
(`provider/gitlab/parse_pipeline.go`) gained the matching entry in its attribute
list, between `optional` and `project`, so the value comes back through
`setOptionalAttr` like every other need option.

Nothing on the unparse side changed: it already wrote `parallel = {...}` into
the need block. Only the schema that reads it back was missing.

## Work Log

### 2026-09-16

Reproduced through the CLI: a pipeline whose `test` job needed
`{job: build, parallel: {matrix: [{ARCH: [amd64]}]}}` unparsed cleanly and then
failed to parse with `An argument named "parallel" is not expected here.`. After
the fix the same file goes YAML to HCL to YAML with the matrix intact.

`provider/gitlab/needs_object_test.go` gained a "need with a parallel matrix"
case, which like the rest of that table checks both the emitted HCL and the YAML
it parses back to. It was confirmed to fail with the two lines removed.

`go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check` all pass.
