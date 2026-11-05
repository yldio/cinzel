---
title: "A comment written on one entry of a list was dropped in both directions"
module: "provider/github"
problem_type: logic_error
component: "provider/github/workflow_yaml.go, provider/github/yaml_comments.go, provider/github/unparse_workflow.go"
severity: medium
root_cause: "a list travels as a plain slice, so neither direction had anywhere to hold a comment written on one of its entries"
symptoms:
  - "'- edited # re-check the title' comes back as a bare 'edited'"
  - "a comment above a list entry is gone after unparse"
  - "the same comment written in HCL does not reach the YAML"
  - "exit code 0"
tags:
  - "comments"
  - "parse"
  - "unparse"
  - "roundtrip"
  - "annotated"
status: fixed
created_date: "2026-11-05"
updated_date: "2026-11-05"
---

# A comment on a list entry has no carrier

## The symptom

A trigger list with a note on one of its entries:

```yaml
  pull_request:
    types:
      - opened
      - edited # re-check the title
      - reopened
      - synchronize
```

came out of unparse as `types = ["opened", "edited", "reopened", "synchronize"]`.
The same note written in the HCL did not reach the YAML. Exit 0, nothing on
stderr, both ways.

## The cause

Comments travel per key, not per entry. On the parse side `annotated` holds a
`head`, a `comment` and a `foot` for the value a key carries; on the unparse
side `nodeComment` holds a `head` and a `line` for the key. A list reaching
either is a `[]any`, and both sides had one slot for the whole of it, which the
attribute's own comment already occupies.

The sequence paths that do exist are for mappings: `collectSeqComments` reads
the comments off each mapping in a sequence, which is what the step writers
walk. A scalar entry has no mapping to be collected into, so nothing read it.

Two further things were in the way.

`TrailingComment` read the text after an expression's end byte and required a
`#` or `//` to start it. A list entry ends before its comma, so every entry but
the last read as having no comment at all.

`hclcomment.WriteLeading` appends to a body. The entries of a list are tokens
inside one expression, so there is no body between the brackets to append a
comment above an entry to.

## The fix

A positional slice of `itemComment{head, line}` beside the list, on both
carriers: `annotated.items` going out, `nodeComment.items` coming back. An
entry with nothing written on it is an empty entry rather than a missing one,
so the slice lines up with the list it describes, and a list where no entry
carries anything gets a nil slice and the one-line form it had before.

Beside the list and not inside it. A wrapper put in the slice would reach every
consumer that reads an entry as a string — `NeedsFromYAML`, `ValidateSchedule`,
the matrix axes — and each of those reads a failed type assertion as an absent
or malformed value, not as a wrapper. That is the shape
`hcl-inline-comment-propagation-to-yaml.md` records being caught once already,
where a wrapped step id read as "no id" and the duplicate-id guard stopped
guarding.

Four pieces:

- `itemComments` (`parse_workflow.go`) reads each entry's comments off the
  `TupleConsExpr`, which is the only thing that knows where an entry was
  written. `annotate` keeps its signature; `annotateExpr` takes the expression.
- `collectItemComments` (`yaml_comments.go`) does the same off a YAML sequence,
  skipping entries that are not scalars: a mapping in a sequence is already
  collected by `collectSeqComments`, and reading it here too would write it
  twice.
- `commentedListTokens` (`unparse_workflow.go`) writes the list one entry to a
  line when entries carry comments, with `hclcomment.LeadingTokens` for the run
  above an entry.
- `TrailingComment` (`internal/hclparser/sources.go`) steps over a leading
  comma. The comma is syntax rather than something its author wrote.

## What this is not

GitLab has the same gap and is not fixed here. Its carriers are a different
pair (`comments` / `nodeComment` in `provider/gitlab/comments.go`) and its
`script` lists are the obvious place it shows.

## The test

`provider/github/list_item_comments_test.go`: each direction, the roundtrip,
and that an uncommented list keeps the one-line form. The first three were
confirmed to fail with the change reverted; the fourth passes either way, which
is the point of it.
