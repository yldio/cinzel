step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "go-release" {
  uses {
    action  = "wangyoucao577/go-release-action"
    version = "v1"
  }

  with {
    name  = "github_token"
    value = "$${{ secrets.GITHUB_TOKEN }}"
  }

  with {
    name  = "goos"
    value = "$${{ matrix.goos }}"
  }

  with {
    name  = "goarch"
    value = "$${{ matrix.goarch }}"
  }

  with {
    name  = "project_path"
    value = "./cmd/acto"
  }

  with {
    name  = "binary_name"
    value = "acto"
  }

  with {
    name  = "goarch"
    value = "$${{ matrix.goarch }}"
  }

  with {
    name  = "goversion"
    value = "1.23.2"
  }

  with {
    name  = "ldflags"
    value = "-s -w -X \"main.version={{.Version}}\""
  }
}

step "changelog" {
  uses {
    action  = "orhun/git-cliff-action"
    version = "v3"
  }

  with {
    name  = "config"
    value = "cliff.toml"
  }

  env {
    name  = "OUTPUT"
    value = "CHANGELOG.md"
  }

  env {
    name  = "GITHUB_REPO"
    value = "$${{ github.repository }}"
  }
}