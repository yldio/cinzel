---
title: "Golden test failures from YAML single quote to double quote conversion"
module: "GitHubProvider"
problem_type: "test_failure"
component: "yaml_output"
severity: "high"
root_cause: "config_error"
symptoms:
  - "Golden tests fail after opening fixture files in editor"
  - "Diff shows single quotes changed to double quotes"
  - "Tests pass in CI but fail locally after editing fixtures"
tags:
  - "yaml"
  - "quotes"
  - "golden-tests"
  - "zed-editor"
  - "ide"
created_date: "2026-03-08"
updated_date: "2026-03-08"
---

## Problem Description

Golden YAML fixture tests started failing after fixture files were opened and saved in the Zed editor. The diffs showed all single-quoted strings had been converted to double-quoted strings.

## Root Cause

The Zed editor normalizes YAML string quoting on save, converting `'value'` to `"value"`. The YAML marshaller was using `SingleQuotedStyle`, producing output that the editor would silently rewrite.

## Solution Implemented

Changed `workflow_yaml.go` to use `DoubleQuotedStyle` exclusively when `stringNeedsQuoting()` returns true:

```go
node := &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: v}
if stringNeedsQuoting(v) {
    node.Style = yamlv3.DoubleQuotedStyle
}
```

Updated all golden fixture files to use double quotes. This aligns production output with what the editor produces.

## Prevention Guidance

- Always use `DoubleQuotedStyle` for quoted YAML strings in this project.
- The `.editorconfig` file defines project formatting conventions.
- If adding a new editor to the workflow, verify its YAML save behavior against the golden fixtures.
- Golden tests use semantic comparison (`assertYAMLSemanticEqual`), not byte comparison, which provides some resilience — but the quote style still matters for the raw fixture files checked into git.
