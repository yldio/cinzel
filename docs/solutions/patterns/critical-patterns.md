---
title: "Critical patterns for cinzel development"
module: "Core"
problem_type: "best_practice"
component: "all"
severity: "critical"
root_cause: "pattern_violation"
symptoms:
  - "Non-deterministic test output"
  - "Golden test failures after unrelated changes"
  - "Double-escaped expressions in output"
tags:
  - "patterns"
  - "determinism"
  - "expressions"
  - "yaml"
  - "hcl"
created_date: "2026-03-08"
updated_date: "2026-03-08"
---

## Critical Patterns

### 1. Deterministic map iteration

Always use deterministic key-order iteration helpers (for example a local `sortedKeys()` helper) when iterating maps whose output is user-visible or tested against golden files.

```go
// WRONG — non-deterministic output
for name, attr := range sb.Attributes {
    out[name] = processAttr(attr)
}

// RIGHT — deterministic output
for _, name := range sortedKeys(sb.Attributes) {
    attr := sb.Attributes[name]
    out[name] = processAttr(attr)
}
```

### 2. Expression escaping

HCL uses `$${{ }}` to represent GitHub Actions `${{ }}` expressions. The conversion is handled automatically by the parser/unparser. Never manually escape or double-escape.

- Parse direction: `hclparser` strips the leading `$` from `$${{ }}`
- Unparse direction: `unparse_emit.go` adds the leading `$` to produce `$${{ }}`

### 3. Single YAML unmarshal

Always unmarshal YAML once via `parseYAMLDocument()`, then classify with `classifyWorkflowDocument()` or `classifyActionDocument()`. Never call `yaml.Unmarshal` twice on the same content.

### 4. YAML quote style

Use `DoubleQuotedStyle` exclusively. The Zed editor converts single quotes to double quotes on save, which breaks golden tests if `SingleQuotedStyle` is used.

### 5. Return value consistency

`parseHCLToWorkflows` returns 4 values. Every error path must return all 4: `return nil, nil, nil, err`. Missing a nil causes compile errors that are tedious to chase across 8+ return sites.
