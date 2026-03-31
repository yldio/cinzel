---
status: pending
priority: p3
issue_id: "005"
tags: [code-review, testing, performance]
dependencies: []
---

# `BenchmarkUnparseWorkflowInMemory` does not cover `yamlJobOrder`

## Problem Statement

The existing `BenchmarkUnparseWorkflowInMemory` calls `workflowToHCL` directly, bypassing `unparseYAMLFile` and therefore never invoking `yamlJobOrder`. The double-parse cost has zero benchmark coverage. If bulk unparse (e.g., AI assist processing dozens of workflow files) ever becomes a concern, there will be no baseline to detect regression.

## Proposed Solutions

### Option A — Extend existing benchmark to call `unparseYAMLFile` (Recommended)

Replace the direct `workflowToHCL` call in the benchmark with `unparseYAMLFile`, which exercises the full path including `yamlJobOrder`.

### Option B — Add a dedicated `BenchmarkYAMLJobOrder` micro-benchmark

Measures only the `yamlJobOrder` function to isolate the second-parse cost.

## Acceptance Criteria

- [ ] Benchmark covers the `yamlJobOrder` code path
- [ ] Baseline numbers are documented

## Work Log

- 2026-03-31: Finding created during code review of job order preservation feature
