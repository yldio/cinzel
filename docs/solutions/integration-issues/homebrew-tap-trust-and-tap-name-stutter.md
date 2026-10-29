---
title: "The README install command stopped working, and the tap name stuttered"
module: "Release CI/CD"
problem_type: "integration_issue"
component: ".goreleaser.yaml, cinzel/steps.hcl, README.md, docs/release/homebrew.md"
severity: "medium"
root_cause: "Homebrew 6.0 requires explicit trust before a short name resolves in a non-official tap, and the tap repository was named after the project it already sat under"
resolution_type: "workflow_improvement"
symptoms:
  - "brew install --cask cinzel fails with Refusing to load ... from untrusted tap after brew tap yldio/cinzel"
  - "The fully qualified name reads yldio/cinzel/cinzel"
tags:
  - "homebrew"
  - "tap"
  - "release"
  - "goreleaser"
  - "install"
created_date: "2026-10-28"
updated_date: "2026-10-28"
---

## Problem

The README told people to install in two steps:

```sh
brew tap yldio/cinzel && brew install --cask cinzel
```

Homebrew 6.0.0 made tap trust mandatory for non-official taps. Tapping a
repository no longer grants trust to what is in it, so the second half fails:

```console
$ brew info mangodisk-cli
Error: Refusing to load formula harry0703/tap/mangodisk-cli from untrusted tap harry0703/tap.
Run `brew trust --formula harry0703/tap/mangodisk-cli` or `brew trust harry0703/tap` to trust it.
```

A short name only resolves once the user has run `brew trust`, which the
README never mentioned, so anyone following it on a current Homebrew got an
error instead of a binary.

Separately, the install line read `yldio/cinzel/cinzel` in its qualified form.
The tap repository was `homebrew-cinzel`, under the `yldio` org, holding one
cask called `cinzel`.

## Root Cause

Two unrelated things that both land in the same command.

Trust: `brew tap` and trust were one action before 6.0.0 and are two now.
Installing a fully qualified name is itself an act of trust in that one item —
Homebrew takes `user/repo/item` as the user naming exactly what they accept —
so it auto-taps, trusts the single cask, and installs, all without a separate
`brew trust`.

Stutter: Homebrew's grammar has no two-part form for an item. Two segments are
always a tap:

```console
$ brew info yldio/cinzel
Error: No available formula with the name "cinzel". Did you mean cinecli?
```

It read `cinzel` as the item name and `yldio` as the user, then looked for a
repository it never got. `HOMEBREW_TAP_CASK_REGEX` in `tap_constants.rb`
requires all three segments. The repetition can only be removed by renaming
one of the three, and the tap repository is the only one that is ours to move
without consequence.

## Attempted Solutions

**Attempted Solution 1:** Submit to `homebrew/core` so plain `brew install
cinzel` works.

Blocked twice over. `Package-Acceptance-Policy.md` asks a self-submission for
225 stars, 90 forks or 90 watchers; the repository has 3, 0 and 0.
`Acceptable-Casks.md` also says open-source command-line software belongs in
core as a formula built from source, not as the cask we publish, so this is a
packaging rewrite and not a rename.

## Solution

1. Rename the tap repository `yldio/homebrew-cinzel` to `yldio/homebrew-tap`,
   which makes the qualified name `yldio/tap/cinzel`. GitHub redirects the old
   name, so anyone who already tapped keeps working.

2. Point `homebrew_casks.repository.name` in `.goreleaser.yaml` at the new
   repository, and update both places in `cinzel/steps.hcl` that name it: the
   app token's `repositories` input and the token verification loop.

3. Replace the two-step install in `README.md`, `plugins/cinzel/README.md`,
   `plugins/cinzel/skills/cinzel/SKILL.md` and both site pages
   (`docs/site/index.html`, `docs/site/docs.html`, where the command is spread
   across syntax-highlighting spans and so does not match a plain grep for the
   command itself) with the one-step fully qualified form:

   ```sh
   brew install --cask yldio/tap/cinzel
   ```

4. Regenerate `.github/workflows` from the HCL.

5. Re-scope the release GitHub App installation to `cinzel,homebrew-tap`. This
   is a GitHub org setting and not in the repository: until it is changed the
   release run fails at the token verification step, by design.

## Verification

- `mise run drift` reports 4 checked, 0 differ
- `mise run lint`
- `mise run test-ci`

## Prevention

The install command in the README is the one piece of this repository that
only users execute, so nothing in CI runs it. When Homebrew changes how names
resolve, the command rots silently. Prefer the fully qualified form: it states
the tap, needs no prior `brew tap`, and grants the narrowest trust.

## Related

- `docs/release/homebrew.md`
- `docs/solutions/integration-issues/release-automation-deprecation-trigger-contract-cleanup.md`
- `docs/solutions/integration-issues/github-app-token-auth-name-mismatch.md`
