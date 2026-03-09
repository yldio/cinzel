stages = ["build"]

template "go_base" {
  image = "golang:1.26"
}

job "build" {
  extends = [template.go_base]
  stage   = "build"
  script  = ["go build ./..."]
}
