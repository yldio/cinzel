step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = "not-a-reference-list"
}

workflow "ci" {
  filename = "steps-wrong-type"
  on "push" {}
  jobs = [job.build]
}
