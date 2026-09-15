---
status: complete
priority: p1
issue_id: "022"
tags: [code-review, correctness, gitlab]
dependencies: []
---

# An empty needs entry writes HCL that cannot be read back

## Problem Statement

A "needs" entry naming no job was written as a reference with nothing after
the dot. The file was written, the command exited 0, and reading it back
failed. The YAML it came from is not recoverable from the HCL, so a pipeline
converted that way is lost unless the original is still around.

## Findings

Three shapes, all accepted on unparse:

```yaml
test:
  needs:
    - ""          # job.
    - job: ""     # job = job.
    - {}          # need {}
```

The first two produced:

```
job "test" {
  depends_on = [
    job.,
  ]
}
```

```
$ cinzel gitlab unparse --file p6.yml --output-directory out
$ echo $?
0
$ cinzel gitlab parse --file out/p6.hcl
An attribute name is required after a dot.
exit status 1
```

The third produced `need {}`, which parse refuses with "needs must contain
non-empty strings or objects naming a job" — the rule already exists on the
parse side and was simply not applied where the file is written.

"extends" in the same function refuses an empty entry with "extends entries
must be non-empty strings". The two branches sit about forty lines apart and
disagree.

## Recommended Action

Refuse an entry naming no job, matching what "extends" already does and what
parse already checks. Keep the reference resolving through the job's assigned
label rather than the sanitized name, so a need still follows a job whose name
HCL cannot spell.

## Technical Details

- `provider/gitlab/unparse_pipeline.go`, `writeJobBlock` needs branch and
  `writeNeedBlock`
- `provider/gitlab/validate.go:117` is the parse-side rule being mirrored

Cross-project and cross-pipeline needs carry a real job name, so neither is
affected. An empty need block is refused only when it names neither a job nor
an upstream pipeline.

A name that sanitizes to a non-empty identifier is still written, so
`needs: ["--"]` becomes `job.__`. That is a reference to a job the pipeline
does not declare, and parse rejects it with "needs unknown job", which is the
right error and already reported. Left alone.

## Acceptance Criteria

- [x] An empty needs name is refused, in all three shapes
- [x] A need naming a job is unaffected, plain and object form
- [x] Cross-project and cross-pipeline needs still unparse
- [x] A need follows the label its job was assigned

## Work Log

- 2026-09-15: Found by probing GitLab reference resolution after the pin
  work. Fixed and closed.
