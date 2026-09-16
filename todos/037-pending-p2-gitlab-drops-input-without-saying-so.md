---
status: pending
priority: p2
issue_id: "037"
tags: [code-review, validation, gitlab]
dependencies: []
---

# Two places accept input and throw part of it away

## Problem Statement

The workflow unparse copies only `name`, `auto_cancel` and `rules` out of the
workflow map; any other key is discarded with no warning, although the
top-level loop warns about unsupported keys.

Separately, `validateServices` runs for each job and for `default` but not for
the top-level `services` that `parseConfig` accepts, so a service with no
`name` passes.

## Findings

- `provider/gitlab/unparse_pipeline.go:286-330`; the top-level warn is at `:471`
- `provider/gitlab/validate.go:85-97`; the top-level key is accepted at
  `provider/gitlab/config.go:232`

## Recommended Action

Iterate the sorted workflow keys and report anything outside the known three.
Add the top-level `services` to the validation alongside `default`.

## Acceptance Criteria

- [ ] An unknown workflow key is reported rather than dropped
- [ ] A top-level service with no name is refused
