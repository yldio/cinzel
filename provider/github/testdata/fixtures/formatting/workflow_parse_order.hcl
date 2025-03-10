step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

workflow "ci" {
  filename = "workflow-parse-order"
  name     = "Build"
  run_name = "Build #1"
  on "push" {}
  jobs = [job.build]
}
