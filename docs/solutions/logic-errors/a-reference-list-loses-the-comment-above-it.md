---
title: "A comment above jobs or steps was dropped on unparse"
module: "provider/github"
problem_type: logic_error
component: "provider/github/unparse_workflow.go"
severity: medium
root_cause: "the writers skip jobs and steps before the line that writes a head comment, and the emitter they reach took no comment"
symptoms:
  - "a comment above the workflow 'jobs' key does not reach the HCL"
  - "a comment above a job's 'steps' key does not reach the HCL"
  - "the same for an action's 'runs.steps'"
  - "exit code 0"
tags:
  - "comments"
  - "unparse"
  - "roundtrip"
status: fixed
created_date: "2026-09-25"
updated_date: "2026-09-25"
---

# A comment above jobs or steps was dropped on unparse

## The symptom

A workflow YAML carrying a note above its `jobs` key, or above a job's `steps`
key, came out of unparse without it:

```yaml
on:
  push: {}
# what this builds
jobs:
  build:
    runs-on: ubuntu-latest
    # what it does
    steps:
      - run: make
```

Neither comment appears in the HCL. Exit code 0, nothing on stderr. The next
parse writes YAML the author no longer recognises, because the two keys most
likely to carry a note are the two that lose it.

## The cause

`jobs` and `steps` are not attributes in the HCL schema. Each becomes a
top-level block referenced from a list, so the writers that walk the YAML keys
step past them:

```go
if key == "jobs" {
    continue
}
```

in `writeWorkflowMetadata`, and the same shape for `steps` in `writeJobBody`.
Both of those `continue` statements sit above the line that calls
`hclcomment.WriteLeading(body, comment.head)`, so the comment was read off the
YAML node, put in the map, and never asked for again.

The list itself is written later by `writeReferenceListAttribute`, which took
the attribute name, the block root and the references, and no comment. There
was nowhere for the head to go even if a caller had had it.

## The fix

`writeReferenceListAttribute` takes a `head string` and writes it before the
list, the way every other emitter in the file does:

```go
func writeReferenceListAttribute(body *hclwrite.Body, attr string, root string, refs []string, head string) error {
	if len(refs) == 0 {
		return nil
	}

	hclcomment.WriteLeading(body, head)
```

Three call sites pass what they have: `comments.at("jobs").head` in
`unparse_workflow.go`, `comments.at("steps").head` for a job, and
`comments.child("runs").at("steps").head` in `unparse_action.go`. The fourth,
`depends_on` in `unparse_emit.go`, passes `""`: it is written from a key the
loop does not skip, so its comment is already on the page above it, and passing
it here would print it twice.

## What this is not

Two things were looked at alongside this and are not part of it.

A comment followed by a blank line is handed back by yaml.v3 as the head of the
following key with a trailing newline on the string, which `WriteLeading` splits
into a second, empty `#` line in the HCL. It is cosmetic: the HCL parses, the
roundtrip is byte-stable over two full passes, and nothing accumulates. Under
the project's criterion it is not cinzel's defect to fix.

The foot comment of a non-last key was suspected of the same loss and is not:
yaml.v3 attaches these shapes as heads, not feet, and reverting the
carried-foot experiment changed no output at all.

## The test

`provider/github/reference_list_comments_test.go` covers both directions of the
skip, one subtest each. Both were confirmed to fail with the emitter's `head`
parameter reverted and to pass with it restored.
