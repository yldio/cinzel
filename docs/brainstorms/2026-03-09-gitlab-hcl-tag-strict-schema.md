# GitLab Provider Strict Schema via HCL Tags

Date: 2026-03-09

## Context

GitLab parsing currently relies on generic body traversal (`hcl:",remain"`) plus custom logic. This works, but strictness is encoded separately and can drift from decode behavior.

## Problem

Manual schema maps are easy to desync from real parser behavior and duplicate what HCL tags can already express (`attr`, `block`, labels, optional fields).

## Direction

Use HCL tagging as the canonical schema contract for parse strictness:

- Define scope structs (`job`, `workflow`, `default`, `include`, `service`, etc.) with `hcl` tags.
- Validate by `gohcl.DecodeBody` against those structs.
- Keep `,remain` only where pass-through is intentional (for example, nested `reports` internals).
- Keep generic parse traversal for conversion logic, but gate it behind tagged decode validation.

## Why this is better

- Single source of truth for allowed attributes/blocks.
- Better diagnostics from HCL decode rather than hand-rolled checks.
- Closer parity with the intended strict-schema philosophy.
- Easier extension: new fields are added once in schema structs.

## Incremental adoption plan

1. Introduce decode-based schema validation by scope.
2. Verify full test suite and fixture matrix remain green.
3. Incrementally migrate more conversion paths from generic map traversal to typed block decoding where this reduces complexity.
4. Add regression fixtures for unknown attrs/blocks per scope.

## Notes

- This does not force a full rewrite in one step.
- It keeps current conversion behavior stable while making strictness declarative.
