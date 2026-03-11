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
      description = "Release tag to publish (for example v1.2.3)"
      type        = "string"
      required    = true
    }
  }

  jobs = [
    job.manual-release,
  ]
}
