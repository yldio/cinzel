# cinzel

<img src="./assets/logo.png" alt="cinzel" width="500px"/>

## Table of Contents

- [About](#about)
- [Installation and usage](#installation-and-usage)
  - [Quick start](#quick-start)
  - [Configuration file](#configuration-file)
  - [AI-assisted generation](#ai-assisted-generation)
  - [Version management (GitHub Actions)](#version-management-github-actions)
- [Providers](#providers)
  - [GitHub Actions](#github-actions)
  - [GitLab CI/CD Pipelines](#gitlab-cicd-pipelines)
- [Changelog](#changelog)
- [Code of Conduct](#code-of-conduct)
- [Contributing](#contributing)
- [License](#license)

## About

**`cinzel`**, pronounced as "*sin-ZEL*" ([IPA](https://en.wikipedia.org/wiki/International_Phonetic_Alphabet): /sĩˈzɛl/), is the Portuguese word for **chisel**.

It's a bidirectional converter between [HCL](https://github.com/hashicorp/hcl) and CI/CD pipeline [YAML](https://yaml.org), with provider-specific mappings (currently GitHub Actions and GitLab CI/CD).

Made with :heart: by [YLD Limited](https://www.yld.com/).

## Installation and usage

Install `cinzel` using one of these options:

- Download a prebuilt binary from [GitHub Releases](https://github.com/yldio/cinzel/releases) (recommended for most users).
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

Read a whole directory instead of one file, and walk into its subdirectories:

```sh
cinzel github parse --directory ./cinzel --recursive
```

Flags shared by `parse` and `unparse`:

| Flag | Description |
| --- | --- |
| `--file`, `-f` | Read one file. |
| `--directory`, `-d` | Read every matching file in a directory. Mutually exclusive with `--file`. |
| `--recursive`, `-r` | Walk subdirectories of `--directory`. |
| `--output-directory` | Where generated files are written. Parse defaults to `.github/workflows` (GitHub) or the working directory (GitLab); unparse defaults to `./cinzel`. |
| `--dry-run` | Print the generated content to stdout instead of writing files. |

`parse` also takes `--yml`, which writes `.yml` files instead of `.yaml`. GitLab
parse always writes `.gitlab-ci.yml`, so the flag only affects GitHub.

### Configuration file

A `.cinzelrc.yaml` in the working directory supplies defaults, which any flag
on the command line overrides:

```yaml
github:
  parse:
    directory: ./cinzel
    output-directory: .github/workflows
    yml: false
```

It applies to `parse` and `unparse` only, per provider. `assist`, `pin`,
`upgrade` and `init` ignore it. The keys under a command are `file`,
`directory`, `output-directory` and, for `parse`, `yml`; anything else is
reported as `warning: ...: unknown key` and skipped, `recursive` included.

Paths in it must be relative and written with forward slashes. The file is
meant to be committed, so it is read on every machine that checks the repo out:
an absolute path names one machine's disk and a leading `~` names one user, and
both are refused with an error naming the key. Forward slashes are converted to
the separator the running system uses, so one spelling works on Linux, macOS
and Windows alike.

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

Requires an API key, read from the environment:

```sh
export ANTHROPIC_API_KEY=sk-ant-...
# or
export OPENAI_API_KEY=sk-...
cinzel github assist --ai openai --prompt "..."
```

`cinzel init` writes a config file holding the default provider and the model
to use for each. It does not store a key: the config is written to disk and
swept up by a backup of your home directory, so the environment is the place
for one.

```sh
cinzel init
```

Refine previous output (targets the latest session by default):

```sh
cinzel github assist --refine "add slack notification on failure" --prompt "add to PR workflow"
```

Refine a specific session:

```sh
cinzel github assist --refine "add caching" --from 20260317-150405
```

Other `assist` flags:

| Flag | Description |
| --- | --- |
| `--output-directory` | Where session folders are created (default `cinzel/assist`). |
| `--dry-run` | Print to stdout instead of writing files. |
| `--acknowledge` | Skip the cost confirmation prompt. |
| `--ai` | `anthropic` or `openai`. |
| `--model` | Model override. |
| `--no-context` | Do not send existing HCL as context. |
| `--context-dir` | Directory to read existing HCL from (default `cinzel`). |

`assist` is available for both providers.

### Version management (GitHub Actions)

`pin` and `upgrade` are GitHub-only; there is no GitLab equivalent.

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

Both take `--file`/`-f` for a single file and `--directory`/`-d` for a
directory (default `cinzel`). `upgrade --parse` writes the regenerated YAML to
`--output-directory` (default `.github/workflows`).

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

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](./CODE_OF_CONDUCT.md)

Please check our [Code of Conduct](./CODE_OF_CONDUCT.md).

## Contributing

Contributions are welcome, as well as suggestions for `cinzel`. Please go over to the [Discussions](https://github.com/yldio/cinzel/discussions) first to understand the current state, features and issues before creating any issue or pull request. :heart:

Please make sure to update tests as appropriate.

## License

This project is licensed under the Apache-2.0 license. See [LICENSE](./LICENSE) for details.
