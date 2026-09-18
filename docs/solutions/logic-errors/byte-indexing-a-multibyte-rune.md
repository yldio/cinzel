---
title: "Byte-indexing a string that holds multibyte runes"
module: "Core"
problem_type: "logic_error"
component: "internal/naming, internal/ai"
severity: "high"
root_cause: "logic_error"
symptoms:
  - "An identifier starting with a non-ASCII digit is written out unprefixed and the HCL it lands in will not parse"
  - "`cinzel unparse` exits 0 and the file it wrote fails on the next `parse`"
  - "HCL context sent to an AI provider ends in U+FFFD"
tags:
  - "unicode"
  - "utf8"
  - "identifiers"
  - "hcl"
  - "truncation"
created_date: "2026-09-18"
updated_date: "2026-09-18"
---

## Problem Description

Two functions read a string by byte where the question was about runes. Both
were correct for ASCII and wrong for everything else, which is why both survived
their own tests.

### 1. The leading-digit check in `SanitizeIdentifier`

HCL refuses an identifier that starts with a digit, so `SanitizeIdentifier`
(`internal/naming/naming.go`) prefixes one with `_`. The check was:

```go
if unicode.IsDigit(rune(out[0])) {
    return "_" + out
}
```

`out[0]` is a byte. `rune(out[0])` on a multibyte rune is that rune's UTF-8
*lead byte* — for `٣` (U+0663, ARABIC-INDIC DIGIT THREE) it is 0xD9, which is
not a digit in any category. So the prefix was skipped and the identifier went
out starting with a digit:

```
YAML:  jobs: { ٣build: ... }
HCL:   job "٣build" { ... }      ← written, exit 0
       cinzel github parse       ← fails: invalid identifier
```

Every job label, template label, variable name and reference in both providers
is named through this one function, so the reach is the whole unparse direction,
not one call site.

### 2. The fallback cut in `truncateAtNewline`

`internal/ai/strip.go` caps the HCL context sent to an AI provider at
`maxContextBytes`. It cuts at the last newline inside the limit, and when there
is no newline it cuts at the limit itself:

```go
cut := s[:maxLen]

if i := strings.LastIndex(cut, "\n"); i > 0 {
    return cut[:i]
}

return cut   // ← lands mid-rune
```

The doc comment claimed the cut avoided "splitting mid-line or mid-rune". The
newline path does; the fallback did not. A limit landing one or two bytes into a
three-byte rune left a partial encoding at the end, which is not a rune and
reaches the provider as U+FFFD — inside HCL structure the provider is being
asked to read.

## Root Cause

A `string` in Go indexes as bytes and ranges as runes. Both sites used the
indexing form to ask a rune-level question:

- "is the first character a digit" is about the first *rune*
- "cut here without breaking a character" is about a *rune boundary*

ASCII hides the difference completely, and every test case for both functions
was ASCII.

## Solution

Read runes with `unicode/utf8` rather than indexing bytes.

```go
// internal/naming/naming.go
if first, _ := utf8.DecodeRuneInString(out); unicode.IsDigit(first) {
    return "_" + out
}
```

```go
// internal/ai/strip.go — after the newline path
// An encoding is at most four bytes, so at most three trailing bytes can be
// a partial one.
for range utf8.UTFMax - 1 {
    if r, size := utf8.DecodeLastRuneInString(cut); r != utf8.RuneError || size != 1 {
        break
    }

    cut = cut[:len(cut)-1]
}
```

`DecodeLastRuneInString` returns `(RuneError, 1)` for an invalid encoding and
`(RuneError, 3)` for a real U+FFFD in the text, so the size check is what keeps
a legitimate replacement character from being eaten.

## Prevention

- Reach for `utf8.DecodeRuneInString` / `DecodeLastRuneInString` whenever the
  question is "which character", never `s[i]` or `rune(s[i])`.
- `for _, r := range s` already yields runes — the bug appears where a range is
  not used, which is usually a first-or-last check.
- Test non-ASCII deliberately. Useful cases: `٣` (U+0663, 2 bytes), `３`
  (U+FF13, fullwidth, 3 bytes), `日` (3 bytes), an emoji (4 bytes), and `café`
  as the control that must *not* be touched.
- A cut by byte count is a rune-boundary question by definition. There is no
  byte limit that is safe on arbitrary text.

### Tests

- `TestSanitizeIdentifierPrefixesEveryLeadingDigit`
  (`internal/naming/naming_test.go`) — ASCII, Arabic-Indic and fullwidth digits,
  plus `café` and `go2` as the two cases that must be left alone.
- `TestTruncateAtNewline` (`internal/ai/provider_test.go`) — cuts one, two and
  three bytes into a multibyte rune, and one on a boundary.

## Related

- [`unicode-emoji-zwj-escape-roundtrip.md`](../runtime-errors/unicode-emoji-zwj-escape-roundtrip.md) — the same class in external libraries: `yaml.v3`'s printability check misses 4-byte UTF-8 sequences
