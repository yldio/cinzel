# GitHub Actions

Converts between HCL and GitHub Actions YAML: workflows, composite actions, and
files that hold nothing but steps.

```sh
cinzel github unparse --file .github/workflows/ci.yaml --output-directory ./cinzel  # YAML -> HCL
cinzel github parse   --file ./cinzel/ci.hcl --output-directory .github/workflows   # HCL -> YAML
```

```sh
cinzel github diff    # what would parse change? writes nothing
```

Add `--dry-run` to print instead of writing. `--file` names one file and
`--directory` a directory of them. `cinzel github -h` lists every flag.

## What it looks like

This workflow:

```yaml
name: CI
on:
  push:
    branches: [main]
jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Test
        run: go test ./...
```

becomes this HCL:

```hcl
workflow "ci" {
  filename = "ci"

  name = "CI"

  on "push" {
    branches = ["main"]
  }

  jobs = [
    job.build_and_test,
  ]
}

job "build_and_test" {
  id = "build-and-test"

  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.checkout,
    step.test,
  ]
}

step "checkout" {
  ignore_id = true

  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "test" {
  ignore_id = true

  name = "Test"

  run = "go test ./..."
}
```

Five things in there are worth knowing about up front, and they cover most of
the surprises:

**Steps are top-level blocks, referred to by name.** A workflow lists jobs and a
job lists steps, both as references — `job.build_and_test`, `step.checkout` —
rather than nesting them. Rename a block and every reference to it follows.
Write a step once and several jobs can use it.

**Names become block labels, and labels cannot hold a `-`.** So
`build-and-test` is written as `job "build_and_test"` with `id =
"build-and-test"` beside it. The `id` is what goes back into the YAML, and a
`needs` pointing at the job follows the rename. Leave the `id` alone and the
name survives the round trip.

**`ignore_id = true` means "this step had no `id:`".** Parse gives a step the
block label as its `id` by default, which is convenient when you are writing
HCL by hand. A step converted from YAML that never had one carries `ignore_id`
so it does not gain an `id` it did not ask for. Delete the line if you want the
default back.

**`needs:` is spelled `depends_on`.** It holds `job.<label>` references, same as
`jobs` and `steps` do.

**`${{ }}` doubles its `$`.** `${{ github.sha }}` in YAML becomes `$${{
github.sha }}` in HCL, so HCL does not try to interpolate it. Parse puts it
back. Never write the escape by hand — the converter owns it in both
directions.

Comments come across too, in both directions: above a block, beside a value,
and closing a body. A note explaining why a version is held back survives the
conversion rather than being dropped on the way.

A multi-line string keeps its `|` block, but two things do not make the trip.
A folded `>` value is folded into one line by the YAML reader before cinzel
sees it, so nothing records that it was wrapped and it comes back as one long
line. And a script written `run: |-` or `run: |+` comes back as `run: |`: a
script is written as an HCL heredoc, a heredoc ends on its closing marker's
own line, and so it holds exactly one trailing newline. Unparse names each
script it rewrites. The value is unchanged in every case, and a shell reads a
missing or extra final newline the same as a present one.

The one that does not is the comment above `env`, `with`, `outputs` or
`secrets`. Each of those is one mapping in YAML and one block per key in HCL,
so the mapping has no block of its own to carry it; writing it above the first
block would turn a note about the whole set into a note about whichever key
sorts first. Unparse drops it and prints where it was, as `file:line: path`.
`permissions`, `defaults` and `concurrency` stay a single block and keep
theirs, as do the comments on a mapping's own keys and the one closing it.

## Values shared between blocks

A `variable` block holds a value other blocks point at, the same way `step` and
`job` blocks are pointed at. It is an HCL value, resolved at conversion time
and gone by the time GitHub sees anything — not a workflow `env` and not a
GitLab CI variable, which is a different thing under the same word.

```hcl
variable "list_os" {
  value = ["ubuntu-24.04", "macos-15", "windows-2022"]
}

job "test" {
  strategy {
    matrix {
      variable {
        name  = "os"
        value = variable.list_os
      }
    }
  }

  runs_on {
    runners = "$${{ matrix.os }}"
  }

  steps = [step.test]
}
```

The matrix expands to the three runners. Another job wanting the same spread
writes `variable.list_os` too, and one edit moves both.

## Composite actions

An action is the other document type. It uses `input`, `output` and `runs`
blocks, and shares the same `step` blocks a workflow uses:

```hcl
step "setup_node" {
  ignore_id = true

  name = "Setup Node"

  uses {
    action  = "actions/setup-node"
    version = "v4"
  }

  with {
    name  = "node-version"
    value = "20"
  }
}

step "build" {
  ignore_id = true

  name = "Build"

  run = "npm run build"
}

action "act" {
  filename = "act"

  description = "Runs setup and build"

  name = "My Composite Action"

  input "node_version" {
    default     = "20"
    description = "Node.js version"
    required    = true
  }

  runs {
    using = "composite"

    steps = [
      step.setup_node,
      step.build,
    ]
  }
}
```

Parse writes this to `<output-directory>/act/action.yml`, since that is where
GitHub looks for one. `runs.using` may be `composite`, `node20` or `docker`.

### Where an action lives

A workflow has one home — GitHub reads `.github/workflows` and nowhere else —
but an action can sit anywhere the repository puts it. A `path` says where:

```hcl
action "setup" {
  filename = "setup"
  path     = "actions-x"
}
```

That writes `actions-x/setup/action.yml`. Unparse fills the path in from
wherever it found the file, so if you keep actions in folders of their own they
convert out and back without you editing anything:

```sh
cinzel github unparse -d actions-x -d actions-y -r --output-directory cinzel
cinzel github parse -d cinzel
```

You need `--recursive` because an action file always sits a level below the
folder you point at: `actions-x/setup/action.yml`.

You can repeat `--directory` on unparse for exactly this, and cinzel adds
`.github/workflows` when you have not named it, so one run takes the whole
repository:

```sh
cinzel github unparse -d actions-x -d actions-y -r --output-directory cinzel
cinzel github parse -d cinzel
# .github/workflows/ci.yaml, actions-x/setup/action.yml and
# actions-y/deploy/action.yml, each back where it came from
```

With `--recursive`, passing one directory that contains another is refused,
since the files underneath would be read twice.

Parse starts an action's path at the repository root, not the output directory,
because that is what the path was recorded against. Pass `--output-directory`
and everything lands under it instead, which is what you want for a scratch
conversion.

An action with no `path` goes under the output directory as it always did, so
nothing you wrote before this existed changes. A `path` is fenced the same way
`filename` is: relative, no `..`, nothing absolute. An action read from outside
the run's root gets no path at all, since an absolute one names one machine's
disk and the HCL is meant to be committed.

Unparse works out what it is looking at on its own: `on` and `jobs` make it a
workflow, `name` and `runs` without those make it an action, and anything else
is read as a file of steps.

## Pinning and upgrading

These two are GitHub-only, and operate on HCL in place:

```sh
cinzel github pin       # rewrite every action tag as the commit SHA it points at
cinzel github upgrade   # bump to the latest version, then pin
```

Both take `--dry-run`, `--file` and `--directory` (default `cinzel`).
`upgrade --parse` regenerates the YAML afterwards. Either command exits
non-zero when an action could not be resolved, so a CI step calling one fails
on a workflow still holding an unpinned action. No token is needed for public
actions; `GITHUB_TOKEN` raises the rate limit from 60 to 5000 an hour.

## What gets checked

The same rules apply in both directions, so HCL that converts is HCL that
converts back:

- A `workflow` needs at least one `on` trigger and at least one job.
- A normal `job` needs `runs_on` and at least one step. A reusable one (with
  `uses`) must have neither, and is the only kind that takes `with` or
  `secrets`.
- An `action` needs `name` and `runs.using`.
- `permissions` scopes and levels are checked against the known set.
- `on.schedule` crons are checked for the 5-field form and sane ranges.
- `${{ }}` is checked for balanced delimiters and a non-empty body.
- A step's `uses` must read as `owner/repo@ref`, `./path`, `$/path` or
  `docker://image`. A `$/` reference names this repository at the commit the
  caller runs from, so it carries no `@ref` and one is refused. It is not
  available on GitHub Enterprise Server.

A file that fails one of these stops with an error naming the problem, rather
than writing YAML that GitHub would reject.

A YAML key the schema does not declare stops the conversion the same way.
`unparse --ignore-unknown` drops those keys instead, naming each with its file
and line, and reports every one rather than stopping at the first. The dropped
key is gone from the HCL and so from the YAML the next parse writes over the
source, and `diff` reports no drift afterwards, so it is for a keyword cinzel
has not caught up with rather than for a typo. The accepted keys are read off
the shape structs in `validate.go`, so a schema that grows a field needs no
second list.

## Coverage

Workflows cover the common triggers and event maps, including the `on: push`
and `on: [push, pull_request]` shorthands and the `on.schedule` list form.

Jobs cover `if`, `timeout-minutes`, `continue-on-error`, `permissions`,
`defaults`, `concurrency`, `container`, `services`, `environment` and
`strategy` (with `matrix`, `include`, `exclude`, `fail-fast` and
`max-parallel`).

Steps cover `run` and the structured `uses { action, version }` form.

Actions cover `input`, `output`, `runs` and `branding` for all three
`runs.using` values.

Not every schema corner is covered. If you hit one, the converter stops and
says so rather than guessing.

## Known limitations

**Round trips are semantically stable, not byte-stable.** Key order and
formatting normalise even where the meaning is untouched.

**Keys inside `container`, `defaults`, `concurrency` and `environment` become
HCL identifiers**, and parse reads `_` back as `-`. A key holding an underscore
or a dot is refused on unparse rather than written out corrupted. No key in
GitHub's own schema for these blocks is affected.

**Two workflows cannot share a name.** A recursive unparse over directories
each holding a `ci.yaml` is refused, naming both files. GitHub reads no
subdirectory of `.github/workflows`, so the two cannot be told apart by path:
writing one as `a/ci` produces a file that converts cleanly and never runs.
Rename one, or convert the directories separately.

The same holds for two actions. A workflow and an action may share a name,
though — they are written to `ci.yaml` and `ci/action.yml`, two files in two
places — so a check workflow beside the composite action it calls converts
without complaint.

## Generated output

Parse opens every file it writes with three lines:

```yaml
# generated-by: cinzel
# cinzel-provider: github
# cinzel-source: cinzel/ci.hcl
```

They are what lets a later run recognise its own output and delete what it no
longer writes, which is how a renamed workflow's old YAML gets cleaned up. The
provider line makes the file cinzel's — one without it was written by hand and
is never touched. The source line makes it one particular HCL file's, so a run
narrowed with `--file` prunes only what that file produced and leaves every
other workflow's output alone.

A step-only parse records no source, since it collects steps from across the
input and has no single file to name. If a file you expected to disappear is
still there, check its markers first, then the output directory's permissions.

`cinzel github diff` reports the same three answers without writing: a file
that differs, one parse would create, and one it would prune. The last is
why the command exists rather than a `diff -ru` against a freshly parsed
copy — an orphaned output looks exactly like a missing file to a directory
comparison.

Beyond that: `on` is written unquoted, indentation is two spaces, and the
top-level key order is `name`, `on`, `jobs`, then the rest. Nothing is added
that the HCL did not say: a workflow declaring no `permissions` is written
without one, and whether a pipeline ought to declare it is between you and
your linter.

## Contributing

Tests live in `provider/github/*_test.go`, with fixtures under
`testdata/fixtures/matrix`. To add a case, drop the files in and it is picked
up:

| Scenario | Files |
| --- | --- |
| Valid parse | `parse/<name>.hcl` + `<name>.golden.yaml` |
| Invalid parse | `parse/<name>.hcl` + `<name>.error.txt` |
| Valid unparse | `unparse/<name>.yaml` + `<name>.roundtrip.golden.yaml` |
| Invalid unparse | `unparse/<name>.yaml` + `<name>.error.txt` |

Keep `.error.txt` to a stable substring rather than the full message.

The HCL schema is the typed structs in [`config.go`](config.go), not an
allowlist table; unparse validates by strict typed YAML decode. Adding a field
means the struct, then the conversion, then a fixture.
