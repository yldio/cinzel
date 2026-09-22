---
title: "A narrowed parse prunes the outputs it was not asked to look at"
module: "GitHub provider"
problem_type: "best_practice"
component: "provider/github, internal/fsutil"
severity: "medium"
root_cause: "design_tension"
symptoms:
  - "`github parse -f one.hcl` deletes the YAML another HCL file in the same project generated"
  - "A directory run writes cd.yaml and ci.yaml; a following single-file run leaves only cd.yaml"
  - "The deleted file carries the cinzel marker, so it is read as stale rather than as someone else's"
tags:
  - "prune"
  - "parse"
  - "output"
  - "github-actions"
created_date: "2026-09-19"
updated_date: "2026-09-19"
---

## What happens

`Parse` records every file the run wrote into `currentOutputs` and hands it to
`fsutil.PruneStaleGeneratedYAML`, which walks the whole output directory and
removes any `.yaml` or `.yml` carrying the provider's marker that the map does
not name.

A run narrowed with `-f` writes one file, so `currentOutputs` holds one entry,
and every other generated workflow in that directory is stale by definition:

```
$ cinzel github parse -d src --output-directory out
$ ls out
cd.yaml  ci.yaml

$ cinzel github parse -f src/cd.hcl --output-directory out
$ ls out
cd.yaml
```

`ci.yaml` is gone, and nothing about the second command mentioned it.

## Why it is not simply a bug

The obvious fix — prune only when the run covered the whole input — is already
contradicted by a test. `TestParsePrunesStaleWorkflowOutputs`
(`provider/github/github_test.go`) writes a stale marked file, runs `Parse`
with `File:` set, and requires that the stale file be deleted. Narrowing the
prune to directory runs turns that test red.

The two behaviours want opposite things from the same call:

- **A rename should clean up after itself.** `workflow "ci" { filename = "x" }`
  renamed to `"y"` has to take `x.yaml` with it, or the old file stays live
  forever. This is what the test pins, and it is the reason the prune exists.
- **A command should not touch what it was not pointed at.** `-f one.hcl` names
  one file, and deleting a sibling's output is a surprise no part of the command
  line asked for.

Both are reasonable. Neither can be read off the arguments alone, because a
single-file run and a whole-directory run produce the same shape of evidence:
a set of files written, and a directory holding more than that set.

## What would settle it

The missing input is ownership: which HCL file produced which YAML. The prune
works from the provider marker, which says "cinzel wrote this" and nothing
about who asked. A marker carrying the source path would let a narrowed run
prune only its own leftovers and leave every other file alone, and the rename
case would still work because the renamed file and its predecessor share a
source.

That is a change to the marker format, which is written into every generated
file and read back by `HasGeneratedMarker`, so it is not a local edit. It is
recorded here rather than done because nothing has yet needed it badly enough
to justify the migration.

## If you are here because a file disappeared

Run `parse` over the directory rather than the file. A directory run names
every output it should, so the prune has nothing to consider stale.

## Related

- [`parse-yml-flag-output-extension-control.md`](../developer-experience/parse-yml-flag-output-extension-control.md) — the other thing that changes which files a parse run claims
