---
status: pending
priority: p3
issue_id: "046"
tags: [code-review, ai, config]
dependencies: []
---

# The built-in default models are out of date

## Problem Statement

The hardcoded defaults name models that are no longer current, and the same
strings are repeated in the config template written by `cinzel init`, so a new
config starts stale too.

## Findings

- `internal/ai/anthropic.go:16` — `claude-sonnet-4-5-20250514`
- `internal/ai/openai.go:15` — `gpt-4o`
- `internal/command/init.go` — both repeated in the template

## Recommended Action

Update the two constants and have the init template read them rather than
restate them.

## Acceptance Criteria

- [ ] Each default is defined once
- [ ] `cinzel init` writes whatever the constants say
