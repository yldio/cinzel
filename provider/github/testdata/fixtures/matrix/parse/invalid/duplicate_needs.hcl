step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

job "release" {
  runs_on {
    runners = "ubuntu-latest"
  }

  needs = [job.build, job.build]
  steps = [step.echo]
}

workflow "ci" {
  filename = "duplicate-needs"
  on "push" {}
  jobs = [job.build, job.release]
}
