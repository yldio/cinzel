---
title: Add --yml flag to control parse output file extension (.yml vs .yaml)
problem_type: feature
component: provider/github, internal/command
symptoms:
  - Users needed .yml output files instead of the default .yaml for GitHub Actions workflows
  - No way to configure output extension via CLI or .cinzelrc.yaml
tags:
  - cli
  - parse
  - github-actions
  - configuration
  - provider-ops
affected_files:
  - provider/provider.go
  - provider/github/io_helpers.go
  - provider/github/github.go
  - internal/command/config.go
  - internal/command/command.go
date: 2026-03-31
---

## Problem Statement

cinzel's `parse` command always emitted GitHub Actions workflow files with the `.yaml` extension.
There was no way for users to get `.yml` output, which some projects require for consistency with
existing conventions.

---

## Solution

Added a `--yml` boolean flag to the `parse` command and a `yml: true` key to `.cinzelrc.yaml`.
Only GitHub workflow output filenames are affected; platform-mandated names (`action.yml`,
`.gitlab-ci.yml`) remain hardcoded.

### `provider/provider.go`

Added `YML bool` to `ProviderOps`:

```go
type ProviderOps struct {
    // ...existing fields...
    YML bool // use .yml extension instead of .yaml
}
```

### `provider/github/io_helpers.go`

Added `workflowExt` helper; updated `resolveParseFilename` to use it:

```go
func workflowExt(opts provider.ProviderOps) string {
    if opts.YML {
        return ".yml"
    }
    return ".yaml"
}
```

`resolveParseFilename` was updated to call `workflowExt(opts)` everywhere it previously
hardcoded `".yaml"`.

### `provider/github/github.go`

Workflow output path changed from hardcoded `.yaml` to `workflowExt(opts)`:

```go
// Before
outputPath := filepath.Join(outputDir, workflowFile.Filename+".yaml")

// After
outputPath := filepath.Join(outputDir, workflowFile.Filename+workflowExt(opts))
```

### `internal/command/config.go`

Added `yml`/`hasYML` to `providerCommandConfig`, parsed `yml: true/false` from `.cinzelrc.yaml`,
and wired the config fallback after the CLI flag:

```go
type providerCommandConfig struct {
    // ...existing fields...
    yml    bool
    hasYML bool
}
```

Config parsing:
```go
case "yml":
    if valueNode.Kind != yaml.ScalarNode || valueNode.Tag != "!!bool" {
        return providerCommandConfig{}, nil, fmt.Errorf("...yml must be a boolean")
    }
    config.yml = valueNode.Value == "true"
    config.hasYML = true
```

CLI-wins precedence (CLI flag checked first via `cmd.IsSet`):
```go
opts := provider.ProviderOps{
    // ...
    YML: cmd.Bool("yml"),
}
// config fallback only applies when the flag was not explicitly passed
if !cmd.IsSet("yml") && conf.hasYML {
    opts.YML = conf.yml
}
```

### `internal/command/command.go`

Added `--yml` BoolFlag to the `parse` command only:

```go
&cli.BoolFlag{
    Name:  "yml",
    Value: false,
    Usage: "Generate .yml files instead of .yaml",
},
```

---

## Design Decisions

| Decision | Reason |
|----------|--------|
| Boolean flag `--yml`, not `--yaml-ext yml` | Simpler ergonomics; only two states needed |
| Config key `yml: true`, not `yaml-ext: yml` | Consistent with the flag name |
| Parse command only | unparse already reads both `.yaml` and `.yml` via `ListFilesWithExtensions` |
| `action.yml` stays hardcoded | GitHub spec mandates this exact filename |
| `.gitlab-ci.yml` stays hardcoded | GitLab spec mandates this exact filename |

---

## Usage

**CLI:**
```bash
cinzel github parse --yml -f cinzel/workflows.hcl
```

**`.cinzelrc.yaml`:**
```yaml
github:
  parse:
    yml: true
```

---

## Prevention Strategies

### What NOT to make configurable

Platform-mandated filenames must never be subject to user configuration:
- `action.yml` — GitHub Actions requires this exact name
- `.gitlab-ci.yml` — GitLab CI requires this exact filename at the repo root

If a future option would change a platform-mandated filename, reject it at design time.

### CLI flag vs config precedence

cinzel uses `cmd.IsSet("flag-name")` (from `urfave/cli`) to distinguish "user passed flag" from
"flag defaulted to zero value". This allows correct three-way precedence:

```
CLI flag (if set) > .cinzelrc.yaml > hardcoded default
```

Always use `cmd.IsSet` when deciding whether the config fallback should apply to a boolean flag —
a plain `cmd.Bool("x") == false` cannot distinguish "not passed" from "passed as false".

---

## New Option Checklist

When adding a boolean option to `ProviderOps` + CLI + config, touch these five places in order:

1. **`provider/provider.go`** — add field to `ProviderOps`, add doc comment
2. **`internal/command/command.go`** — add `BoolFlag` to the relevant command(s)
3. **`internal/command/config.go`** — add `field`/`hasField` to `providerCommandConfig`; parse from YAML; wire into `ProviderOps` using `cmd.IsSet` guard
4. **`provider/<name>/<file>.go`** — consume `opts.Field` at the correct output boundary
5. **Tests** — CLI flag active, CLI flag absent, config-only, CLI-overrides-config

### Related docs

- [`logic-errors/config-input-precedence-ignored-for-parse.md`](../logic-errors/config-input-precedence-ignored-for-parse.md) — covers ProviderOps wiring and the `cmd.IsSet` precedence pattern
