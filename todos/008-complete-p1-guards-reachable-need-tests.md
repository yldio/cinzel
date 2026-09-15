---
status: complete
priority: p1
issue_id: "008"
tags: [code-review, quality, architecture]
dependencies: []
---

# Completeness check in `buildWorkflowJobIndex` defends unreachable state (YAGNI)

## Problem Statement

The new loop at the end of `buildWorkflowJobIndex`:

```go
for jobName := range jobs {
    if _, covered := jobIDMap[jobName]; !covered {
        return nil, nil, nil, fmt.Errorf("job '%s' is defined but was not included in the job order", jobName)
    }
}
```

Both `jobs` (from `doc.Jobs`) and `order` (from `jobOrderFromNode`) are derived from the **same** `yaml.v3.Node` tree in the same parse pass. They cannot diverge. The check defends against a state that is structurally impossible given the current call graph, adds dead code, and will produce a confusing error message if `jobOrderFromNode` has a bug (the message implies a caller contract violation, not an internal bug).

**Note:** This concern does not apply to the fallback path where `order` is `nil` and `jobNames = sortedKeys(jobs)` is used — in that case there is no order/map divergence to check anyway.

## Findings

- `provider/github/unparse_emit.go:52-58`
- Both `order` and `jobs` originate from `parseYAMLDocument` in the same yaml.v3 node tree
- The check fires only when `order` is non-empty, which is exactly the case where divergence is impossible

## Proposed Solutions

### Option A — Remove the completeness loop (Recommended)

Delete lines 52–58 of `unparse_emit.go`. If `jobOrderFromNode` produces wrong output, the existing test suite will catch it.

- **Pros:** -6 LOC; removes dead code; function intent is clearer
- **Cons:** slightly less defensive — an acceptable tradeoff given the invariant is provable

### Option B — Keep the loop but add a comment explaining the invariant

Add: `// Invariant: order and jobs derive from the same yaml.v3 node; this check catches bugs in jobOrderFromNode.`

- **Pros:** explicit invariant documentation
- **Cons:** dead code still present; misleading error message

## Recommended Action

**Do not remove either loop.** The premise above is wrong: `order` and `jobs`
do diverge, because they are built by two different mechanisms over the same
node tree. `jobOrderFromNode` reads raw key nodes, while `Decode` resolves
aliases and drops duplicates. Two inputs reach the supposedly-dead code:

Alias used as a job key — `jobOrderFromNode` records the anchor name, `Decode`
records the resolved value, so the `order` name is missing from `jobs`:

```yaml
on: push
jobs:
  a:
    runs-on: &k ubuntu-latest
    steps: [{run: echo hi}]
  *k :
    runs-on: ubuntu-latest
    steps: [{run: echo two}]
```

gives `order=[a k]` against `jobs=[a ubuntu-latest]`, and the first guard
fires with `job 'k' listed in order but not found in jobs map`.

Empty job key — the `key != ""` guard in `jobOrderFromNode` skips it, `Decode`
keeps it, so `jobs` holds a name `order` never listed:

```yaml
on: push
jobs:
  a:
    runs-on: x
    steps: [{run: hi}]
  "":
    runs-on: x
    steps: [{run: hi}]
```

gives `order=[a]` against `jobs=[a ""]`, and the completeness loop fires with
`job '' is defined but was not included in the job order`.

Both were confirmed by probe against the code at 3a8ee45. What is worth doing
instead is turning these two inputs into regression tests, so the guards have
coverage proving why they exist. That overlaps issue 010, which asks the same
question about the `key != ""` guard and should be resolved with it.

## Technical Details

- Affected file: `provider/github/unparse_emit.go`, `buildWorkflowJobIndex`

## Acceptance Criteria

- [x] Reachability of both guards established by probe
- [x] Alias-key and empty-key inputs are covered by regression tests
- [x] The inputs that reached the guards are fixed at their source
- [x] A crash on a null mapping key is fixed

## Work Log

- 2026-03-31: Finding created during code review
- 2026-09-15: Probed against 3a8ee45. Premise refuted — both guards are
  reachable via alias keys and empty keys. Reclassified from "remove dead
  code" to "add the missing regression tests". Priority lowered to p3: the
  code is correct as written, only its coverage is missing.
- 2026-09-15: Re-probed against 9ed4be6 and found my own earlier triage
  half wrong, plus a crash neither review had seen.

  The alias claim was right that the guard fires, wrong about why. An alias in
  key position does not have an empty `Value` — it holds the anchor's name
  ("k"), with the text it stands for in `.Alias`. So `key != ""` never skipped
  it. `jobOrderFromNode` now resolves the alias, and the order matches the map.

  The empty-key claim was also half wrong. Removing `key != ""` does not break
  anything: the workflow unparses and emits `id = ""`. So the guard was the
  only thing rejecting an unnamed job, and it did it through a message about
  job order that says nothing about the real problem. The parse direction
  already refuses an empty id with `errJobIDNotString`; the validator now says
  the same thing on the way in, and the guard is gone.

  Two more inputs turned up that neither issue mentions:

  A **null mapping key crashed the process**. yaml.v3 decodes a mapping with a
  non-string key into `map[any]any`, and a nil key there panics the goccy
  encoder the validator runs the document through — a segmentation fault from
  the CLI on a six-line file, at any depth, not just in `jobs`. Integer and
  boolean keys did not crash but were dropped silently. `rejectNonStringKeys`
  refuses all of them before the decode.

  A **merge key inside the jobs mapping** put "<<" in the job order and lost
  the merged job. The order now falls back to sorted keys when one is present.

  With those four fixed, both guards in `buildWorkflowJobIndex` are unreachable
  by any input probed. They are left in place: they cost nothing, and the
  history here is that each round of probing found another way to reach them.
