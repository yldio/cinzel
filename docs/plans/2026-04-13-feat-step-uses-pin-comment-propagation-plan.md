---
title: "feat: propagate pin inline comment from step uses version to YAML"
type: feat
status: completed
date: 2026-04-13
---

# feat: propagate pin inline comment from step uses version to YAML

## Problem Statement

The `pin` command adds an inline `# vtag` comment to the `version` attribute of a `uses {}` block:

```hcl
uses {
  action  = "actions/checkout"
  version = "abc123sha..." # v4
}
```

This comment is silently dropped during HCL→YAML conversion. The generated YAML:

```yaml
uses: actions/checkout@abc123sha
```

Should instead be:

```yaml
uses: actions/checkout@abc123sha # v4
```

This mirrors what was done for `permissions` attributes, but the `uses` pipeline is structurally different — it uses `cty.Value` and typed structs, not `map[string]any`.

## Root Cause

`UsesListConfig.Parse()` combines `action` and `version` into a single `cty.StringVal("action@version")` string. The `version` attribute's trailing comment is never read — it's discarded before the combined string is formed. There is no slot to carry comment metadata in `cty.Value`.

## Proposed Solution

### 1. Extract comment from `version` expression (`provider/github/action/uses.go`)

`UsesConfig.Version` is an `hcl.Expression`. Unlike `*hclsyntax.Attribute`, it doesn't have a `SrcRange` field, but it does implement `Range() hcl.Range` which provides `End.Byte`. Use this to read the trailing comment from the source file — same technique as `extractInlineComment`, adapted for `hcl.Expression`.

Add a package-level helper:

```go
func extractExprComment(expr hcl.Expression) string {
    r := expr.Range()
    if r.Filename == "" {
        return ""
    }
    src, err := os.ReadFile(r.Filename)
    if err != nil || int(r.End.Byte) >= len(src) {
        return ""
    }
    rest := src[r.End.Byte:]
    newline := bytes.IndexByte(rest, '\n')
    if newline < 0 {
        newline = len(rest)
    }
    tail := strings.TrimSpace(string(rest[:newline]))
    if !strings.HasPrefix(tail, "#") {
        return ""
    }
    return tail
}
```

### 2. Return comment from `UsesListConfig.Parse()` (same file)

Change signature:

```go
func (config *UsesListConfig) Parse(hv *hclparser.HCLVars) (cty.Value, string, error)
```

After resolving `version`, call `extractExprComment(c.Version)` and return the comment string alongside the `cty.Value`.

### 3. Thread comment through step parse (`provider/github/step/`)

**`step.go`** — add field to `Step`:

```go
UsesComment string
```

**`uses.go`** — update `StepConfig.parseUses()` signature:

```go
func (config *StepConfig) parseUses(hv *hclparser.HCLVars) (cty.Value, string, error)
```

Return the comment from `config.Uses.Parse(hv)`.

**`stepparse.go`** — capture comment and set on `parsedStep`:

```go
parsedUses, usesComment, err := config.parseUses(hv)
...
parsedStep.UsesComment = usesComment
```

### 4. Attach comment in YAML output

Find where step maps are built and locate the `uses` key serialization. Attach `UsesComment` as a `LineComment` on the `yaml.v3` node for that key.

> **Note:** Requires investigating the step→YAML path to confirm whether it goes through `toYAMLNode` (in which case `annotated` wrapping could work) or a separate marshaling path. If the step is serialized as `map[string]any`, wrap the `uses` value with `annotated{value: usesStr, comment: UsesComment}` before the map is passed to the YAML serializer.

## Acceptance Criteria

- [x] `version = "sha" # v4` in a `uses {}` block → `uses: action@sha # v4` in YAML
- [x] No comment on `version` → `uses: action@sha` (no trailing comment, no regression)
- [x] `action` without `version` → no comment attempted
- [x] Golden fixtures updated where steps use pinned SHAs with comments
- [x] Roundtrip stable: YAML → HCL → YAML produces same output
- [x] Tests in `provider/github/action/uses_test.go` and `provider/github/step/uses_test.go` cover comment propagation

## Files

- `provider/github/action/uses.go` — `extractExprComment`, updated `Parse` signature
- `provider/github/action/uses_test.go` — comment extraction and propagation tests
- `provider/github/step/step.go` — add `UsesComment string` to `Step`
- `provider/github/step/uses.go` — updated `parseUses` signature
- `provider/github/step/stepparse.go` — thread comment into `Step.UsesComment`
- Step YAML serialization path (TBD after investigation)
- Golden fixture YAML files (update where applicable)

## Out of Scope

- Comments on `action` attribute (only `version` carries the pin tag)
- Propagating comments from other `uses`-adjacent fields (`with`, `env`)
- Job-level `uses` (reusable workflows) — separate struct, separate path

## Open Question

Confirm step→YAML serialization path before implementing step 4. If steps go through `map[string]any` → `toYAMLNode`, the `annotated` wrapper handles it for free. If they are struct-marshaled separately, a custom marshaler or node builder is needed.

## Sources

- `provider/github/action/uses.go:73` — `UsesListConfig.Parse()`
- `provider/github/step/step.go:27` — `Step.Uses cty.Value`
- `provider/github/step/stepparse.go:81` — `config.parseUses(hv)`
- Related: `provider/github/parse_workflow.go:437` — `extractInlineComment` for `*hclsyntax.Attribute`
- Related plan: `docs/plans/2026-03-09-feat-cinzelrc-provider-config-precedence-plan.md`
