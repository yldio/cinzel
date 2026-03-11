// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

workflow "pull_request" {
  filename = "pull-request"

  name = "Pull Request"

  permissions {
    contents      = "read"
    pull_requests = "write"
  }

  on "pull_request" {}

  jobs = [
    job.pull_request,
  ]
}

workflow "push" {
  filename = "push"

  name = "Merge with Main"

  permissions {
    contents = "read"
  }

  on "push" {
    branches = [
      "main"
    ]
  }

  jobs = [
    job.merge
  ]
}

workflow "release" {
  filename = "release"

  name = "Build Release"

  permissions {
    contents = "read"
  }

  concurrency {
    group              = "release-$${{ github.event_name == 'workflow_dispatch' && github.run_id || github.event.release.tag_name }}"
    cancel_in_progress = true
  }

  on "release" {
    types = [
      "published"
    ]
  }

  on "workflow_dispatch" {}

  jobs = [
    job.release-packages,
  ]
}
