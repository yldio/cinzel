---
title: "feat: Enforce strict GitHub block schemas"
type: feat
status: completed
date: 2026-03-09
origin: docs/brainstorms/2026-03-09-github-strict-block-schema-brainstorm.md
---

# feat: Enforce strict GitHub block schemas

## Overview

Enforce strict, typed schema validation for GitHub provider blocks so only explicitly defined properties/blocks are accepted, while preserving valid reference semantics and provider mapping behavior (`depends_on` in HCL <-> `needs:` in YAML) (see brainstorm: docs/brainstorms/2026-03-09-github-strict-block-schema-brainstorm.md).

## Problem Statement / Motivation

Current GitHub provider parsing still has `remain`/manual paths that can be less strict than typed schema decoding, especially outside already-guarded job attribute paths. This creates risk of silent acceptance for unknown keys/blocks and contract drift between parse/unparse. The brainstorm explicitly selected immediate strict mode, generic unknown-field errors, all block scope coverage, and preserved resource/parameter reference behavior (see brainstorm: docs/brainstorms/2026-03-09-github-strict-block-schema-brainstorm.md).

## Research Decision

Proceeding without external research. This is an internal parser-contract change with strong repository context, clear conventions, and existing fixture matrix patterns.

## Consolidated Research Findings

- Existing strictness baseline already exists for some job attributes: `provider/github/parse_workflow.go`.
- Remaining permissive paths exist in generic block handling and action parsing: `provider/github/parse_workflow.go`, `provider/github/parse_action.go`.
- Unparse currently relies more on semantic validation and generic writers than explicit schema allowlists: `provider/github/unparse_emit.go`, `provider/github/unparse_workflow.go`.
- Dependency mapping contract to preserve:
  - HCL parse `depends_on` -> YAML `needs` (`provider/github/parse_workflow.go`)
  - YAML unparse `needs` -> HCL `depends_on` (`provider/github/unparse_emit.go`)
- Existing institutional learnings to preserve:
  - deterministic map iteration and stable error assertions: `docs/solutions/logic-errors/nondeterministic-map-iteration.md`
  - strict parse/unparse detection chain discipline: `docs/solutions/best-practices/document-type-auto-detection-chain.md`
  - known parsing strategy tradeoffs: `docs/solutions/best-practices/generic-attribute-parsing-for-action-types.md`
  - critical patterns for conversion stability: `docs/solutions/patterns/critical-patterns.md`

## Proposed Solution

Adopt a central typed schema registry (chosen in brainstorm) and enforce it consistently in parse and unparse contracts.

### Scope and Contract (carried from brainstorm)

- All GitHub provider block families are in-scope.
- Unknown attributes and unknown block types are rejected.
- Error style is generic unknown-field/path-oriented errors.
- No compatibility mode or migration flag in this phase.
- Valid references and parameterized expressions remain supported (`job.*`, `step.*`, `variable.*`, expression strings) (see brainstorm: docs/brainstorms/2026-03-09-github-strict-block-schema-brainstorm.md).

### Schema boundary rule

- Paths explicitly documented as free-form containers remain free-form (for example `env`, `with`, and matrix axis maps where applicable).
- Everything else follows explicit typed allowlists.

## Technical Approach

### Phase 1: Schema Baseline

- Define block-level schema declarations for workflow/job/step/action parse paths.
- Align existing manual parse paths to schema checks before conversion.
- Normalize unknown-field errors to deterministic path format.

Planned touchpoints:
- `provider/github/config.go`
- `provider/github/parse_workflow.go`
- `provider/github/parse_action.go`
- `provider/github/errors.go`

### Phase 2: Unparse Parity

- Enforce equivalent schema boundaries during unparse input validation (YAML side), matching parse strictness intent.
- Keep `depends_on` <-> `needs` mapping stable and explicit.
- Preserve existing document classification flow and apply strictness post-classification.

Planned touchpoints:
- `provider/github/unparse_workflow.go`
- `provider/github/unparse_action.go`
- `provider/github/unparse_emit.go`
- `provider/github/validate.go`

### Phase 3: Tests and Docs

- Expand fixture matrix invalid cases for unknown attrs and unknown blocks by major block type.
- Add parity tests for parse/unparse unknown handling and mapping invariants.
- Add reference-preservation regression coverage for `job.*`, `step.*`, `variable.*`, and expressions.
- Update provider docs with strict schema and terminology distinction.

Planned touchpoints:
- `provider/github/validation_test.go`
- `provider/github/github_test.go`
- `provider/github/fixture_matrix_test.go`
- `provider/github/testdata/fixtures/matrix/parse/invalid/*`
- `provider/github/testdata/fixtures/matrix/unparse/invalid/*`
- `provider/github/README.md`

## System-Wide Impact

- **Interaction graph**: parse/unparse entrypoints -> classify document type -> strict schema gate -> semantic validation -> output conversion.
- **Error propagation**: unknown fields should fail early with stable path-based errors before deep conversion.
- **State lifecycle risks**: none (no persistent state/migrations), but high regression risk on accepted syntax surface.
- **API surface parity**: parse and unparse must enforce equivalent schema intent for corresponding block types.
- **Integration test scenarios**:
  - Unknown attr on each major block type fails with deterministic error path.
  - Unknown nested block fails in both directions where block schemas apply.
  - `depends_on` in HCL still maps to YAML `needs:`.
  - YAML `needs:` still emits HCL `depends_on`.
  - Reference and expression cases remain semantically stable after roundtrip.

## SpecFlow Analysis Incorporated

From spec-flow output, this plan explicitly addresses:

- Schema boundary definition (strict vs free-form paths).
- Deterministic error contract and testability.
- Dependency mapping normalization expectations.
- Classification/strictness ordering.
- Reference preservation invariants.

## Open Questions (resolved)

- Unknown-field errors are fail-fast (first violation), consistent with current validation pipeline behavior.
- Canonical path format for strict-schema errors follows existing path-first style (for example `jobs.build`, `jobs.build.steps[0]`, `workflow_yaml`) with stable unknown-field wording.
- Nested free-form policy in v1: only explicitly designated container paths are free-form (for example known value maps like `env`/`with`/matrix axis values); all other keys/blocks are strict by schema.

## Acceptance Criteria

### Functional

- [x] All GitHub provider block families enforce strict typed field/block schemas.
- [x] Unknown attributes are rejected across scoped parse/unparse contracts.
- [x] Unknown block types are rejected across scoped parse/unparse contracts.
- [x] `depends_on` remains the only HCL dependency key; mapping to YAML `needs:` is preserved.
- [x] YAML `needs:` unparse output remains HCL `depends_on`.
- [x] Valid resource/parameter reference patterns continue to work.

### Testing

- [x] Fixture matrix has invalid unknown-attribute and unknown-block cases for workflow/job/step/action-relevant paths.
- [x] Validation tests assert deterministic path-based error messages.
- [x] Roundtrip tests cover reference and expression preservation.
- [x] Existing golden and compatibility suites pass with intentional fixture updates only.

### Documentation

- [x] Provider docs explicitly state strict schema behavior.
- [x] Provider docs preserve terminology distinction (`depends_on` in HCL, `needs:` in YAML).

## Success Metrics

- Zero permissive unknown-field acceptance in covered block scopes.
- No regression in reference/mapping roundtrip behavior.
- Deterministic test outcomes for error fixtures.

## Dependencies & Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Over-strict validation breaks legitimate free-form maps | High | Define/lock free-form path allowlist in schema boundary rules |
| Parse/unparse drift in strictness behavior | Medium | Shared schema registry and parity tests |
| Error-message churn causing flaky fixtures | Medium | Stable path format + deterministic order |
| Reference handling regressions | High | Dedicated reference and expression roundtrip tests |

## Sources & References

### Origin

- **Brainstorm document:** [docs/brainstorms/2026-03-09-github-strict-block-schema-brainstorm.md](docs/brainstorms/2026-03-09-github-strict-block-schema-brainstorm.md) — central schema registry, strict mode scope, generic error style, and reference-preservation boundary.

### Internal References

- Parse workflow/manual attribute handling: `provider/github/parse_workflow.go`
- Action parse permissive paths: `provider/github/parse_action.go`
- Unparse generic key emission: `provider/github/unparse_emit.go`
- Validation pipeline: `provider/github/validate.go`
- Provider schema entry structs: `provider/github/config.go`
- Provider docs baseline: `provider/github/README.md`

### Institutional Learnings

- `docs/solutions/logic-errors/github-job-parser-strict-unknown-attributes-depends-on-mapping.md`
- `docs/solutions/logic-errors/config-input-precedence-ignored-for-parse.md`
- `docs/solutions/logic-errors/nondeterministic-map-iteration.md`
- `docs/solutions/best-practices/document-type-auto-detection-chain.md`
- `docs/solutions/best-practices/generic-attribute-parsing-for-action-types.md`
- `docs/solutions/patterns/critical-patterns.md`
