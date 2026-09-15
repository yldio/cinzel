---
status: pending
priority: p3
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

**Close without the size limit.** `gopkg.in/yaml.v3` already stops both shapes
this issue describes, so the proposed byte cap defends an entry that is not
open. Measured against 3a8ee45:

| input | result | time | peak RSS |
|---|---|---|---|
| billion laughs, 11 levels x10 aliases | `yaml: document contains excessive aliasing` | 0.03s | 17MB |
| 100000-deep nesting | `yaml: line 4: exceeded max depth of 10000` | 0.01s | 26MB |

A byte cap would also not have caught the real exhaustion path, which was
plain valid YAML with many jobs and no aliases at all: 7.5MB took 88s and
1.4GB. That was quadratic identifier construction, fixed in
`perf(unparse): build job identifiers against a set, not a slice`.

What is worth keeping is a test pinning the dependency's guarantee, so an
upgrade that relaxes either limit is caught here rather than in the field.

## Technical Details

- Affected file: `provider/github/unparse_workflow.go`, function `parseYAMLDocument`
- Related path: `provider/github/github.go` → `Unparse` → `unparseYAMLFile`

## Acceptance Criteria

- [x] Alias-expansion and depth exhaustion shown to be already rejected
- [ ] A test pins both yaml.v3 guarantees so a dependency upgrade cannot
      silently drop them

## Work Log

- 2026-03-31: Finding created during code review of single-pass yaml.v3 parse refactor
- 2026-09-15: Probed against 3a8ee45. Both described attacks are already
  rejected by yaml.v3 in under 0.03s. Size limit dropped from scope; the real
  exhaustion path was quadratic identifier construction and is now fixed.
  Remaining work is a regression test over the dependency's guarantees.
