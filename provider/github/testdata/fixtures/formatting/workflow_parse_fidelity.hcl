step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  permissions {
    contents = "read" # only needed for private repos
    actions  = "read" # required by the action
  }

  steps = [step.checkout]
}

workflow "ci" {
  filename = "workflow-parse-fidelity"
  name     = "Fidelity ✅"

  on "push" {}

  defaults {}

  jobs = [job.build]
}
