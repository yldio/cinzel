---
title: "A backslash in a script was read as an escape cinzel had written"
module: "internal/unescape"
problem_type: logic_error
component: "internal/unescape, provider/github, provider/gitlab"
severity: high
root_cause: "parity tells a literal backslash from an escape, and only a double-quoted scalar doubles one"
symptoms:
  - "a script line reading printf \\u00e9 comes back as printf é"
  - "a Windows path C:\\users\\u00e9dir loses six characters"
  - "exit code 0 on both passes, and no warning"
tags:
  - unicode
  - roundtrip
  - escaping
status: fixed
created_date: "2026-09-22"
updated_date: "2026-09-22"
---

# A backslash in a script was read as an escape cinzel had written

## What happened

This `.gitlab-ci.yml`:

```yaml
build:
  script:
    - 'printf \u00e9 now'
```

goes to HCL correctly, and comes back as:

```yaml
build:
  script:
    - printf é now
```

Six characters the author typed are gone, replaced by the one they name. The
same happens to a GitHub `run:` block, where the script becomes an HCL heredoc:
`printf X\u00e9Y` arrives back as `printf XéY`. Both passes exit 0 and neither
warns. A pipeline that ran `printf` with an escape now runs it with a literal,
and a job that copied `C:\users\u00e9dir` copies somewhere else.

## Why

`unescape.Unicode` exists because hclwrite and yaml.v3 escape characters that
are printable in their own output formats: hclwrite escapes anything Go's
`unicode.IsPrint` rejects, which takes in U+200D, and yaml.v3's printable check
only reads three-byte UTF-8, so every emoji above the BMP is escaped too. The
pass runs over finished output and puts those escapes back, which is what keeps
`👮‍♂️ Lint` from becoming `\U0001F46E‍♂️ Lint`. See
`../runtime-errors/unicode-emoji-zwj-escape-roundtrip.md` for how that started.

Running over finished output means it meets escapes it wrote and text the author
wrote, and it has to tell them apart. It did that by counting backslashes: a
writer doubles a literal backslash, so an even run before `uXXXX` is text and an
odd run is an escape.

That is true of a double-quoted scalar and of nothing else. A plain YAML scalar
leaves a backslash single, a single-quoted one does too, and an HCL heredoc has
no escapes at all, so it cannot double anything either. Text in any of those
three arrives at the parity of an escape, and the pass decodes it.

The shape is easy to miss because the source is quoted. `'printf \u00e9 now'` is
single-quoted on the way in, and the value is written back out plain, since
nothing in it needs quotes. The quotes that were there are not the quotes it
leaves in.

## The fix

Parity is kept, and the rune now has to be one a writer would have escaped:

```go
func escaped(r rune) bool {
	return !unicode.IsPrint(r) || r > 0xFFFF
}
```

What the two writers escape is narrow and known, and a probe over every rune
from U+00A0 to U+10FFFF confirms it: encoding each one through yaml.v3 and
through hclwrite produces no escape this predicate excludes. So an escape
outside it never arrives, and anything outside it is text.

Quoting the value instead was tried first and abandoned. Forcing a double quote
around any value holding a backslash fixes the YAML side, at the cost of a
change in three separate quoting functions, and it does not reach the heredoc at
all, because a heredoc cannot double a backslash. One predicate at the single
place that does the rewriting covers both.

## Prevention

A post-processor over finished output is reading its own writing next to
someone else's, and it needs a rule that separates them. Parity is a rule about
one encoding. Before relying on it, ask which of the output's shapes obey it:
here three of four did not, and the two that carry a shell script were among
them.

The narrower question the fix answers is better shaped: not "did something
escape this" but "would anything have escaped this character at all". That one
can be checked exhaustively, and was.

## Tests

- `internal/unescape/backslash_is_text_test.go` puts text and real escapes
  through the pass directly. The Windows path and the doubled backslash pass
  with the predicate removed, so they hold it to only what it should change.
- `provider/gitlab/backslash_roundtrip_test.go` reads the YAML it produces
  rather than matching its text, because a value holding a colon goes out
  double-quoted and writes the same backslash doubled.
- `provider/github/backslash_roundtrip_test.go` covers the heredoc.

One trap worth naming: a test written with a shell heredoc got a decoded `é`
into the Go source, so the test asserted the wrong string and passed against
broken code. Check the file's bytes with `od -c`, not its appearance.
