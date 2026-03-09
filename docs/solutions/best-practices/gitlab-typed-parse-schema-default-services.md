---
title: "GitLab typed parse schema with default/services support"
module: "GitLab Provider"
problem_type: "logic_error"
component: "provider/gitlab"
severity: "high"
root_cause: "parser strictness and behavior were split across generic remain-body traversal and ad-hoc schema checks instead of a single tagged HCL contract"
symptoms:
  - "`hcl:\",remain\"` dominated parse config, making strictness indirect"
  - "Schema intent drift risk between config, parser logic, and tests"
  - "New features (`default`, `service`) required repetitive generic parsing paths"
tags:
  - "gitlab"
  - "hcl"
  - "schema"
  - "parse"
  - "services"
  - "default"
created_date: "2026-03-09"
updated_date: "2026-03-09"
---

## Problem Description

GitLab parsing worked, but schema strictness was not encoded primarily in HCL-tagged structs. The implementation leaned on generic `hcl.Body` traversal and separate validation logic, which made the contract harder to reason about and easier to drift over time.

At the same time, `default` and repeated `service` blocks were being added, increasing the cost of generic parsing and custom checks.

## Root Cause

- Top-level and nested parse shapes were not fully represented as typed HCL decode structs.
- Strictness relied on side validation instead of decode contracts.
- Conversion logic for refs (`depends_on`, `extends`) and optional lists needed special handling that was intertwined with generic parsing.

## Solution Implemented

Refactored parse schema to be HCL-tag-driven and typed, then layered conversion logic on top.

- Typed config/schema structs added for job/template/workflow/default/include/rule/artifacts/cache/service:
  - `provider/gitlab/config.go`
- Parse path rewritten to decode typed blocks and build YAML maps from explicit fields:
  - `provider/gitlab/parse_pipeline.go`
- `default` and repeated `service {}` blocks implemented as first-class parse/unparse/validation features:
  - `provider/gitlab/parse_pipeline.go`
  - `provider/gitlab/unparse_pipeline.go`
  - `provider/gitlab/validate.go`
- README examples and notes updated:
  - `provider/gitlab/README.md`

## Notable Investigation Detail

During refactor, parse failures surfaced as `expected job references` for jobs without `depends_on`.

Cause:
- typed decode provided a non-nil empty expression for optional traversal fields.

Fix:
- guard traversal parsing for nil/null/empty collection expressions before reference resolution.

This stabilized typed parsing without requiring users to set `depends_on` on every job.

## Behavior Locked by Tests

- Existing GitLab fixture matrix and golden tests remained green after refactor.
- Added valid/invalid fixture coverage for `default` + `service` behavior:
  - `provider/gitlab/testdata/fixtures/pipelines/default_services.hcl`
  - `provider/gitlab/testdata/fixtures/pipelines/default_services.golden.yaml`
  - `provider/gitlab/testdata/fixtures/matrix/parse/valid/default_services.hcl`
  - `provider/gitlab/testdata/fixtures/matrix/parse/valid/default_services.golden.yaml`
  - `provider/gitlab/testdata/fixtures/matrix/unparse/valid/default_services.yaml`
  - `provider/gitlab/testdata/fixtures/matrix/unparse/valid/default_services.roundtrip.golden.yaml`
  - `provider/gitlab/testdata/fixtures/matrix/parse/invalid/service_missing_name.hcl`
  - `provider/gitlab/testdata/fixtures/matrix/parse/invalid/service_missing_name.error.txt`
  - `provider/gitlab/testdata/fixtures/matrix/unparse/invalid/services_wrong_type.yaml`
  - `provider/gitlab/testdata/fixtures/matrix/unparse/invalid/services_wrong_type.error.txt`

Verification commands:

- `go test ./provider/gitlab`
- `go test ./...`

## Prevention Guidance

- Keep HCL tags as the primary schema contract; avoid duplicated schema maps unless temporary.
- Reserve `hcl:\",remain\"` only for intentional pass-through islands.
- Add a valid + invalid fixture pair for each new block/attribute.
- Guard optional traversal expressions (`depends_on`, `extends`) against empty decoded expressions.
- Keep roundtrip semantic assertions for every newly typed block.

## Related References

- `docs/brainstorms/2026-03-09-gitlab-hcl-tag-strict-schema.md`
- `docs/plans/2026-03-09-feat-gitlab-pipelines-provider-plan.md`
- `docs/solutions/best-practices/gitlab-provider-baseline-parse-unparse-validation-fixture-matrix.md`
- `docs/solutions/logic-errors/github-strict-schema-parse-unparse-parity-unknown-rejection-stable-mapping.md`
