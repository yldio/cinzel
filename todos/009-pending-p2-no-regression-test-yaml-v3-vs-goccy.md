---
status: pending
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

_To be filled during triage._

## Technical Details

- Affected file: `provider/github/unparse_workflow.go`, `parseYAMLDocument`
- Decoder switched from `github.com/goccy/go-yaml` to `gopkg.in/yaml.v3` for the `map[string]any` decode step

## Acceptance Criteria

- [ ] At least one test exercises `parseYAMLDocument` with a YAML `null` value, a boolean-like string, and a merge key
- [ ] Tests pass and document the expected behavior

## Work Log

- 2026-03-31: Finding created during code review
