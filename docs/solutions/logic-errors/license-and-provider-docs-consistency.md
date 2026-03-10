---
title: "License migration and provider-agnostic docs consistency"
module: "Project metadata and documentation"
problem_type: "logic_error"
component: "LICENSE, README.md, CLAUDE.md, source SPDX headers"
severity: "high"
root_cause: "project-level documentation and headers drifted from current architecture and desired legal license, leaving GitHub-centric wording and AGPL references across the codebase"
symptoms:
  - "README described cinzel as GitHub Actions-focused instead of provider-agnostic"
  - "Root license file was AGPL while project direction required Apache 2.0"
  - "SPDX headers and policy/docs still referenced AGPL-3.0-or-later"
tags:
  - "license"
  - "apache-2.0"
  - "documentation"
  - "provider-agnostic"
  - "spdx"
created_date: "2026-03-10"
updated_date: "2026-03-10"
---

## Problem

Project messaging and legal metadata were inconsistent with the current multi-provider architecture and the requested license model.

- The README quick start and about text were interpreted as GitHub-centric.
- The repository still used AGPL text in `LICENSE`, docs, templates, and SPDX headers.
- This created legal ambiguity and documentation bias that could mislead users and contributors.

## Root Cause

- Historical defaults from the original GitHub-only phase were never fully normalized after GitLab support landed.
- License policy lived in multiple places (license file, SPDX headers, templates, docs, automation), so partial updates were easy to miss.
- Quick-start examples were provider-specific without an explicit provider-neutral command pattern up front.

## Solution

Applied a full consistency migration across legal and documentation surfaces.

1. Migrated project license from AGPL to Apache 2.0.
   - Replaced `LICENSE` with Apache 2.0 text.
   - Updated SPDX headers from `AGPL-3.0-or-later` to `Apache-2.0` across tracked source files.
   - Updated policy/docs and templates that referenced AGPL.

2. Made root documentation provider-agnostic.
   - Updated project description in `README.md` to describe provider-based CI/CD YAML conversion.
   - Updated quick start to lead with generic provider command shape:
     - `cinzel <provider> parse ...`
     - `cinzel <provider> unparse ...`
   - Included both GitHub and GitLab examples.
   - Updated `CLAUDE.md` project overview and architecture references to reflect current provider support.

3. Kept supporting docs aligned with the strict-schema and provider-neutral direction.
   - Cleaned stale docs language where older allowlist/provider-specific guidance could reintroduce drift.

## Verification

- Searched repo for AGPL/Affero references after migration and confirmed no matches in tracked source/docs patterns.
- Ran full test suite after migration:
  - `go test ./...` passed.
- Confirmed README quick start is now generic-first and includes both provider examples.

## Prevention

- Treat license metadata as a single policy surface:
  - `LICENSE`, SPDX headers, docs, templates, and license automation task must be changed together.
- Keep root docs provider-neutral; place provider specifics in provider-scoped docs.
- Add release-time checklist items:
  - license compliance (`mise run license-check`)
  - root-doc wording audit for provider neutrality
  - SPDX scan for unexpected license identifiers.

## Related

- `README.md`
- `CLAUDE.md`
- `LICENSE`
- `mise.toml`
- `docs/solutions/logic-errors/provider-strict-hcl-struct-schema-parity.md`
