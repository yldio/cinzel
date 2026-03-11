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

step "checkout_release_with_credentials" {
  name = "Checkout (full history, push enabled)"

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
    value = "true"
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

  with {
    name  = "tag"
    value = "$${{ steps.tag_version.outputs.new_tag }}"
  }

  with {
    name  = "body"
    value = "$${{ steps.git_cliff.outputs.content }}"
  }
}

step "git_cliff_changelog" {
  id   = "git_cliff"
  name = "Generate changelog"

  // orhun/git-cliff-action v4.7.1
  uses {
    action  = "orhun/git-cliff-action"
    version = "c93ef52f3d0ddcdcc9bd5447d98d458a11cd4f72"
  }

  with {
    name  = "config"
    value = "cliff.toml"
  }

  with {
    name  = "args"
    value = "--offline --verbose --tag $${{ steps.tag_version.outputs.new_tag }}"
  }

  with {
    name  = "github_token"
    value = "$${{ secrets.GITHUB_TOKEN }}"
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

step "commit_release" {
  name = "Commit release changelog"

  // stefanzweifel/git-auto-commit-action v7.1.0
  uses {
    action  = "stefanzweifel/git-auto-commit-action"
    version = "04702edda442b2e678b25b537cec683a1493fcb9"
  }

  with {
    name  = "commit_message"
    value = "chore(release): prepare for $${{ steps.tag_version.outputs.new_tag }}"
  }

  with {
    name  = "file_pattern"
    value = "CHANGELOG.md"
  }

  with {
    name  = "branch"
    value = "$${{ github.ref_name }}"
  }
}

step "tests" {
  name = "Tests"
  run  = "mise run test-ci"
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
    value = "v2.14.3"
  }

  with {
    name  = "args"
    value = "release --clean --release-notes ./release-notes.md"
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

step "release_notes" {
  name = "Generate release notes"
  run  = "git cliff --offline --current --strip header --output ./release-notes.md"
}
