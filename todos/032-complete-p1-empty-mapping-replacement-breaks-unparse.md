---
status: complete
priority: p1
issue_id: "032"
tags: [code-review, correctness, gitlab, output]
dependencies: []
---

# Rewriting ": {}" to ":" turns an empty mapping into null

## Problem Statement

`marshalPipelineYAML` does a byte replacement of `: {}\n` with `:\n` across the
whole document. An empty `include` or `variables` becomes a YAML null, and
unparse then refuses it.

## Findings

`provider/gitlab/pipeline_yaml.go:68`. `include {}` parses to `include: {}`,
is rewritten to `include:`, and unparse fails with
`include must be a string, object, or list`. Verified end to end.

## Recommended Action

Dropped the replacement. The collapsed form was not worth keeping: nothing
read it back, and the second half of the suggestion — teaching unparse to read
a nil as empty — would have made a real `include:` with no value indistinguishable
from an explicitly empty one.

## Technical Details

- `provider/gitlab/pipeline_yaml.go`, `marshalPipelineYAML`
- `internal/yamldoc` never did this rewrite, so GitHub was unaffected. Its
  `CONTEXT.md` recorded GitLab as still carrying the old rewrite; updated
- No fixture held an empty mapping, so no golden file moved
- An empty top-level `variables` is a separate matter: it is dropped on
  unparse because `parseVariableBlocks` writes the key only when a variable
  exists. A job-level `variables = {}` does survive, which is what the
  acceptance criterion covers

## Acceptance Criteria

- [x] An empty `include` block roundtrips
- [x] An empty `variables` block roundtrips (job level; see above)

## Work Log

- 2026-09-16: Reproduced through the CLI — `include {}` emitted `include:` and
  unparse rejected it with "include must be a string, object, or list".
  Dropped the replacement; `TestEmptyMappingStaysAMapping` covers both keys and
  fails without the change.
