step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  mystery {}
  steps = [step.echo]
}

workflow "ci" {
  filename = "unknown-job-block"
  on "push" {}
  jobs = [job.build]
}
