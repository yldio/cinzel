# GitHub Actions

Converts between HCL and GitHub Actions YAML: workflows, composite actions, and
files that hold nothing but steps.

```sh
cinzel github unparse --file .github/workflows/ci.yaml --output-directory ./cinzel  # YAML -> HCL
cinzel github parse   --file ./cinzel/ci.hcl --output-directory .github/workflows   # HCL -> YAML
```

Add `--dry-run` to print instead of writing. `cinzel github -h` lists every flag.

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

Unparse works out which document it is looking at on its own: `on` and `jobs`
make it a workflow, `name` and `runs` without those make it an action, and
anything else is read as a file of steps.

## Pinning and upgrading

These two are GitHub-only, and operate on HCL in place:

```sh
cinzel github pin       # rewrite every action tag as the commit SHA it points at
cinzel github upgrade   # bump to the latest version, then pin
```

Both take `--dry-run`, `--file` and `--directory` (default `cinzel`).
`upgrade --parse` regenerates the YAML afterwards. No token is needed for
public actions; `GITHUB_TOKEN` raises the rate limit from 60 to 5000 an hour.

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
- A step's `uses` must read as `owner/repo@ref`, `./path` or `docker://image`.

A file that fails one of these stops with an error naming the problem, rather
than writing YAML that GitHub would reject.

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

## Generated output

Parse writes `generated-by: cinzel` and `cinzel-provider: github` at the top of
every file. That is how a later run recognises its own output: stale YAML is
pruned only when both markers match. If a file you expected to disappear is
still there, check its markers first, then the output directory's permissions.

Beyond that: `on` is written unquoted, indentation is two spaces, and the
top-level key order is `name`, `on`, `jobs`, then the rest.

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
