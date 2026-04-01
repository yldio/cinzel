---
status: pending
priority: p1
issue_id: "001"
tags: [code-review, architecture, quality]
dependencies: []
---

# Double YAML parse violates CLAUDE.md rule

## Problem Statement

`unparseYAMLFile` parses `yamlBytes` twice: once via `parseYAMLDocument` (using `goccy/go-yaml`) and again inside `yamlJobOrder` (using `gopkg.in/yaml.v3` Node API). This directly violates the `CLAUDE.md` pitfall:

> do not unmarshal YAML twice

In addition to the style violation, two parsers operating on the same bytes can produce inconsistent results (differing YAML 1.1 vs 1.2 semantics, anchor expansion, merge keys), meaning `JobOrder` could reference names that do not correspond to keys in the `Jobs` map.

## Findings

- `provider/github/github.go` → `unparseYAMLFile` calls `parseYAMLDocument(yamlBytes)` then `yamlJobOrder(yamlBytes)`
- `provider/github/unparse_workflow.go:26` → `yamlJobOrder` does a full `yamlv3.Unmarshal` of raw bytes
- CLAUDE.md (Pitfalls section) explicitly forbids double unmarshal

## Proposed Solutions

### Option A — Extract order from `yaml.v3` Node API in a single pass (Recommended)

Replace the first parse in `unparseYAMLFile` with a `yaml.v3` node walk that extracts both `map[string]any` data and job order simultaneously. Remove `yamlJobOrder` entirely.

- **Pros:** single parse, no rule violation, eliminates consistency risk
- **Cons:** requires refactoring `parseYAMLDocument` to use `yaml.v3`; must verify that `yaml.v3` typed decode matches the behavior of `goccy/go-yaml` for all edge cases

### Option B — Accept the double parse but isolate it (Minimal)

Keep `yamlJobOrder` but make it a two-return-value variant of `parseYAMLDocument` so the two parses are collocated and documented as an intentional exception. Add a code comment explaining why `yaml.v3` is needed for order and `goccy/go-yaml` for validation. Update CLAUDE.md pitfall note to reflect the exception.

- **Pros:** minimal code change, explicit and documented tradeoff
- **Cons:** still violates the rule; the inconsistency risk remains

### Option C — Use `yaml.v3` only (Replace goccy dependency in unparse path)

Since `goccy/go-yaml` strict decode is primarily needed for HCL→YAML parse direction, evaluate whether the unparse path actually requires it. If unparse only uses `map[string]any`, switch unparse to `yaml.v3` entirely.

- **Pros:** removes dual-library dependency, single parse
- **Cons:** larger surface change; need to confirm goccy is not needed in unparse path

## Recommended Action

_To be filled during triage._

## Technical Details

- Affected files: `provider/github/github.go`, `provider/github/unparse_workflow.go`
- CLAUDE.md pitfall: "do not unmarshal YAML twice"
- Both parsers agree in practice for typical workflow files, but correctness guarantee is missing

## Acceptance Criteria

- [ ] `yamlBytes` is parsed at most once per call to `unparseYAMLFile`
- [ ] Job order is still preserved correctly after the fix
- [ ] All existing tests pass
- [ ] Roundtrip stability tests pass

## Work Log

- 2026-03-31: Finding created during code review of job order preservation feature
