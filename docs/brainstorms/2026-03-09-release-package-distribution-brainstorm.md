---
title: "Release package distribution"
status: complete
date: 2026-03-09
---

# Release package distribution

Historical note: this brainstorm records the initial `homebrew-cinzel` tap direction. Current implementation targets `yldio/cinzel` for Homebrew automation.

## What We're Building

Add an automated release-to-Homebrew path using a maintained `homebrew-cinzel` tap.

On GitHub Release creation, the pipeline should publish release assets, compute checksums, update the tap formula, and create or update a tap PR so `brew install cinzel` tracks official releases.

This scope includes prerequisite release-readiness fixes found during repo scan (release workflow build path mismatch and CI test command mismatch), because Homebrew automation depends on trustworthy release artifacts.

## Why This Approach

Chosen approach: **Release-driven tap update**.

This is the smallest, least disruptive path from current behavior because release events already exist. It avoids a larger tag-orchestration redesign while still providing full automation from release assets to tap formula updates.

Compared with manual updates, it removes recurring operational work and checksum mistakes. Compared with tag-driven orchestration, it minimizes up-front migration risk and gets value sooner.

## Key Decisions

- **Homebrew target:** maintain a dedicated `homebrew-cinzel` tap (not Homebrew core in this phase).
- **Trigger:** run Homebrew update flow on GitHub `release.created`.
- **Install scope v1:** support macOS + Linux installs from release tarballs.
- **Prerequisites included:** fix release/CI mismatches before enabling tap automation.
- **Operational model:** release artifacts are source of truth; formula updates are derived from release metadata + SHA256.

## Resolved Questions

- Tap vs core: **tap**.
- Trigger model: **GitHub release event**.
- Platform scope: **macOS + Linux**.
- Include prereq fixes: **yes**.

## Open Questions

- None.

## Next Steps

- Proceed to planning with `/ce:plan` to define workflow updates, tap update mechanics, required secrets, and release validation gates.
