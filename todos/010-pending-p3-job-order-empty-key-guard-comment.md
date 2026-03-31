---
status: pending
priority: p3
issue_id: "010"
tags: [code-review, quality]
dependencies: []
---

# `jobOrderFromNode` empty-key guard is undocumented; its intent is unclear

## Problem Statement

```go
for j := 0; j+1 < len(jobs.Content); j += 2 {
    if key := jobs.Content[j].Value; key != "" {
        keys = append(keys, key)
    }
}
```

The `key != ""` guard silently skips job key nodes with an empty `.Value`. There are two interpretations:

1. **It's defensive against malformed YAML** — a key node with empty `.Value` is invalid and should be skipped
2. **It's guarding against alias nodes in key position** — in yaml.v3, alias nodes (`*ref`) in mapping key position have `.Kind == AliasNode` and `.Value == ""`. Skipping them prevents a blank string from entering the order slice.

Without a comment, future maintainers can't tell which it is. The simplicity reviewer argues the guard is unreachable; the security reviewer argues it protects against a real edge case. Both are partially right — the guard is reachable for alias-keyed mappings but the resulting behavior (silently drop then hit the completeness check with a confusing error) is not good.

## Proposed Solutions

### Option A — Add a comment explaining the alias-key edge case

```go
for j := 0; j+1 < len(jobs.Content); j += 2 {
    // Skip alias nodes used as keys (Kind == AliasNode, Value == "").
    // These are unusual in practice but valid YAML.
    if key := jobs.Content[j].Value; key != "" {
        keys = append(keys, key)
    }
}
```

### Option B — Explicitly reject alias keys with an error

```go
keyNode := jobs.Content[j]
if keyNode.Kind == yamlv3.AliasNode || keyNode.Value == "" {
    return nil // or error
}
keys = append(keys, keyNode.Value)
```

## Recommended Action

_To be filled during triage._

## Technical Details

- Affected file: `provider/github/unparse_workflow.go`, `jobOrderFromNode`

## Acceptance Criteria

- [ ] The intent of the empty-key guard is clear from a comment or explicit alias check

## Work Log

- 2026-03-31: Finding created during code review
