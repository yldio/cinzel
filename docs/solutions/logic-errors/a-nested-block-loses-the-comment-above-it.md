---
title: "A key holding a map lost the comment written above it"
module: "provider/github"
problem_type: logic_error
component: "provider/github/unparse_workflow.go"
severity: low
root_cause: "the recursion that opens a block writes only what is inside it, and no caller wrote the head above it"
symptoms:
  - "a comment above 'run:' under 'defaults:' is gone from the HCL"
  - "a comment above 'shell: bash' one line below it survives"
  - "exit 0, no warning"
tags:
  - github
  - unparse
  - comments
status: fixed
created_date: "2026-09-25"
updated_date: "2026-09-25"
---

# A key holding a map lost the comment written above it

## What happened

```yaml
jobs:
  build:
    defaults:
      # why bash
      run:
        shell: bash
```

unparsed to

```hcl
job "build" {
  defaults {
    run {
      shell = "bash"
    }
  }
```

Exit 0. Moving the same comment one line down, above `shell: bash`, kept it.
So two comments a line apart in the same block behaved differently, and the
one that vanished did so without a warning.

## Why

`writeNestedMapAsBlock` walks a mapping and splits on what each value is. A
scalar goes to `writeCommentedAttribute`, which writes the head comment itself
before setting the attribute. A map recurses:

```go
if nestedMap, isMap := toStringAnyMap(value); isMap {
    if err := writeNestedMapAsBlock(blockBody, key, nestedMap, comments.child(key)); err != nil {
        return err
    }
    continue
}
```

The recursion opens a block for that key and writes what is inside it. Nothing
writes above it. `comments.child(key)` descends into the key's own children, so
the head comment on the key itself is passed over rather than lost in the
callee: `comments.at(key).head` still holds it, and no one reads it.

The top level does not have this, because `writeWorkflowMetadata` writes
`comment.head` before its switch, so a comment above `defaults:` at workflow
level survives. The gap is one level in, where the only writer is the recursion.

## The fix

The head goes on before the recursion, in the caller that still has it:

```go
hclcomment.WriteLeading(blockBody, comments.at(key).head)
```

It writes into `blockBody`, the body the block for this key is about to be
appended to, which is where `writeCommentedAttribute` puts a scalar's head as
well. So both branches now write the comment in the same place, and a reader
moving a comment between two adjacent lines gets the same treatment for both.

## Prevention

Same shape as `a-reference-list-loses-the-comment-above-it.md`: a branch that
hands the key to something else skips the line that writes the comment, because
that line lives in the branch it did not take. Whenever a loop splits on the
kind of value, the comment on the key belongs before the split, not inside one
arm of it.

## Tests

`provider/github/nested_block_comment_test.go` asserts the comment above the
nested map key, and carries the adjacent scalar as a control so a fix that
moved the write rather than adding one would fail.
