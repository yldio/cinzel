# cinzel

<img src="./assets/logo.png" alt="cinzel" width="500px"/>

## Table of Contents

- [cinzel](#cinzel)
  - [Table of Contents](#table-of-contents)
  - [About](#about)
  - [Installation and usage](#installation-and-usage)
- [Providers](#providers)
  - [GitHub Actions](#github-actions)
  - [GitLab CI/CD Pipelines](#gitlab-cicd-pipelines)
  - [Changelog](#changelog)
  - [Code of Conduct](#code-of-conduct)
  - [Contributing](#contributing)
  - [License](#license)

## About

**`cinzel`**, pronounced as "*sin-ZEL*" ([IPA](https://en.wikipedia.org/wiki/International_Phonetic_Alphabet): /sĩˈzɛl/), is the Portuguese word for **chisel**.

It's a bidirectional converter between [HCL](https://github.com/hashicorp/hcl) and CI/CD pipeline [YAML](http://www.yaml.de), with provider-specific mappings (currently GitHub Actions and GitLab CI/CD).

Made with :heart: by [YLD Limited](https://www.yld.com/).

## Installation and usage

Install `cinzel` using one of these options:

- Download a prebuilt binary from [GitHub Releases][releases] (recommended for most users).
- Install with Homebrew:

```sh
brew tap yldio/cinzel
brew install --cask cinzel
```

- Install from source with Go:

```sh
go install github.com/yldio/cinzel@latest
```

Confirm installation:

```sh
cinzel --help
```

<!-- For more options on how to install, please go over to the [Wiki](https://github.com/yldio/cinzel/wiki). -->

### Quick start

Use the provider command shape:

```sh
cinzel <provider> parse --file <input.hcl> --output-directory <out-dir>
cinzel <provider> unparse --file <input.yaml> --output-directory <out-dir>
```

Example: GitHub Actions parse/unparse:

```sh
cinzel github parse --file ./test.hcl --output-directory .github/workflows
cinzel github unparse --file ./.github/workflows/test.yaml --output-directory ./cinzel
```

Example: GitLab CI/CD parse/unparse:

```sh
cinzel gitlab parse --file ./pipeline.hcl --output-directory .
cinzel gitlab unparse --file ./.gitlab-ci.yml --output-directory ./cinzel
```

Use `--dry-run` to print generated content to stdout.

### AI-assisted generation

Generate HCL workflow definitions from a natural language prompt:

```sh
cinzel github assist --prompt "golang PR with tests and linting"
```

This calls an LLM (Anthropic by default), generates valid YAML, converts it to HCL via the unparse pipeline, and writes to a timestamped session folder under `./cinzel/assist/`. For GitHub, action versions are automatically pinned to SHAs. Blocks that match your existing HCL are replaced with `// reuses:` comments.

Each prompt creates its own session:

```
cinzel/assist/
  20260317-150405/     # first prompt
    assist.hcl
  20260317-151200/     # second prompt
    assist.hcl
```

Requires an API key:

```sh
export ANTHROPIC_API_KEY=sk-ant-...
# or
export OPENAI_API_KEY=sk-...
cinzel github assist --ai openai --prompt "..."
```

Refine previous output (targets the latest session by default):

```sh
cinzel github assist --refine "add slack notification on failure" --prompt "add to PR workflow"
```

Refine a specific session:

```sh
cinzel github assist --refine "add caching" --from 20260317-150405
```

### Version management (GitHub Actions)

Pin action tags to commit SHAs:

```sh
cinzel github pin                     # pin all actions in ./cinzel/
cinzel github pin --dry-run           # preview without writing
```

Upgrade actions to their latest versions:

```sh
cinzel github upgrade                 # bump to latest + pin SHAs
cinzel github upgrade --dry-run       # preview changes
cinzel github upgrade --parse         # bump + regenerate YAML
```

No GitHub token is required for public actions. Set `GITHUB_TOKEN` for higher rate limits (5000/hr vs 60/hr).

For release operator details about Homebrew automation, see [`docs/release/homebrew.md`](docs/release/homebrew.md).

## Providers

Providers are the CI/CD platforms that `cinzel` can convert between HCL and YAML.

### GitHub Actions

See [`provider/github/README.md`](provider/github/README.md) for the full HCL schema reference and feature coverage.

### GitLab CI/CD Pipelines

See [`provider/gitlab/README.md`](provider/gitlab/README.md) for the GitLab HCL schema and conversion coverage.

## Changelog

Please visit the [Changelog](CHANGELOG.md) for more details.

## Code of Conduct

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](code_of_conduct.md)

Please check our [Code of Conduct](./CODE_OF_CONDUCT.md).

## Contributing

Contributions are welcome, as well as suggestions for `cinzel`. Please go over to the [Discussions](https://github.com/yldio/cinzel/discussions) first to understand the current state, features and issues before creating any issue or pull request. :heart:

Please make sure to update tests as appropriate.

## License

This project is licensed under the Apache-2.0 license. See [LICENSE](./LICENSE) for details.
