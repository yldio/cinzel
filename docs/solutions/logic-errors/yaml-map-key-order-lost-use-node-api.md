---
title: Preserve YAML job key order in cinzel unparse without double-parsing
description: Fix job order loss during YAML→HCL unparse by using yaml.v3 Node API for a single parse that yields both map data and ordered keys
date: 2026-03-31
category: logic-errors
tags: [yaml, go, yaml-v3, key-order, parsing, unparse, roundtrip, hcl, github-actions]
symptoms:
  - Jobs in HCL output are sorted alphabetically instead of matching source YAML order
  - Go map iteration non-determinism causes inconsistent job ordering across runs
  - Fixing order with a second yaml.v3 parse violates the "do not unmarshal YAML twice" rule
components:
  - parseYAMLDocument
  - workflowToHCL
  - buildWorkflowJobIndex
  - YAMLDocument struct
severity: medium
resolved: true
---

# Preserve YAML job key order in unparse without double-parsing

## Symptoms

- `cinzel unparse` emits jobs in alphabetical order instead of the order they appear in the source YAML
- Job order is non-deterministic across runs (Go map iteration randomises output)
- A naive fix adds a second YAML parse, violating the CLAUDE.md rule "do not unmarshal YAML twice"

## Root Cause

`goccy/go-yaml` (and `yaml.v3` plain `Unmarshal`) decode YAML mappings into `map[string]any`, which loses insertion order. There was no ordered representation of the keys after parsing.

The tempting fix — a second `yaml.v3.Unmarshal` just to walk `node.Content` for key order — parses the same bytes twice and creates a correctness risk: if the two parsers produce different views of the document (anchor expansion, merge keys), `jobOrder` and `jobs` can diverge silently.

## Solution

Use `gopkg.in/yaml.v3`'s Node API as the **single** parse entry point. One `yamlv3.Unmarshal` into a `*yamlv3.Node` gives a complete parse tree. From that tree you get two views for free:

- `node.Decode(&doc)` — converts the already-parsed tree into `map[string]any` (no second parse; this is a tree walk)
- `node.Content` — pairs of [key, value] child nodes in source declaration order

```go
// parseYAMLDocument parses YAML content into a document map and extracts
// job names in source order in a single pass using the yaml.v3 Node API.
func parseYAMLDocument(content []byte) (map[string]any, []string, error) {
	var node yamlv3.Node

	if err := yamlv3.Unmarshal(content, &node); err != nil {
		return nil, nil, err
	}

	if len(node.Content) == 0 {
		return nil, nil, nil
	}

	var doc map[string]any

	if err := node.Decode(&doc); err != nil {
		return nil, nil, err
	}

	return doc, jobOrderFromNode(node.Content[0]), nil
}

// jobOrderFromNode extracts job names in source order from a yaml.v3 mapping
// node. It relies on the Node API's preservation of mapping key order, which
// is not available when unmarshaling directly into map[string]any.
func jobOrderFromNode(root *yamlv3.Node) []string {
	if root.Kind != yamlv3.MappingNode {
		return nil
	}

	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == "jobs" {
			jobs := root.Content[i+1]

			if jobs.Kind != yamlv3.MappingNode {
				return nil
			}

			keys := make([]string, 0, len(jobs.Content)/2)

			for j := 0; j+1 < len(jobs.Content); j += 2 {
				if key := jobs.Content[j].Value; key != "" {
					keys = append(keys, key)
				}
			}

			return keys
		}
	}

	return nil
}
```

**Key insight**: After `yamlv3.Unmarshal(content, &node)`, `node` is a DocumentNode and `node.Content[0]` is the root MappingNode. `node.Decode(&doc)` re-uses the already-resident node tree — it does not re-parse the input bytes. The CLAUDE.md rule "do not unmarshal YAML twice" is satisfied.

### Pattern

```
yaml.v3.Node  ──── node.Decode(&typedMap)  →  map[string]any  (values, for conversion)
              └─── walk node.Content        →  []string keys   (order, for output order)
```

### Related structural changes

**Pass job order as an explicit parameter** — do not store it on `YAMLDocument`. Order is a parse-time artifact of the source bytes, not a semantic property of the document:

```go
// Before: field on struct, set externally after construction
type YAMLDocument struct {
    // ...
    JobOrder []string
}

// After: explicit parameter at the call site
func workflowToHCL(doc ghworkflow.YAMLDocument, filename string, jobOrder []string) ([]byte, error)
```

**Distinguish presence vs type errors in `buildWorkflowJobIndex`**:

```go
raw, exists := jobs[jobName]
if !exists {
    return nil, nil, nil, fmt.Errorf("job %q listed in order but not found in jobs map", jobName)
}
jobMap, ok := toStringAnyMap(raw)
if !ok {
    return nil, nil, nil, fmt.Errorf("job %q must be an object", jobName)
}
```

## What Didn't Work

Adding a standalone `yamlJobOrder(content []byte) []string` function that called `yamlv3.Unmarshal` separately — after `parseYAMLDocument` had already called `goccy/go-yaml`. Two parsers, same bytes, different internal models. Violated the CLAUDE.md pitfall "do not unmarshal YAML twice" and introduced a latent correctness risk if the two parsers diverged on anchor expansion or merge keys.

## Prevention

### Detecting a re-introduced double parse

Search for any function that calls both `yaml.Unmarshal` and `node.Decode` (or two separate `Unmarshal` calls) on the same input bytes:

```bash
grep -n "Unmarshal\|\.Decode" provider/github/unparse_workflow.go
```

A PR that adds a new `yaml.Unmarshal` call in the unparse path should be flagged unless there is an explicit comment explaining why a second parse is unavoidable.

### When adding a new "get key order from YAML" requirement

1. Is there already a `yaml.v3.Node` in scope? Walk `node.Content` directly — no new parse.
2. If not, introduce `yaml.v3.Node` as the single parse entry point; get typed data via `node.Decode`.
3. Never add a second `yaml.Unmarshal` just to extract key order.

### Test strategies

**Golden roundtrip with non-alphabetical job order** — most direct regression catch:

- Fixture: a workflow with jobs declared in deliberate reverse-alphabetical order (`deploy`, `build`, `lint`).
- Assert: HCL output preserves `deploy → build → lint`, not sorted `build → deploy → lint`.
- Assert: YAML→HCL→YAML roundtrip produces identical job order.

**Unit test for `jobOrderFromNode`**:

```go
func TestJobOrderFromNode(t *testing.T) {
    yaml := "jobs:\n  deploy:\n  build:\n  lint:\n"
    var node yamlv3.Node
    yamlv3.Unmarshal([]byte(yaml), &node)
    got := jobOrderFromNode(node.Content[0])
    want := []string{"deploy", "build", "lint"}
    // assert got == want
}
```

### CLAUDE.md rule

> **Pitfalls**: do not unmarshal YAML twice

## Related Documentation

- [`docs/solutions/logic-errors/nondeterministic-map-iteration.md`](nondeterministic-map-iteration.md) — Go map iteration as a source of unstable output; canonical fix using `sortedKeys`
- [`docs/solutions/patterns/critical-patterns.md`](../patterns/critical-patterns.md) — Mandated patterns including the single-unmarshal rule
- [`docs/solutions/runtime-errors/unicode-emoji-zwj-escape-roundtrip.md`](../runtime-errors/unicode-emoji-zwj-escape-roundtrip.md) — Another roundtrip stability issue: ZWJ emoji escaping
- [`docs/solutions/logic-errors/github-strict-schema-parse-unparse-parity-unknown-rejection-stable-mapping.md`](github-strict-schema-parse-unparse-parity-unknown-rejection-stable-mapping.md) — Parse/unparse schema parity
- [`docs/solutions/developer-experience/yaml-string-quoting-rules.md`](../developer-experience/yaml-string-quoting-rules.md) — When yaml.v3 Node API output must use double-quoted style
