---
title: "Document type auto-detection ordering for YAML unparse"
module: "GitHubProvider"
problem_type: "best_practice"
component: "unparse"
severity: "medium"
root_cause: "logic_error"
symptoms:
  - "Action YAML incorrectly classified as workflow"
  - "Step-only YAML incorrectly classified as action"
  - "Unparse produces wrong HCL block type"
tags:
  - "unparse"
  - "classification"
  - "auto-detection"
  - "workflow"
  - "action"
  - "step-only"
created_date: "2026-03-08"
updated_date: "2026-03-08"
---

## Problem Description

The unparse direction must auto-detect whether a YAML document is a workflow, an action, or step-only. The classification order matters because some documents could partially match multiple types.

## Root Cause

A YAML document with `name` and `runs` could be either a workflow step or an action. The classification must check for workflow-specific keys first, then action-specific keys, then fall back to step-only.

## Solution Implemented

The detection chain in `github.go` (`unparseYAMLFile`):

1. **Workflow**: `classifyWorkflowDocument()` — has `on` and/or `jobs` keys
2. **Action**: `classifyActionDocument()` — has `name` and `runs`, but NOT `on` or `jobs`
3. **Step-only**: fallback — `parseStepsFromYAML()`

```go
// Single unmarshal
doc, err := parseYAMLDocument(yamlBytes)

// Try workflow first (most specific)
workflowDoc, err := classifyWorkflowDocument(doc)
if workflowDoc != nil { return workflowToHCL(...) }

// Try action second
if actionDoc := classifyActionDocument(doc); actionDoc != nil {
    return actionToHCL(...)
}

// Fallback to step-only
steps, err := parseStepsFromYAML(yamlBytes)
```

## Prevention Guidance

- The `isActionDocument()` check explicitly excludes documents with `on` or `jobs` keys to avoid false positives.
- When adding new document types (e.g., reusable workflow definitions), add them to the chain BEFORE the step-only fallback.
- Always check the most specific type first (workflow has the most distinguishing keys), then progressively less specific.
- Test edge cases: a minimal document with just `name` and `runs` should classify as action, not workflow.
