---
title: "A trailing comment scan that stops at the first one, and takes the CR with it"
module: "internal/pin"
problem_type: logic_error
component: "internal/pin"
severity: medium
root_cause: "the scan for the comment on a version line returns after one comment, where a closed /* */ leaves the rest of the line still to read, and its end offset includes the CR of a CRLF ending"
symptoms:
  - "after upgrade, a version line reads '# v5 # v4' and names a version the SHA is not"
  - "a comment written before the tag comment survives the rewrite and sits beside the new one"
  - "pinning a CRLF checkout rewrites one line as LF, leaving mixed endings and a diff on lines nobody edited"
  - "exit code 0 throughout"
tags:
  - "pin"
  - "upgrade"
  - "comments"
  - "line-endings"
status: "fixed"
created_date: "2026-09-22"
updated_date: "2026-09-22"
related:
  - docs/solutions/logic-errors/HANDOFF-github-internal-review.md
  - docs/solutions/logic-errors/github-pin-comment-placement-and-empty-permissions-default.md
---

# A trailing comment scan that stops at the first one

Found by the sweep recorded in `HANDOFF-github-internal-review.md`, working
through `internal/pin`.

`pin` and `upgrade` rewrite a version assignment in place and give it a comment
naming the tag the SHA came from:

```hcl
version = "de0fac2e4500dabe0009e67214ff5f5447ce83dd" # v4
```

The old line is replaced whole, its trailing comment included, because leaving
the old comment standing is how a line comes to read `# v5 # v4` and claim a
version the SHA is not. `trailingCommentEnd` finds where that old comment ends.

## Two faults in one function

**It stopped after one comment.** A `#` or `//` runs to the newline, so finding
one is finding all of them. A `/* */` does not: it ends where it closes, and the
rest of that line is still part of the trailing comment. The scan returned at the
close, so anything after it stayed:

```hcl
version = "cccc…" /* pinned */ # v4
```

upgraded to

```hcl
version = "bbbb…" # v5 # v4
```

which is the exact line `TestUpgradingTwiceLeavesOneComment` was written to
prevent. That test passes a plain `# v4`, where one pass is enough, so the case
sat underneath it.

**It took the CR with the comment.** On a CRLF file the offset ran to the `\n`,
which puts the preceding `\r` inside the replaced span. The replacement carries
no `\r`, so that one line came out LF in a file that is CRLF throughout. A pin on
a Windows checkout produced mixed endings and a diff on lines nobody edited.

## The fix

Scan in a loop. A closed `/* */` advances the end and goes round again; a `#` or
`//` reaches the newline and is the last thing on the line, so it stops. An
unterminated `/*` still stops without taking anything further, because swallowing
the rest of the file would delete every block below.

Then step back over a single trailing `\r`, which belongs to the line ending and
not to the comment.

`internal/pin/pin.go`, `trailingCommentEnd`.

## Testing it

`TestASecondCommentAfterABlockCommentIsAlsoReplaced` covers `/* */` followed by
`#`, by `//`, and by a second `/* */`. `TestPinKeepsCRLFLineEndings` pins a CRLF
file with no comment, a hash comment and a block comment, and counts line endings
that came out LF.

The CRLF test asserts the whole rewritten line rather than a prefix of it. A
first version checked `strings.Contains` for the `version = "…" # v4` prefix, and
a deliberately wrong fix — stepping back one character unconditionally instead of
only over a `\r` — passed it, and passed the entire existing suite. The prefix
check cannot see a tail left standing after the part it matched. Asserting the
line up to and including its `\r\n` fails that wrong fix on the block-comment
case.

## Prevention

A scan for "the comment at the end of this line" is a loop, not a lookup,
wherever the language has a comment form that closes before the newline.

When a rewrite replaces a span of text, the span ends at the last character the
rewrite is responsible for. A line ending is the file's, not the line's.

## The author's note went with it

A third fault sat behind the other two, and it is the reason the fix to them is
shaped the way it is. The rewrite replaces the whole trailing comment, so

```hcl
version = "v4" # do not move: v5 drops node16
```

came back as

```hcl
version = "aaaa…" # v4
```

and the instruction not to move was gone, with the pin reporting success. The
tag comment is the tool's to overwrite. Everything else on that line is the
author's.

`authorNote` now reads the old comment and carries what the author wrote into
the new one, behind the tag. Three things decide its shape:

The tag goes first, which is what makes the rewrite idempotent. A line this
tool wrote reads back as the same note with the leading tag dropped, so pinning
a file repeatedly leaves one note and one tag however many times it runs.

Which leading word counts as the tool's depends on the version the line holds.
On a full SHA a pin has already been over the line, so a leading tag is the one
it left, and keeping it is the superseded `# v5 # v4` naming a version the SHA
is not. On a tag, nothing has pinned the line yet, so the only word dropped is
one naming that same tag — a note opening `v5 drops node16` keeps its first
word. A first version dropped any tag-shaped leading word and ate it.

That left a hole, found by probing repeated upgrades rather than repeated pins.
A tag is whatever a repository releases, and GitHub takes names semver does not.
An upgrade writes back what `LatestTag` returned, checked only against
`safeNamePattern`, so the tag a pass leaves does not always look like `v5`.
Recognising only the semver shape left the rest unrecognised, kept as the
author's note, and growing on every run: three upgrades of a repository
releasing `stable`, `latest` and `edge` gave

```hcl
version = "aaaa…" # edge latest stable
```

which is the superseded comment this whole scan exists to prevent, reached by
another route. Widening the pattern to `safeNamePattern` fixed that and broke
the note: `/* pinned */` lost its word too.

The tool's tag is identified by the comment's shape instead. `versionLine`
writes exactly one `#` comment, tag first, on a line holding a SHA. That triple
is recognisable without guessing what a tag looks like, and there is no pattern
that recognises every tag and nothing else.

The test runs per comment, not once over the whole run, because the tag is not
always first: `/* pinned */ # v4` holds the note first and the superseded tag
behind it.

Every comment collapses into one run of text. The note is written back as a
single `#`, which runs to the newline, so a block comment written over two
lines carried across verbatim would end the comment halfway and leave its tail
as HCL.

The three tests written for the first two faults asserted the old comment was
gone, which is the contract this change reverses. Each had one assertion that
was load-bearing and one that only encoded "destroyed". They now assert the
whole rewritten line, which is what tells a note carried over apart from a
comment left standing beside the new one — the same exactness the CRLF lesson
above is about. Four reverts were run against them, each compiling: no note at
all, the greedy tag drop, no collapse, and the unconditional step-back. Each
fails.
