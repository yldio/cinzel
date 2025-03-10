---
title: "Generic attribute parsing handles all action runs.using types without special-casing"
module: "GitHubProvider"
problem_type: "best_practice"
component: "parse_action"
severity: "medium"
root_cause: "incomplete_setup"
symptoms:
  - "README claims action types are unsupported when they actually work"
  - "Unnecessary TODO items for node/docker action support"
tags:
  - "actions"
  - "composite"
  - "node20"
  - "docker"
  - "generic-parsing"
  - "architecture"
created_date: "2026-03-08"
updated_date: "2026-03-08"
---

## Problem Description

The README listed "Node.js and Docker action types in HCL → YAML parse direction" as a "Not yet" item, implying these action types were unsupported. Investigation revealed they already worked.

## Root Cause

The `parseActionRunsBlock` function only special-cases `steps` (which needs reference list resolution for composite actions). All other attributes (`main`, `pre`, `post`, `image`, `entrypoint`, `args`, etc.) flow through the generic `parseAttr` path. This means any `runs.using` type works automatically.

Similarly, `actionToHCL` (unparse direction) only checks `using == "composite"` to decide whether to extract step blocks. All other runs attributes are written generically.

## Solution Implemented

- Added test fixtures for `node_action` and `docker_action` proving both directions work.
- Replaced the misleading "Not yet" README section with "Known limitations" (only genuine limitations: edge-case coverage, byte-stable roundtrip).

## Prevention Guidance

- Before documenting something as unsupported, write a test to verify. Generic code paths often handle more cases than expected.
- When adding new action types in the future, only add special handling if the type has structural differences (like composite's `steps` references). Don't add switch cases for types that the generic path already covers.
- The architecture principle: **special-case the structure, not the type**. Composite is special because of reference resolution, not because of its `using` value.
