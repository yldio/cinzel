---
title: "A matrix axis renamed on the way to YAML, its reference left behind"
module: "GitHub provider"
problem_type: "logic_error"
component: "provider/github"
severity: "high"
root_cause: "logic_error"
symptoms:
  - "`matrix { go_version = [...] }` emits `go-version:` while the step still reads `${{ matrix.go_version }}`"
  - "actionlint: `property \"go_version\" is not defined in object type {go-version: ...}`"
  - "The same axis comes out spelled two ways depending on whether it was written as an attribute or a `variable` block"
  - "A YAML→HCL→YAML roundtrip does not return the axis name it started with"
tags:
  - "matrix"
  - "naming"
  - "roundtrip"
  - "github-actions"
  - "actionlint"
created_date: "2026-09-18"
updated_date: "2026-09-18"
---

## Problem Description

`parseBodyMap` (`provider/github/parse_workflow.go`) put every HCL key through
`naming.ToYAMLKey`, which replaces `_` with `-`. That is right for the keys
GitHub owns — `runs_on` must become `runs-on`, `fail_fast` must become
`fail-fast` — and wrong for a matrix axis, which is named by whoever wrote the
workflow.

An axis is referenced by its name in `${{ matrix.X }}`, and cinzel copies that
expression through as written. Renaming the axis and not the reference left the
reference resolving against an axis that no longer existed:

```hcl
matrix {
  go_version = ["1.24", "1.25"]
}

step "setup" {
  run = "go version $${{ matrix.go_version }}"
}
```

```yaml
matrix:
  go-version:          # ← renamed
    - "1.24"
steps:
  - run: go version ${{ matrix.go_version }}   # ← not renamed
```

actionlint on that output:

```
property "go_version" is not defined in object type {go-version: number; node-version: number} [expression]
```

The run fails at the reference, and nothing before it complains: `cinzel github
parse` exits 0.

### The two spellings

The rename reached attributes and not `variable` blocks, so the same axis came
out differently depending on how it was written:

| HCL | YAML emitted |
|---|---|
| `matrix { go_version = ["1.24"] }` | `go-version:` |
| `matrix { variable { name = "go_version" ... } }` | `go_version:` |

`include` / `exclude` entries were already left alone, so an include naming
`go_version` disagreed with the top-level axis in the same matrix.

## Root Cause

One rename applied to two kinds of key that only look alike:

- **Provider-owned keys** — `runs_on`, `fail_fast`, `max_parallel`,
  `timeout_minutes`. GitHub defines the spelling; HCL cannot hold the hyphen in
  an attribute position, so the rename is the whole point.
- **Author-owned names** — matrix axis names. cinzel never defined them and
  carries their references verbatim, so renaming one breaks the pair.

The unparse side already had this right. `writeMatrixBlock`
(`provider/github/strategy_matrix_unparse.go`) writes an axis name with
`cty.StringVal(axis.Name)` and `writeAttributeAny(matrixBody, axis.Name, ...)`
— no `toHCLKey`, unlike every surrounding call. So the rename was also the one
thing a roundtrip could not undo.

## Solution

Scope the rename out of the matrix, and only the matrix:

```go
// provider/github/parse_workflow.go
func yamlKeyIn(scope, name string) string {
	if scope == "matrix" {
		return name
	}

	return naming.ToYAMLKey(name)
}
```

Both call sites in `parseBodyMap` — the attribute branch and the generic block
branch — go through it. `parseBodyMap` already carried `scope` for other
decisions (`depends_on` in a job, `matrix` in a strategy), so no new plumbing
was needed.

`on` event keys are deliberately **not** exempt. `branches_ignore` →
`branches-ignore` and `pull_requests` → `pull-requests` are GitHub's own
spellings, so the rename is correct there.

## Prevention

Before putting a key through a naming transform, ask who named it:

- **The provider named it** — transform it. The spelling is fixed externally and
  HCL needs its own form.
- **The author named it** — leave it. Their name is likely referenced somewhere
  cinzel copies verbatim, and a transform breaks the pair.

The tell is a reference syntax. `${{ matrix.X }}`, `${{ env.X }}`,
`${{ needs.X.outputs.Y }}` all name something the author declared. If one side
of such a pair goes through a transform, both must, or neither.

Where an unparse and a parse disagree about a name, the unparse side is usually
the one to trust: it had to produce something a human would write.

### Tests

`TestParseKeepsMatrixAxisNames` (`provider/github/matrix_axis_name_test.go`)
covers both ways of writing an axis, an `include` entry, and two provider-owned
keys that must still be renamed. Without the fix it reports:

```
axis was renamed to "go-version:", which no ${{ matrix.* }} reference names
```

### actionlint as the oracle

Whether emitted YAML is *semantically* valid is not something a golden file
answers — a golden test happily locks in the broken spelling. actionlint
resolves `${{ matrix.X }}` against the axes actually defined, which is exactly
the property at stake. `mise run lint` already runs it over `.github/workflows`;
pointing it at a scratch repo holding one generated file is enough to check a
change like this.

## Related

- [`critical-patterns.md`](../patterns/critical-patterns.md) — expression escaping, the other place a `${{ }}` expression must survive untouched
