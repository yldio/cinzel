---
status: pending
priority: p3
issue_id: "051"
tags: [code-review, gitlab, output]
dependencies: []
---

# A pipeline that is only a spec block writes a trailing {}

## Problem Statement

The root document is appended unconditionally, so a pipeline holding nothing
but a `spec` block produces `spec: …` followed by `---` and `{}`.

## Findings

`provider/gitlab/pipeline_yaml.go:52`.

## Recommended Action

Skip the root document when the pipeline map is empty and at least one other
document was written.

## Acceptance Criteria

- [ ] A spec-only pipeline emits one document
- [ ] A pipeline with both a spec and jobs is unchanged
