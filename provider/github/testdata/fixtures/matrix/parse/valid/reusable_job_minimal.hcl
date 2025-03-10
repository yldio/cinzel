job "release" {
  uses = "org/repo/.github/workflows/release.yaml@v1"

  with {
    name  = "environment"
    value = "prod"
  }

  secrets = "inherit"
}

workflow "wf" {
  filename = "reusable-job-minimal"
  on "push" {}
  jobs = [job.release]
}
