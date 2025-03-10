# GitHub Actions

## Usage

```sh
cinzel github -h
```

### Parse HCL to YAML

Convert HCL files containing `workflow`, `job`, `step`, and `action` blocks into GitHub Actions YAML files.

```sh
cinzel github parse --file ./cinzel/steps.hcl --output-directory .github/workflows
```

### Unparse YAML to HCL

Convert GitHub Actions workflow YAML, composite action YAML, or step-only YAML back to HCL.

```sh
cinzel github unparse --file ./.github/workflows/steps.yaml --output-directory ./cinzel
```

Use `--dry-run` to print generated files to stdout.

## HCL shape: workflows

```hcl
step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "test" {
  run = "go test ./..."
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.checkout,
    step.test,
  ]
}

workflow "ci" {
  filename = "ci"

  on "push" {
    branches = ["main"]
  }

  jobs = [
    job.build,
  ]
}
```

## HCL shape: composite actions

```hcl
step "setup" {
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
  name = "Build"
  run  = "npm run build"
}

action "my_action" {
  filename    = "my-action"
  name        = "My Composite Action"
  description = "Runs setup and build"

  input "node_version" {
    description = "Node.js version"
    required    = true
    default     = "20"
  }

  output "result" {
    description = "Build result"
    value       = "$${{ steps.build.outputs.result }}"
  }

  runs {
    using = "composite"
    steps = [step.setup, step.build]
  }
}
```

Actions are written to `<output-directory>/<filename>/action.yml` during parse.

## Notes

- `workflow.jobs`, `job.steps`, and `action.runs.steps` are explicit references (`job.<id>` and `step.<id>`).
- YAML unparse generates stable HCL identifiers (sanitized when YAML keys contain `-`).
- Document type is auto-detected during unparse: workflow (has `on`/`jobs`), action (has `name`/`runs`, no `on`/`jobs`), or step-only.

## Validation rules

The same rules are enforced in both directions (HCL → YAML and YAML → HCL):

- A `workflow` must define at least one `on` trigger and at least one job reference.
- A normal `job` (without `uses`) must define `runs_on` and at least one referenced step.
- A reusable `job` (with `uses`) cannot define `runs_on` or `steps`.
- `with` and `secrets` are only valid for reusable jobs (`uses` set).
- `permissions` scopes and levels are validated against the known GitHub Actions set.
- `on.schedule` cron expressions are validated (5-field format, value ranges).
- `${{ }}` expression syntax is checked for balanced delimiters and non-empty bodies.
- Step `uses` references are validated for correct format (`owner/repo@ref`, `./path`, or `docker://image`).
- An `action` must define `name` and `runs.using`.

## Testing

Test coverage includes golden fixtures, a fixture-driven compatibility matrix under `provider/github/testdata/fixtures/matrix`, strict validation checks, semantic roundtrip tests, and benchmarks under `provider/github/*_test.go`.

## Fixture matrix workflow

- Add new parse scenarios in `provider/github/testdata/fixtures/matrix/parse`.
- Add new unparse scenarios in `provider/github/testdata/fixtures/matrix/unparse`.
- For valid parse scenarios, add `<name>.hcl` and `<name>.golden.yaml`.
- For invalid parse scenarios, add `<name>.hcl` and `<name>.error.txt`.
- For valid unparse scenarios, add `<name>.yaml` and `<name>.roundtrip.golden.yaml`.
- For invalid unparse scenarios, add `<name>.yaml` and `<name>.error.txt`.
- Keep `.error.txt` messages focused on stable substrings, not full error text.

## Coverage

- HCL `workflow`, `job`, `step`, and `action` reference graph with explicit references (`workflow.jobs`, `job.steps`, `job.needs`, `action.runs.steps`).
- Action support for all `runs.using` types: `composite`, `node20`, and `docker`, in both parse and unparse directions.
- `action` blocks support `input`, `output`, `runs`, and `branding` sub-blocks, written to `<filename>/action.yml`.
- Common workflow triggers and event maps, including empty event blocks for trigger-only events.
- YAML workflow `on` shorthand forms (`on: push` and `on: [push, pull_request]`) are normalized during unparse.
- YAML `on.schedule` list form is normalized to HCL `cron` list in `on "schedule"` blocks.
- Standard jobs (`runs-on`, `steps`) and reusable jobs (`uses`, `with`, `secrets`) with strict validation.
- Job-level common keys parse/unparse: `if`, `timeout-minutes`, `continue-on-error`, `permissions`, `defaults`, `concurrency`, `container`, `services`, `environment`, and `strategy` (`matrix`, `include`, `exclude`, `fail-fast`, `max-parallel`).
- Step `run` and structured `uses { action, version }` parsing and unparsing.
- GitHub expression strings (`${{ ... }}`) are preserved across workflow, job, and step fields in parse/unparse flows.
- Expression syntax validation (balanced `${{ }}` delimiters, non-empty bodies).
- Permissions validation (known scopes and levels).
- Cron expression validation (5-field format, value ranges).
- Uses reference validation (`owner/repo@ref`, `./path`, `docker://image`).
- Strategy matrix normalization for supported shapes (`matrix.variable` in HCL and matrix axes in YAML).
- Auto-detection of document type during unparse (workflow, action, or step-only).
- Semantic roundtrip stability for covered fixtures (HCL → YAML → HCL → YAML).

## Known limitations

- Not every GitHub Actions schema edge case or uncommon field combination is covered. The most common workflow, job, step, and action fields are supported.
- Roundtrip output is semantically stable but not byte-stable: key ordering and formatting may normalize even when the meaning is preserved.

## Output guarantees

- Parse output uses unquoted `on` (never `"on"`).
- Parse output uses 2-space YAML indentation.
- Parse output top-level key order is deliberate (`name`, `on`, `jobs`, then remaining keys).
- Unparse output formats HCL with clear section separators and trailing commas in reference lists.
- Identifier normalization is stable: YAML names are sanitized to valid HCL identifiers when needed.
- Expression escaping is stable in HCL output (`${{ ... }}` in YAML becomes `$${{ ... }}` in HCL string literals).
