---
title: "GitHub job parser strict unknown attributes with depends_on mapping"
module: "GitHub Provider"
problem_type: "logic_error"
component: "provider/github parse_workflow"
severity: "high"
root_cause: "manual remain-body parsing without unknown-attribute enforcement"
symptoms:
  - "HCL `needs` accepted or handled inconsistently during rename to `depends_on`"
  - "Unknown job attributes (typos/custom keys) were not uniformly rejected"
  - "Strict HCL contract was not guaranteed while mapping to YAML `needs:`"
tags:
  - "hcl"
  - "validation"
  - "depends_on"
  - "needs"
  - "github-actions"
created_date: "2026-03-09"
updated_date: "2026-03-09"
---

## Problem

During the GitHub HCL rename from `needs` to `depends_on`, we needed strict behavior: HCL must support only `depends_on`, while GitHub YAML still uses `needs:`.

The parser uses `hcl:",remain"` plus manual attribute handling, so unknown fields are not automatically rejected. Without explicit validation, this can silently accept invalid keys or produce inconsistent behavior.

## Root Cause

`provider/github/parse_workflow.go` parsed job attributes manually but did not enforce an explicit allowlist for all supported job attributes before conversion.

Because of that, strict unknown-attribute rejection depended on ad-hoc handling instead of a schema-first rule.

## Solution

1. Added strict job attribute allowlist validation in `provider/github/parse_workflow.go`.
2. Kept canonical mapping behavior:
   - HCL parse: `depends_on` -> YAML `needs`
   - YAML unparse: `needs` -> HCL `depends_on`
3. Removed special-case migration/backward messages and relied on generic unknown-attribute failures.

Current behavior:

- `depends_on` in HCL is valid and maps correctly.
- `needs` in HCL fails as unknown attribute in job scope.
- Any other unsupported job key (for example `myprop`) fails the same way.

## Key Files

- `provider/github/parse_workflow.go`
- `provider/github/unparse_emit.go`
- `provider/github/job/parsed.go`
- `provider/github/validation_test.go`
- `provider/github/github_test.go`
- `provider/github/testdata/fixtures/matrix/parse/invalid/legacy_needs_attribute.hcl`
- `provider/github/testdata/fixtures/matrix/parse/invalid/legacy_needs_attribute.error.txt`

## Verification

- Added/updated validation tests for:
  - legacy `needs` rejected as unknown job attribute
  - unknown job attribute rejected (`myprop`)
  - dependency checks still enforced with `depends_on`
- Added mapping tests for both directions:
  - HCL `depends_on` -> YAML `needs:`
  - YAML `needs:` -> HCL `depends_on`
- Fixture matrix includes invalid legacy `needs` case with expected error text.
- Full suite passed:
  - `go test ./provider/github`
  - `go test ./...`

## Prevention

- For any parser using `hcl:",remain"`, always add explicit allowlists for attrs/blocks.
- Reject unrecognized fields generically instead of special-casing specific deprecated names.
- Add negative tests for both:
  - deprecated/renamed keys
  - arbitrary unknown keys
- Keep mapping tests that assert terminology translation across formats (`depends_on` <-> `needs:`).

## Related References

- `docs/plans/2026-03-09-feat-rename-github-needs-to-depends-on-plan.md`
- `docs/brainstorms/gitlab-provider.md`
- `docs/solutions/logic-errors/config-input-precedence-ignored-for-parse.md`
- `docs/solutions/patterns/critical-patterns.md`
