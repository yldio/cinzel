---
status: complete
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

- [x] An unknown workflow key is reported rather than dropped
- [x] A top-level service with no name is refused

## Technical Details

Two unrelated changes in the same area.

`provider/gitlab/unparse_pipeline.go` walks the sorted workflow keys before
writing the block and warns on stderr for anything outside `name`,
`auto_cancel` and `rules`, matching the wording and the channel the top-level
loop already uses. The key is still dropped — `hclWorkflowBlock`
(`provider/gitlab/config.go:133`) has nowhere to put it — but it is no longer
dropped silently.

`provider/gitlab/validate.go` runs `validateServices` over the top-level
`services` too. `hclConfig` accepts that key alongside the one under `default`,
and only the default was checked, so an object with no `name` reached the YAML.
The owner label is "pipeline", so the error reads
`pipeline service entries must include 'name'`.

## Work Log

### 2026-09-16

Reproduced both through the CLI. A pipeline with `workflow.something_else`
unparsed with exit 0 and no mention of the key; `services = [{ alias = "db" }]`
parsed to a YAML document holding a nameless service.

Tests in `provider/gitlab/dropped_input_test.go`:
`TestUnknownWorkflowKeyIsReported` captures stderr around Unparse (2 cases, one
asserting the three known keys stay quiet) and `TestTopLevelServicesAreValidated`
drives Parse over inline HCL (3 cases). Both were confirmed to fail with the
fixes removed.

`go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check` all pass.
