stages = ["build"]

job "build" {
  stage  = "build"
  script = ["echo build"]
  foo    = "bar"
}
