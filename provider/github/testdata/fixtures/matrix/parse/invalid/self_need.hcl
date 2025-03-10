step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  needs = [job.build]
  steps = [step.echo]
}

workflow "ci" {
  filename = "self-need"
  on "push" {}
  jobs = [job.build]
}
