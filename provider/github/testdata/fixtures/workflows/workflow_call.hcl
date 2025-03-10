step "build" {
  run = "echo build"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.build]
}

workflow "wf" {
  filename = "workflow_call"

  on "workflow_call" {
    input "ref" {
      type     = "string"
      required = true
    }

    output "artifact-url" {
      value = "$${{ jobs.build.outputs.url }}"
    }
  }

  jobs = [job.build]
}
