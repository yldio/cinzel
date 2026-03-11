// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

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
    step.mise_setup,
    step.tests,
    step.coverage,
  ]
}

job "merge" {
  name = "Merge with main"

  timeout_minutes = 5

  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.checkout,
  ]
}

job "release-packages" {
  name = "Release packages"
  if   = "$${{ github.event_name == 'workflow_dispatch' || (!github.event.release.prerelease && !github.event.release.draft) }}"

  timeout_minutes = 20

  permissions {
    contents = "write"
  }

  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.checkout_release,
    step.mise_setup,
    step.goreleaser,
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
