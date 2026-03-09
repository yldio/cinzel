---
title: "Provider strict HCL schema parity via tagged decode"
module: "Providers"
problem_type: "logic_error"
component: "provider/github, provider/gitlab"
severity: "high"
root_cause: "schema validation logic drifted from parser contracts when allowed-key maps and remain-body parsing were maintained separately"
symptoms:
  - "strictness behavior differed by provider or scope"
  - "new fields required duplicated updates in parser + schema maps"
  - "harder to guarantee unknown-field rejection parity"
tags:
  - "providers"
  - "hcl"
  - "schema"
  - "strict-validation"
  - "github"
  - "gitlab"
created_date: "2026-03-09"
updated_date: "2026-03-09"
---

## Problem Description

The repository had reached a state where schema strictness and parse behavior could drift. GitLab moved toward typed parse structs, while GitHub still validated HCL using manual allowed-attribute/block maps.

The goal was to align providers on one principle: strictness should come from HCL-tagged decode structs.

## Root Cause

- Manual schema maps were a second source of truth.
- `hcl:",remain"`-first parsing obscured explicit schema contracts.
- Providers evolved at different speeds, increasing parity risk.

## Solution Implemented

- GitLab plan/status cleaned and updated to reflect completed phases and implemented features:
  - `docs/plans/2026-03-09-feat-gitlab-pipelines-provider-plan.md`
- GitHub parse schema moved to typed decode via `provider/github/config.go` + typed parser code, removing parse-side schema maps entirely.
- GitHub no longer keeps a dedicated schema-validation file for parse contracts; parse strictness comes from typed decode during config parsing.
- GitHub unparse schema checks migrated from manual `allowed*Keys` tables to strict typed YAML decode (`goccy/go-yaml` strict mode) in `provider/github/validate.go`.
- Native decoder diagnostics are now the expected source of unknown-key errors; fixtures/tests assert stable substrings from those diagnostics.
- GitLab typed parse + default/services work was documented:
  - `docs/solutions/best-practices/gitlab-typed-parse-schema-default-services.md`

## Validation Evidence

All tests passed after the change:

- `go test ./provider/github`
- `go test ./...`

## Prevention Guidance

- Keep schema contracts in HCL-tagged structs; avoid parallel allowed-key maps.
- Prefer strict typed YAML decode for unparse validation instead of manual key allowlists.
- Use labeled-block structs where labels are part of schema (`on <event>`, `service <id>`, `input <id>`, `output <id>`).
- Treat `hcl:",remain"` as an exception-only tool for intentional pass-through islands.
- Keep fixture-matrix invalid cases for unknown attrs/blocks and assert stable decoder substrings.

## Related References

- `docs/brainstorms/2026-03-09-gitlab-hcl-tag-strict-schema.md`
- `docs/plans/2026-03-09-feat-gitlab-pipelines-provider-plan.md`
- `docs/solutions/best-practices/gitlab-typed-parse-schema-default-services.md`
- `docs/solutions/logic-errors/github-strict-schema-parse-unparse-parity-unknown-rejection-stable-mapping.md`
