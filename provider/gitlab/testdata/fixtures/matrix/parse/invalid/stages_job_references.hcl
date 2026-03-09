stages = [job.build, job.test]

job "build" {
  stage  = "build"
  script = ["echo build"]
}
