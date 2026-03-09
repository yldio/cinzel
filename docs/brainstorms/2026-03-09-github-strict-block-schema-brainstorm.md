---
title: "GitHub provider strict block schema"
status: complete
date: 2026-03-09
---

# GitHub provider strict block schema

## What We're Building

Define explicit typed HCL schemas for GitHub provider blocks so unsupported properties are rejected by HCL decoding (properties must be declared, otherwise parsing fails).

Scope chosen for this brainstorm:

- All GitHub provider HCL blocks
- Both directions (parse and unparse)
- Generic unknown-field errors
- Immediate strict mode (no compatibility aliases)

## Why This Approach

Current parsing patterns use remaining-body/manual handling in places, which can allow behavior that is less strict than fully typed decoding. Moving block properties to typed decode schemas makes unknown attributes fail by default, catches typos early, and keeps parse/unparse behavior aligned.

## Approaches Considered

### A) Central schema registry (chosen)

One source of truth using typed decode structs per GitHub block type, so accepted properties are explicitly declared in schema and unknown properties fail at parse time.

Why chosen:

- Lowest drift risk between parse and unparse
- Consistent error semantics across all blocks
- Easiest to test and document as a product rule

### B) Per-file allowlists

Rejected because duplication increases drift risk and maintenance cost.

### C) Pre-validation wrapper

Rejected because it can obscure block-specific error context and split validation responsibilities.

## Key Decisions

- GitHub provider blocks use explicit typed schemas; unknown attributes and unknown block types are invalid by parser rules.
- Strictness applies to parse and unparse flows.
- Error style is generic unknown-field (no deprecated-key custom messaging).
- No transition mode: strict behavior is immediate.
- Terminology stays explicit: HCL uses `depends_on`; YAML keeps GitHub-native `needs:`.
- Strict schemas must not break valid references to other resources/parameters (for example `job.*`, `step.*`, `variable.*`, expression strings).

## Success Criteria

- Any unsupported property or block in GitHub provider blocks fails deterministically through typed schema decoding.
- Parse/unparse enforce the same schema contract for equivalent block shapes.
- Fixture matrix includes invalid unknown-field cases per major block type.
- Existing valid resource/parameter reference patterns continue to parse/unparse correctly.
- Documentation clearly states strict schema expectations.

## Constraints

- Preserve deterministic behavior and ordering conventions.
- Keep validation messages stable enough for fixture-based tests.
- Keep provider-specific strictness inside `provider/github` boundaries.
- Keep parse errors consistent across block types for predictable fixture assertions.
- Keep reference semantics intact while tightening schema (strict fields, unchanged reference behavior).

## Open Questions

- None.

## Resolved Questions

- Scope: all GitHub blocks.
- Direction: both parse and unparse.
- Error style: generic unknown-field errors.
- Rollout: immediate strict mode.
- Design: central schema registry.
- Strictness mechanism: typed HCL schemas declare allowed properties; unknown fields fail parsing.
- Compatibility boundary: preserve valid cross-resource/parameter references.
