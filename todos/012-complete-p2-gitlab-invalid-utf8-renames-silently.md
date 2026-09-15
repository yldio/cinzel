---
status: complete
priority: p2
issue_id: "012"
tags: [code-review, correctness]
dependencies: ["011"]
---

# GitLab rewrites invalid UTF-8 instead of refusing it

## Problem Statement

A byte that cannot start a UTF-8 sequence is read by goccy as though it
could. The encoder later writes it out as U+FFFD, so a job named with such a
byte comes back under a different name. Nothing reports it.

## Findings

Probed at 1e25e11 by round-tripping through both directions.

```
input job name bytes : b'\xff\xfe'
output job name bytes: b'\xef\xbf\xbd\xef\xbf\xbd'
```

Two bytes become six, and the job is now called something else. The same
substitution happens in a script value and in a stage name, which means a
job can end up pointing at a stage that no longer exists.

The GitHub provider refuses the identical input:

```
yaml: invalid leading UTF-8 octet
```

Three neighbouring shapes were probed at the same time and are sound in both
providers:

- A NUL byte. GitLab escapes it to `\0`, keeps the value, and round-trips
  stably. GitHub refuses it outright. Either is defensible.
- A lone surrogate escape. Both refuse it.
- `${...}` in a value. Escaped to `$${...}` on the way out and restored on
  the way back, in both providers.

## Recommended Action

Call `utf8.Valid` on the input in `checkYAMLSoundness`, alongside the alias
and key checks already there. It is stdlib and it answers exactly the
question being asked, so nothing has to match on decoder message text.

## Technical Details

- `provider/gitlab/unparse_pipeline.go`, `checkYAMLSoundness`

## Acceptance Criteria

- [x] Invalid UTF-8 is refused rather than substituted
- [x] Valid multibyte text, including four-byte sequences, still works

## Work Log

- 2026-09-15: Created and closed. Found while probing four input boundaries
  neither provider had tests for. Only the UTF-8 substitution turned out to
  be a defect; the other three are recorded above so the next pass does not
  re-probe them.
