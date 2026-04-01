---
name: Preserve job order in HCL→YAML parse
description: Fixed job ordering bug where HCL job list declaration order was lost during YAML conversion, causing alphabetical sorting instead of preserving source order
type: logic-errors
tags: [go, yaml, hcl, cinzel, job-order, map-ordering, github-actions]
components: [provider/github/parse_workflow.go, provider/github/workflow_yaml.go]
symptoms: [jobs output in alphabetical order instead of declared order, HCL jobs list reference order ignored during parse, roundtrip YAML→HCL→YAML changes job sequence]
---

## Problem

When parsing HCL to YAML, jobs declared as `jobs = [job.test, job.build, job.deploy]` appeared in the YAML output in alphabetical order (`build`, `deploy`, `test`) rather than the declared order (`test`, `build`, `deploy`).

## Root Cause

In `parse_workflow.go`, jobs were collected into a `map[string]any`. Go maps have non-deterministic iteration order. When the YAML serializer (`workflowMapNode` → `genericMapNode`) emitted the jobs, it called `sort.Strings(keys)` to get a stable but alphabetical output — losing the original declaration order.

The declaration order was available in `workflow.JobRefs` (a `[]string` extracted from `jobs = [job.a, job.b, ...]`), but was not threaded through to the serializer.

## Solution

**Sentinel key pattern**: store the ordered slice alongside the jobs map in the workflow body, extract it in the serializer, and write jobs in that order.

### Step 1 — Capture order in `parse_workflow.go`

After building the jobs map, store `jobsOrder` as a private sentinel alongside it:

```go
if len(workflow.JobRefs) > 0 {
    jobs := make(map[string]any)

    for _, jobID := range workflow.JobRefs {
        jobContent, exists := parsedJobs[jobID]
        if !exists {
            return nil, nil, nil, fmt.Errorf("error in workflow '%s': cannot find job '%s'", wf.ID, jobID)
        }
        jobs[jobID] = jobContent.Body
    }

    workflow.Body["jobs"] = jobs
    workflow.Body["jobsOrder"] = workflow.JobRefs  // ← new: preserve declaration order
}
```

### Step 2 — Extract sentinel in `workflowMapNode` (`workflow_yaml.go`)

At the top of `workflowMapNode`, extract and exclude the sentinel before iterating keys:

```go
func workflowMapNode(workflow map[string]any) (*yamlv3.Node, error) {
    node := &yamlv3.Node{Kind: yamlv3.MappingNode}
    seen := map[string]struct{}{}

    // "jobsOrder" is a private sentinel set by the parser to preserve the
    // HCL-defined job sequence; it must never appear in the YAML output.
    jobOrder, _ := workflow["jobsOrder"].([]string)
    seen["jobsOrder"] = struct{}{}

    for _, key := range workflowKeyOrder {
        value, ok := workflow[key]
        if !ok {
            continue
        }

        if key == "jobs" && len(jobOrder) > 0 {
            if jobsMap, ok := value.(map[string]any); ok {
                if err := appendOrderedJobsMap(node, jobsMap, jobOrder); err != nil {
                    return nil, err
                }
                seen[key] = struct{}{}
                continue
            }
        }

        if err := appendMappingPair(node, key, value); err != nil {
            return nil, err
        }
        seen[key] = struct{}{}
    }
    // ... remaining keys
```

### Step 3 — `appendOrderedJobsMap` helper

Writes jobs in declaration order, with sorted fallback for any jobs not in the order slice:

```go
func appendOrderedJobsMap(node *yamlv3.Node, jobs map[string]any, jobOrder []string) error {
    mapNode := &yamlv3.Node{Kind: yamlv3.MappingNode}
    seen := make(map[string]struct{}, len(jobOrder))

    for _, id := range jobOrder {
        v, ok := jobs[id]
        if !ok {
            continue
        }
        if err := appendMappingPair(mapNode, id, v); err != nil {
            return err
        }
        seen[id] = struct{}{}
    }

    // Append any jobs not in jobOrder, sorted for stability
    remaining := make([]string, 0, len(jobs)-len(seen))
    for k := range jobs {
        if _, ok := seen[k]; !ok {
            remaining = append(remaining, k)
        }
    }
    sort.Strings(remaining)
    for _, k := range remaining {
        if err := appendMappingPair(mapNode, k, jobs[k]); err != nil {
            return err
        }
    }

    keyNode := &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: "jobs"}
    node.Content = append(node.Content, keyNode, mapNode)
    return nil
}
```

## Test Added

Added a byte-exact snapshot test to `TestParseFormattingSnapshots` in `snapshot_format_test.go`:

```go
{
    name:       "job order preserved in parse direction",
    inputFile:  filepath.Join("testdata", "fixtures", "formatting", "workflow_parse_job_order.hcl"),
    outputFile: "workflow-parse-job-order.yaml",
    expected:   filepath.Join("testdata", "fixtures", "formatting", "workflow_parse_job_order.golden.yaml"),
},
```

The fixture defines 3 jobs in non-alphabetical order (`test`, `build`, `deploy`):

```hcl
jobs = [job.test, job.build, job.deploy]
```

The golden file verifies exact output order:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    ...
  build:
    ...
  deploy:
    ...
```

## Prevention

### Why byte-equality snapshots are required for order bugs

`assertYAMLSemanticEqual` (used in roundtrip tests) unmarshals both YAMLs into `any` and compares via `reflect.DeepEqual`. Go map comparison is order-insensitive, so a semantic test would **pass even if job order is wrong**. Only a byte-exact snapshot comparison catches ordering regressions.

**Rule**: any feature that preserves source order must have a byte-exact snapshot test (not a semantic one).

### Sentinel key pattern safety checklist

When using the sentinel key pattern:
1. Use a lowercase (unexported-style) key name — `jobsOrder` not `JobsOrder`
2. Mark the key as `seen` immediately in every serializer that touches the map
3. Add a comment explaining the sentinel's purpose
4. Write a test that verifies the sentinel key does NOT appear in any output

### Alternative: typed ordered map

For new code, consider an explicit ordered map type instead of the sentinel:

```go
type orderedJobs struct {
    order []string
    jobs  map[string]any
}
```

This makes the ordering requirement visible in the type system rather than implicit in a hidden key.

## Related

- [`yaml-map-key-order-lost-use-node-api.md`](yaml-map-key-order-lost-use-node-api.md) — same class of problem in the unparse direction (YAML→HCL), solved with the yaml.v3 Node API single-pass parse
- [`nondeterministic-map-iteration.md`](nondeterministic-map-iteration.md) — general Go map iteration ordering pitfalls
