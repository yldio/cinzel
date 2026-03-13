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

  name = "Auto Release"

  permissions {
    contents = "read"
  }

  on "push" {
    branches = [
      "main"
    ]
  }

  jobs = [
    job.auto_release,
  ]
}

workflow "release_published" {
  filename = "release-published"

  name = "Build Release (Published)"

  permissions {
    contents = "read"
  }

  concurrency {
    group              = "release-$${{ github.event.release.tag_name }}"
    cancel_in_progress = true
  }

  on "release" {
    types = [
      "published"
    ]
  }

  jobs = [
    job.release-packages,
  ]
}

workflow "release" {
  filename = "release"

  name = "Release"

  permissions {
    contents = "read"
  }

  on "workflow_dispatch" {
    input "tag" {
      description = "Release tag (leave empty for auto-calculation)"
      type        = "string"
      required    = false
    }
  }

  jobs = [
    job.manual-release,
  ]
}
