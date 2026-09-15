---
status: complete
priority: p1
issue_id: "013"
tags: [code-review, security]
dependencies: []
---

# A workflow or action filename writes outside the output directory

## Problem Statement

`filename` is joined onto `--output-directory` and written without being
checked. A filename carrying `../` walks out of the directory the caller
asked for, and an absolute filename ignores it altogether. Nothing reports
either.

## Findings

Probed at 2c48b19 through the built CLI.

```
workflow "esc" {
  filename = "../../../../../../tmp/t4/pwned"
  ...
}

$ cinzel github parse --file in/w.hcl --output-directory /tmp/t4/out
RC=0
/tmp/t4/pwned.yaml      <- six levels above the output directory
```

An absolute filename lands under the output directory prefixed with the
whole path, `/tmp/t5/out2/tmp/t5/absolute-pwned.yaml`, which is not where the
caller asked for it either.

What counts as rooted, and as a separator, differs by platform, and the same
HCL is read on all of them. `/x` is not absolute on Windows, `C:x` is not
absolute anywhere, and `..\x` is one filename on Linux. A check written
against `filepath` alone answers for the machine running the tool rather than
for the file.

The action writer has the same shape at `provider/github/github.go:129`,
where the filename becomes a directory name, so `../../x` there creates the
directory as well as the file.

Reached from any HCL the tool is pointed at, which for this repository
includes HCL a contributor opens a pull request with, and for `cinzel assist`
includes HCL a model wrote.

GitLab is not affected: its output name is fixed, `.gitlab-ci.yml` on parse
and the input file's base name on unparse.

## Recommended Action

Fold both separators to `/`, then judge the result with `path` rather than
`filepath`, so the answer does not depend on where the tool runs. Refuse a
rooted name, a drive letter, and anything resolving above the output
directory. Allow a plain subdirectory: the action writer already puts every
action under its own folder, so banning the separator would break it.

## Technical Details

- `provider/github/io_helpers.go`, `checkFilenameStaysInside`
- called from `parseHCLWorkflows` and `parseHCLActions`, before anything is
  written

## Acceptance Criteria

- [x] `../`, a deep `../`, `sub/../../`, an absolute path, a bare `..`, a
      backslash parent and a drive-relative `C:x` are all refused, for both a
      workflow and an action, on every platform
- [x] A plain name, a subdirectory and a `./` prefix still work
- [x] A refused filename leaves no file behind

## Work Log

- 2026-09-15: Created and closed. Found while probing file-boundary handling
  after the terminal sanitiser from #57 was confirmed to cover GitLab errors
  as well, which it does: an ANSI escape in a job name reaches stderr as
  `\x1b` text, not as a raw control byte.
- 2026-09-15: The first fix used `filepath` throughout and passed on macOS
  and Linux while letting `/tmp/escaped` through on Windows, which CI caught.
  Rewritten to fold separators and judge on `path`, with the two Windows
  shapes added as cases.
