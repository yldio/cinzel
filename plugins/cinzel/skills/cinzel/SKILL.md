---
name: cinzel
description: Write CI/CD pipelines as HCL instead of YAML, using the cinzel CLI. Covers GitHub Actions workflows and composite actions, and GitLab CI/CD pipelines, in both directions — converting an existing YAML pipeline to HCL, editing the HCL, converting it back, and pinning GitHub action versions to commit SHAs. Use when a repo has a `cinzel/` directory or a `.cinzelrc.yaml`, when the user asks to convert a pipeline to or from HCL, or when they ask to pin or upgrade GitHub action versions.
allowed-tools:
  - Bash(cinzel *)
---

# cinzel

Converts CI/CD pipelines between YAML and HCL, both directions. Two providers:
`github` and `gitlab`.

- **unparse** = YAML → HCL
- **parse** = HCL → YAML

The names read from HCL's point of view, which catches people out. `parse`
reads HCL and writes YAML.

## Before anything

```sh
cinzel --version
```

Not installed: `brew tap yldio/cinzel && brew install --cask cinzel`, or
`go install github.com/yldio/cinzel@latest`. Do not hand-write HCL in a repo
that has no cinzel — the generated YAML is what CI actually runs, and without
the tool nobody can regenerate it.

## Is this repo a cinzel repo?

A `cinzel/` directory holding `.hcl` files, or a `.cinzelrc.yaml`. If so, the
HCL is the source and the YAML under `.github/workflows/` is generated: edit
the HCL and re-run `parse`. Editing the YAML directly is lost on the next run.

Generated YAML carries `generated-by: cinzel` and `cinzel-provider: <name>` in
a header comment. That marker is how cinzel recognises its own output when it
prunes stale files, so do not strip it.

## Converting a pipeline you already have

```sh
cinzel github unparse --file .github/workflows/ci.yaml --output-directory ./cinzel
cinzel gitlab unparse --file .gitlab-ci.yml --output-directory ./cinzel
```

Edit the HCL, then convert back:

```sh
cinzel github parse --file ./cinzel/ci.hcl --output-directory .github/workflows
cinzel gitlab parse --file ./cinzel/.gitlab-ci.hcl --output-directory .
```

`--dry-run` on either prints instead of writing. Use it first when you are not
sure what a change produces.

Shared flags: `--file`/`-f` one file, `--directory`/`-d` a directory (not with
`--file`), `--recursive`/`-r` walk into subdirectories, `--output-directory`,
`--dry-run`. Parse also takes `--yml` for a `.yml` extension; GitLab parse
always writes `.gitlab-ci.yml` regardless.

## What the HCL looks like

GitHub — steps and jobs are top-level blocks, referred to by name:

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

GitLab — jobs are blocks, repeatable things (`rule`, `cache`, `service`,
`include`) are nested blocks:

```hcl
stages = ["build", "test"]

job "build_app" {
  id = "build-app"

  artifacts {
    paths = ["app"]
  }

  image  = "golang:1.26"
  script = ["go build -o app ./..."]
  stage  = "build"
}

job "test" {
  depends_on = [
    job.build_app,
  ]

  rule {
    if = "$CI_PIPELINE_SOURCE == \"merge_request_event\""
  }

  script = ["go test ./..."]
  stage  = "test"
}
```

## The five things that trip people up

**A name with a `-` becomes a block label with `_`, and `id` keeps the
original.** `build-and-test` is `job "build_and_test"` with `id =
"build-and-test"`. The `id` is what goes back to YAML. Deleting it renames the
job in CI and breaks anything depending on it.

**Lists of jobs and steps are references, not strings.** `job.build_app`,
`step.checkout`. Rename a block and every reference follows. A reference to a
block that does not exist is an error, not a silent skip.

**`needs:` is written `depends_on`** in both providers.

**`$` doubles.** `${{ github.sha }}` → `$${{ github.sha }}`, `${CI_COMMIT_SHA}`
→ `$${CI_COMMIT_SHA}`. Never write the escape by hand — the converter owns it
both ways, and hand-escaping produces a literal `$$` in the YAML. A bare
`$CI_PIPELINE_SOURCE` with no braces needs nothing.

**`ignore_id = true` on a GitHub step** records that the step had no `id:` in
the YAML. Parse otherwise gives a step its block label as an `id`. Leave it
where unparse put it; remove it only if you want that step to gain an `id`.

## Pinning GitHub action versions

GitHub only, and operates on the HCL, not the YAML:

```sh
cinzel github pin                # rewrite every action tag as the SHA it points at
cinzel github upgrade            # bump to latest, then pin
cinzel github upgrade --parse    # and regenerate the YAML
```

Both take `--dry-run`, `--file`/`-f`, `--directory`/`-d` (default `cinzel`).
No token needed for public actions; `GITHUB_TOKEN` in the environment raises
the rate limit from 60/hour to 5000.

Run `parse` after `pin` or the YAML still carries the old tags.

## Defaults in a file

`.cinzelrc.yaml` in the working directory, so the flags need not be typed:

```yaml
github:
  parse:
    directory: ./cinzel
    output-directory: .github/workflows
```

Command-line flags win. It covers `parse` and `unparse` only — `assist`, `pin`,
`upgrade`, `init` ignore it. Paths must be relative with forward slashes; the
file is committed and read on every machine, so an absolute path or a `~` is
refused.

## When something fails

cinzel refuses rather than writing a file that cannot be read back. An error
naming an unknown keyword means the schema does not cover it — check the
provider README in the cinzel repo before assuming it is a bug.

cinzel does not lint. It converts. Whether GitHub or GitLab accepts the
resulting pipeline is `actionlint`'s question, not cinzel's.

## Full reference

In the cinzel repo, if it is checked out: `provider/github/README.md` and
`provider/gitlab/README.md`. Otherwise
<https://github.com/yldio/cinzel>.
