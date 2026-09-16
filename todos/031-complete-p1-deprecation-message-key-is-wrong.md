---
status: complete
priority: p1
issue_id: "031"
tags: [code-review, correctness, github, action]
dependencies: []
---

# deprecation_message fails in both directions

## Problem Statement

Parse writes the YAML key `deprecation-message`. GitHub and this project's own
strict shape both use `deprecationMessage`, so cinzel's output is rejected by
cinzel's own validation on the way back. In the other direction a correct
`deprecationMessage:` unparses to an HCL attribute of that literal name, which
the schema (`deprecation_message`) rejects on reparse.

## Findings

- `provider/github/parse_action.go:88` writes `deprecation-message`
- `provider/github/validate.go:76` expects `deprecationMessage`
- `provider/github/config.go:98` declares `hcl:"deprecation_message,optional"`

## Recommended Action

Emit `deprecationMessage` at parse_action.go:88, and map that key back to
`deprecation_message` in the unparse attribute-name conversion.

## Technical Details

- `parse_action.go` now writes `deprecationMessage`
- `unparse_action.go` gains `inputAttrHCLKey`, used in place of `toHCLKey` for
  input attributes only. The general rule swaps hyphens for underscores, which
  leaves a camel-case key untouched, so the emitted attribute was a literal
  `deprecationMessage` the schema does not declare
- Scoped to input attributes rather than added to `naming.ToHCLKey`: this is
  the only camel-case key GitHub uses, and widening the general rule would
  start renaming keys elsewhere that are meant to pass through
- No golden holds a deprecation message, so nothing moved

## Acceptance Criteria

- [x] An action input with a deprecation message roundtrips
- [x] Generated YAML passes the strict shape check

## Work Log

- 2026-09-16: Fixed as recommended. Reproduced both halves through the CLI
  first; the test drives the full HCL to YAML to HCL to YAML cycle and
  compares the two YAML documents, so a regression on either side fails it.
