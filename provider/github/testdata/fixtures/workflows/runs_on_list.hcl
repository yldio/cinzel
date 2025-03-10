step "echo" {
  run = "echo hello"
}

job "build" {
  runs_on {
    runners = ["self-hosted", "linux", "arm64"]
  }

  steps = [step.echo]
}

workflow "runs_on_list" {
  filename = "runs_on_list"

  on "push" {}

  jobs = [job.build]
}
