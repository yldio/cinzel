---
status: complete
priority: p1
issue_id: "029"
tags: [code-review, correctness, gitlab, roundtrip]
dependencies: []
---

# An object need keeps the sanitized job name and cannot be reparsed

## Problem Statement

`remapJobRefs` maps a sanitized HCL identifier back to the original job key,
but only for `[]any` entries that are strings. A need written as an object
(`{job: build-app, artifacts: true}`) is a `map[string]any` and is skipped, so
the reference stays `build_app`.

## Findings

`provider/gitlab/parse_pipeline.go:623` (`parseNeedBlocks`) and `:1251`
(`remapJobRefs`). Unparsing
`{"build-app":…, "test":{needs:[{job:"build-app",artifacts:true}]}}` emits
`need { job = job.build_app }`; reparsing fails with
`job 'test' needs unknown job 'build_app'`. Any job name that needs sanitizing
plus an object-form need cannot roundtrip.

## Recommended Action

In `remapJobRefs`, when the entry is a `map[string]any`, remap its `"job"`
string through the same key table.

## Technical Details

- `remapJobRefs` now switches on the entry type. The string branch is
  unchanged; the map branch reads `"job"` and remaps through the same table
- `"extends"` shares the loop and takes the same branch, which costs nothing:
  an extends entry is always a string, so the map case never fires for it
- Reproduced and confirmed fixed through the CLI, not only in test

Out of scope, noticed while verifying: a **cross-project** need
(`job: other-job` with `project:`) also comes back as `other_job`. That is a
different mechanism — the name belongs to another pipeline, so there is no key
table entry to remap through, and `jobRefID` falls back to the sanitized form.
Fixing it needs the original name carried on the need itself, which is a schema
change rather than a remap fix.

## Acceptance Criteria

- [x] Object-form need to a hyphenated job roundtrips
- [x] String-form needs keep working

## Work Log

- 2026-09-16: Fixed. Added the failing case to the existing
  `TestNeedsAsObjects` table rather than a new file, and a matching
  string-form case so the branch that already worked stays pinned.
