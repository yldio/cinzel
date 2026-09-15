---
status: complete
priority: p2
issue_id: "017"
tags: [code-review, correctness]
dependencies: ["016"]
---

# GitLab unparse reports success having converted nothing

## Problem Statement

Unparse returned nil whether it converted a pipeline or not. A run that read
no files, and a run that read files and found no pipeline in any of them,
both ended at exit 0 with an empty output directory. Nothing said so.

## Findings

Probed at 2477964. Both paths are reached by ordinary mistakes, not by
contrived input:

```
--directory with the pipeline one level down, --recursive left off   RC=0
--directory pointed at the wrong directory                           RC=0
--directory holding YAML that is not a pipeline                      RC=0
```

The GitHub provider refuses the same two inputs:

```
no YAML files found in input
error in file '...': not a valid type
```

GitLab's own Parse refuses an empty result too, with `errNoDefinitions`
already declared for it. Only Unparse was silent.

Two neighbouring shapes were probed at the same time and are sound:

- A nonexistent `--directory`. Already refused, by stat.
- The canonical `.gitlab-ci.yml` writing a hidden `.gitlab-ci.hcl`. It reads
  back through both `--file` and `--directory`, so it round-trips. Odd to
  look at, not a defect.

## Recommended Action

Refuse an input holding no YAML with `errNoYAMLFiles`, matching GitHub, and
refuse a run that found no pipeline with `errNoDefinitions`, which GitLab
already declares and uses on the Parse side.

The second flag has to be set when the pipeline is found rather than when
the file is written, or `--dry-run` fails on valid input: it converts the
pipeline and skips only the write it was told to skip.

## Technical Details

- `provider/gitlab/gitlab.go`, `Unparse`
- `provider/gitlab/errors.go`, `errNoYAMLFiles` added

`TestNonPipelineDocumentIsStillSkipped` asserted a lone non-pipeline file
exits 0. Skipping is still what happens to the document, and no HCL is
written for it, but on its own it is also the whole run. The test now
asserts both halves, and a second test covers a non-pipeline document beside
a real one, which is the case skipping exists for.

## Acceptance Criteria

- [x] Input holding no YAML is refused
- [x] YAML holding no pipeline is refused
- [x] A pipeline still converts, dry run included
- [x] A non-pipeline document beside a pipeline is still skipped, not fatal

## Work Log

- 2026-09-15: Created and closed. Found by checking whether the unparse
  direction in 016 had the same shape in the other provider.
