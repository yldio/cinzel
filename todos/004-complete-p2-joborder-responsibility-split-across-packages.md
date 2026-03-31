---
status: pending
priority: p2
issue_id: "004"
tags: [code-review, architecture, quality]
dependencies: [001]
---

# `JobOrder` responsibility is split: field on struct but set at call site

## Problem Statement

`YAMLDocument.JobOrder` is a field on the struct in `provider/github/workflow/yaml_document.go`, but `NewYAMLDocument` never sets it. It is set at the call site in `unparseYAMLFile` after the fact:

```go
workflowDoc.JobOrder = yamlJobOrder(yamlBytes)
```

This means:
1. Any other caller of `NewYAMLDocument` will silently get `JobOrder == nil`
2. The struct is partially constructed by its constructor — violating encapsulation
3. The contract is not enforced or documented

## Findings

- `provider/github/workflow/yaml_document.go:37` — `NewYAMLDocument` returns `YAMLDocument` without `JobOrder`
- `provider/github/github.go:209` — `workflowDoc.JobOrder = yamlJobOrder(yamlBytes)` set externally
- `provider/github/unparse_workflow.go:87` — `workflowToHCL` reads `doc.JobOrder` expecting it to be set

## Proposed Solutions

### Option A — Pass order as parameter to `workflowToHCL` instead of a struct field (Recommended)

Remove `JobOrder` from `YAMLDocument` entirely. Change `workflowToHCL` signature:

```go
func workflowToHCL(doc ghworkflow.YAMLDocument, filename string, jobOrder []string) ([]byte, error)
```

At the call site:
```go
return workflowToHCL(*workflowDoc, baseName, yamlJobOrder(yamlBytes))
```

- **Pros:** explicit data flow; `YAMLDocument` is a pure data struct; no surprise nil fields
- **Cons:** `workflowToHCL` signature changes (internal function, no external callers)

### Option B — Accept `order []string` in `NewYAMLDocument`

Thread the raw bytes or pre-computed order into `NewYAMLDocument`.

- **Pros:** constructor is complete
- **Cons:** `NewYAMLDocument` is in a different package and takes `mapper func` — adding a raw bytes param is awkward

### Option C — Document the two-step construction with a comment

Keep current code, add a comment on the field explaining it must be set by the caller.

- **Pros:** minimal change
- **Cons:** invisible contract; easy to violate silently

## Recommended Action

_To be filled during triage._

## Technical Details

- Affected files: `provider/github/workflow/yaml_document.go`, `provider/github/github.go`, `provider/github/unparse_workflow.go`

## Acceptance Criteria

- [ ] `YAMLDocument` is fully constructed in one place, or `JobOrder` is not a struct field
- [ ] No caller can inadvertently get `JobOrder == nil` when order was available

## Work Log

- 2026-03-31: Finding created during code review of job order preservation feature
