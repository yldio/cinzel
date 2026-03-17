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

  with {
    name  = "github_token"
    value = "$${{ steps.release_app_token.outputs.token }}"
  }
}

step "release_app_token" {
  id   = "release_app_token"
  name = "Create release app token"

  // actions/create-github-app-token v3.0.0
  uses {
    action  = "actions/create-github-app-token"
    version = "f8d387b68d61c58ab83c6c016672934102569859"
  }

  with {
    name  = "app-id"
    value = "$${{ secrets.RELEASE_APP_ID }}"
  }

  with {
    name  = "private-key"
    value = "$${{ secrets.RELEASE_PRIVATE_KEY }}"
  }

  with {
    name  = "repositories"
    value = "cinzel,homebrew-cinzel"
  }
}

step "mise_setup" {
  name = "Setup mise"

  // jdx/mise-action v4.0.0
  uses {
    action  = "jdx/mise-action"
    version = "c1ecc8f748cd28cdeabf76dab3cccde4ce692fe4"
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
  if   = "$${{ steps.resolve_release_tag.outputs.skip != 'true' }}"

  // mathieudutour/github-tag-action v6.2
  uses {
    action  = "mathieudutour/github-tag-action"
    version = "a22cf08638b34d5badda920f9daf6e72c477b07b"
  }

  with {
    name  = "github_token"
    value = "$${{ steps.release_app_token.outputs.token }}"
  }

  with {
    name  = "custom_tag"
    value = "$${{ steps.resolve_release_tag.outputs.tag }}"
  }
}

step "resolve_release_tag" {
  id   = "resolve_release_tag"
  name = "Resolve release tag"
  run  = <<EOF
set -euo pipefail

input_tag="$${{ github.event.inputs.tag }}"
input_tag="$${input_tag#v}"
calculated_tag="$${{ steps.calculate_next_version.outputs.nextStrict }}"
bump="$${{ steps.calculate_next_version.outputs.bump }}"

if [ -n "$input_tag" ]; then
  echo "using manual tag: $input_tag"
  echo "tag=$input_tag" >> "$GITHUB_OUTPUT"
elif [ "$bump" != "none" ] && [ -n "$calculated_tag" ]; then
  echo "using calculated tag: $calculated_tag (bump: $bump)"
  echo "tag=$calculated_tag" >> "$GITHUB_OUTPUT"
else
  echo "no releasable commits and no manual tag provided"
  echo "skip=true" >> "$GITHUB_OUTPUT"
fi
EOF
}

step "ensure_release_app" {
  name = "Ensure release app credentials"
  run  = <<EOF
set -euo pipefail

if [ -z "$RELEASE_APP_ID" ] || [ -z "$RELEASE_PRIVATE_KEY" ]; then
  echo "RELEASE_APP_ID and RELEASE_PRIVATE_KEY secrets are required for release automation"
  exit 1
fi
EOF

  env {
    name  = "RELEASE_APP_ID"
    value = "$${{ secrets.RELEASE_APP_ID }}"
  }

  env {
    name  = "RELEASE_PRIVATE_KEY"
    value = "$${{ secrets.RELEASE_PRIVATE_KEY }}"
  }
}

step "create_release" {
  id   = "create_release"
  name = "Create a GitHub release"

  // ncipollo/release-action v1.21.0
  uses {
    action  = "ncipollo/release-action"
    version = "339a81892b84b4eeb0f6e744e4574d79d0d9b8dd"
  }

  with {
    name  = "token"
    value = "$${{ steps.release_app_token.outputs.token }}"
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
  id   = "git_cliff_changelog"
  name = "Generate full changelog"

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
    value = "$${{ steps.release_app_token.outputs.token }}"
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

step "git_cliff_release_notes" {
  id   = "git_cliff"
  name = "Generate release notes"

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
    value = "--offline --verbose --unreleased --tag $${{ steps.tag_version.outputs.new_tag }}"
  }

  with {
    name  = "github_token"
    value = "$${{ steps.release_app_token.outputs.token }}"
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
    value = "chore(release): prepare for $${{ steps.tag_version.outputs.new_tag }} [skip ci]"
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

step "calculate_next_version" {
  id   = "calculate_next_version"
  name = "Calculate next version"

  // ietf-tools/semver-action v1.11.0
  uses {
    action  = "ietf-tools/semver-action"
    version = "c90370b2958652d71c06a3484129a4d423a6d8a8"
  }

  with {
    name  = "token"
    value = "$${{ github.token }}"
  }

  with {
    name  = "branch"
    value = "main"
  }

  with {
    name  = "patchList"
    value = "fix, perf, refactor"
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
    value = "release --clean"
  }

  env {
    name  = "GITHUB_TOKEN"
    value = "$${{ steps.release_app_token.outputs.token }}"
  }

  env {
    name  = "HOMEBREW_TAP_GITHUB_TOKEN"
    value = "$${{ steps.release_app_token.outputs.token }}"
  }
}
