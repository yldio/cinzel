---
title: "A script written |- or |+ comes back as |"
module: "provider/github"
problem_type: logic_error
component: "provider/github, provider/github/step"
severity: low
root_cause: "a heredoc holding exactly one trailing newline, and setAsHeredoc trimming the value's own before writing it"
symptoms:
  - "run: |- comes back as run: |, with a trailing newline the source did not have"
  - "run: |+ comes back as run: |, with its trailing blank lines gone"
  - "a folded > value comes back as one long line"
  - "exit code 0, and the second roundtrip is stable, so a diff shows it once and never again"
tags:
  - "block-scalars"
  - "roundtrip"
  - "heredoc"
status: "reported, not fixed"
created_date: "2026-10-26"
updated_date: "2026-10-26"
related:
  - docs/solutions/logic-errors/a-comment-above-a-fan-out-mapping-has-no-block.md
  - docs/solutions/logic-errors/literal-block-cannot-state-its-own-first-line.md
---

# A script loses its chomping indicator

YAML writes a multi-line string six ways. The indicator says whether the
newlines fold, and the chomping indicator says what happens to the ones at the
end:

```
|    "x\n"      clip: one trailing newline, however many were written
|-   "x"        strip: none
|+   "x\n\n"    keep: all of them
>    "x\n"      fold, clip
>-   "x"        fold, strip
>+   "x\n\n"    fold, keep
```

All six are valid YAML and GitHub Actions accepts all six.

## What survives

A literal value keeps its style anywhere cinzel writes a plain string:
`env`, `with`, `inputs`, `outputs`, a job `if`. Those become quoted HCL
strings holding the newlines, and the emitter writes any value with a newline
in it back as a literal block. A `|-` input default makes the trip exactly,
trailing newline and all.

## What does not

A step's `run` is the one value written as a heredoc
(`setAsHeredoc`, `provider/github/step/stepdecode.go`). A heredoc ends on its
closing marker's own line, so the value it holds always ends in exactly one
newline — which is `|` and nothing else. `setAsHeredoc` opens with
`strings.Trim(content, "\n")`, so `|-` and `|+` arrive at the same tokens `|`
does and come back as `|`.

A folded value is lost everywhere, not only in `run`. YAML folds `>` into a
single line as it reads the file, so by the time any of this code runs there
is no newline left in the value and nothing to say it was ever folded.

## Why it is reported rather than fixed

Nothing runs differently. GitHub writes `run` to a file and hands it to a
shell, and a shell reads a missing or extra final newline the same as a
present one — checked against bash with a trailing backslash continuation, a
comment as the last line, and a script containing its own heredoc. Output
identical in every case.

Fixing it means a marker comment carrying the style through the HCL, the way
`# cinzel-style:` was prototyped for the folded case, or giving up heredocs
for scripts and writing them as quoted strings with escapes. The first adds
syntax a user can delete or copy by accident; the second makes every script
unreadable in the HCL. Neither buys correct behaviour, only a tidier diff.

So `warnChompedRun` (`provider/github/unparse_workflow.go`) names each script
it rewrites, with its path, and the conversion goes on.

## The thing to watch

The heredoc path is lossy by design, and only `run` goes through it today.
Anything routed through it later — a GitLab `script`, say — inherits the same
normalisation, where a trailing newline may not be as harmless.
