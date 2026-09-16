---
status: complete
priority: p1
issue_id: "028"
tags: [code-review, correctness, gitlab, parse]
dependencies: []
---

# A job named "stages" deletes the pipeline's stages

## Problem Statement

The job loop runs after the reserved top-level keys are written into the same
map, and overwrites whatever is there. A job whose label is `stages`,
`variables`, `default`, `include`, `spec` or `image` destroys that key.

## Findings

`provider/gitlab/parse_pipeline.go:168`. With `stages = ["build"]` and a
`job "stages"` block the output was

```yaml
stages:
  script: [x]
  stage: build
```

The stage list is gone and the command exits 0.

## Recommended Action

Before assigning, refuse a job whose name is already a key in the pipeline map:
`job '%s' collides with pipeline keyword '%s'`.

## Technical Details

- `provider/gitlab/errors.go`: `errJobNamedAfterKeyword`
- The check is in the final `for name, job := range jobs` loop, against the
  live `pipeline` map rather than a hardcoded keyword list. That covers every
  key the pipeline actually holds, including templates and `spec`, and stays
  right as keys are added
- A job reaching the name through `id` is caught too: the loop runs over the
  written key, not the block label
- A keyword the pipeline does not use is not a collision. `job "stages"` in a
  file with no `stages` attribute parses, and GitLab reads it as a job

## Acceptance Criteria

- [x] A job named after a reserved key fails with a clear error
- [x] Ordinary job names are unaffected
- [x] The name reached through `id` is caught the same way
- [x] A keyword the pipeline does not use is still a valid job name

## Work Log

- 2026-09-16: Fixed. Checked against the live map rather than a keyword list,
  so the check does not go stale.
