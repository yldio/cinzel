---
title: "GitHub strict schema parity across parse and unparse"
module: "GitHub Provider"
problem_type: "logic_error"
component: "provider/github schema validation"
severity: "high"
root_cause: "inconsistent strictness between manual HCL parsing and non-typed unparse validation"
symptoms:
  - "Unknown HCL attributes or blocks were accepted in some parse paths"
  - "Unknown YAML keys could pass through unparse paths"
  - "Strictness changes risked breaking depends_on/needs mapping or references"
tags:
  - "schema"
  - "parse"
  - "unparse"
  - "validation"
  - "depends_on"
  - "needs"
created_date: "2026-03-09"
updated_date: "2026-03-09"
---

## Problem Description

GitHub provider strictness was uneven: some parse/unparse paths enforced unknown-field rejection, while others were permissive due to manual body handling and generic YAML processing.

Goal: make strict schema behavior symmetric in both directions without regressing:

- dependency mapping (`depends_on` in HCL <-> `needs:` in YAML)
- valid reference/parameter semantics (`job.*`, `step.*`, expression strings)

## Root Cause

- HCL parse used manual `remain`-style processing in places, without a single shared strict schema gate.
- YAML unparse validation was not contract-driven by strict typed decode across workflow/job/step/action scopes.
- Strictness was not centralized, so behavior drifted by code path.

## Solution Implemented

1. Enforced parse-side strictness via typed HCL decode contracts in:
   - `provider/github/parse_workflow.go`
   - `provider/github/parse_action.go`
   - `provider/github/config.go`
2. Enforced unparse-side strictness via strict typed YAML decode in:
   - `provider/github/validate.go`
   - `provider/github/unparse_action.go`
3. Preserved mapping and semantics:
   - HCL `depends_on` -> YAML `needs`
   - YAML `needs` -> HCL `depends_on`
   - HCL `needs` rejected as unknown attribute

## Verification

Key updated tests and fixtures:

- `provider/github/validation_test.go`
  - unknown workflow/job/action attrs/blocks
  - unknown YAML workflow/job/step/action keys
- `provider/github/github_test.go`
  - explicit mapping checks both directions (`depends_on` <-> `needs`)
- `provider/github/testdata/fixtures/matrix/parse/invalid/legacy_needs_attribute.hcl`
- `provider/github/testdata/fixtures/matrix/parse/invalid/legacy_needs_attribute.error.txt`
- `provider/github/testdata/fixtures/matrix/parse/invalid/unknown_job_block.hcl`
- `provider/github/testdata/fixtures/matrix/parse/invalid/unknown_job_block.error.txt`

Validation commands passed:

- `go test ./provider/github`
- `go test ./...`

Implementation note: this document originally referenced `provider/github/schema_validation.go`; current implementation removed that file in favor of typed parse contracts and strict YAML decode.

## Prevention Guidance

- Use typed decode contracts as the schema source of truth; avoid allowlist schema tables.
- Keep `hcl:",remain"` only for intentional pass-through islands.
- For every new accepted key/block, add:
  - one valid case
  - one nearby typo/unknown invalid case
- Keep parse/unparse parity tests for the same conceptual field surface.
- Preserve deterministic error wording and ordering for fixture stability.

## Related References

- `docs/brainstorms/2026-03-09-github-strict-block-schema-brainstorm.md`
- `docs/plans/2026-03-09-feat-github-strict-block-schema-enforcement-plan.md`
- `docs/plans/2026-03-09-feat-rename-github-needs-to-depends-on-plan.md`
- `docs/solutions/logic-errors/github-job-parser-strict-unknown-attributes-depends-on-mapping.md`
