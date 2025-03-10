---
title: "YAML string quoting rules — when to quote and what style to use"
module: "GitHubProvider"
problem_type: "developer_experience"
component: "yaml_output"
severity: "medium"
root_cause: "config_error"
symptoms:
  - "actions/checkout@v4 appears as quoted string in output"
  - "Boolean-like strings not quoted causing YAML parse errors"
  - "Numbers interpreted as integers instead of strings"
tags:
  - "yaml"
  - "quoting"
  - "stringNeedsQuoting"
  - "workflow_yaml"
created_date: "2026-03-08"
updated_date: "2026-03-08"
---

## Problem Description

YAML output needs careful quoting decisions. Unquoted `true` becomes a boolean, unquoted `20` becomes an integer, but quoting `actions/checkout@v4` is unnecessary and looks noisy.

## Root Cause

YAML's type inference is context-dependent. The `stringNeedsQuoting()` function in `workflow_yaml.go` must identify all values that would be misinterpreted without quotes, while avoiding quoting values that are unambiguously strings.

## Solution Implemented

`stringNeedsQuoting()` checks for:

1. **Empty strings** — must be quoted (`""`)
2. **Boolean-like values** — `true`, `false`, `yes`, `no`, `on`, `off` (case-insensitive)
3. **Null-like values** — `null`, `~`
4. **Numeric-looking strings** — anything that `strconv.ParseFloat` accepts
5. **YAML special characters** — `:`, `#`, `[`, `]`, `{`, `}`, `,`, `&`, `*`, `?`, `|`, `-`, `<`, `>`, `=`, `!`, `%`

**Notably excluded**: The `@` character does NOT trigger quoting. This keeps `actions/checkout@v4` clean.

When quoting is needed, `DoubleQuotedStyle` is always used.

## Prevention Guidance

- If adding new quoting triggers, run the full golden test suite — false positives will show up as unnecessary quotes in output.
- The `@` exclusion was specifically chosen for GitHub Actions `uses` references. If a future provider needs `@` quoted, handle it at the provider level, not in the shared quoting function.
- Test with `actions/setup-node@v4`, `"20"` (numeric string), `true` (boolean), and `ubuntu-latest` (contains `-` but doesn't need quoting since `-` is only special at line start in YAML block context).
