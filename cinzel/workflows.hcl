// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

workflow "pull_request" {
  filename = "pull-request"

  name = "Pull Request"

  on "pull_request" {}

  jobs = [
    job.pull_request,
  ]
}

workflow "push" {
  filename = "push"

  name = "Merge with Main"

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

  on "release" {
    types = [
      "created"
    ]
  }

  jobs = [
    job.releases-matrix
  ]
}
