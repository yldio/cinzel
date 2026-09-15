---
status: complete
priority: p2
issue_id: "023"
tags: [code-review, correctness, gitlab]
dependencies: ["022"]
---

# A template keyed "." writes an id the parse direction refuses

## Problem Statement

A hidden key is a template and its name is what follows the dot. A key of "."
leaves nothing, so the template was written with an empty id. The command
exited 0 and reading the file back failed.

## Findings

```yaml
".":
  script: [make dot]
test:
  extends: ["."]
  script: [make test]
```

```
$ cinzel gitlab unparse --file q3.yml --output-directory out
warning: unsupported top-level key '.' passed through
$ echo $?
0
```

```hcl
template "template" {
  id     = ""
  script = ["make dot"]
}
```

```
$ cinzel gitlab parse --file out/q3.hcl
error in template 'template': 'id' must be a non-empty string
exit status 1
```

The job loop in the same function already refuses its own unnamed case with
this error and says why in a comment. The template loop, twenty lines above
it, does not.

## Recommended Action

Refuse a hidden key with nothing after the dot, matching the job loop.

## Technical Details

- `provider/gitlab/unparse_pipeline.go`, the template loop in `pipelineToHCL`
- `errBlockIDNotString` is the error the job loop already returns

Reached by two shapes: the key on its own, and a job extending it. Both end
at the same loop, so one guard covers both.

Found while fixing 022, which is the same class in the needs path: a
reference or id written empty, exit 0, and refused on the way back.

## Acceptance Criteria

- [x] A template keyed "." is refused
- [x] The refusal reaches the case where a job extends it
- [x] A named template still unparses
- [x] An extends reference still follows the template's assigned label

## Work Log

- 2026-09-15: Found while probing the template path after 022. Fixed and
  closed.
