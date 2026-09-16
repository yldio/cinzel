---
status: pending
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

- [ ] A parallel-matrix need roundtrips
