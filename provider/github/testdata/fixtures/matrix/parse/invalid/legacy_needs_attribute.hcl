step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  needs = [job.release]
  steps = [step.echo]
}

workflow "ci" {
  filename = "legacy-needs"
  on "push" {}
  jobs = [job.build]
}
