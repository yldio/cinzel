stages = ["build", "test", "deploy"]

job "build" {
  stage  = "build"
  script = ["echo build"]
}

job "test" {
  stage      = "test"
  depends_on = [job.build]
  script     = ["echo test"]
}

job "deploy" {
  stage      = "deploy"
  depends_on = [job.build, job.test]
  script     = ["echo deploy"]
}
