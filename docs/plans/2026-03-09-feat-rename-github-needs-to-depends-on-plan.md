---
title: "feat: Rename GitHub HCL needs to depends_on"
type: feat
status: active
date: 2026-03-09
origin: docs/brainstorms/gitlab-provider.md
---

# feat: Rename GitHub HCL needs to depends_on

## Overview

Rename the GitHub provider's HCL dependency attribute from `needs` to `depends_on` for consistency with the brainstorm decision and Terraform-style HCL conventions (see brainstorm: docs/brainstorms/gitlab-provider.md).

## Problem Statement / Motivation

The brainstorm standardized dependency references as `depends_on` across providers, but GitHub HCL still uses `needs` in existing schema/tests. Keeping mixed terms increases cognitive overhead and weakens provider consistency.

## Proposed Solution

Adopt `depends_on` as the only supported GitHub HCL dependency attribute.

### Strict policy

- Parse direction (HCL -> YAML):
  - Accept only `depends_on`.
  - Reject legacy `needs` with a clear migration error.
- Unparse direction (YAML -> HCL):
  - Always emit `depends_on`.
- YAML output/input remains `needs:` (GitHub Actions schema).

## Technical Considerations

- Update GitHub job config schema and decode logic where dependency attributes are parsed.
- Update unparse emitters to output `depends_on` in generated HCL.
- Maintain deterministic ordering and existing formatting conventions.
- Keep sentinel errors in package-local `errors.go` files.

## System-Wide Impact

- **Interaction graph**: HCL job attribute parsing -> job model -> YAML marshalling (`needs:`) and reverse unparse path.
- **Error propagation**: conflicting attributes (`needs` + `depends_on`) should fail at validation with actionable messaging.
- **API surface parity**: align naming with GitLab plan and future cross-provider docs.
- **Integration scenarios**:
  - Existing fixtures with `needs` still parse correctly.
  - Unparse of YAML `needs` emits HCL `depends_on`.
  - Roundtrip remains semantically stable.

## Acceptance Criteria

- [ ] GitHub HCL parse accepts `depends_on` and maps to YAML `needs`.
- [ ] Legacy `needs` is rejected in parse with a clear migration error.
- [ ] Defining both `needs` and `depends_on` in one job returns a clear error.
- [ ] GitHub unparse emits `depends_on` (not `needs`) in HCL output.
- [ ] Golden tests and roundtrip tests are updated and passing.
- [ ] Fixture matrix includes valid/invalid cases for aliasing and conflicts.
- [ ] Provider README/docs reflect canonical `depends_on` usage.
- [ ] Documentation explicitly distinguishes HCL `depends_on` from GitHub YAML `needs:` and includes a one-block before/after migration example.

## Implementation Phases

### Phase 1: Schema and Validation

- Add `depends_on` to GitHub job schema.
- Remove legacy `needs` from supported HCL schema.
- Add explicit validation errors for `needs` usage and for (`needs` + `depends_on`) conflicts.

### Phase 2: Parse/Unparse Behavior

- Parse both aliases to one internal dependency representation.
- Ensure unparse writes only `depends_on`.

### Phase 3: Tests and Docs

- Update/expand fixtures and matrix tests.
- Add migration note to docs showing old/new syntax.
- Add explicit "Terminology distinction" subsection in provider docs:
  - HCL attribute: `depends_on`
  - GitHub YAML key: `needs`
  - Mapping rule: `depends_on` <-> `needs:`

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaks existing HCL using `needs` | Medium | Return explicit migration error with old/new syntax example |
| Ambiguous behavior when both are set | Medium | Explicit validation error |
| Test churn from expected output changes | Low | Update golden fixtures and roundtrip assertions together |

## Sources & References

- **Origin brainstorm:** [docs/brainstorms/gitlab-provider.md](docs/brainstorms/gitlab-provider.md) — decision to use `depends_on` in HCL.
- **Related plan:** [docs/plans/2026-03-09-feat-gitlab-pipelines-provider-plan.md](docs/plans/2026-03-09-feat-gitlab-pipelines-provider-plan.md)
- **Configuration foundation plan:** [docs/plans/2026-03-09-feat-cinzelrc-provider-config-precedence-plan.md](docs/plans/2026-03-09-feat-cinzelrc-provider-config-precedence-plan.md)
