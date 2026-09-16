---
status: pending
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

- [ ] Both decisions recorded
- [ ] Code and CLAUDE.md agree
