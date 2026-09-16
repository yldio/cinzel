---
title: "Embedded actionlint validation on parse"
status: deferred
date: 2026-09-16
---

# Embedded actionlint validation on parse

## What This Would Be

A `--lint` flag on `cinzel github parse`, off by default. When set, every
generated workflow document is run through actionlint in memory before it is
written. Errors are printed and the run exits non-zero with nothing written.

Deferred on 2026-09-16: investigated, viable, not wanted yet. This file records
what was measured so the next attempt does not start from zero.

## Why It Came Up

cinzel writes YAML that GitHub can reject. The duplicate step id fixed in
`685cf54` is one case: cinzel wrote the file and exited 0, and only actionlint
run afterwards said otherwise. Each such hole is currently closed one at a time
inside cinzel, duplicating checks a mature linter already has.

## What Was Verified

Against `github.com/rhysd/actionlint v1.7.12`, as a library, not a subprocess:

```go
l, _ := actionlint.NewLinter(io.Discard, &actionlint.LinterOptions{})
errs, err := l.Lint("w.yaml", content, nil)
```

`Lint` takes the bytes directly, so no temp file and no disk write. Run over a
workflow holding a duplicate step id and a bad context property it reports both:

```
10:13 step ID "same" duplicates. previously defined at line:8,col:13 [id]
13:24 property "nosuchthing" is not defined in object type {...} [expression]
```

The `# generated-by: cinzel` marker does not disturb it — those are comment
lines, and positions simply shift by two.

## Where It Would Go

`provider/github/github.go`, in the workflow loop, between
`fsutil.PrependGeneratedMarker` and `fsutil.WriteFile`. The bytes are already
in hand there.

Not in the action loop. actionlint only understands workflows; handed a
composite action it invents errors, because it reads the file as a workflow:

```
1:1 "jobs" section is missing in workflow [syntax-check]
2:1 unexpected key "description" for "workflow" section [syntax-check]
3:1 unexpected key "runs" for "workflow" section [syntax-check]
```

GitLab gets nothing. actionlint is GitHub-only.

## Cost

Roughly 30 lines in `github.go` plus a flag in `internal/command/config.go`.

The dependency footprint is the real price, and it is paid by everyone building
cinzel, not only by users who pass the flag. Fifteen new transitive packages:

```
github.com/bmatcuk/doublestar/v4 github.com/clipperhouse/uax29/v2
github.com/fatih/color github.com/mattn/go-colorable
github.com/mattn/go-isatty github.com/mattn/go-runewidth
github.com/mattn/go-shellwords github.com/robfig/cron/v3
go.yaml.in/yaml/v4 golang.org/x/sync golang.org/x/sys
```

`go.yaml.in/yaml/v4` would be a third YAML library in the tree, alongside
`gopkg.in/yaml.v3` and `goccy/go-yaml`.

## Licensing

No obstacle. actionlint and every new transitive dependency except three are
MIT; `golang.org/x/sync` and `golang.org/x/sys` are BSD-3-Clause;
`go.yaml.in/yaml/v4` is Apache-2.0, the same licence as cinzel.

All permissive, all compatible downstream into Apache-2.0. Weaker obligations
than `hashicorp/hcl/v2`, already a direct dependency under MPL-2.0.

What this would owe is attribution: MIT and BSD-3 both ask that the copyright
notice travel with binary distributions. cinzel has no `NOTICE` file today and
`mise run license-check` only checks headers on cinzel's own source, so no
tooling gate would catch the omission. That is a policy decision about the whole
dependency tree rather than something this change forces.

Not checked: whether `go.yaml.in/yaml/v4` ships a NOTICE file, which
Apache-2.0 section 4(d) would then require be carried.

## The Unsolved Part

Positions. actionlint reports line and column in the generated YAML. The user
wrote HCL. `10:13` points into a file they did not write, and under `--lint`
with a failing check, did not even get written.

Three ways out, none taken:

1. Report the YAML position as-is. Cheap, and close to useless — the user has
   to reason about output they cannot see.
2. Map YAML positions back to HCL. cinzel tracks no provenance through
   `parseHCLToWorkflows`, so this is the expensive piece, and it is expensive
   whether or not actionlint is involved.
3. Print the offending YAML snippet inline and skip mapping. A middle ground
   that is at least honest about what it shows.

## What Already Exists

`--dry-run` runs the same parse path and writes nothing, so cinzel's own checks
already report before any file is touched:

```
$ cinzel github parse --file x.hcl --dry-run
error in job 'build': two steps in one job write the same step id: 's1' and 's2' both write 'same'
exit 1
```

It differs in two ways from what `--lint` would give: it prints the generated
YAML to stdout on success, and it stops at the first error rather than
collecting them all. cinzel fails fast at every check in `parse_workflow.go`.

This repo also runs actionlint over `.github/workflows/*.yaml` in
`mise run lint`, so for cinzel itself an embedded copy adds nothing. The value
is for users who do not have actionlint installed.

## Note

Adopting this makes the `errDuplicateStepID` check from `685cf54` redundant.
actionlint catches the same case with a better message.
