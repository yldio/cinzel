---
title: "GitLab CI/CD Pipelines Provider"
type: feat
status: completed
date: 2026-03-09
origin: docs/brainstorms/2026-03-09-gitlab-provider.md
---

# GitLab CI/CD Pipelines Provider

## Overview

Add a GitLab CI/CD Pipelines provider to cinzel, enabling bidirectional conversion between HCL and GitLab CI/CD YAML. This is the second provider, validating the multi-provider architecture. The GitLab provider uses its own idiomatic HCL shape — no shared abstractions with the GitHub provider.

## Problem Statement / Motivation

cinzel currently only supports GitHub Actions. The `provider.Provider` interface was designed for multiple providers but has never been tested with a second implementation. GitLab CI/CD is the most requested alternative and has a fundamentally different pipeline model (stages vs DAG, monolithic scripts vs steps), making it a strong test of the architecture's flexibility.

## Proposed Solution

Implement `provider/gitlab/` following the same structural patterns as `provider/github/` but with GitLab-specific HCL schema, YAML marshalling, and validation. The provider will be wired into the CLI via `cinzel gitlab parse` and `cinzel gitlab unparse`.

## Prerequisites Status

- [x] `.cinzelrc.yaml` command-scoped config foundation completed (`docs/plans/2026-03-09-feat-cinzelrc-provider-config-precedence-plan.md`).
- [x] HCL dependency keyword alignment to `depends_on` completed in GitHub provider (`docs/plans/2026-03-09-feat-rename-github-needs-to-depends-on-plan.md`).
- [x] Strict schema validation baseline established in GitHub provider parse/unparse contracts (`docs/plans/2026-03-09-feat-github-strict-block-schema-enforcement-plan.md`).
- [x] GitLab provider implementation completed.

### HCL Shape

```hcl
stages = ["build", "test", "deploy"]

variable "deploy_env" {
  name        = "DEPLOY_ENV"
  value       = "production"
  description = "Target environment"
}

job "build" {
  stage  = "build"
  image  = "golang:1.26"
  script = ["go build -o app ./..."]

  artifacts {
    paths     = ["app"]
    expire_in = "1 hour"
  }
}

job "test" {
  stage  = "test"
  image  = "golang:1.26"
  script = ["go test ./..."]

  cache {
    key   = "go-modules"
    paths = ["vendor/"]
  }

  rule {
    if   = "$CI_PIPELINE_SOURCE == 'merge_request_event'"
    when = "on_success"
  }
}

job "deploy" {
  stage      = "deploy"
  depends_on = [job.build, job.test]
  script     = ["./deploy.sh"]

  rule {
    if   = "$CI_COMMIT_BRANCH == 'main'"
    when = "manual"
  }

  rule {
    when = "never"
  }
}

workflow {
  rule {
    if   = "$CI_COMMIT_BRANCH == 'main'"
    when = "always"
  }

  rule {
    if   = "$CI_PIPELINE_SOURCE == 'merge_request_event'"
    when = "always"
  }
}
```

## Technical Considerations

### Critical design decisions resolved in brainstorm

1. **Provider-specific HCL** — no shared `step` blocks across providers (see brainstorm: `docs/brainstorms/2026-03-09-gitlab-provider.md`)
2. **Provider selection implicit from CLI** — no `ci = "gitlab"` attribute
3. **`template` block** for hidden jobs — `.go_base:` in YAML ↔ `template "go_base"` in HCL
4. **`variable` with explicit `name`** — HCL label is internal, `name` attribute is the YAML key
5. **Repeated `rule` blocks** (singular) — first match wins, order preserved
6. **`depends_on`** for DAG dependencies — maps to `needs:` in YAML

### Critical gaps identified by SpecFlow analysis

#### 1. `${}` interpolation collision (CRITICAL)

HCL natively interprets `${...}` as string interpolation. GitLab uses `${CI_VARIABLE}` syntax. This collides.

**Resolution:** Use `$${VAR}` in HCL (double-dollar escape), emitting `${VAR}` in YAML. This mirrors the GitHub `$${{ }}` → `${{ }}` pattern. The unbraced form `$CI_VAR` passes through as-is (HCL does not interpret `$NAME` without braces). During unparse, `${VAR}` in YAML becomes `$${VAR}` in HCL.

#### 2. `naming.ToYAMLKey` incompatibility (CRITICAL)

GitLab YAML uses underscores (`before_script`, `allow_failure`). The existing `naming.ToYAMLKey()` converts underscores to hyphens — this would corrupt GitLab output.

**Resolution:** The GitLab provider does NOT use `naming.ToYAMLKey()` or `naming.ToHCLKey()`. Keys pass through unchanged since both HCL and GitLab YAML use underscores natively. `naming.SanitizeIdentifier()` and `naming.UniqueIdentifier()` are still reusable.

#### 3. Single-file output model (CRITICAL)

GitLab outputs one `.gitlab-ci.yml`. When `--directory` provides multiple HCL files, they must merge into a single output. Duplicate job names across files produce an error.

**Resolution:** Parse all HCL files, collect all jobs/variables/stages into one document, validate for duplicates, then marshal to a single file. Default output path: `.gitlab-ci.yml` (under output directory `.`).

#### 4. Extended feature handling during unparse (IMPORTANT)

Real `.gitlab-ci.yml` files contain features beyond the minimal baseline (`extends`, `include`, `default`, hidden jobs, `services`).

**Resolution:** Implemented explicit support for `template`/`extends`, repeated `include` blocks, `default`, and repeated `service` blocks while retaining safe pass-through behavior for still-unsupported keys.

#### 5. Document classification (IMPORTANT)

GitLab has no single discriminator like GitHub's `on`+`jobs`.

**Resolution:** A document is GitLab CI if it contains: (a) any top-level key with a `script` sub-key, OR (b) `stages` as a list, OR (c) `workflow` with `rules`. This heuristic covers the vast majority of real files.

#### 6. `workflow` block semantics (IMPORTANT)

GitLab's `workflow` is optional pipeline-level config (rules, name). Unlike GitHub's `workflow` which is a file container.

**Resolution:** Unlabeled `workflow {}` block. At most one per pipeline. Contains `rule` blocks and optional `name` attribute.

### Architecture impacts

- **No changes to `provider.Provider` interface** — it works as-is
- **No changes to `internal/command/`** — CLI auto-wires the new provider
- **No changes to `internal/hclparser/`** — expression evaluation is generic
- **`internal/yamlwriter/`** — `toYAMLNode()` and `stringNeedsQuoting()` are reusable. `marshalWorkflowYAML()` and `workflowKeyOrder` are GitHub-specific; GitLab needs its own equivalent
- **`internal/naming/`** — `SanitizeIdentifier()` and `UniqueIdentifier()` are reusable. `ToYAMLKey()`/`ToHCLKey()` are NOT used by the GitLab provider
- **`conversion.go`** — duplicate `ctyToAny()`/`anyToCty()` into the GitLab provider initially. Extract to `internal/conversion/` only if the implementations remain identical after both providers are complete

## Integration Test Scenarios

- Parse GitLab HCL while GitHub HCL exists in the same directory — providers must not interfere.
- Unparse a `.gitlab-ci.yml` found alongside GitHub workflow YAMLs — document classification must correctly route.

## Acceptance Criteria

### Phase 1: Foundation

- [x] `provider/gitlab/` package exists with `doc.go`, `gitlab.go`, `config.go`, `errors.go`, `models.go`
- [x] `gitlab.New()` returns a `provider.Provider` implementation
- [x] `cinzel gitlab parse --help` and `cinzel gitlab unparse --help` work
- [x] Default output directories: parse → `.` (producing `.gitlab-ci.yml`), unparse → `./cinzel`
- [x] Provider smoke test passes (`TestProviderWiringSmoke`)
- [x] `$${VAR}` escape prototype verified — HCL parser handles it identically to `$${{ }}` (hard gate for Phase 2)

### Phase 2: Parse (HCL → YAML)

- [x] `stages` top-level attribute → `stages:` YAML list
- [x] `variable` blocks → `variables:` map with optional `description`
- [x] `job` blocks → top-level YAML jobs with `script`, `image`, `stage`, `before_script`, `after_script`, `tags`
- [x] `rule` blocks → `rules:` list within jobs (ordered, first-match-wins). Supported attributes: `if`, `when`, `allow_failure`, `changes` (as string list), `exists` (as string list). The object form of `changes` (`paths`/`compare_to`) is deferred to v0.2.
- [x] `artifacts` block → `artifacts:` with `paths`, `exclude`, `expire_in`, `name`, `untracked`, `when`. Nested `reports` sub-key is passed through as a generic block.
- [x] `cache` block → `cache:` with `key`, `paths`, `untracked`, `when`, `policy`
- [x] `depends_on` → `needs:` simple string list
- [x] `workflow` block → `workflow:` with `rules` and optional `name`
- [x] `$${VAR}` escape → `${VAR}` in YAML output
- [x] `$VAR` passes through unchanged
- [x] YAML key ordering: `stages`, `variables`, `workflow`, `default`, then jobs alphabetically
- [x] Validation: job requires `script`, `stage` must reference declared stage, no duplicate job names, no `depends_on` cycles
- [x] Multiple HCL files merge into single `.gitlab-ci.yml`
- [x] `--dry-run` prints to stdout

### Phase 3: Unparse (YAML → HCL)

- [x] Document classification: identifies GitLab CI YAML via heuristic
- [x] Reserved keywords (`stages`, `variables`, `workflow`, `default`) → appropriate HCL blocks/attributes
- [x] Top-level keys with `script` → `job` blocks
- [x] Hidden jobs (`.name:`) ↔ `template` blocks
- [x] `needs:` → `depends_on = [job.x]` reference list
- [x] `rules:` → repeated `rule` blocks
- [x] `${VAR}` in YAML → `$${VAR}` in HCL
- [x] Explicit unparse support for `extends`, `include`, `default`, `services`; remaining out-of-scope keys continue via generic pass-through with warning
- [x] `--dry-run` prints to stdout

### Phase 4: Testing & Documentation

- [x] Golden tests for all v0.1 features (`testdata/fixtures/pipelines/`)
- [x] Roundtrip tests proving HCL → YAML → HCL → YAML semantic stability
- [x] Fixture matrix: `testdata/fixtures/matrix/{parse,unparse}/{valid,invalid}/`
- [x] Invalid input tests with `.error.txt` expected messages
- [x] Benchmark tests: `BenchmarkParsePipeline`, `BenchmarkUnparsePipeline`, `BenchmarkRoundtripPipeline`
- [x] `provider/gitlab/README.md` with HCL schema reference
- [x] Root `README.md` updated with GitLab provider entry

## Success Metrics

- All golden tests pass
- All roundtrip tests prove semantic stability
- Real-world `.gitlab-ci.yml` files (from popular open-source projects) can be unparsed and re-parsed without data loss for v0.1 features
- No regressions in GitHub provider tests

## Dependencies & Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| `${}` escape doesn't work identically to `$${{ }}` | Blocks all expression handling | Prototype early in Phase 1 |
| `naming.ToYAMLKey` accidentally used | Corrupts all GitLab output | Provider does not import `naming.ToYAMLKey` — CI lint can verify |
| Single-file merge produces non-deterministic output | Flaky tests | Use deterministic key-order helpers from day one |
| Document classification false positives | Wrong provider handles a file | Classification only runs when user explicitly uses `cinzel gitlab unparse` |
| `conversion.go` duplication | Maintenance burden | Extract to `internal/conversion/` if patterns match after implementation |

## Implementation Phases

### Phase 1: Foundation (scaffold + smoke test)

Files to create:
- `provider/gitlab/doc.go`
- `provider/gitlab/gitlab.go` — struct, `New()`, `Parse()`, `Unparse()` stubs, constants
- `provider/gitlab/config.go` — `parseConfig` struct with GitLab HCL schema
- `provider/gitlab/errors.go` — sentinel errors
- `provider/gitlab/models.go` — `PipelineYAMLFile` struct
- `provider/gitlab/io_helpers.go` — input/output resolution
- `provider/gitlab/gitlab_test.go` — smoke test

Wire in `cinzel.go`:
```go
providers := []provider.Provider{
    github.New(),
    gitlab.New(),
}
```

Prototype `$${VAR}` escape to confirm HCL parser handles it correctly.

### Phase 2: Parse direction (HCL → YAML)

Files to create:
- `provider/gitlab/parse_pipeline.go` — main parse logic, `parseHCLToPipeline()`
- `provider/gitlab/pipeline_yaml.go` — YAML node builder with GitLab key ordering
- `provider/gitlab/validate.go` — validation (script required, stage declared, no cycles)
- `provider/gitlab/conversion.go` — `ctyToAny()`/`anyToCty()` (or extract to internal)

Test fixtures:
- `testdata/fixtures/pipelines/basic_pipeline.hcl` + `.golden.yaml`
- `testdata/fixtures/pipelines/variables_and_stages.hcl` + `.golden.yaml`
- `testdata/fixtures/pipelines/rules_and_artifacts.hcl` + `.golden.yaml`
- `testdata/fixtures/pipelines/depends_on.hcl` + `.golden.yaml`
- `testdata/fixtures/pipelines/workflow_rules.hcl` + `.golden.yaml`
- `testdata/fixtures/pipelines/expression_escaping.hcl` + `.golden.yaml`

Test files:
- `provider/gitlab/golden_test.go`
- `provider/gitlab/fixture_matrix_test.go` (parse side)

### Phase 3: Unparse direction (YAML → HCL)

Files to create:
- `provider/gitlab/unparse_pipeline.go` — YAML→HCL conversion, document classification
- `provider/gitlab/unparse_emit.go` — HCL generation helpers

Test fixtures:
- Roundtrip fixtures reuse parse fixtures
- `testdata/fixtures/matrix/unparse/valid/` — YAML→HCL→YAML roundtrip pairs
- `testdata/fixtures/matrix/unparse/invalid/` — malformed YAML with `.error.txt`
- `testdata/fixtures/pipelines/real_world_passthrough.yaml` + `.golden.hcl` — file with v0.2 features

Test files:
- `provider/gitlab/roundtrip_test.go`
- `provider/gitlab/fixture_matrix_test.go` (unparse side)

### Phase 4: Polish

- `provider/gitlab/benchmark_test.go`
- `provider/gitlab/README.md`
- Update root `README.md`
- Run `mise run license` for headers
- Run `mise run fmt`
- Full test suite green: `go test ./...`

## Alternative Approaches Considered

1. **Shared `step` abstraction across providers** — rejected because GitLab has no step concept; forced abstraction would be leaky (see brainstorm: `docs/brainstorms/2026-03-09-gitlab-provider.md`, decision 1)
2. **`ci = "gitlab"` attribute in HCL** — rejected because CLI subcommand already identifies the provider (see brainstorm: decision 2)
3. **`job ".my_template"` for hidden jobs** — rejected because HCL identifiers can't start with `.` and the intent should be explicit in the block type (see brainstorm: decision 3)
4. **`needs` keyword** — replaced with `depends_on` to align with HCL conventions (Terraform uses `depends_on`); see brainstorm update

## Sources & References

### Origin

- **Brainstorm document:** [docs/brainstorms/2026-03-09-gitlab-provider.md](docs/brainstorms/2026-03-09-gitlab-provider.md) — Key decisions carried forward: provider-specific HCL (no shared abstractions), `template` block for hidden jobs, `depends_on` over `needs`, repeated `rule` blocks, `$${VAR}` escape pattern.

### Internal References

- Provider interface: `provider/provider.go`
- CLI wiring: `internal/command/command.go:77-170`
- GitHub provider (reference implementation): `provider/github/github.go`
- HCL config schema pattern: `provider/github/config.go`
- Parse flow pattern: `provider/github/parse_workflow.go`
- YAML marshalling pattern: `provider/github/workflow_yaml.go`
- Document classification pattern: `provider/github/unparse_workflow.go`
- Naming utilities: `internal/naming/naming.go`
- Sorted iteration: `internal/maputil/maputil.go`

### Institutional Learnings

- Deterministic map iteration: `docs/solutions/logic-errors/nondeterministic-map-iteration.md`
- YAML quote style: `docs/solutions/test-failures/yaml-single-quote-golden-mismatch.md`
- Return value consistency: `docs/solutions/build-errors/parse-workflow-return-value-mismatch.md`
- Document classification ordering: `docs/solutions/best-practices/document-type-auto-detection-chain.md`
- Generic attribute parsing: `docs/solutions/best-practices/generic-attribute-parsing-for-action-types.md`
- Avoid false "Not yet" claims: `docs/solutions/documentation-gaps/not-yet-sections-create-false-expectations.md`
