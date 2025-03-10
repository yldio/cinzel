step "setup" {
  uses {
    action = "./.github/actions/setup"
  }
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.setup]
}

workflow "local_action" {
  filename = "local_action"

  on "push" {}

  jobs = [job.build]
}
