stages = ["build"]

workflow {
  name = "Main pipeline"

  rule {
    if   = "$${CI_COMMIT_BRANCH} == \"main\""
    when = "always"
  }

  rule {
    when = "never"
  }
}

job "build" {
  stage  = "build"
  script = ["echo build"]
}
