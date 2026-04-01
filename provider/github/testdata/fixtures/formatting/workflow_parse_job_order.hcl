step "echo" {
  run = "echo hi"
}

job "test" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

job "deploy" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

workflow "ci" {
  filename = "workflow-parse-job-order"
  name     = "CI"

  on "push" {}

  jobs = [job.test, job.build, job.deploy]
}
