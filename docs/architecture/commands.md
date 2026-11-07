# Commands

## AI assist (`cinzel <provider> assist`)

Pipeline: prompt → LLM → YAML → strip fences → split docs → temp files →
Unparse → merge and dedup HCL → session folder.

Output is `cinzel/assist/{timestamp}/assist.hcl`, one session per prompt.
`--refine` targets the latest session; `--from {timestamp}` picks an older one.

Merging compares against the existing `cinzel/*.hcl`. A block identical to one
already there becomes a `// reuses:` comment; a different block with the same
signature gets a `// note:` saying where the other one is. Generated actions
are pinned to SHAs afterwards.

`StripHCLContext` replaces every string value in the context with `"..."` by
walking the HCL AST, so the prompt carries structure and not content.

Config is `os.UserConfigDir()/cinzel/config.yaml`, written by `cinzel init`. It
holds no API key: those come from the environment. Resolution order is CLI
flags, then env vars, then the config file, then the built-in defaults.

A config that still holds an `api_key` is read, because one written before
`cinzel init` stopped asking for keys may have one. That is a migration path,
not a way to configure a key, and `configWarnings` says so on every load.

## Version management (`cinzel github pin/upgrade`)

`pin` resolves action tags to SHAs through the GitHub API, cached for 24 hours.
The tag stays on the line it pinned: `version = "<sha>" # <tag>`.

`upgrade` finds the latest release, compares it against what is there by tag or
by SHA, and updates both the version and its comment.

No token is needed for public actions. `GITHUB_TOKEN` raises the rate limit.

Either command exits non-zero when it could not resolve an action, so a CI step
calling `cinzel github pin` fails on a workflow still holding an unpinned one.
Each failure is named as it happens, and the summary line counts them. With
`--parse`, `upgrade` writes the regenerated YAML for what did upgrade before
reporting the failure.

`--check` turns either command into the question instead of the work. It
implies `--dry-run`, so nothing is written — a check that pinned what it was
asked to look for would pass on its own side effect — and it ends non-zero
when the run found something to do: an action still on a tag for `pin`, a
newer release for `upgrade`. A failed resolve outranks the finding, because a
run that could not reach an action does not know its state either way.

`--dry-run` alone keeps exiting 0 whatever it reports. It answers what a run
would do; `--check` answers whether it needs to.

`upgrade` reads a cooldown from `github.upgrade.cooldown-days`, or from
`--cooldown-days` when that is set. `LatestRelease` returns the publication
date alongside the tag, and a release still inside the window is reported with
the time left and skipped, counted apart from both upgraded and failed.

Three things fall out of where the check sits. It comes after "already
current", so an action on the newest tag is not reported as held for the
length of the window. A release with no publication date is taken rather than
held, since GitHub leaves `published_at` null on a draft and an unknown date
is not evidence of newness. And a held release does not fail `--check`,
because a repository with a policy would otherwise have a red check for as
long as every window lasted.

Upgrade reads its own configuration key rather than going through
`loadProviderCommandConfig`, which validates against the parse and unparse
schema and would warn "unknown key" on everything written under `upgrade`.

`--only-action` narrows a run to one action. The match happens before
anything is resolved, so a targeted run makes one request rather than one per
action in the file. Every other action is still reported, as `skipped`, and
counted: dropping them would make a targeted run over a large directory look
like a one-action file.

A target matching nothing ends the run with `errNoSuchAction`. Every result
comes back skipped in that case, which is no upgrade and no failure and reads
as a clean run, so a typo would report success having looked at nothing — the
same shape as `fmt` reporting a clean check over a path holding no HCL.

`--json` on either command reports the run as a document on stdout. The
per-action lines and the report share one writer, so under the flag the
progress is redirected to stderr — interleaved on stdout it would be
something no parser accepts, which is the one thing the flag exists to
produce. `upgrade --parse` prints "Regenerating YAML..." and that goes the
same way.

Both formats call `countPins`/`countUpgrades` and `pinOutcome`/
`upgradeOutcome`, so the summary and the exit code cannot disagree between
them. A pipeline moving to `--json` changes how the run is reported, not
whether it passes.

Each end of a move is reported as a release and a commit — `from`/`fromSHA`
and `to`/`toSHA`. `from` used to carry whatever the version line held, which
made it a tag on an unpinned action and a 40-character SHA on a pinned one,
so a consumer had to know which before it could read the field. A pinned line
names its release only in the trailing comment, which `commentTag` reads
before `keepNote` drops it.

A held release carries `published` and `eligibleAt` as dates rather than the
remaining duration. A duration is read against the clock at the moment of the
run, so a report written now and read tomorrow says something that is no
longer true. `UpgradeResult` carries both for the same reason.

Both report each action through `resolvedVerb`/`upgradedVerb`, which read
"would pin" and "would upgrade" under a dry run. Those per-action lines are
the whole of what a failed `--check` step shows, and in the past tense every
one of them claimed a change that was not on disk.

## Diff (`cinzel <provider> diff`)

`planParse` works out every file a parse would write; `ParsePlan.Diff` reads
each back and compares, then asks `fsutil.StaleGeneratedYAML` what the prune
would remove. Nothing is written.

The plan carries the prune's inputs rather than its answer. Staleness depends
on what is on disk, and under parse that means after the writes:
`sameAsCurrentOutput` compares by file identity, which is what makes a
case-only rename come out right. A provider recording no current outputs —
GitLab, which writes one pipeline and prunes nothing — is not asked, since the
walk would then read every generated file in the directory as stale.

Comparison is byte-exact once CRLF is normalised to LF. Semantic comparison
would pass a file someone hand-reformatted, which is the thing the check is
for; without normalisation every file fails on a Windows runner, because
generated files are written LF and git converts on checkout.

Exit 0 identical, 1 drifted (`ErrDiffFound`), 2 could not run
(`ErrCannotRun`, read by `main`). Only diff carries the second: every other
command still exits 1. `cli.Exit` is not used anywhere — it calls `os.Exit`
inside `Run`, which would skip the error printing in `Execute` and take a test
binary with it.

`Diff` is not on `provider.Provider`. The command reaches it through a
one-method `differ` interface, so the two test doubles implementing the
provider interface do not have to answer for it.

Removed and added lines are coloured on a terminal, through
`fsutil.UseColor`, which asks the writer it was handed rather than `os.Stdout`:
cinzel writes through an injected `io.Writer`, so a run collected into a buffer
is plain and a test does not have to own a stream. `NO_COLOR` set to anything,
including the empty string, turns it off, as does `TERM=dumb`.

The marker occupies its own column (`markerRemoved`, `markerAdded`,
`markerContext` in `internal/fsutil/difftext.go`). A unified diff conventionally
writes it against the content, which in YAML moves the indentation of the lines
whose indentation is being read.

## Format (`cinzel fmt`)

`hclwrite.Format` over every `.hcl` file in the paths `fmtPaths` works out.
Rewritten only when the bytes change, so an already-formatted tree writes
nothing.

A flag is the whole answer when one is given. Otherwise the paths come from
`.cinzelrc.yaml`, which is where a repository has already written down where
its HCL is: `unparse.output-directory` is where HCL is written, and
`parse.directory` and `parse.file` are where it is read from, across both
providers, deduplicated and sorted. `cinzel` is the fallback when the config
names none of them. Reading only that default meant a repository keeping its
HCL anywhere else had `cinzel fmt` format nothing and say so in a line that
read like success.

Finding no HCL at all is `errNoHCLFiles`, not a clean run. "0 checked" at
exit 0 passes a `--check` step having checked no file, which is the one
answer that must not come back green. A path that does not exist is marked
`UserInput`: it was typed or configured by the author, and unmarked it is
reported as cinzel's own bug, under an invitation to open an issue about a
missing directory.

Each file is parsed with `hclsyntax.ParseConfig` first. `hclwrite.Format`
works on a token stream rather than a parsed document, so it reformats HCL it
cannot parse as readily as HCL it can, and writing that back leaves a broken
file laid out neatly enough to look deliberate.

`--check` names the files that would change and exits 1 (`errUnformatted`)
without writing. That is the CI shape: a drifted file fails the step instead
of being quietly fixed in a checkout nobody commits.

No provider is involved, so it registers at the top level next to `init`
rather than under `github`/`gitlab`.
