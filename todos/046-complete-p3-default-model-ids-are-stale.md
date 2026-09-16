---
status: complete
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

- [x] Each default is defined once
- [x] `cinzel init` writes whatever the constants say

## Technical Details

The Anthropic default was worse than stale. `claude-sonnet-4-5-20250514`
carries Sonnet 4.5's name and Sonnet 4's date, and appears in neither
`anthropic-sdk-go`'s model list nor anywhere else: it names no model at all.
Every assist run against Anthropic was starting from a string the API would
refuse.

Both constants now read from their SDK rather than restating a string, so a
dependency bump carries the default with it instead of leaving it to rot:

    anthropicDefaultModel = string(anthropic.ModelClaudeSonnet4_6)
    openaiDefaultModel    = string(openai.ChatModelGPT5_4)

`ai.DefaultModels()` exposes them as a provider-keyed map, and
`providerDefaults()` in `internal/command/init.go` renders the template's
provider block from it. The template's two hardcoded lines became one `%s`.

## Work Log

2026-09-16

Searched for every copy of the strings before changing anything: four sites in
three files, exactly as the todo listed. Checked both SDKs' model lists in the
module cache to pick the replacements, which is where the fabricated Anthropic
id showed up — the SDK has `claude-sonnet-4-5-20250929`, never an 0514 of that
name.

Wrote `TestInitTemplateUsesTheCodeDefaults`, which renders the template and
checks every provider's block against `ai.DefaultModels()`, then walks the
rendered lines and fails on any `model:` naming something no provider defaults
to. The second half is what catches a future edit that hardcodes a name again.

Proving it fails without the fix needed care: stashing the whole change made
the test not compile, which proves nothing. Instead the new API was kept and
`providerDefaults()` was pointed back at a literal map holding the two old
strings. All four assertions fired, naming both missing blocks and both stale
ids.

Verified with `go test -count=1 ./...`, `mise run lint`, `mise run drift` and
`mise run license-check`.
