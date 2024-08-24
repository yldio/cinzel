workflow "pull_request" {
  filename = "pull-request"

  name = "Pull Request"

  on {
    events = "pull_request"
  }

  jobs = [
    job.pull_request,
  ]
}

job "pull_request" {
  name = "$${{ matrix.os }}"

  strategy {
    matrix {
      name = "os"
      value = [
        "ubuntu-20.04",
        "macos-14",
        "windows-2022"
      ]
    }
  }

  runs {
    on = "$${{ matrix.os }}"
  }

  timeout_minutes = 5

  steps = [
    step.checkout,
    step.go_setup,
    step.tests,
    step.coverage,
  ]
}

step "checkout" {
  name = "Checkout"

  uses {
    action  = "actions/checkout"
    version = "v4.1.7"
  }
}

step "go_setup" {
  name = "Setup Go environment"
  if   = "$${{ matrix.os }} != ubuntu-20.04"

  uses {
    action  = "actions/setup-go"
    version = "v5.0.2"
  }

  with {
    name  = "go-version-file"
    value = "./go.mod"
  }
}

step "tests" {
  name = "Tests"
  run  = "make test-ci"
}

step "coverage" {
  name = "Coverage"
  if   = "$${{ matrix.os }} == ubuntu-20.04"

  uses {
    action  = "gwatts/go-coverage-action"
    version = "v2.0.0"
  }
}