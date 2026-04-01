---
status: pending
priority: p2
issue_id: "002"
tags: [code-review, quality, error-handling]
dependencies: [001]
---

# Job in `order` but absent from `jobs` map gives misleading error

## Problem Statement

In `buildWorkflowJobIndex`, when `order` is non-empty it is used verbatim as the iteration list. If `yamlJobOrder` returns a name that is not a key in the `jobs` map (due to parser disagreement, anchor expansion, or merge-key edge cases), `toStringAnyMap(jobs[jobName])` returns `(nil, false)` and the caller gets:

```
job 'x' must be an object
```

This is misleading — the actual problem is that the job does not exist in the parsed map, not that it has the wrong type.

## Findings

- `provider/github/unparse_emit.go:30-34`
- `jobs[jobName]` returns `nil` (zero value of `any`) when key is absent
- `toStringAnyMap(nil)` returns `(nil, false)` — same as a wrongly-typed value
- Error message `"job 'x' must be an object"` conflates two distinct failure modes

## Proposed Solutions

### Option A — Pre-check presence before type assertion (Recommended)

```go
raw, exists := jobs[jobName]
if !exists {
    return nil, nil, nil, fmt.Errorf("job '%s' listed in order but not found in jobs map", jobName)
}
jobMap, ok := toStringAnyMap(raw)
if !ok {
    return nil, nil, nil, fmt.Errorf("job '%s' must be an object", jobName)
}
```

- **Pros:** distinct, actionable error messages; minimal change
- **Cons:** none

### Option B — Silently skip missing jobs with a warning log

- **Pros:** more lenient
- **Cons:** produces silently incomplete output; not consistent with project error philosophy

## Recommended Action

_To be filled during triage._

## Technical Details

- Affected file: `provider/github/unparse_emit.go`
- Function: `buildWorkflowJobIndex`

## Acceptance Criteria

- [ ] When a job name from `order` is absent in `jobs`, error message clearly states it was not found
- [ ] When a job name is present but wrong type, original "must be an object" error is preserved
- [ ] Existing tests pass

## Work Log

- 2026-03-31: Finding created during code review of job order preservation feature
