stages = ["test"]

job "test" {
  stage  = "test"
  script = ["go test ./..."]

  artifacts {
    paths     = ["coverage.out"]
    expire_in = "1 week"
  }

  cache {
    key   = "go-modules"
    paths = ["vendor/"]
    when  = "always"
  }

  rule {
    if   = "$${CI_PIPELINE_SOURCE} == \"merge_request_event\""
    when = "on_success"
  }

  rule {
    when = "never"
  }
}
