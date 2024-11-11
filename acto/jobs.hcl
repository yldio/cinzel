job "releases-matrix" {
  name = "Release Go Binary"

  runs_on {
    runners = "ubuntu-latest"
  }

  strategy {
    matrix {
      variable {
        name = "goos"
        value = [
          "linux",
          "windows",
          "darwin"
        ]
      }

      variable {
        name = "goarch"
        value = [
          "386",
          "amd64",
          "arm64"
        ]
      }

      exclude = [
        { goarch = "386", goos = "darwin" },
        { goarch = "arm64", goos = "windows" },
      ]
    }
  }

  steps = [
    step.checkout,
    step.go-release,
    step.changelog
  ]
}