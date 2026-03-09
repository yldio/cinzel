stages = ["build"]

include {
  local = ".gitlab/base.yml"
}

include {
  template = "Jobs/Build.gitlab-ci.yml"
}

include {
  project = "group/platform"
  file    = ".gitlab-ci.yml"
  ref     = "main"
}

job "build" {
  stage  = "build"
  script = ["echo build"]
}
