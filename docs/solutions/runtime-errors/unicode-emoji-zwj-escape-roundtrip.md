---
title: Unicode Emoji ZWJ Sequence Escaping in YAML↔HCL Roundtrip
problem_type: logic-errors
component: provider/github, provider/gitlab
symptoms:
  - "YAML `name: 👮‍♂️ Lint` → HCL `name = \"👮\\u200d♂️ Lint\"` (ZWJ escaped)"
  - "HCL → YAML `name: \"\\U0001F46E‍♂️ Lint\"` (emoji escaped as \\U)"
  - "Multi-codepoint emoji sequences corrupted after YAML→HCL→YAML roundtrip"
tags:
  - unicode
  - emoji
  - yaml
  - hcl
  - roundtrip
  - golang
affected_files:
  - provider/github/unparse_workflow.go
  - provider/github/unparse_action.go
  - provider/github/github.go
  - provider/github/workflow_yaml.go
  - provider/gitlab/unparse_pipeline.go
  - provider/gitlab/pipeline_yaml.go
  - provider/github/roundtrip_test.go
date: 2026-03-31
---

## Problem Statement

Emoji sequences using Unicode Zero Width Joiner (U+200D, a ZWJ — Zero Width Joiner) corrupted
during YAML→HCL→YAML roundtrip. A workflow named `👮‍♂️ Lint` would, after one roundtrip, become
`\U0001F46E‍♂️ Lint` in YAML output.

**Example:**
```yaml
# Input YAML
name: 👮‍♂️ Lint
```
```hcl
# After YAML→HCL (unparse)
name = "👮\u200d♂️ Lint"  # ZWJ escaped by hclwrite
```
```yaml
# After HCL→YAML (parse) — CORRUPTED
name: "\U0001F46E‍♂️ Lint"  # Supplementary-plane emoji escaped by yaml.v3
```

---

## Root Cause Analysis

Two separate library bugs compounded in the roundtrip pipeline:

### Bug 1: `hclwrite` escapes ZWJ (U+200D)

**Location:** `github.com/hashicorp/hcl/v2@v2.24.0/hclwrite/generate.go:350`

`escapeQuotedStringLit` uses Go's `unicode.IsPrint(r)` to decide whether to escape a rune.
Per Go's `unicode` package, category Cf (Format) characters are not printable. U+200D (ZERO WIDTH
JOINER) is category Cf, so it is escaped as `\u200d` in HCL string literals.

When HCL is parsed back, `\u200d` restores correctly. But the next YAML serialisation step then
encounters the ZWJ next to a supplementary-plane emoji...

### Bug 2: `yaml.v3` escapes supplementary-plane emoji

**Location:** `gopkg.in/yaml.v3@v3.0.1/yamlprivateh.go:85` (`is_printable`)

This function was ported from C's libyaml. Its UTF-8 byte range check only handles characters up
to 3-byte sequences (≤ U+FFFF):

```go
// Only covers bytes 0xC2–0xED (≤ 3-byte UTF-8 sequences)
b[i] > 0xC2 && b[i] < 0xED
```

Supplementary-plane emoji like 👮 (U+1F46E) require 4-byte UTF-8 sequences starting at 0xF0–0xF4.
Since the check misses them, they are treated as non-printable, triggering double-quoted mode and
`\U0001F46E` escaping in `yaml_emitter_write_double_quoted_scalar`.

### Combined Effect

```
👮‍♂️ (U+1F46E U+200D U+2642 U+FE0F)
  │
  ▼ yaml.v3 unmarshal → string (correct)
  │
  ▼ hclwrite.SetAttributeValue (unparse step)
    → escapeQuotedStringLit → `👮\u200d♂️`  ← ZWJ escaped
  │
  ▼ hclwrite.Format → `name = "👮\u200d♂️ Lint"`
  │
  ▼ hcl parse back → `👮\u200d♂️` (restored, ZWJ present again)
  │
  ▼ yaml.v3 Marshal → encounters U+1F46E (4-byte), treats as non-printable
    → double-quoted mode → `\U0001F46E\u200d\u2642\uFE0F Lint`  ← CORRUPT
```

---

## Investigation Steps

1. **Reproduced** with minimal YAML: `name: 👮‍♂️ Lint` → unparse → parse
2. **Traced HCL layer**: Found `hclwrite/generate.go:350` uses `unicode.IsPrint`; category Cf =
   non-printable → ZWJ escaped
3. **Traced YAML layer**: Found `yaml.v3/yamlprivateh.go:85` `is_printable` only handles 3-byte
   UTF-8; 4-byte sequences fall through to `\U` escaping
4. **Confirmed** both bugs are in external (vendored) libraries — not modifiable
5. **Chose post-processing** at output boundaries rather than patching external libraries

---

## Solution

Post-process all HCL and YAML output at the respective generation boundary, replacing `\uXXXX` /
`\UXXXXXXXX` escape sequences with their raw UTF-8 equivalents for codepoints ≥ U+00A0 (YAML 1.2
printable range). ASCII (U+0000–U+007F) and C1 controls (U+0080–U+009F) remain escaped.

### HCL Output Post-Processor

Added to `provider/github/unparse_workflow.go` (used by all HCL generators via `unescapeHCLUnicode`):

```go
func unescapeHCLUnicode(src []byte) []byte {
    return reHCLUnicodeEscape.ReplaceAllFunc(src, func(match []byte) []byte {
        n, err := strconv.ParseInt(string(match[2:]), 16, 32)
        if err != nil || n <= 0x9F || !utf8.ValidRune(rune(n)) {
            return match
        }
        var buf [utf8.UTFMax]byte
        l := utf8.EncodeRune(buf[:], rune(n))
        return append([]byte(nil), buf[:l]...)
    })
}

var reHCLUnicodeEscape = regexp.MustCompile(`\\u[0-9a-fA-F]{4}|\\U[0-9a-fA-F]{8}`)
```

Applied in every HCL generation return:
```go
return unescapeHCLUnicode(hclwrite.Format(f.Bytes())), nil
```

### YAML Output Post-Processor

Added to `provider/github/workflow_yaml.go` and `provider/gitlab/pipeline_yaml.go`:

```go
func unescapeYAMLUnicode(src []byte) []byte {
    return reYAMLUnicodeEscape.ReplaceAllFunc(src, func(match []byte) []byte {
        n, err := strconv.ParseInt(string(match[2:]), 16, 32)
        if err != nil || n <= 0x9F || !utf8.ValidRune(rune(n)) {
            return match
        }
        var buf [utf8.UTFMax]byte
        l := utf8.EncodeRune(buf[:], rune(n))
        return append([]byte(nil), buf[:l]...)
    })
}

var reYAMLUnicodeEscape = regexp.MustCompile(`\\U[0-9A-Fa-f]{8}|\\u[0-9A-Fa-f]{4}`)
```

Applied in `marshalWorkflowYAML`:
```go
// before return
return unescapeYAMLUnicode(out), nil
```

### Files Modified

| File | Change |
|------|--------|
| `provider/github/unparse_workflow.go` | Added `unescapeHCLUnicode` + applied to all HCL returns |
| `provider/github/unparse_action.go` | Applied `unescapeHCLUnicode` |
| `provider/github/github.go` | Applied `unescapeHCLUnicode` |
| `provider/github/workflow_yaml.go` | Added `unescapeYAMLUnicode` + applied to YAML output |
| `provider/gitlab/unparse_pipeline.go` | Added `unescapeHCLUnicode` + applied to all HCL returns |
| `provider/gitlab/pipeline_yaml.go` | Added `unescapeYAMLUnicode` + applied to YAML output |
| `provider/github/roundtrip_test.go` | Added `TestEmojiRoundtripStability` |

---

## Prevention Strategies

### 1. Test Coverage (Primary Prevention)

The `TestEmojiRoundtripStability` test in `provider/github/roundtrip_test.go` provides direct
regression coverage:

```go
func TestEmojiRoundtripStability(t *testing.T) {
    emojiName := "\U0001F46E\u200D\u2642\uFE0F Lint"  // 👮‍♂️ Lint
    // Verifies: no \u200d in HCL, no \U in YAML, emoji preserved verbatim
}
```

### 2. When to Apply the Same Pattern

Apply the same `unescapeXXXUnicode` post-processing pattern at **any output boundary** where:
- An external serialisation library is in use
- That library uses `unicode.IsPrint` or equivalent heuristics
- User-controlled strings may contain non-BMP / category-Cf characters

Common triggers: emoji, RTL marks (U+200E/200F), word joiners (U+2060), BOM (U+FEFF).

### 3. Threshold: U+009F

The post-processors only unescape codepoints `> 0x9F`. This preserves:
- ASCII (U+0000–U+007F): standard escaping rules apply
- C1 controls (U+0080–U+009F): genuinely unsafe in YAML/HCL

Everything ≥ U+00A0 is valid UTF-8 in both YAML 1.2 and HCL, so it can safely be unescaped.

### 4. New Provider Checklist

When implementing a new provider or serialisation path, apply both unescapers:
- HCL generation: wrap `hclwrite.Format(f.Bytes())` with `unescapeHCLUnicode(...)`
- YAML generation: wrap final `[]byte` output with `unescapeYAMLUnicode(...)`

---

## Known Limitations

- This fix does not cover other binary/non-UTF-8 content (which should remain escaped)
- If `hclwrite` or `yaml.v3` upstream fixes their libraries, these post-processors become no-ops
  (harmless, since they only act on escape sequences that wouldn't be present)
- The YAML post-processor targets double-quoted `\UXXXXXXXX` form; single-quoted YAML strings
  cannot contain escape sequences (yaml.v3 won't produce them for format chars)

---

## External Library Notes

| Library | Version | Issue | Status |
|---------|---------|-------|--------|
| `github.com/hashicorp/hcl/v2` | v2.24.0 | `hclwrite` uses `unicode.IsPrint` → escapes Cf chars | Upstream, not fixed |
| `gopkg.in/yaml.v3` | v3.0.1 | `is_printable` misses 4-byte UTF-8 (supplementary-plane) | Upstream bug, not fixed |
