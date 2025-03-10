# GitLab Pipelines Provider — Brainstorm

Date: 2026-03-09

## Context

cinzel currently supports GitHub Actions as its only provider. The architecture (`provider.Provider` interface) was designed for multiple providers from the start. GitLab CI/CD Pipelines is the second provider.

## Key decisions

### 1. Provider-specific HCL (no shared abstractions)

Each provider gets its own idiomatic HCL shape. No shared `step` blocks across providers — GitLab has no step concept (jobs are monolithic `script` blocks). Trying to abstract across fundamentally different execution models creates leaky abstractions that serve neither well.

- GitHub: `workflow`, `job`, `step`, `action` blocks
- GitLab: `job`, `template`, `variable`, `rule` blocks + top-level `stages`

### 2. Provider selection is implicit from CLI

```sh
cinzel github parse --file ./ci.hcl
cinzel gitlab parse --file ./ci.hcl
```

No `ci = "gitlab"` attribute in HCL. The CLI subcommand is the source of truth. Considered and rejected adding a `ci` variable — it would be redundant with the subcommand.

### 3. `template` block (not hidden job convention)

GitLab's dot-prefixed hidden jobs (`.my_template`) are a YAML convention. In HCL, use an explicit `template` block type:

```hcl
template "go_base" {
  image         = "golang:1.26"
  before_script = ["go mod download"]
}

job "test" {
  extends = [template.go_base]
  stage   = "test"
  script  = ["go test ./..."]
}
```

Parse mapping: `template "go_base"` → `.go_base:` in YAML.
Unparse mapping: `.go_base:` in YAML → `template "go_base"` in HCL.

Advantages over `job ".my_template"`:
- Intent is explicit in block type
- HCL identifiers can't start with `.`
- Clean reference syntax: `extends = [template.go_base]`
- Validation is distinct: templates can't have `stage`, jobs must have `script`

### 4. Variable blocks use explicit `name` attribute

Aligned with GitHub provider's `env` pattern. HCL label is an internal identifier, not the output key:

```hcl
variable "deploy_env" {
  name        = "DEPLOY_ENV"
  value       = "production"
  description = "Target environment"
}
```

This decouples HCL naming constraints from CI variable naming conventions (SCREAMING_SNAKE_CASE).

### 5. Rules as repeated singular `rule` blocks

GitLab's `rules:` list maps to repeated `rule` blocks in HCL. First match wins (order preserved from file):

```hcl
rule {
  if   = "$CI_COMMIT_BRANCH == 'main'"
  when = "on_success"
}

rule {
  if            = "$CI_PIPELINE_SOURCE == 'merge_request_event'"
  when          = "manual"
  allow_failure = true
}

rule {
  when = "never"
}
```

Follows HCL convention: singular block name when it repeats (like Terraform's `variable`, `resource`).

### 6. `depends_on` as reference list

Uses `depends_on` instead of `needs` — idiomatic HCL (matches Terraform convention) and reads more naturally than GitHub's `needs` jargon. Maps to `needs:` in both GitLab and GitHub YAML during parse:

```hcl
job "deploy" {
  depends_on = [job.build, job.test]
}
```

### 7. Project config file — `.cinzelrc.yaml`

Provider-specific preferences live in a project-level config, not in HCL or CLI flags:

```yaml
github:
  output-directory: .github/workflows

gitlab:
  output-directory: .
  single-file: true
  filename: .gitlab-ci.yml
```

Precedence: CLI flag > config file > provider defaults.

Solves: output conventions per provider, default directories, team-level agreement (checked into git).

## GitLab vs GitHub: structural differences

| Concept | GitHub Actions | GitLab CI |
|---------|---------------|-----------|
| Execution unit | Job has multiple `steps` | Job has one `script` block |
| Ordering | `needs` (pure DAG) | `stages` (ordered phases) + optional `needs` |
| Triggers | `on:` event types | `workflow:rules` + job-level `rules` |
| Inheritance | None | `extends` + hidden jobs + `default` |
| Composition | Reusable workflows, actions | `include` (local, remote, template, component) |
| File output | Multiple `.yaml` files | Single `.gitlab-ci.yml` |
| Job namespace | Under `jobs:` key | Top-level (flat) |
| Variables | `${{ }}` expressions | `$CI_*` predefined variables |

## Proposed HCL shape

```hcl
stages = ["build", "test", "deploy"]

variable "deploy_env" {
  name        = "DEPLOY_ENV"
  value       = "production"
  description = "Target environment"
}

template "go_base" {
  image         = "golang:1.26"
  before_script = ["go mod download"]
}

job "build" {
  stage  = "build"
  image  = "golang:1.26"
  script = ["go build -o app ./..."]

  artifacts {
    paths = ["app"]
  }
}

job "test" {
  extends = [template.go_base]
  stage   = "test"
  script  = ["go test ./..."]

  rule {
    if   = "$CI_PIPELINE_SOURCE == 'merge_request_event'"
    when = "on_success"
  }
}

job "deploy" {
  stage  = "deploy"
  depends_on = [job.build, job.test]
  script     = ["./deploy.sh"]

  environment {
    name = "production"
    url  = "https://app.example.com"
  }

  rule {
    if   = "$CI_COMMIT_BRANCH == 'main'"
    when = "manual"
  }

  rule {
    when = "never"
  }
}
```

## Minimum viable scope (v0.1)

1. `stages` — ordered list
2. `variable` — global variables with name/value/description
3. `job` — `script`, `image`, `stage`, `before_script`, `after_script`, `tags`
4. `rule` — job-level `if`/`when`/`changes`/`allow_failure`
5. `artifacts` and `cache`
6. `depends_on` — DAG references between jobs
7. `workflow` block — pipeline-level `rule` blocks

## Deferred to v0.2

- `template` blocks and `extends`
- `default` block
- `include` (local, remote, template, project, component)
- `trigger` (multi-project/child pipelines)
- `parallel` and `parallel:matrix`
- `environment`, `release`, `secrets`
- `services`
- `.cinzelrc.yaml` config file
