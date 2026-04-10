---
title: "github pin inline comment format and workflow permissions least-privilege default"
date: 2026-04-10
category: logic-errors
tags: [cinzel, hcl, github-actions, yaml, permissions, security, pin, upgrade, sentinel-swap]
symptoms:
  - "cinzel github pin adds comment above uses block instead of inline on version line"
  - "comment style uses // instead of # on version line after pinning"
  - "empty permissions {} HCL block emits permissions: read-all in YAML output"
  - "no permissions block in HCL produces no permissions key in YAML output at all"
  - "security scanner (zizmor/Scorecard) warns: overly broad permissions: uses read-all permissions"
  - "permissions: {} stripped to permissions: (null) by global empty-map post-processor"
affected_components:
  - "internal/pin/pin.go"
  - "internal/pin/upgrade.go"
  - "provider/github/parse_workflow.go"
  - "provider/github/workflow_yaml.go"
problem_type: logic-errors
related:
  - docs/solutions/patterns/assist-pin-upgrade-feature-implementation.md
  - docs/solutions/logic-errors/preserve-hcl-job-order-in-yaml-output.md
---

# github pin inline comment format and empty permissions least-privilege default

Two independent logic errors in the `github pin`/`upgrade` commands and the workflow permissions parser, both producing semantically wrong output from valid input.

## Symptoms

**Pin comment placement:**
After running `cinzel github pin`, the generated HCL had a `//` comment placed *above* the `uses {}` block instead of inline on the `version` line:

```hcl
// actions/checkout v4   ← wrong: detached, wrong style
uses {
  action  = "actions/checkout"
  version = "abc123sha..."
}
```

**Empty permissions escalation:**
A security scanner flagged generated workflow YAML files:

```
Warning: commitlint.yml:10: overly broad permissions: uses read-all permissions
Warning: security.yml:11: overly broad permissions: uses read-all permissions
Error: Process completed with exit code 13.
```

This occurred even when the source HCL had an empty `permissions {}` block — i.e. the author explicitly declared no permissions, yet the output granted broad read access.

**Missing permissions key:**
Workflows with no `permissions` block at all produced YAML with no `permissions` key — causing GitHub Actions to inherit the repo's default token permissions (often `contents: write` or broader) rather than denying all access.

## Root Cause

### Pin comment placement

`PinFile` and `UpgradeFile` called `upsertUsesComment()` to inject a freestanding `// action tag` comment above the `uses {}` block. Two problems:

1. **Wrong position** — the comment was detached from the field it documented; editors and diff tools could separate it from the `version` line.
2. **Wrong style** — `//` is valid HCL but `#` is the conventional inline comment style, matching the GitHub Actions YAML convention (`uses: actions/checkout@sha # v4`).

### Empty permissions → read-all

In `parse_workflow.go`, the code treated an empty body map as a cue to apply the `read-all` shorthand:

```go
if len(child) == 0 {
    out["permissions"] = "read-all"   // ← inverted semantics
} else {
    out["permissions"] = child
}
```

An empty `permissions {}` block means *deny all* — the most restrictive state. The code did the opposite, expanding it to the broadest read shorthand.

### No permissions block → missing key

`parseWorkflowConfig` only set `out["permissions"]` inside the `cfg.PermBlocks` loop. With no block present, the key was never set, so the YAML had no `permissions` field at all — silently inheriting the repo's token permissions.

### `permissions: {}` erased by global post-processor

`marshalWorkflowYAML` in `workflow_yaml.go` stripped all empty maps as a formatting pass:

```go
out := bytes.ReplaceAll(buf.Bytes(), []byte(": {}\n"), []byte(":\n"))
```

This turned `permissions: {}\n` into `permissions:\n` — which YAML decodes as `null`, not an empty map. A naive targeted fix (`ReplaceAll("permissions:\n", "permissions: {}\n")`) would break non-empty permissions blocks, since `permissions:\n  contents: read` also starts with `permissions:\n`. Go's `regexp` package has no negative lookahead, so the distinction cannot be expressed as a regex.

## Solution

### Problem 1 — inline `#` comment on version line

**`internal/pin/pin.go`** and **`internal/pin/upgrade.go`** (same change in both):

```go
// Before
oldLine := fmt.Sprintf(`version = %q`, ref.Version)
newLine := fmt.Sprintf(`version = %q`, sha)
updated = strings.Replace(updated, oldLine, newLine, 1)
updated = upsertUsesComment(updated, ref.Action, ref.Version)  // removed

// After
oldLine := fmt.Sprintf(`version = %q`, ref.Version)
newLine := fmt.Sprintf(`version = %q # %s`, sha, ref.Version)
updated = strings.Replace(updated, oldLine, newLine, 1)
```

The `upsertUsesComment` function was deleted entirely. Output now:

```hcl
uses {
  action  = "actions/checkout"
  version = "abc123sha..." # v4
}
```

### Problem 2 — empty permissions → `{}`

**`provider/github/parse_workflow.go`** (two locations: workflow-level and job-level):

```go
// Before
if len(child) == 0 {
    out["permissions"] = "read-all"
} else {
    out["permissions"] = child
}

// After
out["permissions"] = child
```

Empty permissions now round-trips faithfully:

```hcl
# HCL input
permissions {}
```

```yaml
# YAML output
permissions: {}   # all scopes denied — correct least-privilege
```

### Problem 3 — no permissions block → always emit `permissions: {}`

**`provider/github/parse_workflow.go`** — add a default after the PermBlocks loop in `parseWorkflowConfig`:

```go
for _, block := range cfg.PermBlocks {
    child, err := parseBodyMap(block.Body, hv, "permissions")
    if err != nil {
        return nil, err
    }
    out["permissions"] = child
}

if _, ok := out["permissions"]; !ok {
    out["permissions"] = map[string]any{}
}
```

### Problem 4 — sentinel-swap to survive the global empty-map strip

**`provider/github/workflow_yaml.go`** — protect `permissions: {}` before the global replacement, restore after:

```go
var permissionsEmptySentinel = []byte("\x00PERM_EMPTY\x00\n")

// Inside marshalWorkflowYAML:
raw := buf.Bytes()
raw = bytes.ReplaceAll(raw, []byte("permissions: {}\n"), permissionsEmptySentinel)
raw = bytes.ReplaceAll(raw, []byte(": {}\n"), []byte(":\n"))
out := bytes.ReplaceAll(raw, permissionsEmptySentinel, []byte("permissions: {}\n"))
```

The sentinel (`\x00PERM_EMPTY\x00`) is a byte sequence that cannot appear in valid YAML, making false matches impossible. The three passes: protect → strip other empty maps → restore.

## Prevention

### Never treat empty as "grant everything"

`read-all` and `write-all` are expansion shorthands that broaden access. Only emit them when the source data explicitly contains that string — never as a fallback for an empty or zero-value input. The invariant: if the user wrote fewer scopes, the output must have fewer or equal scopes, never more.

### Global post-processors are fragile for security-sensitive fields

String replacements on serialized YAML have no structural awareness — they cannot distinguish `permissions: {}` (an explicit deny-all) from `cache: {}` (a harmless empty map). Any post-processor that strips or rewrites output must explicitly protect fields with security semantics, or be redesigned to operate at the node level.

When a global strip must be kept, use the **sentinel-swap pattern**: protect the target value with a byte sequence that cannot appear in valid YAML, run the strip, restore the sentinel. Prefer this over regex when Go's RE2 engine lacks the lookahead needed to express the distinction.

### Round-trip tests catch escalation bugs

A one-way golden test can pass even when permissions are silently escalated. A round-trip test (HCL → YAML → HCL) will fail if the output is semantically different from the input. Add explicit unit tests — not just golden file updates — for security-sensitive fields:

```
permissions {} → permissions: {}       (not read-all, not omitted)
no permissions → permissions: {}       (not omitted)
permissions { contents = "read" } → permissions:\n  contents: read   (scopes preserved)
```

Golden file bulk-updates during refactors can silently regress these invariants without a reviewer noticing. Dedicated unit tests like `TestParsePermissionsDefault` catch them immediately.

### Test boundary conditions for every enum-like field

- empty block
- single scope (`contents: read`)
- all scopes `read`
- all scopes `write`
- mixed scopes

### GitHub Actions least-privilege best practices

- Prefer explicit per-scope grants over shorthands (`contents: read` not `read-all`).
- `read-all` grants read on every scope including sensitive ones (id-token, secrets metadata).
- Security scanners (zizmor, Scorecard, StepSecurity) flag `read-all`/`write-all` as findings — generated workflows using these shorthands will fail CI security checks.
- The safest default for generated workflows is `permissions: {}` at the top level with explicit grants added per-job.

### Inline `#` comment is more robust than a block comment

The `# vtag` inline pattern is preferred over a freestanding comment above the block because:

1. **Mirrors YAML convention** — tools like Dependabot and `pin-github-action` use `# vX.Y.Z` inline.
2. **Survives reformatting** — a comment on the same line cannot be accidentally detached during editing.
3. **Tooling-friendly** — upgrade tools locate the version by scanning the SHA line, no lookahead needed.
4. **Visually scannable** — SHA and tag are visible together without scrolling.
