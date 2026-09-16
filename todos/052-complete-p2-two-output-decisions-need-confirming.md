---
status: complete
priority: p2
issue_id: "052"
tags: [code-review, question, github, output]
dependencies: []
---

# Two output decisions that look deliberate but contradict something

## Problem Statement

Neither is a confirmed bug. Both need a decision before anything is changed.

**Injected `permissions: {}`.** Parse writes an empty permissions map when the
author wrote none, and the document builder keeps it explicit. GitHub reads
`{}` as "no permissions", which is not what omitting the key means. Every step
likewise gains an `id:` it did not have. If this is the intended default, it
belongs in the docs; if not, it changes the meaning of generated workflows.

**Key order.** `workflowKeyOrder` puts `jobs` after `permissions`, `env`,
`defaults` and `concurrency`. CLAUDE.md states the order is name, run-name, on,
jobs, then the rest sorted. One of the two is wrong.

## Findings

- `provider/github/parse_workflow.go:471-473`
- `provider/github/workflow_yaml.go`, `workflowKeyOrder` and `docValue`
- `CLAUDE.md`, "YAML output"

## Recommended Action

Decide each, then either change the code or write the decision into CLAUDE.md.
Golden files move with whichever way the key order goes.

## Acceptance Criteria

- [x] Both decisions recorded
- [x] Code and CLAUDE.md agree

## Technical Details

Three decisions in the end, all resolved in favour of the code. CLAUDE.md
changed; no code and no golden file moved.

**Injected `permissions: {}`** — `provider/github/parse_workflow.go:493`. Git
history settles it: the line arrived in `fc0b359`, "fix(parse): always emit
permissions: {} in workflow YAML output (#9)". Emitting it was the point of
that commit, not an oversight in it.

**Step `id:`** — `provider/github/step/stepparse.go:52-56`, which falls back to
the block label and carries the comment "Keep backward compatibility with
legacy behavior where step label is used as the emitted step id unless
explicitly ignored". `ignore_id` is the opt-out for anyone who does not want
it.

**Key order** — the user's call: the code is right and the doc was wrong.
`name, run-name, on, permissions, env, defaults, concurrency, jobs` reads the
way a hand-written workflow does, short top-level keys first and jobs as the
long tail. The alternative would have moved 5 of the 6 GitHub golden files to
put `permissions` after `jobs`, which no one writes by hand.

CLAUDE.md's "YAML output" section now lists all eight keys and adds a
"Defaults on parse" note for the two injected values.

## Work Log

2026-09-16

Read the two sites, then blamed them rather than judging them on sight:
`git log -S` on the permissions line found the commit whose subject is the
decision. The step id had its answer in a comment at the call site.

Only the key order was genuinely open, so that was the one put to the user.
Answered: keep the code, fix the doc.

This unblocks Batch 6. 044 (quoting gaps) no longer waits on a key-order
decision, since the order is not moving.
