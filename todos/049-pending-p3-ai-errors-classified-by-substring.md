---
status: pending
priority: p3
issue_id: "049"
tags: [code-review, ai, errors]
dependencies: []
---

# Provider errors are classified by searching their text

## Problem Statement

`classifyError` looks for `"401"`, `"429"` and `"billing"` inside the error
message, so a prompt or a model name containing those characters is classified
wrongly, and a real status carried in a typed error is missed.

Nearby, `StripFences` leaves an unterminated fence in the output verbatim, so a
truncated response is handed to the YAML parser with its opening fence still
attached.

## Findings

`internal/ai/provider.go`.

## Recommended Action

Use the SDKs' typed API errors and read the status code. In `StripFences`,
treat an opening fence with no closer as running to the end of the text.

## Acceptance Criteria

- [ ] A 429 is classified from the status, not the message text
- [ ] A truncated fenced response yields the YAML without the fence line
