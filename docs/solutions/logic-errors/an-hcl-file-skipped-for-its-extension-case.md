---
title: "An HCL file in a directory was skipped for the case of its extension"
module: "internal/fsutil"
problem_type: logic_error
component: "internal/fsutil, provider/github, provider/gitlab"
severity: high
root_cause: "the directory walk compared the extension exactly where every neighbouring path folds case"
symptoms:
  - "gitlab writes a pipeline missing one input file's jobs, at exit 0"
  - "github deletes the YAML it generated for that file, as stale, at exit 0"
  - "the same file read without complaint when named with -f"
tags:
  - fsutil
  - parse
  - prune
status: fixed
created_date: "2026-09-25"
updated_date: "2026-09-25"
---

# An HCL file in a directory was skipped for the case of its extension

## What happened

A directory holding `a.hcl` and `b.HCL`:

```
$ cinzel gitlab parse -f in --output-directory out
$ cat out/.gitlab-ci.yml
stages:
  - build
compile:
  script:
    - echo a
  stage: build
```

The `link` job that `b.HCL` declares is not there, and nothing was said about
it. Exit 0.

GitHub is worse, because its output directory is pruned. Run it once with
`alpha.hcl` and `beta.hcl`, rename the second to `beta.HCL`, run it again:

```
$ ls out
alpha.yaml
beta.yaml
$ mv in/beta.hcl in/beta.HCL && cinzel github parse -f in --output-directory out
$ ls out
alpha.yaml
```

`beta.yaml` is gone. The prune walks the output directory and removes any
generated file the run did not write, and the run no longer reads the file that
produces it. A case-only rename of an input deletes its output.

## Why

`ParseHCLInput` walks a directory with:

```go
if filepath.Ext(current) != ".hcl" {
    return nil
}
```

Nothing around it compares that way. `ListFilesWithExtensions`, thirty lines
down in the same file and the walker the unparse direction uses, lower-cases
both sides. `UniqueOutputName` folds case, and says why in its own comment.
`claimFilename` and `checkOutputPaths` in `provider/github/io_helpers.go` fold
case, and both say why: a name differing only in case is the same file on macOS
and Windows.

`ParseHCLInput`'s own single-file branch checks no extension at all, so
`-f beta.HCL` reads the file the directory walk refuses to see. The same file,
the same tool, two answers depending on how it was reached.

The skip is silent by construction. Returning `nil` from a `WalkDir` callback
means "not interesting", which is the right answer for a `.md` sitting in the
directory and the wrong one here.

## The fix

Fold the comparison, matching the neighbours:

```go
if !strings.EqualFold(filepath.Ext(current), ".hcl") {
    return nil
}
```

`EqualFold` rather than `strings.ToLower`, because the comparison is against a
constant and the result is not kept.

## Prevention

The tell was available without a probe: one file holding two extension
comparisons, written two different ways, for two directions of the same
conversion. A predicate that decides whether input is seen at all deserves the
same reading in every place it is asked.

The silence is the other half. A walk that answers "not interesting" cannot
distinguish a file the user did not mean from one it failed to recognise, and
the second is invisible in exactly the runs where it matters.

## Tests

`internal/fsutil/hcl_extension_case_test.go` puts `upper.HCL`, `mixed.Hcl` and
`lower.hcl` through `ParseHCLInput` and requires each to appear in the returned
sources. The third is the control: it passes with or without the fix.
