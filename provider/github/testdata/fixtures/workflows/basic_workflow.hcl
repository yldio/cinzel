step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "test" {
  run = "go test ./..."
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.checkout,
    step.test,
  ]
}

workflow "ci" {
  filename = "basic_workflow"

  on "push" {
    branches = ["main"]
  }

  jobs = [job.build]
}
