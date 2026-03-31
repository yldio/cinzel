workflow "multi_on" {
  filename = "multi-on"

  name = "CI"

  on "pull_request" {
  }

  on "push" {
    branches = ["main"]
  }

  jobs = [
    job.build,
  ]
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.build_step_1,
  ]
}

step "build_step_1" {
  run = "echo hi"
}
