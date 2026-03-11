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
  if   = "$${{ !github.event.release.prerelease && !github.event.release.draft }}"

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
    step.release_notes,
    step.goreleaser,
  ]
}

job "manual-release" {
  name = "Manual release"

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
    step.tests,
    step.tag_version,
    step.git_cliff_changelog,
    step.commit_release,
    step.create_release,
  ]
}
