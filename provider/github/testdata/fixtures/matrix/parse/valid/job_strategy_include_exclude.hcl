step "test" {
  run = "go test ./..."
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  strategy {
    fail_fast    = false
    max_parallel = 2

    matrix {
      variable = [{
        name  = "os"
        value = ["ubuntu-latest", "macos-latest"]
      }]

      include = [{
        os   = "ubuntu-latest"
        node = "20"
      }]

      exclude = [{
        os   = "macos-latest"
        node = "20"
      }]
    }
  }

  steps = [step.test]
}

workflow "ci" {
  filename = "job-strategy-include-exclude"
  on "push" {}
  jobs = [job.build]
}
