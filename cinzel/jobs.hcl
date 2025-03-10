// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

job "pull_request" {
  name = "$${{ matrix.os }}"

  strategy {
    matrix {
      variable {
        name  = "os"
        value = variable.list_os
      }
    }
  }

  runs_on {
    runners = "$${{ matrix.os }}"
  }

  timeout_minutes = 5

  steps = [
    step.checkout,
    step.go_setup,
    step.tests,
    step.coverage,
  ]
}

job "merge" {
  name = "Merge with main"

  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.checkout,
  ]
}

job "releases-matrix" {
  name = "Release Go Binary"

  runs_on {
    runners = "ubuntu-latest"
  }

  strategy {
    matrix {
      variable {
        name = "goos"
        value = [
          "linux",
          "windows",
          "darwin"
        ]
      }

      variable {
        name = "goarch"
        value = [
          "386",
          "amd64",
          "arm64"
        ]
      }

      exclude = [
        { goarch = "386", goos = "darwin" },
        { goarch = "arm64", goos = "windows" },
      ]
    }
  }

  steps = [
    step.checkout,
    step.go-release,
  ]
}

job "changelog" {
  name = "Update CHANGELOG"

  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.changelog
  ]
}
