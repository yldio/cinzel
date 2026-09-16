---
status: pending
priority: p3
issue_id: "050"
tags: [code-review, gitlab, unparse]
dependencies: []
---

# A top-level key that is not an identifier is written as-is

## Problem Statement

The passthrough branch writes the key unchanged, so `my weird key` becomes
`my weird key = "v"`, which is not valid HCL. Only reachable through the path
that already warns about unsupported keys.

## Findings

`provider/gitlab/unparse_pipeline.go:509`.

## Recommended Action

When `naming.SanitizeIdentifier` changes the key, fail rather than warn — the
file produced cannot be read back either way, and an error says so at the point
it happens.

## Acceptance Criteria

- [ ] A non-identifier top-level key is refused with a clear message
