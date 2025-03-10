workflow "pull_request" {
  filename = "pull-request"

  name = "Pull Request"

  on "pull_request" {
  }

  jobs = [
    job.pull_request,
  ]
}

job "pull_request" {
  name = "$${{ matrix.os }}"

  runs_on {
    runners = "$${{ matrix.os }}"
  }

  strategy {
    matrix {
      variable {
        name  = "os"
        value = variable.list_os
      }
    }
  }

  steps = [
    step.checkout,
  ]
}

step "checkout" {
  id = "checkout"

  name = "Checkout"

  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

variable "list_os" {
  value = ["ubuntu-24.04", "macos-15", "windows-2022"]
}
