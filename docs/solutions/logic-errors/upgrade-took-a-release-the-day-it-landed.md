---
title: "upgrade took a release the day it was published"
module: "internal/pin"
problem_type: logic_error
component: "internal/pin"
severity: medium
root_cause: "LatestTag returned only a name, so nothing in the upgrade path could tell a release published an hour ago from one published a year ago"
symptoms:
  - "a scheduled upgrade job moves to a release the day it lands"
  - "a tag moved or pulled shortly after publication is already in the pipeline"
  - "there is no way to say how new is too new"
tags:
  - pin
  - upgrade
  - policy
  - supply chain
status: fixed
created_date: "2026-11-06"
updated_date: "2026-11-06"
---

# upgrade took a release the day it was published

## What happened

`upgrade` moves every action to the newest release, with no notion of how new
that is. A release published minutes before the run was in the pipeline
minutes after it, and the window in which a tag gets moved or pulled is the
same window in which the upgrade had already taken it.

There was nothing to configure, because there was nothing to configure with:
`LatestTag` returned a string.

```go
LatestTag(ctx context.Context, owner, repo string) (string, error)
```

A name says nothing about age. Both "published an hour ago" and "published a
year ago" arrive as `"v5"`.

## The fix

`LatestRelease` returns the tag and the publication date, which the API was
already sending in the same response:

```go
type Release struct {
	Tag       string
	Published time.Time
}
```

`Cooldown` is one duration and a clock, and `UpgradeFile` consults it after
resolving the latest release. The zero value holds nothing, which is what
every existing caller passes and how upgrade behaved before.

```
held actions/deploy-pages: d6db90164ac5 → v5.0.1, published 65d ago, 24d left
```

Three decisions sit in where the check goes and what it treats as held.

The cooldown is checked **after** "already current". An action sitting on the
newest tag would otherwise be reported as held for as long as the window
lasted, which reads as an upgrade waiting that is not.

A release with **no publication date** is taken rather than held. GitHub
leaves `published_at` null on a draft, and an unknown date is not evidence
that something is new — holding on one would stall an upgrade for a reason
the run cannot print.

A held release **does not fail `--check`**. The policy working is not a
problem to go red over, and a check that failed on one would leave a
repository with a cooldown permanently red for the length of every window,
where the only route to green is abandoning the policy.

The report names each held action with the time left. A run that silently
declines an upgrade is indistinguishable from one that found none, and the
next person to look wonders why the bump never happens.

## Where the policy lives

`.cinzelrc.yaml`, under `github.upgrade.cooldown-days`, with
`--cooldown-days` overriding it for one run. The file is the default because a
cooldown is something a repository keeps: a scheduled upgrade job that has to
remember a flag is one edit away from silently not having it.

Upgrade reads its own key rather than going through
`loadProviderCommandConfig`, which validates against the parse and unparse
schema and would warn "unknown key" on everything written under `upgrade`.

A quoted `"14"` is refused rather than read as 0. It is a string to YAML and
identical to a reader, so taking it would make the one spelling that looks the
same behave differently. A negative window is refused too: 0 already means no
cooldown, so a negative number is one whose author meant something the policy
cannot carry out.

## Prevention

When a decision is about *how* something is rather than *what* it is, check
that the data carrying it has the attribute at all. `LatestTag` could not have
grown a policy without changing its signature, and the absence was invisible
because the function did exactly what its name said.

## Tests

`internal/pin/cooldown_test.go` fixes the clock through `Cooldown.Now`, so no
test depends on how long it takes to run. It covers a release inside the
window, one past it, the zero policy, a release with no date, and an action
already on the latest tag.

`internal/command/cooldown_config_test.go` covers reading the key, every
spelling that means "no policy", the values that are refused, and flag
precedence — including `--cooldown-days 0`, which has to be distinguishable
from the flag being absent, and is why the override tests `IsSet` rather than
a non-zero value.

Verified against the real API: a 90-day window over this repository's own HCL
held `actions/deploy-pages` at 65 days with 24 left and took
`actions/upload-pages-artifact` at 209.
