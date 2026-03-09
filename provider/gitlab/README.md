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
```

## Notes

- HCL uses `depends_on`; YAML uses `needs:`.
- `$${VAR}` in HCL becomes `${VAR}` in YAML.
- `${VAR}` in YAML becomes `$${VAR}` in HCL output.
- Parse output is one file: `.gitlab-ci.yml` in the selected output directory.
