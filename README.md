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

It's a bidirectional converter between [HCL](https://github.com/hashicorp/hcl) and [GitHub Actions](https://github.com/features/actions) [YAML](http://www.yaml.de) — supporting workflows, jobs, steps, and composite actions.

Made with :heart: by [Yld](https://www.yld.com/).

## Installation and usage

You can install `cinzel` either by:

- downloading [the released binary][releases];

- by *Homebrew*;

```sh
brew install cinzel
```

- or by `go install`.

```sh
go install github.com/yldio/cinzel@latest
```

<!-- For more options on how to install, please go over to the [Wiki](https://github.com/yldio/cinzel/wiki). -->

### Quick start

Transform HCL `workflow`/`job`/`step` blocks into GitHub workflow YAML:

```sh
cinzel github parse --file ./test.hcl --output-directory .github/workflows
```

Transform GitHub workflow YAML back into HCL:

```sh
cinzel github unparse --file ./.github/workflows/test.yaml --output-directory ./cinzel
```

Use `--dry-run` to print generated content to stdout.

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

This project is licensed under the AGPL-3.0-or-later license. See [LICENSE](./LICENSE) for details.
