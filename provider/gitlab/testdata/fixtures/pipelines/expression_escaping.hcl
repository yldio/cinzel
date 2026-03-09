stages = ["build"]

job "build" {
  stage = "build"
  script = [
    "echo $${CI_COMMIT_BRANCH}",
    "echo $CI_PIPELINE_SOURCE",
  ]

  rule {
    if   = "$${CI_PIPELINE_SOURCE} == \"push\""
    when = "on_success"
  }
}
