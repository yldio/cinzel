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

## Unparse: what is converted and what is skipped

The detection chain classifies each YAML file as a workflow, an action or a set
of steps, in that order. A file matching none of the three is skipped with a
warning on stderr, and a run that converts nothing at all ends non-zero.

Skipping and failing are different outcomes, deliberately. A file that is none
of the three is passed over and the run goes on: cinzel never had it, so nothing
is lost. A file that is recognised and then fails to convert stops the run, and
the run writes nothing at all — not the files that converted before it, not the
ones after.

A skip is named on stderr in both providers. GitLab stayed silent about it for
longer, which made a directory holding one pipeline beside four other YAML files
report what a full conversion reports.

That is why every file is converted into memory before any of them is written,
in both providers. Writing inside the conversion loop meant a failure left the
files ahead of it on disk and never read the ones behind it, so which half of a
directory survived was decided by where the failing file sorted. Every failure
is collected and reported together, rather than the first one the run met.

Step-only is the last link, so every file the earlier two refuse arrives there:
a dependabot config, an issue template, anything else sharing the directory. It
recognises a step rather than accepting what is left — every value has to be a
mapping, and every key in it has to be one a step declares. Without that, any
mapping of mappings converted, with the keys no step declares silently dropped.

The test stops there. A step that sets neither `uses` nor `run` runs nothing,
but a step-only file is cinzel's own library of step definitions rather than
something GitHub reads, and parse writes such a step whenever a block carries
only a name. Requiring one of the two refused cinzel's own output, and since the
chain judges the document as a whole, one such step took every working step
beside it down with it.

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

## Generated markers and pruning

Every YAML file parse writes opens with cinzel's own header:

```yaml
# generated-by: cinzel
# cinzel-provider: github
# cinzel-source: src/ci.hcl
```

The header is load-bearing for a delete. After writing, `Parse` calls
`fsutil.PruneStaleGeneratedYAML`, which walks the output directory and removes
generated files the run did not write — how a renamed workflow's old output
gets cleaned up. Two things keep that walk from deleting more than it should.

The provider line makes the file cinzel's. A file without it was written by
hand and is never touched, whatever its name.

The source line makes the file some particular HCL file's. The prune deletes a
marked file only when the source it records was among the files the run read,
so a run narrowed with `-f` cleans up after its own input and leaves every
other file's output alone. A rename is still cleaned up: the file that renamed
it was read. Paths are spelled by `fsutil.SourceKey` — relative to the working
directory with forward slashes, absolute for a file outside it — because both
sides of that comparison have to spell one file one way.

A file with no source line is pruned on the provider line alone. That is what
every file written before the line existed looks like, so they prune as they
always did and the line appears as files are regenerated.

Three cases record no source. GitLab writes one pipeline built from every file
read, so there is no single file to name, and nothing there prunes. A step-only
parse collects steps from across the input the same way. Neither has a block to
ask.

`fsutil.WithoutGeneratedMarker` strips all three lines on the way back, or
unparsing a generated file would copy cinzel's note into the HCL — and the
source line names the very file it would be written into.

See `docs/solutions/best-practices/a-narrowed-parse-prunes-what-it-did-not-look-at.md`.
