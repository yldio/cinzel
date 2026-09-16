---
status: complete
priority: p1
issue_id: "027"
tags: [code-review, correctness, github, parse]
dependencies: []
---

# Two jobs with the same id silently become one

## Problem Statement

`parseHCLToWorkflows` writes each job into the map with
`jobs[jobContent.Key] = jobContent.Body` and never checks whether the key is
taken. Two `job` blocks carrying the same `id` produce one YAML job. The second
wins, nothing is reported, and the command exits 0.

## Findings

`provider/github/parse_workflow.go:119-137`. Verified by parsing a file with
two job blocks that both set `id = "same"`.

A second collision of the same class was found while fixing this one and is
covered here too: two `job` blocks sharing a *label* — `job "build"` declared
twice — overwrote one another in `parsedJobs` before any `id` was read, so the
first block was gone before the key check could see it.

`jobs = [job.first, job.first]` needed nothing: the existing reference
validation already rejects it.

## Recommended Action

Error before the assignment when the key already exists, naming both block
labels.

## Technical Details

- `provider/github/errors.go`: `errDuplicateJobLabel`, `errDuplicateJobKey`
- The label check sits at the top of the `for _, j := range cfg.Jobs` loop,
  before `parseJobConfig`, so it fires on the block that shadows
- The key check is per workflow, in the `workflow.JobRefs` loop: each workflow
  is its own file, so the same key reached from another workflow is not a
  collision. `TestTheSameJobKeyInTwoWorkflowsIsFine` pins that
- Both messages name the offenders, and the key message also names the key,
  which is what the file would have held

## Acceptance Criteria

- [x] Duplicate job id fails with a message naming the id
- [x] Distinct ids are unaffected
- [x] A duplicate block label fails, naming the label
- [x] The same key in two separate workflows still parses

## Work Log

- 2026-09-16: Fixed. Found the label collision alongside the reported id
  collision and closed both.
