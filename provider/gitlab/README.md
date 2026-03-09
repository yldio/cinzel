# GitLab CI/CD Pipelines

## Usage

```sh
cinzel gitlab -h
```

### Parse HCL to YAML

Convert GitLab-oriented HCL blocks into a single `.gitlab-ci.yml` file.

```sh
cinzel gitlab parse --file ./cinzel/pipeline.hcl --output-directory .
```

### Unparse YAML to HCL

Convert `.gitlab-ci.yml` (or other GitLab CI YAML) into HCL.

```sh
cinzel gitlab unparse --file ./.gitlab-ci.yml --output-directory ./cinzel
```

Use `--dry-run` to print generated files to stdout.

## HCL shape

```hcl
stages = ["build", "test", "deploy"]

variable "deploy_env" {
  name        = "DEPLOY_ENV"
  value       = "production"
  description = "Target environment"
}

job "build" {
  stage  = "build"
  image  = "golang:1.26"
  script = ["go build -o app ./..."]
}

job "test" {
  extends    = [template.go_base]
  stage      = "test"
  depends_on = [job.build]
  script     = ["go test ./..."]

  rule {
    if   = "$${CI_PIPELINE_SOURCE} == \"merge_request_event\""
    when = "on_success"
  }
}

workflow {
  rule {
    if   = "$${CI_COMMIT_BRANCH} == \"main\""
    when = "always"
  }
}

include {
  local = ".gitlab/base.yml"
}

default {
  image = "alpine:3.20"

  service {
    name  = "postgres:16"
    alias = "db"
  }
}

template "go_base" {
  image = "golang:1.26"
}
```

## Notes

- HCL uses `depends_on`; YAML uses `needs:`.
- `$${VAR}` in HCL becomes `${VAR}` in YAML.
- `${VAR}` in YAML becomes `$${VAR}` in HCL output.
- Parse output is one file: `.gitlab-ci.yml` in the selected output directory.
- `template.<id>` and `job.<id>` references in `extends` map to YAML `extends` entries.
- Repeated `include {}` blocks map to YAML `include:` entries.
- Repeated `service {}` blocks map to YAML `services:` entries under `default` or a `job`.
- Parse schema is defined by typed HCL structs in `provider/gitlab/config.go`; `hcl:",remain"` is used only for intentional pass-through islands.
- Unparse schema validation favors strict typed YAML decode over manual key allowlist tables.
