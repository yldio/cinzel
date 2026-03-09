stages = ["test"]

job "test" {
  stage  = "test"
  script = ["echo test"]

  service {
    alias = "db"
  }
}
