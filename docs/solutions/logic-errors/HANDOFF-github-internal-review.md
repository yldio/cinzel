---
title: "Handoff: unparse writes a null parse cannot read, and what is left to review"
module: "provider/github"
problem_type: logic_error
component: "provider/github, internal"
severity: high
root_cause: "the attribute emit funnel has no nil guard, where its gitlab twin does"
symptoms:
  - "unparse writes 'defaults = null' or 'strategy = null', and the next parse refuses the file it just wrote"
  - "Eight more job and workflow keywords write a null parse silently drops, so HCL pass 1 and pass 2 differ"
  - "Exit code 0 throughout"
tags:
  - "null"
  - "roundtrip"
  - "handoff"
  - "review"
status: "part done — the null finding, the reader question and provider/github are closed, internal/ is still open"
created_date: "2026-09-20"
updated_date: "2026-09-20"
---

# Handoff: internal/ and provider/github

> The null finding below is fixed, in the commit that adds
> `docs/solutions/logic-errors/github-unparse-writes-a-null-parse-cannot-read.md`.
> Read that note rather than this section: the scope turned out wider than the
> table here (`continue-on-error` too, and the step `env`/`with` case, which
> goes the other way), and dropping the nulls exposed a second fault the table
> could not show. Everything under "Probed and clean" and "Not looked at" still
> stands.

This is a work-in-progress note, not a solution. It records what a probe of
`provider/github` found and what has not been looked at, so the next session
starts from the findings rather than from scratch.

## The criterion

Two user instructions govern what counts as a defect here, both still in force:

> are you doing validations? that is the job for actionlint, you're just
> concern to convert hcl to yml and viceversa

> just keep what makes sense for the domain of cinzel, don't try to do what
> actionlint can provide

The operative test: **a defect is cinzel's only if cinzel writes output its own
parser rejects, or a roundtrip loses information.** "GitHub would reject this"
is actionlint's job, not cinzel's. Do not add schema validation.

## What has been reviewed

`provider/gitlab` was swept in the session before this one. Four defects were
found and fixed, in `d1ff759`, `b8d3260`, `98ceea5` and `8c2acee`. Read those
commits first — the last one is the direct parallel to the finding below.

`internal/` and `provider/github` were swept in an earlier session, before the
four gitlab defects were known. Neither has been re-probed against the shapes
those defects turned out to have. That is the gap this note exists to close.

## Finding: unparse writes nulls, two of which parse cannot read

The same shape as `8c2acee` in gitlab, and worse in two cases.

`writeCommentedAttribute` in `provider/github/unparse_workflow.go:506` is the
single funnel every attribute goes through. The gitlab twin, at
`provider/gitlab/unparse_emit.go:41`, opens with:

```go
if raw == nil {
    return nil
}
```

The github one has no such guard, so a YAML key written with no value becomes
`<attr> = null` in the HCL.

Reproduced with:

```yaml
name: CI
on:
  push:
    branches: [main]
jobs:
  build:
    runs-on: ubuntu-latest
    timeout-minutes:
    if:
    steps:
      - name: Test
        run: go test ./...
```

Unparse writes `if = null` and `timeout_minutes = null`. Parse then drops both,
so HCL pass 1 and HCL pass 2 differ by two lines carrying nothing. Confirmed by
`unparse -> parse -> unparse` and diffing the two HCL files.

**Two of them are worse: the HCL does not parse at all.** `defaults` and
`strategy` are blocks in the schema, not attributes, so writing them as an
attribute produces HCL cinzel refuses to read:

```
An argument named "defaults" is not expected here.
Did you mean to define a block of type "defaults"?
```

That is unparse writing a file its own parse cannot open, at exit 0.

Scope, established by probe, one keyword per run:

| Level | Keyword | Result |
| --- | --- | --- |
| job | `defaults` | writes null, **HCL unreadable** |
| job | `strategy` | writes null, **HCL unreadable** |
| job | `timeout-minutes`, `if`, `name`, `concurrency`, `container`, `environment`, `permissions` | writes null, parse drops it |
| job | `env` | refused on unparse, correctly |
| workflow | `defaults` | writes null, **HCL unreadable** |
| workflow | `permissions`, `concurrency`, `run-name` | writes null, parse drops it |
| step | every keyword tried | refused on unparse, correctly |

Step level is already safe. The typed decode rejects a null before it reaches
the emitter, which is why `env:` and every step keyword error out.

### Before fixing

Check whether GitHub reads any of these nulls as "clear what this would
otherwise inherit", the way GitLab does for `rules`, `artifacts`, `cache`,
`services`, `only`, `except` and `needs`. In gitlab those seven keep their
nulls deliberately, pinned by `provider/gitlab/null_collection_test.go` and
decided in `3a15b2c`. If GitHub has an equivalent, the fix is not a blanket
guard. If it does not, the guard matches gitlab and the two providers stop
differing for no reason.

### Testing the fix

`provider/gitlab/null_not_written_test.go` is the model — copy its shape.

Two traps, both of which produced a green test against broken code while the
gitlab fix was being written:

1. **hclwrite pads attribute names to a common width within a block.** `image
   = null` can carry two spaces. `strings.Contains(hcl, "image = null")` misses
   it. Normalise first: `strings.Join(strings.Fields(hcl), " ")`.
2. **Neuter the fix and confirm the new test fails** before believing it. A
   test that asserts nothing looks identical to a test that passes.

Add a two-pass stability test as well: unparse, parse, unparse, require the two
HCL files byte-identical. That is what caught the residue in gitlab.

Note one unrelated difference the two-pass diff also shows: pass 2 gains an
empty `permissions {}` block that pass 1 has none of, because parse writes
`permissions: {}` into the YAML by default. Decide separately whether that is
wanted; it is not part of this finding.

## Probed and clean

- Standard YAML tags (`!!str`) convert.
- An unknown custom tag (`!custom [a, b]`) is refused, not corrupted. This is
  the shape that was broken in gitlab (`d1ff759`); github already handles it.
- An unknown `%FOO` directive is refused before any conversion, so the
  two-reader hole fixed in `b8d3260` does not appear here. ~~**Worth a deeper
  look**: gitlab's hole was that one reader accepted a document the other
  refused, and only the refusal path was checked. Whether github's single
  goccy-plus-yamlv3 split has the same asymmetry was not established.~~ Looked
  at. It does not have that asymmetry — `parseYAMLDocument` returns its error
  and the caller stops, and all three of `b8d3260`'s shapes are refused at exit
  1. The split is elsewhere: the step-only path reads the raw bytes a second
  time and loses digits doing it. See
  `github-step-only-path-loses-digits.md`.
- No warning in `provider/github` or `internal/` makes a claim about the input
  the way the gitlab template warning did (`98ceea5`). The ones in
  `internal/pin` and `internal/command` report a failure that did happen.

## Not looked at

Nothing below has been probed in this session. Listed by size so the next
session can pick.

`internal/`, about 12k lines over 13 packages:

- `internal/command` (3164) — flag and config handling. `.cinzelrc.yaml`
  parsing was reviewed in `20f8a92`; the rest was not.
- ~~`internal/pin` (2631) — network-facing, the GitHub API resolver, upgrade
  logic. Largest untouched surface.~~ The rewrite path was probed over 21 shapes
  through `PinFile` and `UpgradeFile`: comment forms, attribute order, nesting
  depth, subdirectory and local actions, branch and short-SHA versions,
  interpolated actions, and a missing `version`. Two findings, both in
  `trailingCommentEnd` — see
  `pin-trailing-comment-stops-at-the-first-one.md`. One question left open
  rather than decided: the rewrite replaces the whole trailing comment, so an
  author's own note on the version line is lost, and it is not clear whether
  that note should be kept beside the tag or is correctly treated as belonging
  to the version being replaced. The resolver, the cache and the API error
  paths were read but not probed: they need a server stub this session did not
  build.
- `internal/ai` (1678) — assist, prompt construction, HCL stripping.
- ~~`internal/hclparser` (1925) — note `internal/hclparser/errors.go` has zero
  `UserInput` markings, where every other errors.go marks most of them. Some of
  those errors are certainly the author's fault and are sending people to the
  issue tracker over their own typo, which is exactly what `20f8a92` fixed in
  `internal/command`. Cheap win, probably.~~ Done, and not cheap: the marking
  was only half of it, because `provider/github/step/stepparse.go` wrote the
  open-an-issue line in by hand at twelve sites and put back what the mark took
  off. Probing it also turned up a template holding one interpolation being
  refused where the same reference written bare resolves. See
  `hclparser-typos-sent-to-the-issue-tracker.md`. The rest of the package was
  not read.
- `internal/fsutil` (753), `internal/yamldoc` (617), `internal/yamlwriter`
  (358), `internal/hclcomment` (195), `internal/naming` (199),
  `internal/cinzelerror` (513), `internal/unescape` (93), `internal/maputil`
  (92), `internal/test` (90).

`provider/github`, 5510 lines plus four subpackages (`action/`, `job/`,
`step/`, `workflow/`):

- ~~`parse_workflow.go` (1490) is the largest single file and was not read.~~
  Done. Probed clean on every non-comment shape tried — duplicate env keys and
  duplicate permissions blocks are both refused, `depends_on`, matrix with
  `variable` blocks, include/exclude, container, defaults, services, outputs,
  reusable workflows, concurrency and empty collections all roundtrip stable.
  One finding, in the comment path: the step and job readers each read half of
  an env or with block's comments. See
  `github-env-block-comments-read-by-halves.md`.
- `unparse_workflow.go` (765) was read only around the emit funnel, and now the
  foot writer — see the note above.
- `validate.go` (572) — remember the criterion above before touching it.

## Method that worked

Stub or neuter the guard, run the roundtrip, see whether HCL -> YAML -> HCL
returns what it started with. Then neuter the *fix* to confirm the new test
actually fails, and restore. Build with `go build -o /tmp/cinzelbin .` and probe
through the CLI rather than through unit tests: several of the findings above
only appear end to end.

One keyword per run. A probe file with several null keywords in it stops at the
first one that errors and tells you nothing about the rest.
