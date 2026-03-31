---
status: pending
priority: p2
issue_id: "003"
tags: [code-review, correctness, quality]
dependencies: [001]
---

# Jobs in `jobs` map but absent from `order` are silently dropped

## Problem Statement

In `buildWorkflowJobIndex`, when `order` is non-empty, only the jobs named in `order` are included in `entries`. If `yamlJobOrder` returns a subset of job names (parser inconsistency, partial parse, future anchor expansion adding synthetic keys), the remaining jobs are silently omitted from the HCL output. No error is raised and no warning is emitted. The user gets a silently incomplete conversion.

## Findings

- `provider/github/unparse_emit.go:21-47`
- `jobNames := order` — if `order` is shorter than `jobs`, extra jobs are dropped
- No check `len(entries) == len(jobs)` after the loop
- Silent data loss is especially dangerous in a YAML→HCL converter where the output may be used as the authoritative source

## Proposed Solutions

### Option A — Validate completeness after the loop (Recommended)

After the loop, check that all jobs in the map were covered:

```go
for jobName := range jobs {
    if _, seen := jobIDMap[jobName]; !seen {
        return nil, nil, nil, fmt.Errorf("job '%s' is in the workflow but was not included in the job order", jobName)
    }
}
```

- **Pros:** surfaces inconsistency immediately; no silent data loss
- **Cons:** fails rather than falling back — appropriate for a converter tool

### Option B — Append uncovered jobs at the end

For jobs in `jobs` but not in `order`, append them after the ordered ones using `sortedKeys` for determinism.

- **Pros:** no data loss; graceful degradation
- **Cons:** output order may not match source; unclear contract

### Option C — Log a warning and continue

- **Pros:** lenient
- **Cons:** silent in non-verbose mode; data loss still occurs

## Recommended Action

_To be filled during triage._

## Technical Details

- Affected file: `provider/github/unparse_emit.go`
- Function: `buildWorkflowJobIndex`

## Acceptance Criteria

- [ ] If `order` is non-empty and a job in `jobs` is not in `order`, the function returns an error or appends the job (per chosen option)
- [ ] No silent data loss in HCL output
- [ ] Existing tests pass

## Work Log

- 2026-03-31: Finding created during code review of job order preservation feature
