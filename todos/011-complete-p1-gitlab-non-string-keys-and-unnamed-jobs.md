---
status: complete
priority: p1
issue_id: "011"
tags: [code-review, correctness]
dependencies: ["006", "008"]
---

# GitLab accepts keys it cannot represent, and jobs with no name

## Problem Statement

Todos 006 through 010 fixed four defect classes in the GitHub provider. The
GitLab provider carries the same code shapes and was never probed for them.
It turns out to share two of the four, and both corrupt output rather than
failing.

## Findings

Probed at c6a1553 by running each shape through `Unparse` and then back
through `Parse`.

**Non-string keys collapse.** goccy renders a key by its text, so `~`,
`null` and `NULL` are all the key `null`:

```yaml
variables:
  ~: a
  null: b
  NULL: c
```

emits `null = "c"` — two of the three values gone, no error. The emitted
HCL then reads back as the string `"null"`, so what comes out is not what
went in either. `true`, `123` and `1.50` behave the same way; `1.50`
additionally arrives as `"1.5"`.

A key that is a sequence is the one shape goccy does reject, with
`found an invalid key for this map`.

**An unnamed job is emitted and cannot come back.** `"": {stage: build}`
produces:

```hcl
job "job" {
  id     = ""
  ...
}
```

Running that through `Parse` fails with `'id' must be a non-empty string`.
`Unparse` writes a file the tool itself refuses.

The other two GitHub defects are not present here. Large whole numbers keep
their digits: `99999999999999999999` emits intact, because goccy does not
resolve them to float64 the way yaml.v3 does. There is no alias-in-key
misreading either: `*k` in key position resolves to the anchor's value and
the job is named correctly.

## Recommended Action

Reject both, matching what the GitHub provider already does.

The key check rides along with the yaml.v3 pre-pass added in 006, so it
costs no extra parse. A merge key is left alone: it is how a pipeline shares
a block between jobs.

The unnamed job is refused in `pipelineToHCL` rather than in
`validatePipeline`, which only runs in the parse direction and so never sees
the unparse path.

## Technical Details

- `provider/gitlab/unparse_pipeline.go`, `checkYAMLSoundness` and
  `rejectNonStringKeys`
- `provider/gitlab/unparse_pipeline.go`, `pipelineToHCL`, at the point the
  job identifiers are built

## Acceptance Criteria

- [x] A non-string key is rejected rather than silently collapsed
- [x] Quoted keys that look like scalars still work
- [x] Merge keys still work
- [x] An unnamed job is rejected in both directions

## Work Log

- 2026-09-15: Created and closed. Probed the GitLab provider for each of the
  four defect classes fixed in the GitHub provider by todos 006 to 010, and
  found two of them present. One test assertion written during this work was
  wrong rather than the code: quoted `"null"` and `"true"` keys round-trip
  correctly, HCL simply writes them unquoted on the way out.
