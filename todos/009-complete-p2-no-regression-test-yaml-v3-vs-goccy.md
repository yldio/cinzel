---
status: complete
priority: p2
issue_id: "009"
tags: [code-review, testing, quality]
dependencies: []
---

# No targeted regression test for yaml.v3 vs goccy/go-yaml decode behavioral differences

## Problem Statement

`parseYAMLDocument` switched from `goccy/go-yaml` (strict, YAML 1.2) to `yaml.v3`'s `node.Decode` for mapping the YAML document into `map[string]any`. The golden roundtrip tests cover happy-path workflow files but not the edge cases where the two libraries can produce different results:

- `null` / `~` values in maps
- Boolean-like strings (`on`, `off`, `yes`, `no`, `true`, `false`) as map values
- Integer-keyed maps (uncommon but valid YAML)
- Merge keys (`<<`) in job definitions
- Anchors/aliases (`&foo`, `*foo`) in the jobs mapping

If the new decode behavior silently changes any of these, the golden tests will not catch it because those edge cases are not present in the fixtures.

## Findings

- `provider/github/unparse_workflow.go:68` — `node.Decode(&doc)` replaces `yaml.Unmarshal` from `goccy/go-yaml`
- No test in `provider/github/` exercises YAML edge cases at the decode boundary
- Most relevant CLAUDE.md doc: "use strict typed decode (goccy/go-yaml)" — the switch to `yaml.v3` for the unparse decode path should be explicitly validated

## Proposed Solutions

### Option A — Add a unit test for `parseYAMLDocument` with edge-case YAML inputs (Recommended)

```go
func TestParseYAMLDocumentEdgeCases(t *testing.T) {
    cases := []struct{ name, yaml string; wantKey string; wantVal any }{
        {"null value", "jobs:\n  build:\n    timeout: ~\n", "timeout", nil},
        {"bool-like string on", "on:\n  push:\n", "on", ...},
        // etc.
    }
}
```

- **Pros:** explicit contract; catches future regressions
- **Cons:** need to know exact `yaml.v3` decode semantics for each case

### Option B — Add edge-case YAML fixtures to `testdata/matrix/unparse/`

Extend the matrix unparse test infrastructure with YAML files exercising the above edge cases, paired with expected HCL golden output.

- **Pros:** end-to-end; catches decode AND HCL emit differences
- **Cons:** more fixture overhead

## Recommended Action

Option A, plus a fix. Probing the edge cases the issue lists turned up a real
defect rather than only a test gap.

`yaml.v3` resolves a run of digits too large for `int64` or `uint64` to a
`float64`, and a float that wide cannot hold every digit. An ID written as
`99999999999999999999` came back as `1e+20` and was emitted as
`100000000000000000000` — a different number, with nothing said. `goccy` keeps
the digits as text, so the GitLab provider was never affected; the GitHub
unparse path picked this up when it switched decoders.

`keepWholeNumbersExact` retags such a scalar as a string before the decode.
Only a plain run of digits is touched: a quoted or explicitly tagged scalar
carries a non-zero `Style`, and `.inf` and `.nan` are not whole numbers, so
both keep the float the file asked for.

The rest of the list was already correct and is now pinned by tests: `~`,
`null` and an empty value all decode to null; `off`, `NO` and `yes` stay
strings under YAML 1.2; `1:30` is not sexagesimal; anchors, aliases and merge
keys are expanded by the decoder before the writer sees them.

## Technical Details

- Affected file: `provider/github/unparse_workflow.go`, `parseYAMLDocument`
- Decoder switched from `github.com/goccy/go-yaml` to `gopkg.in/yaml.v3` for the `map[string]any` decode step

## Acceptance Criteria

- [x] At least one test exercises `parseYAMLDocument` with a YAML `null` value, a boolean-like string, and a merge key
- [x] Tests pass and document the expected behavior
- [x] The precision loss the probe found is fixed, with a test that fails without the fix

## Work Log

- 2026-03-31: Finding created during code review
- 2026-09-15: Probed all the listed cases end to end against 2fc5fe5. Found one
  real defect (large whole numbers losing precision) and confirmed the rest
  already behaved correctly. Fixed the defect in `parseYAMLDocument` and pinned
  the whole list in `provider/github/decode_boundary_test.go`.

  The two large-number subtests were verified to fail with
  `keepWholeNumbersExact` removed, then restored. Merge-key handling was probed
  separately: a merge onto a scalar is rejected by the decoder with `map merge
  requires map or sequence of maps as the value`, which is clear enough to
  leave alone.
