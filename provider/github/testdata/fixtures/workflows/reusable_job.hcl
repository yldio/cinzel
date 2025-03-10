job "release" {
  uses = "org/repo/.github/workflows/release.yaml@v1"

  with {
    name  = "target"
    value = "prod"
  }

  secrets = "inherit"
}

workflow "wf" {
  filename = "reusable_job"

  on "push" {
    tags = ["v*"]
  }

  jobs = [job.release]
}
