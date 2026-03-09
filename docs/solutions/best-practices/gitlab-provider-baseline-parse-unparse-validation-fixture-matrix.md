---
title: "GitLab provider baseline parse/unparse with validation fixture matrix"
module: "GitLab Provider"
problem_type: "logic_error"
component: "provider/gitlab"
severity: "high"
root_cause: "missing baseline provider parity and validation guards across parse/unparse directions"
symptoms:
  - "GitLab provider behavior drift risk between HCL->YAML and YAML->HCL"
  - "Invalid stage/reference types not explicitly locked by tests"
  - "No fixture matrix to prevent regressions in mapping and error behavior"
tags:
  - "gitlab"
  - "parse"
  - "unparse"
  - "validation"
  - "fixtures"
created_date: "2026-03-09"
updated_date: "2026-03-09"
---

## Problem Description

The repository needed a real second provider implementation to validate multi-provider architecture. GitLab CI has a different model than GitHub, so the baseline had to prove both conversion directions, strict-enough validation, deterministic output, and mapping invariants.

A concrete failure mode highlighted during implementation was invalid stage definitions such as:

```hcl
stages = [job.build, job.test]
```

This must fail because `stages` must be literal stage names, not job references.

## Root Cause

- No production GitLab provider existed yet.
- Parse and unparse logic had no baseline parity harness.
- Validation contracts (type constraints, needs graph rules, stage references) were not locked via matrix fixtures.

## Solution Implemented

Implemented a baseline GitLab provider with both directions plus validation and fixture coverage:

- Provider wiring and CLI registration:
  - `cinzel.go`
  - `provider/gitlab/gitlab.go`
- Parse direction (HCL -> `.gitlab-ci.yml`):
  - `provider/gitlab/parse_pipeline.go`
  - `provider/gitlab/pipeline_yaml.go`
  - `provider/gitlab/validate.go`
- Unparse direction (YAML -> HCL):
  - `provider/gitlab/unparse_pipeline.go`
  - `provider/gitlab/unparse_emit.go`
- Shared conversion helpers:
  - `provider/gitlab/conversion.go`
- Fixture matrix and targeted tests:
  - `provider/gitlab/fixture_matrix_test.go`
  - `provider/gitlab/gitlab_test.go`
  - `provider/gitlab/testdata/fixtures/matrix/...`

## Behavior Locked by Tests

- Parse/unparse command help and provider smoke wiring.
- Parse output defaults and dry-run output path.
- `${VAR}` / `$${VAR}` handling across directions.
- `depends_on` <-> `needs` mapping.
- Hidden jobs to `template` pass-through during unparse.
- Invalid inputs with expected errors (`.error.txt`) including stage/reference typing.

## Validation Evidence

Implemented in commit `16b3a95` (`feat: add gitlab provider parse and unparse baseline`).

Test commands passing:

- `go test ./provider/gitlab`
- `go test ./...`

## Prevention Guidance

- Keep parse and unparse contracts symmetric with explicit tests for both directions.
- Add one valid and one invalid fixture for each new feature field.
- Preserve deterministic ordering and semantic YAML comparison in matrix tests.
- Treat list typing strictly for `stages`; reject traversals/mixed types.
- Keep mapping invariants (`depends_on` <-> `needs`) under dedicated tests to avoid drift.

## Related References

- `docs/plans/2026-03-09-feat-gitlab-pipelines-provider-plan.md`
- `docs/brainstorms/2026-03-09-gitlab-provider.md`
- `docs/plans/2026-03-09-feat-github-strict-block-schema-enforcement-plan.md`
- `docs/solutions/logic-errors/github-strict-schema-parse-unparse-parity-unknown-rejection-stable-mapping.md`
