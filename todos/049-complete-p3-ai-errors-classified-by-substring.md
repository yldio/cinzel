---
status: complete
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

- [x] A 429 is classified from the status, not the message text
- [x] A truncated fenced response yields the YAML without the fence line

## Technical Details

`classifyError` now asks `apiStatusCode` first. Both SDKs wrap an HTTP failure
in a typed error carrying the status: `*anthropic.Error` and `*openai.Error`,
each unwrapped in turn with `errors.As` so a wrapped error is still read.

When a status is known it decides the answer outright and the message text is
not consulted at all. That covers both directions of the old defect: a 402 had
no substring to match on and fell through to the generic message, and a server
fault whose response body echoed "rate_limit" or a model name carrying "429"
was reported as a rate limit. An SDK error prints its response body, so that
second case is not hypothetical.

The text search stays as the fallback for a provider error that is neither SDK
type.

`StripFences` gained `openFencePattern`, which matches an opening fence with no
closer. Two paths use it: a response with no closed fence at all (matched
directly), and a response that closed one fence then was cut off inside a
second (`trailingOpenFence` scans the text after the last closed fence and
appends what it finds as another document).

## Work Log

2026-09-16

Like 042, this one cannot be driven through the CLI without a live API call, so
it was reproduced by calling `StripFences` and `classifyError` directly.

Reverted the fix with `git stash push` on `internal/ai/provider.go` and ran the
new tests against the old code: three `TestStripFences` cases failed (the
opening fence line was returned attached to the document) and the 402 case of
`TestClassifyError` failed with

    got:  "LLM API error (test-provider): POST \"https://api.example.com/v1/messages\": 402 Payment Required "
    want to contain: "quota exceeded"

Restored, and all pass.

One snag while writing the fixtures: a bare `&openai.Error{StatusCode: ...}`
panics inside `(*Error).Error()`, which dereferences `r.Request` and
`r.Response`. A real SDK error always carries both, so the helpers
`apiRequestAndResponse`, `openaiError` and `anthropicError` build them with
`httptest.NewRequest`.

Divergence from the recommended action: the text search was kept as a fallback
rather than replaced, since a provider error that is neither SDK type still
reaches this function.

Verified with `go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check`.
