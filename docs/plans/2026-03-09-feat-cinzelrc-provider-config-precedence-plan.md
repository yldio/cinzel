---
title: "feat: Add .cinzelrc.yaml config foundation"
type: feat
status: completed
date: 2026-03-09
origin: docs/brainstorms/gitlab-provider.md
---

# feat: Add .cinzelrc.yaml config foundation

## Overview

Introduce project-level `.cinzelrc.yaml` config with deterministic precedence before GitLab provider implementation. This gives provider-specific output defaults and a stable config path GitLab can use from day one (see brainstorm: docs/brainstorms/gitlab-provider.md).

## Problem Statement / Motivation

The brainstorm decided on `.cinzelrc.yaml` and precedence (`CLI > config > provider defaults`), but the current GitLab plan deferred it to v0.2. That sequencing risks rework and delays team-level output conventions. Implementing config first reduces churn and follows the original decision (see brainstorm: docs/brainstorms/gitlab-provider.md).

## Brainstorm Carry-Forward Matrix

This plan is config-first, but it preserves all brainstorm conclusions and scope boundaries:

1. **Provider-specific HCL** remains unchanged; `.cinzelrc.yaml` stores provider preferences only and does not introduce cross-provider HCL abstractions (see brainstorm: docs/brainstorms/gitlab-provider.md).
2. **Provider selection remains implicit from CLI** (`cinzel github ...`, `cinzel gitlab ...`); config does not add `ci = "..."` in HCL (see brainstorm: docs/brainstorms/gitlab-provider.md).
3. **`template` block decision** remains valid but out-of-scope for this plan; config work must not block future template mapping (see brainstorm: docs/brainstorms/gitlab-provider.md).
4. **`variable name` decision** remains provider-schema behavior and is unaffected; config does not redefine variable semantics (see brainstorm: docs/brainstorms/gitlab-provider.md).
5. **Repeated singular `rule` blocks** remain the target model for GitLab HCL and are not altered by config introduction (see brainstorm: docs/brainstorms/gitlab-provider.md).
6. **`depends_on` naming** remains the HCL dependency keyword; config is orthogonal to dependency graph semantics (see brainstorm: docs/brainstorms/gitlab-provider.md).
7. **`.cinzelrc.yaml` with precedence** is brought forward from deferred scope into immediate scope by this plan (see brainstorm: docs/brainstorms/gitlab-provider.md).

## Proposed Solution

Add a config-loading layer in CLI option resolution that computes effective `provider.ProviderOps` before `Parse`/`Unparse`.

### Config file shape (v1)

```yaml
github:
  parse:
    output-directory: .github/workflows
  unparse:
    output-directory: ./cinzel

gitlab:
  parse:
    output-directory: .
    single-file: true
    filename: .gitlab-ci.yml
  unparse:
    output-directory: ./cinzel
```

### Precedence

1. CLI flags (explicitly set flags)
2. `.cinzelrc.yaml` provider section
3. Provider defaults

Resolution rule for v1: precedence is based on flag presence (`cmd.IsSet(...)`), not zero values. A set flag always wins, including intentionally empty values.

### Discovery and validation rules

- Config file path is `.cinzelrc.yaml` in cwd for v1.
- Missing config file is non-error.
- Invalid YAML or wrong field types fail fast with actionable error messages.
- Unknown keys are ignored with warning in v1 (forward-compatible posture).
- Only active provider semantics are validated; malformed inactive provider sections are tolerated unless the YAML is structurally invalid.
- `output-directory` is command-scoped (`parse` vs `unparse`) and does not share a single provider-level key.

## Technical Considerations

- **CLI integration point**: `internal/command/command.go` currently maps flags directly via `toProviderOpts`; this is the insertion point for effective option resolution.
- **Provider contract stability**: `provider/provider.go` already has sufficient fields for `output-directory`; no interface change is required for v1.
- **Default behavior policy**: backward compatibility is not a hard requirement for this pre-production phase.
- **Future GitLab compatibility**: include command-scoped `single-file` and `filename` keys for `gitlab.parse` now, even if full GitLab provider is not merged yet.
- **Determinism**: map-derived warnings/output must use deterministic key ordering.

## System-Wide Impact

- **Interaction graph**: CLI command parses flags -> loads config (if present) -> resolves effective options -> provider Parse/Unparse reads effective ops -> fsutil writes outputs.
- **Error propagation**: config parse/validation errors should surface before provider execution, returning a single actionable error through existing CLI error handling.
- **State lifecycle risks**: no persistent state; main risk is wrong output path selection causing files in unexpected locations.
- **API surface parity**: both `parse` and `unparse` commands for each provider must use the same precedence resolver.
- **Integration test scenarios**:
  - Config absent -> command still runs with provider defaults.
  - Config present + no CLI override -> config values applied.
  - Config present + CLI override -> CLI wins.
  - CLI flag present but empty -> CLI still wins over config/defaults.
  - Invalid config -> deterministic failure before provider logic.
  - Mixed provider sections -> active provider works even if inactive section has semantic issues.

## SpecFlow Findings Applied

This plan incorporates these spec-flow gaps/edge cases:

- Explicit config discovery rule (cwd-scoped for v1) to avoid monorepo ambiguity.
- Clear treatment of explicit vs omitted CLI flags for precedence.
- Defined unknown-key behavior (warn, do not fail) and type mismatch behavior (fail).
- Clarified GitLab `single-file` policy for this phase (config key supported; effective behavior remains single-file oriented).
- Added cross-provider isolation acceptance criteria.

## Acceptance Criteria

### Functional

- [x] `.cinzelrc.yaml` is loaded when present in cwd.
- [x] Effective option precedence is `CLI > config > provider defaults`.
- [x] Missing `.cinzelrc.yaml` still allows parse/unparse using provider defaults.
- [x] Config `github.parse.output-directory` applies to `github parse` when `--output-directory` is omitted.
- [x] Config `github.unparse.output-directory` applies to `github unparse` when `--output-directory` is omitted.
- [x] Config supports `gitlab.parse.output-directory`, `gitlab.parse.single-file`, and `gitlab.parse.filename` keys for upcoming provider rollout.
- [x] Config supports `gitlab.unparse.output-directory` for future unparse defaults.
- [x] Both `parse` and `unparse` use the same resolution logic.
- [x] Resolution uses CLI flag presence, not zero-value checks.

### Error and UX

- [x] Invalid YAML in `.cinzelrc.yaml` returns a clear error with file path.
- [x] Invalid field type returns clear path-based error (example: `github.parse.output-directory must be string`).
- [x] Unknown config keys emit deterministic warnings (stderr) but do not fail execution.
- [x] Inactive provider semantic errors do not block active provider commands.
- [x] Warning format is consistent (`warning: <field-path>: <message>`) and sorted by field path.

### Testing

- [x] Unit tests cover precedence permutations in `internal/command/` resolution helpers.
- [x] Integration tests cover parse/unparse behavior with and without config.
- [x] Tests explicitly document and lock expected v1 config-resolution behavior.
- [x] Dry-run tests verify displayed file paths reflect resolved output directory.
- [x] Test fixtures are named and committed for each permutation (example: `internal/command/testdata/config/cli-overrides-config.yaml`).
- [x] Tests include "flag set but empty" permutations for parse/unparse.

## Success Metrics

- Existing GitHub test suites pass after intentional expectation updates tied to config-first behavior.
- New config resolution test matrix passes for all precedence combinations.
- No duplicate option-resolution logic across parse/unparse command paths.

## Dependencies & Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Ambiguous precedence for empty CLI values | Unexpected output paths | Use `cmd.IsSet(...)` presence semantics and test empty-flag cases |
| Config parsing done in multiple places | Inconsistent behavior | Centralize effective-op resolution helper in CLI layer |
| Expectation drift for existing GitHub behavior | Confusing test or user results | Document intentional behavior changes and update fixtures/tests in same PR |
| Premature coupling to unfinished GitLab provider | Rework later | Keep schema additive; avoid provider interface changes |
| Non-deterministic warning ordering | Flaky tests/log assertions | Sort keys before emitting warnings |

## Implementation Phases

### Phase 1: Config Foundation

- Add config model and loader for `.cinzelrc.yaml`.
- Add precedence resolver for effective provider options.
- Wire resolver into both parse/unparse command paths.

Planned touchpoints:
- `internal/command/command.go`
- `provider/provider.go` (only if additive fields become necessary; avoid if possible)
- `internal/fsutil/` (only if shared config read helper is needed)

### Phase 2: Validation and UX

- Add typed validation errors with field-path context.
- Add unknown-key warning behavior.
- Ensure error format is compatible with current CLI error wrapping.

Planned touchpoints:
- `internal/command/command.go`
- `internal/cinzelerror/`

### Phase 3: Test Matrix and Docs

- Add precedence and behavior-contract tests.
- Add user-facing docs for config schema and precedence.
- Update GitLab provider plan to depend on this completed foundation.

Planned test matrix files:
- `internal/command/testdata/config/no-config.yaml`
- `internal/command/testdata/config/valid-config.yaml`
- `internal/command/testdata/config/invalid-type.yaml`
- `internal/command/testdata/config/unknown-keys.yaml`
- `internal/command/testdata/config/cli-overrides-config.yaml`
- `internal/command/testdata/config/cli-empty-overrides-config.yaml`

Planned touchpoints:
- `internal/command/command_test.go` (or equivalent new test file)
- `README.md`
- `docs/plans/2026-03-09-feat-gitlab-pipelines-provider-plan.md`

## Open Questions

- Should v1 discovery stay cwd-only or move to walk-up search in v1.1?
- Should unknown keys be promoted from warning to failure behind a strict mode later?
- Should command-scoped keys expand beyond `output-directory` in v1.1 (`recursive`, `dry-run` defaults)?

## Sources & References

### Origin

- **Brainstorm document:** [docs/brainstorms/gitlab-provider.md](docs/brainstorms/gitlab-provider.md) — carried forward decisions on provider-specific HCL, CLI-driven provider selection, repeated `rule` blocks, `depends_on`, and `.cinzelrc.yaml` precedence.

### Internal References

- CLI command wiring and provider option mapping: `internal/command/command.go:77`
- Provider option contract: `provider/provider.go:7`
- GitHub defaults and output resolution baseline: `provider/github/github.go:19`
- Input/output filesystem behavior: `internal/fsutil/fsutil.go`
- Existing GitLab plan to re-sequence: `docs/plans/2026-03-09-feat-gitlab-pipelines-provider-plan.md`

### Institutional Learnings

- Deterministic map iteration: `docs/solutions/logic-errors/nondeterministic-map-iteration.md`
- Critical patterns baseline: `docs/solutions/patterns/critical-patterns.md`
- YAML quote style consistency: `docs/solutions/test-failures/yaml-single-quote-golden-mismatch.md`
- Detection-chain discipline for unparse flows: `docs/solutions/best-practices/document-type-auto-detection-chain.md`
- Signature/return path consistency when refactoring core parse flows: `docs/solutions/build-errors/parse-workflow-return-value-mismatch.md`
