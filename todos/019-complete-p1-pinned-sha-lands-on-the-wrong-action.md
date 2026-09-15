---
status: complete
priority: p1
issue_id: "019"
tags: [code-review, correctness, supply-chain]
dependencies: []
---

# A pinned SHA can land on a different action

## Problem Statement

`PinFile` and `UpgradeFile` wrote a resolved SHA by searching the file for
text reading `version = "<tag>"` and replacing the first hit. That line is
the right one only while every action before it was also rewritten. As soon
as one is skipped, its line stays matchable and absorbs the next action's
SHA.

The result inverts what pinning is for. The skipped action ends up pinned to
a commit from a repository it has nothing to do with, and the action that
was actually resolved is left on a mutable tag. Both commands exit 0 and the
summary counts the action as pinned.

## Findings

Three separate faults, all from the same three lines.

**A failed resolve misplaces the SHA.** Two actions on `v4`, the first
private or unreachable. `acme/private` fails, so no replace happens for it;
`actions/checkout` resolves and its SHA replaces the first `version = "v4"`,
which belongs to `acme/private`:

    version = "ffffffff..." # v4   <- acme/private, never resolved
    version = "v4"                 <- actions/checkout, resolved

**An upgrade needs no failure at all.** `UpgradeFile` also skips an action
already on the latest tag, and that branch is reached on an ordinary healthy
run. Two actions on `v4`, the first current, the second due `v5`: the `v5`
SHA lands on the current action.

**Upgrading twice stacks comments.** The replacement carried its own comment
and the old one was left in place, producing `# v5 # v4` — a line naming a
version the SHA is not.

`findActionRefs` already computes the offset of every version match and
discards it, so the information needed to write in the right place was
being thrown away.

## Fix

`ActionRef` keeps the `start` and `end` of its version assignment, `end`
extended past a trailing comment by `trailingCommentEnd`. Both commands
collect `versionEdit` values instead of rewriting as they go, then
`applyVersionEdits` splices them back to front so an earlier edit cannot
shift a later one's offsets.

## Verification

`internal/pin/version_rewrite_test.go`. Each gate was proven to fail with
only the rewrite reverted to the string search:

- `TestFailedPinLeavesItsOwnLineAlone` — acme/private took the checkout SHA
- `TestUpgradeMovesTheActionItResolved` — acme/current took the stale SHA
- `TestUpgradingTwiceLeavesOneComment` — found 2 comments, v4 still named

`TestEveryResolvedActionGetsItsOwnSHA` passes both with and without the
guard, so it was checked against a deliberately wrong fix: splicing edits
front to back instead of back to front. It fails there, where the other
three still pass. That is the case it exists to catch.

`UpgradeFile` had no test before this — the only one was a `t.Skip`.
