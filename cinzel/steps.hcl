// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

step "checkout" {
  name = "Checkout"

  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "tag_version" {
  id   = "tag_version"
  name = "Bump version and push tag"

  uses {
    action  = "mathieudutour/github-tag-action"
    version = "6.2"
  }

  with {
    name  = "github_token"
    value = "$${{ secrets.GITHUB_TOKEN }}"
  }
}

step "create_release" {
  id   = "create_release"
  name = "Create a GitHub release"

  uses {
    action  = "ncipollo/release-action"
    version = "1.9"
  }

  with {
    name  = "github_token"
    value = "$${{ secrets.GITHUB_TOKEN }}"
  }
}

step "go_setup" {
  name = "Setup Go environment"
  if   = "$${{ matrix.os != 'ubuntu-20.04' }}"

  uses {
    action  = "actions/setup-go"
    version = "v5.0.2"
  }

  with {
    name  = "go-version-file"
    value = "./go.mod"
  }
}

step "tests" {
  name = "Tests"
  run  = <<EOF
make test-ci
short_sha=$(git rev-parse --short "$GITHUB_SHA")
echo "tag=dev-$short_sha" >> $GITHUB_OUTPUT
EOF
}

step "coverage" {
  name = "Coverage"
  if   = "$${{ matrix.os == 'ubuntu-20.04' }}"

  uses {
    action  = "gwatts/go-coverage-action"
    version = "v2.0.0"
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
    value = "./cmd/cinzel"
  }

  with {
    name  = "binary_name"
    value = "cinzel"
  }

  with {
    name  = "goarch"
    value = "$${{ matrix.goarch }}"
  }

  with {
    name  = "goversion"
    value = "1.26"
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
