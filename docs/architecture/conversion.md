# Conversion

Parse is HCL → YAML, unparse is YAML → HCL. Providers live in
`provider/github` and `provider/gitlab`.

## Expressions

HCL writes `$${{ }}` for what GitHub Actions reads as `${{ }}`. The parser
strips the leading `$` and `unparse_emit.go` puts it back. Never escape by
hand, and never double-escape.

## Document detection on unparse

- `on` + `jobs` → workflow
- `name` + `runs` → action
- neither → step-only

The chain runs in `provider/github/github.go`; the workflow case is
`classifyWorkflowDocument` (`provider/github/unparse_workflow.go`), which
delegates to `ghworkflow.NewYAMLDocument`.

## Output paths

- actions → `<dir>/<name>/action.yml`
- workflows → `<dir>/<name>.yaml`

## Defaults applied on parse

- a workflow with no permissions gets `permissions: {}`
- a step with no `id` gets one from its block label, unless `ignore_id` is set

## Schema

The HCL schema lives only in `provider/<name>/config.go`. No ad-hoc key maps,
and no second copy of the schema inside validation. `hcl:",remain"` is for
intentional pass-through only.

YAML validation decodes into a typed struct with `goccy/go-yaml`, strictly.
Not an allowlist of keys.

Adding a field means three edits, in order: structs, conversion, tests.

## YAML output

Output is built through `internal/yamldoc`, an ordered, comment-carrying
document. Key order, inline comments and the difference between a collapsed
empty map and an explicit one are properties of the document, so the encoder
never recovers them by rewriting encoded bytes.

Double quotes only. Single quotes break golden tests, because some editors
rewrite them to double on save.

What gets quoted is `needsQuoting` in `internal/yamldoc/encode.go`: the empty
string and `~`, the YAML 1.1 bool and null words in any case, anything that
parses as a number, anything with leading or trailing whitespace, and the YAML
special characters. `@` is not one of them, so `actions/checkout@v4` stays
bare. GitLab keeps its own `stringNeedsQuoting`
(`provider/gitlab/pipeline_yaml.go`) because the shared rules would quote a
large share of real pipelines.

Workflow key order is `workflowKeyOrder` (`provider/github/workflow_yaml.go`):
name, run-name, on, permissions, env, defaults, concurrency, jobs, then
everything else sorted. `jobs` goes last, the way a hand-written workflow
reads: the short top-level keys first, then the long tail.

Job order inside `jobs` is the order they were declared in HCL, carried as a
`jobOrder` slice from `parseHCLToWorkflows` into `marshalWorkflowYAML`. An
empty order sorts them.

## Comments

Comments travel with values through the pipeline in an `annotated` wrapper, and
`plain` (`provider/github/workflow_yaml.go`) strips them at that one boundary,
so nothing outside the emitter sees a wrapper. Head, trailing and foot
positions are all carried, so a comment above an attribute, one beside it and
one closing a body each land where they were written.

See `docs/solutions/logic-errors/hcl-inline-comment-propagation-to-yaml.md` for
why it is a wrapper rather than a side map.
