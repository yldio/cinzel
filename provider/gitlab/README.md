# GitLab CI/CD

Converts between HCL and `.gitlab-ci.yml`.

```sh
cinzel gitlab unparse --file .gitlab-ci.yml --output-directory ./cinzel   # YAML -> HCL
cinzel gitlab parse   --file ./cinzel/.gitlab-ci.hcl --output-directory . # HCL -> YAML
```

Add `--dry-run` to print instead of writing. `cinzel gitlab -h` lists every flag.

## What it looks like

This pipeline:

```yaml
stages: [build, test]

build-app:
  stage: build
  image: golang:1.26
  script:
    - go build -o app ./...
  artifacts:
    paths: [app]

test:
  stage: test
  needs: [build-app]
  script:
    - go test ./...
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
```

becomes this HCL:

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

Four things in there are worth knowing about up front, and they cover most of
the surprises:

**A job's name becomes a block label, and labels cannot hold a `-`.** So
`build-app` is written as `job "build_app"` with `id = "build-app"` beside it.
The `id` is what goes back into the YAML, and anything pointing at the job —
`needs`, `extends` — follows the rename. Leave the `id` alone and the name
survives the round trip.

**`needs:` is spelled `depends_on`, and it holds references, not strings.**
`job.build_app` is a real reference: rename the block and this moves with it. A
`needs:` entry that carries options rather than a bare name becomes a `need {}`
block instead.

**Repeatable things are blocks; everything else is an attribute.** `rule {}`,
`cache {}`, `service {}` and `include {}` can appear more than once, so they are
blocks. `cache.key` and `service.variables` are attributes, because the shape
follows [`config.go`](config.go), not the value. If you are unsure which a
keyword is, unparse a pipeline that uses it and read what comes out.

**`$` doubles up.** `${CI_COMMIT_BRANCH}` in YAML becomes `$${CI_COMMIT_BRANCH}`
in HCL, so HCL does not try to interpolate it itself. Parse puts it back. A bare
`$CI_PIPELINE_SOURCE` with no braces, as in the example above, needs no
escaping.

## The rest of the shape

```hcl
variable "deploy_env" {
  name        = "DEPLOY_ENV"
  value       = "production"
  description = "Target environment"
}

workflow {
  rule {
    if   = "$${CI_COMMIT_BRANCH} == \"main\""
    when = "always"
  }
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

job "test" {
  extends = [template.go_base]
  stage   = "test"
  script  = ["go test ./..."]
}

include {
  local = ".gitlab/base.yml"
}
```

A `template` block is a hidden job: `template "go_base"` is `.go_base:` in YAML,
and `extends` refers to it as `template.go_base`.

A `spec:` header becomes a `spec` block and is written back as its own YAML
document ahead of the pipeline, which is where GitLab reads it.

A job does not always need a `script`. A `trigger` job has none by definition,
one that `extends` a template inherits one, and a job in the steps syntax
carries its commands under `run`.

## Things that behave in a particular way

**An empty or null keyword is kept, not dropped.** GitLab reads `cache: []` and
`rules: null` as "clear whatever this would inherit", which is different from
leaving the keyword out. Both survive the round trip — an empty collection is
written as `cache = []` or `depends_on = []` where the schema otherwise uses
blocks. A null on a keyword GitLab does not allow one on is dropped, since
writing it back would produce YAML GitLab's own schema rejects.

**`only:` and `except:`** are carried through in both their list and object
forms, though `rules:` is the modern spelling.

**A single command can be a bare string.** `script: make` works as well as
`script: [make]`, in both directions, matching GitLab's own schema.

**Top-level `image:`, `before_script:`, `after_script:`, `cache:` and
`services:`** are accepted where GitLab reads them as the `default` of the same
name. They stay where they were written rather than being folded into a
`default` block, because GitLab does not document which one wins when a pipeline
has both.

**Every document of a multi-document pipeline is read**, not just the first. A
pipeline that is nothing but `include:` entries converts like any other; a YAML
file that is not a pipeline at all is skipped.

**Parse writes one file**, `.gitlab-ci.yml`, into the output directory, with
`generated-by: cinzel` and `cinzel-provider: gitlab` at the top. Those markers
are what lets a later run recognise its own output.

## Coverage

The schema in [`config.go`](config.go) covers the documented GitLab keywords and
is checked against GitLab's own editor schema
(`app/assets/javascripts/editor/schema/ci.json`). That includes `dependencies`,
`identity`, `manual_confirmation`, `inherit`, `secrets`, `id_tokens`, `hooks`,
`pages`, `run` and `dast_configuration` on a job; `start_in`, `interruptible`,
`variables`, `needs` and `auto_cancel` on a rule; `expose_as`, `public` and
`access` on `artifacts`; `unprotect` on `cache`; `docker` and `kubernetes` on a
service; `rules` and `integrity` on an include; and `options` and `expand` on a
variable.

Converting a keyword the schema does not declare stops with an error naming it,
rather than writing a file that cannot be read back.
