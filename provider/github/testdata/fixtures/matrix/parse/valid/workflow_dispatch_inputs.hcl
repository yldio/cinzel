step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

workflow "wf" {
  filename = "workflow-dispatch-inputs"

  on "workflow_dispatch" {
    input "target" {
      type     = "string"
      required = true
    }
  }

  jobs = [job.build]
}
