---
status: pending
priority: p2
issue_id: "008"
tags: [code-review, quality, architecture]
dependencies: []
---

# Completeness check in `buildWorkflowJobIndex` defends unreachable state (YAGNI)

## Problem Statement

The new loop at the end of `buildWorkflowJobIndex`:

```go
for jobName := range jobs {
    if _, covered := jobIDMap[jobName]; !covered {
        return nil, nil, nil, fmt.Errorf("job '%s' is defined but was not included in the job order", jobName)
    }
}
```

Both `jobs` (from `doc.Jobs`) and `order` (from `jobOrderFromNode`) are derived from the **same** `yaml.v3.Node` tree in the same parse pass. They cannot diverge. The check defends against a state that is structurally impossible given the current call graph, adds dead code, and will produce a confusing error message if `jobOrderFromNode` has a bug (the message implies a caller contract violation, not an internal bug).

**Note:** This concern does not apply to the fallback path where `order` is `nil` and `jobNames = sortedKeys(jobs)` is used — in that case there is no order/map divergence to check anyway.

## Findings

- `provider/github/unparse_emit.go:52-58`
- Both `order` and `jobs` originate from `parseYAMLDocument` in the same yaml.v3 node tree
- The check fires only when `order` is non-empty, which is exactly the case where divergence is impossible

## Proposed Solutions

### Option A — Remove the completeness loop (Recommended)

Delete lines 52–58 of `unparse_emit.go`. If `jobOrderFromNode` produces wrong output, the existing test suite will catch it.

- **Pros:** -6 LOC; removes dead code; function intent is clearer
- **Cons:** slightly less defensive — an acceptable tradeoff given the invariant is provable

### Option B — Keep the loop but add a comment explaining the invariant

Add: `// Invariant: order and jobs derive from the same yaml.v3 node; this check catches bugs in jobOrderFromNode.`

- **Pros:** explicit invariant documentation
- **Cons:** dead code still present; misleading error message

## Recommended Action

_To be filled during triage._

## Technical Details

- Affected file: `provider/github/unparse_emit.go`, `buildWorkflowJobIndex`

## Acceptance Criteria

- [ ] Completeness loop is removed OR clearly documented as a self-invariant check
- [ ] All existing tests pass after removal

## Work Log

- 2026-03-31: Finding created during code review
