---
title: "Style guidance rule interpretation and enforcement"
module: "Code style and documentation standards"
problem_type: "logic_error"
component: "CLAUDE.md, provider/github, provider/gitlab"
severity: "medium"
root_cause: "style rules were partially ambiguous and applied inconsistently across block-start spacing, control-flow spacing, and comment attachment"
symptoms:
  - "mixed spacing patterns around return/if/for"
  - "comment/doc comment blocks separated from code by blank lines"
  - "outdated style guidance tied to specific helper usage"
tags:
  - "style"
  - "go"
  - "formatting"
  - "documentation"
  - "consistency"
created_date: "2026-03-11"
updated_date: "2026-03-11"
---

## Problem

The project had drift in Go style conventions. Spacing around `return`, `if`, and `for` was inconsistent across files, comment blocks were sometimes detached from the code they documented, and style guidance in docs was not precise enough for consistent interpretation.

## Root Cause

- Rules were present but not strict enough on block-level exceptions.
- Historical edits introduced mixed patterns.
- Some docs over-specified implementation details (for example one specific helper) instead of codifying the underlying style intent.

## Solution

Style policy was clarified and then enforced across the codebase.

1. Clarified style contract in `CLAUDE.md`:
   - blank line before `return`, non-error-guard `if`, and `for` only when not first statement in current block
   - no blank line when those statements are first in block (`func`, `if`/`else`, `for`, `switch`/`case`, nested block)
   - comments/doc comments must be directly attached to declarations/statements
   - exported declarations must have Go doc comments

2. Applied codebase-wide normalization:
   - removed leading blank lines before first-statement control-flow in blocks
   - inserted missing separation where control-flow followed prior executable code
   - removed blank lines between comments and documented code
   - added missing exported docs in non-test files

3. Updated related documentation to match the new conventions:
   - replaced prescriptive helper naming with deterministic-order helper guidance in:
     - `docs/solutions/patterns/critical-patterns.md`
     - `docs/solutions/logic-errors/nondeterministic-map-iteration.md`
     - `docs/plans/2026-03-09-feat-gitlab-pipelines-provider-plan.md`

4. Landed style-focused commits:
   - `10809ba` `style: trim blank lines at block starts`
   - `0cc0cba` `style: attach comments directly to code`
   - `db725e7` `style: enforce control-flow spacing conventions`

## Verification

- Formatting applied with `gofmt` after mechanical edits.
- Full test suite passed after style normalization:
  - `go test ./...`
- Follow-up scans confirmed no remaining comment-to-code blank-line gaps and no undocumented exported functions in non-test Go files.

## Prevention

- Keep style rules explicit in `CLAUDE.md` with block-level exceptions.
- Run `gofmt` and `go test ./...` for style sweeps.
- During review, explicitly check:
  - first statement in block has no leading spacer line
  - non-first control-flow statements are visually separated
  - comments are attached directly to code
  - exported declarations include proper doc comments.

## Related

- `CLAUDE.md`
- `docs/solutions/patterns/critical-patterns.md`
- `docs/solutions/logic-errors/nondeterministic-map-iteration.md`
- `docs/solutions/logic-errors/provider-strict-hcl-struct-schema-parity.md`
