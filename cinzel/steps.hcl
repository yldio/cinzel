// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

step "checkout" {
  name = "Checkout"

  uses {
    action  = "actions/checkout"
    version = "3d3c42e5aac5ba805825da76410c181273ba90b1" # v7.0.1
  }

  with {
    name  = "persist-credentials"
    value = "false"
  }
}

step "checkout_release" {
  name = "Checkout (full history)"

  uses {
    action  = "actions/checkout"
    version = "3d3c42e5aac5ba805825da76410c181273ba90b1" # v7.0.1
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

step "verify_release_token" {
  name = "Verify the release token before anything changes"

  run = <<EOF
set -euo pipefail

scoped=$(gh api /installation/repositories --paginate --jq '.repositories[].full_name')

for repo in "$GITHUB_REPOSITORY" yldio/homebrew-cinzel; do
  if ! printf '%s\n' "$scoped" | grep -qxF "$repo"; then
    echo "the release token is not scoped to $repo"
    echo "it reaches: $scoped"
    exit 1
  fi
done

# Never writes, but the server still runs its permission check on
# git-receive-pack, so a read-only token fails here rather than at the tag.
git push --dry-run origin "HEAD:refs/heads/$GITHUB_REF_NAME"
EOF

  env {
    name  = "GH_TOKEN"
    value = "$${{ steps.release_app_token.outputs.token }}"
  }

  env {
    name  = "GIT_CONFIG_COUNT"
    value = "1"
  }

  env {
    name  = "GIT_CONFIG_KEY_0"
    value = "url.https://x-access-token:$${{ steps.release_app_token.outputs.token }}@github.com/.insteadOf"
  }

  env {
    name  = "GIT_CONFIG_VALUE_0"
    value = "https://github.com/"
  }
}

step "release_app_token" {
  id   = "release_app_token"
  name = "Create release app token"

  uses {
    action  = "actions/create-github-app-token"
    version = "bcd2ba49218906704ab6c1aa796996da409d3eb1" # v3.2.0
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

  with {
    name  = "permission-contents"
    value = "write"
  }
}

step "mise_setup" {
  name = "Setup mise"

  uses {
    action  = "jdx/mise-action"
    version = "c2a87611a18de5b3828c5652fe268e992400cb5c" # v4.3.0
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

  uses {
    action  = "mathieudutour/github-tag-action"
    version = "af99e60ce8132224b8e6ebab5023449fe256ed46" # v7
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

  uses {
    action  = "orhun/git-cliff-action"
    version = "a9a95522b26fe6403f7bb24031f21fb573d0f5ff" # v4.9.1
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

  uses {
    action  = "orhun/git-cliff-action"
    version = "a9a95522b26fe6403f7bb24031f21fb573d0f5ff" # v4.9.1
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

  uses {
    action  = "stefanzweifel/git-auto-commit-action"
    version = "4a55954c782fc1ea30b9056cd3e7a2b40ca8887d" # v7.2.0
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

  env {
    name  = "GIT_CONFIG_COUNT"
    value = "1"
  }

  env {
    name  = "GIT_CONFIG_KEY_0"
    value = "url.https://x-access-token:$${{ steps.release_app_token.outputs.token }}@github.com/.insteadOf"
  }

  env {
    name  = "GIT_CONFIG_VALUE_0"
    value = "https://github.com/"
  }
}

step "tests" {
  name = "Tests"
  run  = "mise run test-ci"
}

step "lint" {
  name = "Lint"
  if   = "$${{ matrix.os == 'ubuntu-24.04' }}"
  run  = "mise run lint"
}

step "drift" {
  name = "Generated workflows match their HCL"
  if   = "$${{ matrix.os == 'ubuntu-24.04' }}"
  run  = "mise run drift"
}

step "coverage" {
  name = "Coverage"
  if   = "$${{ matrix.os == 'ubuntu-24.04' }}"

  uses {
    action  = "gwatts/go-coverage-action"
    version = "2845595538a59d63d1bf55f109c14e104c6f7cb3"
  }

  env {
    name  = "GIT_CONFIG_COUNT"
    value = "1"
  }

  env {
    name  = "GIT_CONFIG_KEY_0"
    value = "url.https://x-access-token:$${{ github.token }}@github.com/.insteadOf"
  }

  env {
    name  = "GIT_CONFIG_VALUE_0"
    value = "https://github.com/"
  }
}

step "calculate_next_version" {
  id   = "calculate_next_version"
  name = "Calculate next version"

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
  uses {
    action  = "goreleaser/goreleaser-action"
    version = "f06c13b6b1a9625abc9e6e439d9c05a8f2190e94" # v7.2.3
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
