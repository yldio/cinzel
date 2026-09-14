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
- A `needs:` entry that is an object rather than a plain job name becomes a `need {}` block, whose `job` is a `job.<id>` reference so it tracks a sanitized name like `depends_on` does.
- A job needs no `script` of its own: a `trigger` job has none by definition, and a job that `extends` a template inherits one. Both are written as `job` blocks.
- `$${VAR}` in HCL becomes `${VAR}` in YAML.
- `${VAR}` in YAML becomes `$${VAR}` in HCL output.
- Parse output is one file: `.gitlab-ci.yml` in the selected output directory.
- A pipeline of nothing but `include:` entries is unparsed like any other; a YAML file that is not a pipeline at all is still skipped.
- `only:` and `except:`, the older way to spell `rules:`, are carried through in both their list and object forms.
- A `script`, `before_script` or `after_script` may be a single command as a bare string, which is what GitLab's own schema takes, as well as a list.
- An explicitly empty `needs:`, `cache:`, `services:` or `rules:` survives the roundtrip. GitLab reads one as "override whatever this would inherit", which is not the same as leaving the keyword out, so each is written as an empty attribute where the schema otherwise uses blocks (`cache = []`, `depends_on = []`).
- A `rules:`, `artifacts:`, `cache:`, `only:` or `except:` written as null, and an empty `extends: []`, survive the roundtrip the same way. GitLab reads either spelling as clearing an inherited value. A null on any other keyword is dropped rather than written back, since GitLab's own schema does not accept one there.
- A job written in the steps syntax carries its commands under `run` instead of `script`, and needs no `script` of its own. It still needs a stage.
- `image:`, `before_script:`, `after_script:`, `cache:` and `services:` are also accepted at the top level, where GitLab reads each as the `default` of the same name. Each stays where it was written rather than being folded into a `default` block, since GitLab does not document which wins when a pipeline has both.
- A `variable` block takes `expand` and `options` alongside `value` and `description`; a variable that carries nothing else stays a plain scalar.
- Parse output includes cinzel provider markers in YAML headers (`generated-by` and `cinzel-provider`).
- `template.<id>` and `job.<id>` references in `extends` map to YAML `extends` entries.
- A job or template name that is not a valid HCL identifier is sanitized to make the block referenceable (`build-app` becomes the label `build_app`) and the original name is kept in an `id` attribute, so the name and any `needs` or `extends` pointing at it survive the roundtrip.
- Repeated `include {}` blocks map to YAML `include:` entries.
- The HCL schema in `provider/gitlab/config.go` covers the documented GitLab keywords, including `dependencies`, `identity`, `manual_confirmation`, `inherit`, `secrets`, `id_tokens`, `hooks`, `pages`, `run` and `dast_configuration` on a job, `start_in` and `interruptible` on a rule, `expose_as`/`public`/`access` on `artifacts`, `unprotect` on `cache`, `docker` and `kubernetes` on a service, `rules` and `integrity` on an include, and `options` on a variable. It is checked against GitLab's own editor schema, `app/assets/javascripts/editor/schema/ci.json`.
- A `rule` block takes `variables`, `needs`, `start_in`, `interruptible` and `auto_cancel` alongside `if`, `when`, `allow_failure`, `changes` and `exists`, for both workflow and job rules.
- Whether a nested map becomes an HCL block or an object attribute follows the schema in `provider/gitlab/config.go`, not the value's shape: `artifacts.reports` is a block, while `cache.key`, `service.variables`, `default.retry` and `include.inputs` are attributes.
- Repeated `service {}` blocks map to YAML `services:` entries under `default` or a `job`.
- Repeated `cache {}` blocks map to a YAML `cache:` list under `default` or a `job`; a single block stays a `cache:` object.
- Parse schema is defined by typed HCL structs in `provider/gitlab/config.go`; `hcl:",remain"` is used only for intentional pass-through islands.
- Unparse schema validation favors strict typed YAML decode over manual key allowlist tables.
