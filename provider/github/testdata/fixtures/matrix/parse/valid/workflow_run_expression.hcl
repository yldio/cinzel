step "echo" {
  run = "echo hi"
}

job "build" {
  if = "$${{ github.ref == 'refs/heads/main' }}"

  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

workflow "ci" {
  filename = "workflow-run-expression"
  run_name = "$${{ github.workflow }} #$${{ github.run_number }}"

  on "workflow_run" {
    workflows = ["Build"]
    types     = ["completed"]
  }

  jobs = [job.build]
}
