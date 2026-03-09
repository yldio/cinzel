---
title: "Config input precedence ignored for parse when using .cinzelrc.yaml"
module: "CLI"
problem_type: "logic_error"
component: "command_config_resolution"
severity: "high"
root_cause: "logic_error"
symptoms:
  - "`github parse --dry-run` fails with '`file` or `directory` must be set' even when config defines parse.directory"
  - "Config output-directory is applied, but config file/directory input is ignored"
  - "Dry-run path tests pass with CLI --file but fail in real config-only usage"
tags:
  - "config"
  - "precedence"
  - "dry-run"
  - "cli"
  - "provider-opts"
created_date: "2026-03-09"
updated_date: "2026-03-09"
---

## Problem Description

After introducing `.cinzelrc.yaml` command-scoped config, `github parse --dry-run` still failed with:

`file` or `directory` must be set

This happened even when `.cinzelrc.yaml` contained:

```yaml
github:
  parse:
    directory: ./cinzel
```

## Root Cause

The CLI config resolver only mapped `output-directory` from config into `provider.ProviderOps` and did not map input source fields (`file`, `directory`).

Validation in provider parse path still requires one input source, so config-only usage failed despite valid config.

## Solution Implemented

Updated command config resolution to apply config input sources when CLI input flags are not explicitly set:

- Added `file` and `directory` support in `internal/command/config.go`
- Applied precedence by flag presence (`cmd.IsSet("file") || cmd.IsSet("directory")`)
- Added config validation to reject setting both `file` and `directory` in the same command scope

Key behavior:

- If CLI sets `--file` or `--directory`, CLI wins.
- If CLI does not set either, config `file`/`directory` is applied.
- `output-directory` precedence remains unchanged.

## Test Coverage Added

Added tests in `internal/command/config_test.go` for:

- Config `parse.directory` is applied
- CLI `--file` overrides config `directory`
- Config `file + directory` conflict fails fast
- Dry-run works with config-only `parse.directory`

## Prevention Guidance

- When adding config precedence, test both:
  - output-related options
  - required input options (`file`/`directory`)
- Add at least one real invocation test that mirrors user behavior (`parse --dry-run` without CLI input flags)
- In plan acceptance criteria, include a CLI contract section distinguishing:
  - required inputs
  - optional outputs
  - what can be sourced from config

## Related References

- `internal/command/config.go`
- `internal/command/config_test.go`
- `docs/plans/2026-03-09-feat-cinzelrc-provider-config-precedence-plan.md`
- `docs/solutions/patterns/critical-patterns.md`
