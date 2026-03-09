stages = ["build", "test"]

job "build" {
  stage  = "build"
  script = ["echo build"]
}

job "test" {
  stage      = "test"
  depends_on = [job.build]
  script     = ["echo test"]
}

workflow {
  rule {
    if   = "$${CI_COMMIT_BRANCH} == \"main\""
    when = "always"
  }
}
