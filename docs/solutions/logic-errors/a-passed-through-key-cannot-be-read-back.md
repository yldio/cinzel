---
title: "A passed-through top-level key produced HCL parse refused"
module: "provider/gitlab"
problem_type: logic_error
component: "provider/gitlab/config.go, provider/gitlab/parse_pipeline.go"
severity: high
root_cause: "unparse wrote an unknown top-level key as a bare attribute, and parseConfig had no ',remain' to read it back"
symptoms:
  - "unparse writes 'my_extra = \"kept\"' with a warning, at exit 0"
  - "the next parse of that same file fails with 'An argument named \"my_extra\" is not expected here', at exit 1"
  - "the two commands the README puts side by side do not compose"
tags:
  - "roundtrip"
  - "passthrough"
  - "schema"
status: fixed
created_date: "2026-09-25"
updated_date: "2026-09-25"
---

# A passed-through top-level key produced HCL parse refused

## The symptom

A GitLab pipeline carrying a top-level key cinzel has no block for:

```yaml
my_extra: kept
stages:
  - build
build:
  stage: build
  script:
    - make
```

Unparse writes it out as a bare attribute, warns, and exits 0:

```
warning: unsupported top-level key 'my_extra' passed through
```

```hcl
stages = ["build"]

job "build" {
  script = ["make"]
  stage  = "build"
}
my_extra = "kept"
```

Parsing that file back fails:

```
An argument named "my_extra" is not expected here.
```

Exit 1. The README documents exactly these two steps, in this order, as the way
to start from a pipeline you already run: unparse it, then "Edit the HCL, then
convert it back." So the failure lands on the second command of the documented
way in, for any pipeline holding one key outside the schema.

Every scalar shape reached it: a string, a number, a boolean and a list. Only a
map value did not, because that is read as a job and fails earlier with a
different error.

## The cause

`parseConfig` in `provider/gitlab/config.go` named every key it knew and had no
`hcl:",remain"`, so gohcl refused the file over the one attribute left. The
write direction had no matching gap: `pipelineToHCL` writes an unrecognised key
straight out, and `pipeline_yaml.go` already has a `remaining` loop that puts
such a key back into the YAML. Only the read side was missing.

Half of the guard was already there. `passthrough_key_test.go` refuses a key
that is not a valid HCL identifier, precisely because it "produced HCL nothing
can read back". A key that *is* a valid identifier produced HCL nothing could
read back either, and nothing stopped it.

## The fix

`parseConfig` gains the remain body, matching `hclReportsBlock` above it:

```go
Body hcl.Body `hcl:",remain"`
```

and `parsePassthroughAttrs` reads its attributes into the pipeline map.

Two things about that reader are load-bearing.

It uses `JustAttributes`, not a cast to `*hclsyntax.Body`. The remain body is
the same body, so the raw `Attributes` map still holds every key the schema
consumed; `JustAttributes` is the hidden-aware view and returns only what is
left. `parseGenericBodyMap` in the same file does take the raw cast, which is
correct there because a reports block has no schema to hide anything.

Its diagnostics are dropped. The only one it raises here is that the body still
holds blocks, which are the job, template and include blocks the schema took,
and returning that would refuse every pipeline that has a job.

A key already in the map is skipped. Nothing reaches that today, since the
schema's own keys are hidden by the time the reader runs, but the map is what
the YAML is built from and overwriting a parsed block with a raw attribute
would lose the block.

## Why this and not a refusal

Refusing the key on unparse was the other option, and it would have matched the
non-identifier guard. It was not taken because passthrough is deliberate: the
warning, the emitter's `remaining` loop and the last subtest of
`passthrough_key_test.go`, which asserts a valid identifier passes through with
no error, all say so. Reading the key back makes the existing behaviour honest
rather than reversing a decision.

The warning is left as it is. It now describes what happens.

## The test

`provider/gitlab/passthrough_roundtrip_test.go`. Four subtests over the value
shapes, and one that the schema's own keys are not read a second time. The four
were confirmed to fail with `parsePassthroughAttrs` removed, and the fifth to
pass both with and without it, which is what makes it a control rather than a
restatement. The HCL is also byte-stable over two unparse passes.
