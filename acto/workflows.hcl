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