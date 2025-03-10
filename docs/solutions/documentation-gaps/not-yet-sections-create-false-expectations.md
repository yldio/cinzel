---
title: "README 'Not yet' sections create false expectations about missing functionality"
module: "Documentation"
problem_type: "documentation_gap"
component: "readme"
severity: "medium"
root_cause: "inadequate_documentation"
symptoms:
  - "README claims features are missing when they work"
  - "Contributors avoid areas they think are incomplete"
  - "Users assume functionality gaps that don't exist"
tags:
  - "readme"
  - "documentation"
  - "not-yet"
  - "known-limitations"
created_date: "2026-03-08"
updated_date: "2026-03-08"
---

## Problem Description

The `provider/github/README.md` had a "Not yet" section listing three items:
1. Full GitHub Actions schema parity for all edge cases
2. Byte-stable roundtrip output
3. Node.js and Docker action types in parse direction

Item 3 was already fully supported. Items 1 and 2 are permanent design trade-offs, not planned work.

## Root Cause

"Not yet" implies items are on a roadmap and will be implemented. This framing is misleading when applied to:
- Features that already work (just weren't tested/documented)
- Inherent limitations that will never be "fixed"

## Solution Implemented

1. Added test fixtures proving Node.js and Docker actions work (removed the false claim).
2. Replaced "Supported now" / "Not yet" with "Coverage" / "Known limitations".
3. The two remaining items are honestly framed as permanent design trade-offs.

## Prevention Guidance

- Before documenting something as unsupported, write a test to verify.
- Use "Known limitations" for permanent design trade-offs. Use issues/milestones for actual planned work.
- Avoid "Not yet" / "Coming soon" / "TODO" sections in README files — they go stale quickly and mislead users.
- If a limitation is fundamental to the architecture (like format normalization in a transpiler), call it out as a design decision, not a missing feature.
