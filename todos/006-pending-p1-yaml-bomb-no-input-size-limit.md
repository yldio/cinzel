---
status: pending
priority: p1
issue_id: "006"
tags: [code-review, security, performance]
dependencies: []
---

# No input size limit before YAML alias expansion (billion laughs DoS)

## Problem Statement

`parseYAMLDocument` calls `yamlv3.Unmarshal(content, &node)` with no upper bound on input size or alias depth. `gopkg.in/yaml.v3` expands YAML aliases eagerly during unmarshal. A crafted input using deeply nested anchors (e.g., billion-laughs pattern) causes exponential memory allocation before any application check runs.

While `cinzel` is a CLI tool operating on local files, two attack surfaces exist:
1. `cinzel assist` feeds LLM-generated YAML through the unparse path — the LLM output is untrusted
2. A repository CI workflow could trigger `cinzel unparse` against a YAML file controlled by a third party

## Findings

- `provider/github/unparse_workflow.go` — `parseYAMLDocument` calls `yamlv3.Unmarshal` with no size check
- No max-byte guard at any call site in `Unparse` or `unparseYAMLFile`
- `node.Decode(&doc)` after unmarshal expands the node tree into `map[string]any`, compounding allocation

## Proposed Solutions

### Option A — Byte-length guard at `parseYAMLDocument` call site (Recommended)

Add a length check in `parseYAMLDocument` before the unmarshal:

```go
const maxYAMLBytes = 1 << 20 // 1 MB
if len(content) > maxYAMLBytes {
    return nil, nil, fmt.Errorf("YAML input exceeds maximum size of %d bytes", maxYAMLBytes)
}
```

- **Pros:** simple, fast, zero false positives for real workflow files (largest real workflow is <100KB)
- **Cons:** arbitrary limit; must be documented

### Option B — Cap alias expansion with a wrapper decoder

No built-in alias limit in yaml.v3. Would require forking or wrapping.

- **Pros:** more precise
- **Cons:** significant implementation cost; yaml.v3 is not easily intercepted

## Recommended Action

_To be filled during triage._

## Technical Details

- Affected file: `provider/github/unparse_workflow.go`, function `parseYAMLDocument`
- Related path: `provider/github/github.go` → `Unparse` → `unparseYAMLFile`

## Acceptance Criteria

- [ ] Input larger than a defined maximum is rejected with a clear error before `yamlv3.Unmarshal`
- [ ] Limit is documented
- [ ] Existing tests pass

## Work Log

- 2026-03-31: Finding created during code review of single-pass yaml.v3 parse refactor
