stages = ["build", "test"]

variable "deploy_env" {
  name        = "DEPLOY_ENV"
  value       = "production"
  description = "Target environment"
}

job "build" {
  stage  = "build"
  script = ["go build ./..."]
}

job "test" {
  stage  = "test"
  script = ["go test ./..."]
}
