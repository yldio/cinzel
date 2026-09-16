---
status: complete
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

- [x] A spec-only pipeline emits one document
- [x] A pipeline with both a spec and jobs is unchanged

## Technical Details

`provider/gitlab/pipeline_yaml.go` — `marshalPipelineYAML` now decides whether
the root document is worth writing:

```go
if len(root.Content) > 0 || len(docs) == 0 {
	docs = append(docs, &yamlv3.Node{Kind: yamlv3.DocumentNode, Content: []*yamlv3.Node{root}})
}
```

The second half of the condition keeps the existing behaviour for a pipeline
that is empty outright — it still writes `{}`, which a reader accepts, rather
than a zero-byte file.

`provider/gitlab/spec_only_test.go` — `TestSpecOnlyPipelineEmitsOneDocument`,
three cases: a spec alone (no `---`, no `{}`), a spec with a job (both
documents kept), and an empty pipeline (`{}` still written).

## Work Log

### 2026-09-16

Reproduced through the CLI. `/tmp/r51/spec.hcl` holding only a spec block
parsed to:

```yaml
spec:
  inputs:
    env:
      default: staging
---
{}
```

With the fix the same file writes just the spec document, and a file holding a
spec plus `job "build"` still writes both, the job in the second.

Proved the test fails without the fix by restoring the unconditional append:
the spec-only case fails on both assertions. Full sweep clean —
`go test -count=1 ./...`, `mise run lint`, `mise run drift`,
`mise run license-check`.
