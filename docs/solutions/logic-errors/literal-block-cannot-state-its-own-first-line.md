---
title: "A run script indented with tabs became YAML no reader accepts"
module: "internal/yamldoc"
problem_type: "logic_error"
component: "internal/yamldoc"
severity: "high"
root_cause: "every multi-line string got a literal block, including the ones a block cannot represent"
symptoms:
  - "parse exits 0; unparse then refuses the file it wrote: found a tab character where an indentation space is expected"
  - "The refusal names cinzel's issue tracker, over a script the author wrote"
  - "A leading blank line is dropped silently instead, at exit 0 on both passes"
tags:
  - "yaml"
  - "literal-block"
  - "roundtrip"
created_date: "2026-09-22"
updated_date: "2026-09-22"
---

## Problem Description

`scalarNode` gave a literal block to every string holding a newline. A literal
block writes its content indented and takes the indentation from the first line,
so the first line is the one thing it cannot describe freely. Two shapes break
it:

**A first line starting with a tab.** The block writes two spaces and then the
author's tab. A YAML reader takes that as indentation, and a tab is not
indentation, so the whole document fails to parse:

```yaml
run: |-
  	echo one
  	echo two
```

`parse` exits 0 having written it. `unparse` then refuses it with
`found a tab character where an indentation space is expected`, and because the
error carries no user-input marking, it points the author at cinzel's issue
tracker over their own script.

**An empty first line.** The leading blank lines sit above the indentation
indicator that would have to describe them, so a reader drops them. That one
fails quietly: both passes exit 0 and the script is a line shorter.

A tab further down is fine — by then the indentation is already fixed — which is
why this looks like it works until someone indents line 1.

## The Fix

`literalBlockHolds` decides whether the string's first line is one a block can
state. If it is not, the value goes out double-quoted, which states the string
exactly and is merely less pleasant to read.

The change is deliberately about the first line only. Tabs elsewhere, spaces
anywhere, and trailing blank lines all keep their block, because a block holds
them correctly and rewriting them would churn every golden file for nothing.

## Verification

Two reverts, each compiling:

1. Always a literal block, the original behaviour: 5 of the 5 quoted-case
   subtests fail, three on `cannot be read back` and two on
   `came back changed`.
2. Never a literal block, a deliberately wrong fix that also makes the quoted
   cases pass: 3 control subtests fail on `lost its literal block`.

The second revert is why the control test exists. Without it, quoting every
multi-line string would pass the suite while rewriting every `run:` in the
repository.

End to end: an HCL step with `run = "\techo hi\n"` now parses to
`run: "\techo hi\n"`, unparses back to a heredoc holding the tab, at exit 0 on
both passes.
