---
status: complete
priority: p1
issue_id: "025"
tags: [code-review, correctness, output, github, gitlab]
dependencies: []
---

# Rewriting encoded bytes turns a literal backslash-u into broken YAML

## Problem Statement

`unescapeUnicode` in `internal/yamldoc/encode.go` and `unescapeYAMLUnicode` in
`provider/gitlab/pipeline_yaml.go` run a regex over the encoder's output bytes
and replace `\uXXXX` / `\UXXXXXXXX` with the raw rune. The regex cannot tell an
escape the encoder produced from one the author wrote inside a string, so a
script that contains the four characters `é` is rewritten into something
the YAML parser refuses.

`encode.go` also states in its own doc comment that nothing in the output is
recovered by rewriting encoded bytes, which is what this does.

## Findings

```hcl
step "writejson" {
  run = "echo '{\"msg\": \"caf\\u00e9\"}' > out.json"
}
```

emits

```yaml
run: "echo '{\"msg\": \"caf\é\"}' > out.json"
```

Reparsing that file fails: `yaml: found unknown escape character`. Reproduced
through the CLI in both providers.

## Recommended Action

Not the deletion the report suggested. The yaml.v3 encoder already sets its
unicode flag, and the escapes come from its `is_printable` helper, which only
recognises 3-byte UTF-8; hclwrite escapes category-Cf runes for its own
reasons. Deleting the post-processors brings back the emoji corruption
`TestEmojiRoundtripStability` covers.

The real defect is backslash parity. A writer emits a literal backslash
doubled, so an authored `\u00e9` reaches the output as `\\u00e9` and the
regex matches its second half. Counting the backslashes before the escape and
rewriting only an odd-length run leaves authored text alone.

## Technical Details

- New `internal/unescape`, `Unicode`: one parity-aware implementation
- Replaced four byte-identical copies of the same bug:
  - `internal/yamldoc/encode.go`, `unescapeUnicode`
  - `provider/gitlab/pipeline_yaml.go`, `unescapeYAMLUnicode`
  - `provider/github/unparse_workflow.go`, `unescapeHCLUnicode`
  - `provider/gitlab/unparse_pipeline.go`, `unescapeHCLUnicode`
- The regex takes the leading run of backslashes into the match so the parity
  can be read; without it the second backslash of `\\uXXXX` starts a match

## Acceptance Criteria

- [x] A `run` containing a literal `\u00e9` roundtrips unchanged, both directions
- [x] Non-ASCII in a string still comes out as raw UTF-8, not an escape
- [x] Golden and roundtrip tests pass for both providers

## Work Log

- 2026-09-16: Reproduced through the CLI. Found the cause to be parity, not the
  rewrite itself, so the suggested deletion would have regressed the emoji
  roundtrip. Unified the four copies behind `internal/unescape` and fixed the
  parity there. Full suite green.
