// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

step "checkout" {
  name = "Checkout"

  // actions/checkout v6.0.2
  uses {
    action  = "actions/checkout"
    version = "de0fac2e4500dabe0009e67214ff5f5447ce83dd"
  }

  with {
    name  = "persist-credentials"
    value = "false"
  }
}

step "checkout_release" {
  name = "Checkout (full history)"

  // actions/checkout v6.0.2
  uses {
    action  = "actions/checkout"
    version = "de0fac2e4500dabe0009e67214ff5f5447ce83dd"
  }

  with {
    name  = "fetch-depth"
    value = "0"
  }

  with {
    name  = "persist-credentials"
    value = "false"
  }
}

step "mise_setup" {
  name = "Setup mise"

  // jdx/mise-action v3.6.3
  uses {
    action  = "jdx/mise-action"
    version = "5228313ee0372e111a38da051671ca30fc5a96db"
  }

  with {
    name  = "install"
    value = "true"
  }

  with {
    name  = "cache"
    value = "true"
  }
}

step "tag_version" {
  id   = "tag_version"
  name = "Bump version and push tag"

  // mathieudutour/github-tag-action v6.2
  uses {
    action  = "mathieudutour/github-tag-action"
    version = "a22cf08638b34d5badda920f9daf6e72c477b07b"
  }

  with {
    name  = "github_token"
    value = "$${{ secrets.GITHUB_TOKEN }}"
  }
}

step "create_release" {
  id   = "create_release"
  name = "Create a GitHub release"

  // ncipollo/release-action v1.20.0
  uses {
    action  = "ncipollo/release-action"
    version = "b7eabc95ff50cbeeedec83973935c8f306dfcd0b"
  }

  with {
    name  = "github_token"
    value = "$${{ secrets.GITHUB_TOKEN }}"
  }
}

step "tests" {
  name = "Tests"
  run  = <<EOF
mise run test-ci
short_sha=$(git rev-parse --short "$GITHUB_SHA")
echo "tag=dev-$short_sha" >> $GITHUB_OUTPUT
EOF
}

step "coverage" {
  name = "Coverage"
  if   = "$${{ matrix.os == 'ubuntu-24.04' }}"

  // gwatts/go-coverage-action v2.0.0
  uses {
    action  = "gwatts/go-coverage-action"
    version = "2845595538a59d63d1bf55f109c14e104c6f7cb3"
  }
}

step "goreleaser" {
  // goreleaser/goreleaser-action v7.0.0
  uses {
    action  = "goreleaser/goreleaser-action"
    version = "ec59f474b9834571250b370d4735c50f8e2d1e29"
  }

  with {
    name  = "distribution"
    value = "goreleaser"
  }

  with {
    name  = "version"
    value = "latest"
  }

  with {
    name  = "args"
    value = "$${{ github.event_name == 'workflow_dispatch' && 'release --clean --snapshot --skip=publish --skip=announce --skip=validate' || 'release --clean' }}"
  }

  env {
    name  = "GITHUB_TOKEN"
    value = "$${{ secrets.GITHUB_TOKEN }}"
  }

  env {
    name  = "HOMEBREW_TAP_GITHUB_TOKEN"
    value = "$${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}"
  }
}

step "release_observability" {
  name = "Release observability summary"
  if   = "$${{ always() }}"
  run  = <<EOF
set -e

summary_path="$GITHUB_STEP_SUMMARY"

if [ -z "$summary_path" ]; then
  echo "GITHUB_STEP_SUMMARY is not available"
  exit 0
fi

{
  echo "## Release artifact summary"
  echo
  echo "- Event: $GITHUB_EVENT_NAME"
  echo "- Ref: $GITHUB_REF"
  echo "- SHA: $GITHUB_SHA"
  echo
  echo "### Checksums"

  if [ -f "dist/checksums.txt" ]; then
    echo
    echo '```text'
    cat "dist/checksums.txt"
    echo '```'
  else
    echo
    echo "checksums file not found at dist/checksums.txt"
  fi

  echo
  echo "### Formula output"

  formula_found="false"
  for formula in dist/*.rb; do
    if [ -f "$formula" ]; then
      formula_found="true"
      echo
      echo "- Generated formula: $formula"
    fi
  done

  if [ "$formula_found" = "false" ]; then
    echo
    echo "no generated formula files found under dist/*.rb"
  fi
} >> "$summary_path"
EOF
}

step "changelog" {
  // orhun/git-cliff-action v4.7.1
  uses {
    action  = "orhun/git-cliff-action"
    version = "c93ef52f3d0ddcdcc9bd5447d98d458a11cd4f72"
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
