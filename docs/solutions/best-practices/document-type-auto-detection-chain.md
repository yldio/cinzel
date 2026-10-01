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
updated_date: "2026-09-18"
---

> The chain is unchanged and still lives in `provider/github/github.go:240-257`.
> `classifyWorkflowDocument` (`unparse_workflow.go:193`) now asks
> `ghworkflow.NewYAMLDocument` whether the document is a workflow rather than checking
> keys itself, so the schema answers the question. `isActionDocument`
> (`unparse_action.go:32`) is still the key check the prevention section below names.

## Problem Description

The unparse direction must auto-detect whether a YAML document is a workflow, an action, or step-only. The classification order matters because some documents could partially match multiple types.

## Root Cause

A YAML document with `name` and `runs` could be either a workflow step or an action. The classification must check for workflow-specific keys first, then action-specific keys, then fall back to step-only.

## Solution Implemented

The detection chain in `github.go` (`unparseYAMLFile`):

1. **Workflow**: `classifyWorkflowDocument()` — has `on` and/or `jobs` keys
2. **Action**: `classifyActionDocument()` — has `name` and `runs`, but NOT `on` or `jobs`
3. **Step-only**: fallback — `parseStepsFromYAML()`, which now asks whether
   each value is shaped like a step (`looksLikeStep`, `io_helpers.go`) rather
   than only whether it is a mapping

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

> **Later.** Being the last link made the step-only path a catch-all: a step is
> a mapping, the decoder reads the keys it knows and ignores the rest, so any
> mapping of mappings converted. `something: {do: another}` became a step named
> "something" with `do` dropped, at exit 0. The fallback now requires every
> value to carry only keys a step declares and to set `uses` or `run`, and a
> file that reaches the end of the chain is warned about rather than skipped in
> silence.

## Prevention Guidance

- The `isActionDocument()` check explicitly excludes documents with `on` or `jobs` keys to avoid false positives.
- When adding new document types (e.g., reusable workflow definitions), add them to the chain BEFORE the step-only fallback.
- The step-only fallback is not a catch-all. It is the last link, so everything the earlier links refuse arrives there, and what it accepts it converts. It has to recognise a step rather than accept what is left.
- Always check the most specific type first (workflow has the most distinguishing keys), then progressively less specific.
- Test edge cases: a minimal document with just `name` and `runs` should classify as action, not workflow.
