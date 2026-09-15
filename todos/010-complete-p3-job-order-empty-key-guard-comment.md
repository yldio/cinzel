---
status: complete
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

## Recommended Action

Interpretation 2 is the correct one, and the guard is reachable — probed
against 3a8ee45. An alias node in key position has `.Value == ""`, so the
guard skips it while `Decode` still resolves it into the jobs map. A literal
empty key `"":` does the same. Both cases are worked through with inputs in
issue 008, which should be resolved together with this one: 008 supplies the
regression tests, this issue supplies the comment explaining what the tests
are pinning.

Option A, with the comment naming both inputs rather than only aliases.

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

- [x] The intent of the empty-key guard is resolved — the guard is gone and what it stood for is handled where it belongs

## Work Log

- 2026-03-31: Finding created during code review
- 2026-09-15: Guard confirmed reachable by probe via alias keys and empty
  keys. Interpretation 2 confirmed. Merged with issue 008.
- 2026-09-15: Both interpretations turned out to be wrong, and the guard is
  gone rather than commented.

  Interpretation 2 rested on an alias key having `Value == ""`. It does not: an
  alias node in key position holds the anchor's name, so the guard never saw
  one. Alias resolution is now explicit in `jobKeyName`.

  Interpretation 1 was closer but the guard was doing it by accident. A literal
  `"":` key is the only thing it skipped, and skipping it is what later
  produced `job '' is defined but was not included in the job order` — a
  message about ordering for what is really an unnamed job. The validator now
  refuses an empty job name outright, matching what the parse direction already
  did, and the guard has nothing left to do.

  Resolved with 008, which carries the rest of the detail and the two further
  defects the re-probe turned up.
