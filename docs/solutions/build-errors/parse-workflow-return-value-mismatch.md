---
title: "Build failure from parseHCLToWorkflows signature change missing return values"
module: "GitHubProvider"
problem_type: "build_error"
component: "parse_workflow"
severity: "high"
root_cause: "logic_error"
symptoms:
  - "Compile error: not enough return values"
  - "Build fails after adding ActionYAMLFile return type"
  - "Multiple files affected by single signature change"
tags:
  - "compile-error"
  - "return-values"
  - "signature-change"
  - "parse"
created_date: "2026-03-08"
updated_date: "2026-03-08"
---

## Problem Description

After adding composite action support, `parseHCLToWorkflows` was changed from returning 3 values to 4 values `([]WorkflowYAMLFile, map[string]any, []ActionYAMLFile, error)`. The build broke because 8+ error-path return statements still had 3 values.

## Root Cause

When adding a new return value to a function with many early-return error paths, it's easy to miss some. The function had returns scattered across nested conditionals and loop bodies.

## Solution Implemented

Updated every `return nil, nil, err` to `return nil, nil, nil, err` across all error paths in `parse_workflow.go`. Also updated callers in `github.go` and `benchmark_test.go` to accept the 4th value.

## Prevention Guidance

- When changing a function signature with many return sites, search for the old return pattern: `grep -n "return nil, nil, err" parse_workflow.go`
- The compiler catches these, but the error messages point to individual lines — easy to fix one and miss others.
- Consider whether extracting sub-functions could reduce the number of return sites before making the change.
- Run `go build ./...` immediately after the signature change, before making any other modifications.
