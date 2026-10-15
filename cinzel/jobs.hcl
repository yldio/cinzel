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
    step.lint,
    step.drift,
    step.coverage,
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
    step.ensure_release_app,
    step.release_app_token,
    step.checkout_release,
    step.mise_setup,
    step.goreleaser,
  ]
}

job "release" {
  name = "Release"

  timeout_minutes = 20

  permissions {
    contents = "write"
  }

  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.ensure_release_app,
    step.release_app_token,
    step.checkout_release,
    step.verify_release_token,
    step.mise_setup,
    step.tests,
    step.calculate_next_version,
    step.resolve_release_tag,
    step.tag_version,
    step.git_cliff_changelog,
    step.git_cliff_release_notes,
    step.commit_release,
    step.create_release,
  ]
}

job "pages" {
  name = "Pages"

  timeout_minutes = 5

  permissions {
    contents = "read"
    pages    = "write"
    id_token = "write"
  }

  concurrency {
    group              = "pages"
    cancel_in_progress = false
  }

  environment {
    name = "github-pages"
    url  = "$${{ steps.pages_deploy.outputs.page_url }}"
  }

  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.checkout,
    step.pages_configure,
    step.pages_upload,
    step.pages_deploy,
  ]
}
