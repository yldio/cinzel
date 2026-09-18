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
  - "A workflow cut in half, or reduced to one line, by a marker inside its own run block"
tags:
  - "patterns"
  - "determinism"
  - "expressions"
  - "yaml"
  - "hcl"
  - "regexp"
created_date: "2026-03-08"
updated_date: "2026-09-18"
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

Do not insert into a map inside a `range` over that same map either. Whether the
new key is visited by the same range is undefined, so a rename chain is applied
once on some runs and twice on others. Collect the changes, end the range, then
apply them — `relabel` (`provider/gitlab/parse_pipeline.go`) is the worked
example.

```go
// WRONG — the inserted key may or may not be visited again
for label, child := range mapping.children {
    delete(mapping.children, label)
    mapping.children[keys[label]] = child
}
```

Deleting from the map being ranged over is defined and safe; inserting is not.

### 2. Expression escaping

HCL uses `$${{ }}` to represent GitHub Actions `${{ }}` expressions. The conversion is handled automatically by the parser/unparser. Never manually escape or double-escape.

- Parse direction: `hclparser` strips the leading `$` from `$${{ }}`
- Unparse direction: `unparse_emit.go` adds the leading `$` to produce `$${{ }}`

### 3. Single YAML unmarshal

Always unmarshal YAML once via `parseYAMLDocument()`, then classify with `classifyWorkflowDocument()` or `classifyActionDocument()`. Never call `yaml.Unmarshal` twice on the same content.

### 4. YAML quote style

Use `DoubleQuotedStyle` exclusively. The Zed editor converts single quotes to double quotes on save, which breaks golden tests if `SingleQuotedStyle` is used. The decision of what to quote lives in `needsQuoting` (`internal/yamldoc/encode.go:158`), with GitLab keeping its own `stringNeedsQuoting` (`provider/gitlab/pipeline_yaml.go:329`) because the shared rules would quote a large share of real pipelines.

### 5. A structural marker is only one at column 0

A markdown fence and a YAML document separator are both structure, and both are
ordinary text the moment they are indented. A workflow writes them: a `run: |`
block holds a heredoc with `---` in it, or a README fragment wrapped in
backticks. Matching either anywhere on the line read the workflow's own content
as the end of the workflow.

```go
// WRONG — an indented "```" inside a run block ends the fence
regexp.MustCompile("(?s)```(?:ya?ml)?\\s*\n(.*?)```")

// RIGHT — (?m) makes ^ and $ line anchors, so only column 0 counts
regexp.MustCompile("(?ms)^```(?:ya?ml)?[ \\t]*\n(.*?)^```[ \\t]*$")
```

`splitYAMLDocuments` (`internal/command/assist.go`) applies the same rule to
`---`, and `StripFences` (`internal/ai/provider.go`) to fences. Both failures
look the same from outside: `cinzel assist` reports a conversion error, or
writes half a workflow, against an LLM response that was entirely valid.

Use `(?m)` for line anchors and `\A` / `\z` when the anchor really is the whole
text — `^` and `$` stop meaning that as soon as `(?m)` is on.

### 6. Return value consistency

`parseHCLToWorkflows` returns 4 values. Every error path must return all 4: `return nil, nil, nil, err`. Missing a nil causes compile errors that are tedious to chase across 8+ return sites.
