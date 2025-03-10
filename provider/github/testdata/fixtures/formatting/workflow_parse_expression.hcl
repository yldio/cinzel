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
  filename = "workflow-parse-expression"
  run_name = "$${{ github.workflow }} #$${{ github.run_number }}"
  on "push" {}
  jobs = [job.build]
}
