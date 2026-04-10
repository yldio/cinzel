---
title: "HCL inline comments not propagated to YAML output"
problem_type: feature_implementation
component: provider/github
symptoms:
  - HCL inline comments on attributes silently dropped in generated YAML
  - Feature initially scoped only to permissions blocks, leaving other attributes without comment propagation
  - Type assertion error after wrapping values: `permissions scope "actions" must have a string value`
tags:
  - hcl
  - yaml
  - comments
  - permissions
  - type-assertion
  - annotated
  - parse
  - workflow
date: 2026-04-10
status: solved
---

## Problem

HCL attributes with trailing inline comments:

```hcl
permissions {
  contents = "read" # only needed for private repos
  actions  = "read" # required by the action
}
```

produced YAML with the comments silently dropped:

```yaml
permissions:
  contents: read
  actions: read
```

The initial fix introduced a permissions-specific mechanism (`parsePermissionsBlock` + a `permissionsComments` sentinel key + `permissionsMapNode`), but this was bespoke, duplicated logic, and didn't cover any other HCL attribute.

---

## Root Cause

The parse pipeline converts HCL to a `map[string]any` that flows through validation and ultimately into the `yaml.v3` node serializer. The intermediate map representation has no slot for comment metadata — values are plain Go types (`string`, `bool`, `map[string]any`, etc.) — so any comment extracted from the HCL source was immediately discarded.

---

## Solution

### 1. Introduce an `annotated` wrapper type (`provider/github/workflow_yaml.go`)

A lightweight struct that carries a comment alongside any value through the `map[string]any` pipeline:

```go
type annotated struct {
    value   any
    comment string
}
```

Two helpers strip the wrapper at validation boundaries:

```go
func unwrapAnnotated(v any) any {
    if a, ok := v.(annotated); ok {
        return a.value
    }
    return v
}

func unwrapAnnotatedMap(m map[string]any) map[string]any {
    out := make(map[string]any, len(m))
    for k, v := range m {
        out[k] = unwrapAnnotated(v)
    }
    return out
}
```

### 2. Handle `annotated` in `toYAMLNode` (same file)

The terminal serializer unwraps the value and attaches the comment as a `LineComment` on the `yaml.v3` node — which renders as a trailing `# comment` on the same output line:

```go
case annotated:
    node, err := toYAMLNode(v.value)
    if err != nil {
        return nil, err
    }
    node.LineComment = v.comment
    return node, nil
```

### 3. Wrap at parse time in `parseBodyMap` (`provider/github/parse_workflow.go`)

In the `default` attribute case, check for a trailing comment and wrap if present:

```go
default:
    val, err := parseAttr(attr.Expr, hv)
    if err != nil {
        return nil, err
    }
    yamlKey := naming.ToYAMLKey(name)
    if c := extractInlineComment(attr); c != "" {
        out[yamlKey] = annotated{value: val, comment: c}
    } else {
        out[yamlKey] = val
    }
```

`extractInlineComment` reads the source file using `attr.SrcRange.End.Byte` to locate the byte immediately after the attribute value, then scans the rest of the line for a `#` prefix.

### 4. Simplify permissions parsing

With `parseBodyMap` now handling comments, the permissions-specific code is redundant. Replace `parsePermissionsBlock` with a plain `parseBodyMap` call:

```go
for _, block := range cfg.PermBlocks {
    child, err := parseBodyMap(block.Body, hv, "permissions")
    if err != nil {
        return nil, err
    }
    out["permissions"] = child
}
```

This removes `parsePermissionsBlock`, `permissionsMapNode`, the `permissionsComments` sentinel key, and the special-case branch in `workflowMapNode`.

### 5. Fix validators (`provider/github/validate.go`)

`ValidatePermissions` does `levelRaw.(string)` on map values. After wrapping, this assertion fails. Unwrap before validating at both workflow and job level:

```go
if perms, ok := workflow.Body["permissions"]; ok {
    if permsMap, ok := perms.(map[string]any); ok {
        perms = unwrapAnnotatedMap(permsMap)
    }
    if err := ghworkflow.ValidatePermissions(perms); err != nil {
        return withPath("workflow."+workflow.ID+".permissions", err)
    }
}
```

---

## Why the General Approach Over the Permissions-Specific One

| Permissions-specific | General `annotated` |
|---|---|
| Magic `permissionsComments` sentinel key pollutes map | No sentinel; comment travels with the value |
| Requires `permissionsMapNode` rendering path | `toYAMLNode` handles it in one new case |
| Only permissions attributes carry comments | Any HCL attribute with `# comment` propagates |
| More code, less reuse | Less code, works for free on future attributes |

---

## Error Encountered During Implementation

```
error in workflow 'ci': workflow.ci.permissions: permissions scope "actions" must have a string value
```

`ValidatePermissions` iterates `map[string]any` and asserts each value to `string`. After the `annotated` wrapper was introduced, the assertion received an `annotated` struct instead of a plain string. Fixed by calling `unwrapAnnotatedMap` before validation.

---

## Prevention

### Pitfalls

- **`map[string]any` type assertions are silent**: Any code doing `v.(string)` on pipeline values breaks when a new wrapper type is introduced. The compiler gives no warning; it only surfaces at runtime with specific inputs.
- **Blast radius is invisible**: You cannot grep to find all consumers of pipeline map values statically — every function that touches the map may need updating.
- **Validator/serializer ordering**: Unwrapping must happen before validation. This ordering is load-bearing but usually implicit.

### Testing Strategy

- For every function that type-asserts pipeline map values, add a parallel test that passes `annotated`-wrapped values alongside plain values.
- Include a roundtrip test that injects commented HCL attributes and asserts they appear as inline comments in the generated YAML.

### Design Guidance

- **Unwrap at a defined boundary**: establish one normalisation step (e.g. `unwrapAnnotatedMap`) that all validators call, rather than unwrapping ad hoc.
- **Use `annotated` when**: metadata must travel with the value through the pipeline and the consumer count is small and controlled.
- **Use a separate metadata map when**: consumers are numerous or externally defined — validators never need to see the metadata at all.

---

## Related

- [`github-pin-comment-placement-and-empty-permissions-default.md`](./github-pin-comment-placement-and-empty-permissions-default.md) — documents the sentinel-swap pattern for `permissions: {}` and inline `# vtag` comments on version lines
- [`yaml-map-key-order-lost-use-node-api.md`](./yaml-map-key-order-lost-use-node-api.md) — documents `yaml.v3` Node API usage for order-preserving serialization
- [`preserve-hcl-job-order-in-yaml-output.md`](./preserve-hcl-job-order-in-yaml-output.md) — documents the `jobsOrder` sentinel key pattern, a predecessor to the `annotated` approach
