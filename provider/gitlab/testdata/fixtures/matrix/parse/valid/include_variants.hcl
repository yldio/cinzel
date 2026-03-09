stages = ["build"]

include {
  local = ".gitlab/base.yml"
}

include {
  template = "Jobs/Build.gitlab-ci.yml"
}

job "build" {
  stage  = "build"
  script = ["echo build"]
}
