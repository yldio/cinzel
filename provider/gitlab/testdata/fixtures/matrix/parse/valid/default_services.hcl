stages = ["test"]

default {
  image = "alpine:3.20"

  service {
    name  = "postgres:16"
    alias = "db"
  }
}

job "test" {
  stage  = "test"
  script = ["echo test"]

  service {
    name = "redis:7"
  }
}
