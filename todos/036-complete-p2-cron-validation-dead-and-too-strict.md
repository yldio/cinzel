---
status: complete
priority: p2
issue_id: "036"
tags: [code-review, validation, github, cron]
dependencies: []
---

# Cron is checked on one side only, and rejects valid expressions

## Problem Statement

Two separate defects in the same area.

The schedule check only runs when the value is a map, but on the parse path
`DenormalizeScheduleEvent` has already made it a `[]any`, so the branch never
fires. `99 99 99 99 99` passes HCL to YAML and is then refused going back.

`ValidateCron` accepts numeric fields only. GitHub allows names, so `MON-FRI`,
`SUN`, `JAN` and day-of-week `7` are rejected although they are valid.

## Findings

- `provider/github/validate.go:152-160` — the map-only branch
- `provider/github/workflow/cron.go` — `validateCronField`, `validateCronRange`,
  `validateCronNumber`

## Recommended Action

Handle the `[]any` case and validate each element's `cron`. Accept the
three-letter month and day names, and treat day-of-week `7` as Sunday.

## Acceptance Criteria

- [x] `0 9 * * MON-FRI` is accepted
- [x] `99 99 99 99 99` is refused on the parse path, not only on unparse

## Technical Details

Two changes, one per defect.

`provider/github/validate.go` now calls `validateScheduleEvent`, which takes the
event in either shape. The parse path runs `DenormalizeScheduleEvent`
(`provider/github/workflow/yaml_document.go`) first, which turns the event into
`[]any` of `{"cron": ...}` entries, so the map-only guard never matched and the
check was dead there. The map branch stays for the unparse path, where the event
is still normalized. Each list entry is validated separately and the error names
its index.

`provider/github/workflow/cron.go` gained `cronValue`, which resolves a field
value through a per-field name table before falling back to `strconv.Atoi`.
`monthNames` covers `JAN`-`DEC`, `dayNames` covers `SUN`-`SAT`, and only those
two fields carry a table, so `MON` in the month field is still refused.
`day-of-week` max moved from 6 to 7 because both 0 and 7 name Sunday. Names are
matched case-insensitively.

A step (`*/n`, `a-b/n`) is a count rather than a point in the field, so the
wildcard-step branch checks it with `strconv.Atoi` before handing it on, and
`*/MON` is refused.

## Work Log

### 2026-09-16

Reproduced both halves through the CLI: `99 99 99 99 99` parsed to YAML with
exit 0, and `0 9 * * MON-FRI` was refused with `invalid range start "MON"`.
After the fix the first exits 1 with
`workflow.in.on.schedule: schedule[0]: cron minute field: value 99 out of range`
and the second writes `- cron: "0 9 * * MON-FRI"`.

Tests: `provider/github/schedule_validation_test.go` drives `Parse` for the
list-shape path (4 cases); `provider/github/workflow/cron_test.go` gained seven
valid name cases and four refusals, including a day name in the month field and
a name used as a step. Both sets were confirmed to fail with the fix removed —
the parse test on the two error cases, the cron test on all six name cases.

`go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check` all pass. The existing `dow out of range` case moved
from `7` to `8`, which is now the first value past the field.
