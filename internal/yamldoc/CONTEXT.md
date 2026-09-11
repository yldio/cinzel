# yamldoc

The intermediate representation between a provider's source model and its
YAML output.

## Why it exists

The GitHub emitter used to hand `map[string]any` to yaml.v3 and then recover,
from the encoded bytes, three facts the map could not hold:

- **key order** — recovered by sorting, plus a `jobsOrder` key smuggled into
  the map and deleted on the way out
- **inline comments** — carried in an `annotated{value, comment}` wrapper that
  every consumer of the map had to know to unwrap
- **empty maps** — `bytes.ReplaceAll(raw, ": {}\n", ":\n")`, with
  `permissions: {}` protected by swapping in a NUL sentinel first

Each of those leaked. The byte rewrite could not tell a mapping from text
inside a scalar, so a `run` script containing `overrides: {}` was silently
corrupted. The wrapper was invisible to `walkStrings`, so a malformed `${{`
on a commented attribute skipped validation. goccy, used on the step-only
path, saw the wrapper's unexported fields and wrote `uses: {}`.

A `Doc` holds those facts directly, so the encoder reads them rather than
reconstructing them.

## Vocabulary

- **document** — an ordered mapping (`Doc`). Items keep the order they were
  `Set` in. For parsed input that is source order.
- **value** — one YAML value plus its inline comment. Its kind is null,
  scalar, map or sequence.
- **null vs empty map** — `Null()` encodes as a bare `key:`, `Map(New())` as
  `key: {}`. The difference is a property of the value, decided by whoever
  builds the document, not by the encoder.
- **ordering policy** — the caller's business, not the package's.
  `provider/github` applies `workflowKeyOrder` at the root and the declared
  job order under `jobs`, sorting whatever is left; the package itself has no
  opinion.
- **projection** — `plain` in `provider/github` strips the `annotated`
  wrapper for consumers that read parsed content instead of emitting it
  (validators). Values reaching a validator are never wrapper types.

## Boundaries

`Encode` still uses yaml.v3's Node API and inherits one of its defects: its
`is_printable` helper only recognises 3-byte UTF-8, so a 4-byte character
(an emoji) forces double-quoted style and downgrades a `|` literal block to
a quoted one-liner. `unescapeUnicode` restores the escaped characters but
cannot restore the block style. That is an encoder defect, not a
representation one; fixing it means replacing or patching yaml.v3.

The GitLab emitter (`provider/gitlab/pipeline_yaml.go`) still does the old
byte rewrite and has its own copy of the unicode unescape. It has not been
moved, and it does not have the corruption bug: it never forces literal style,
so a multi-line script containing `overrides: {}` is double-quoted and the
rewrite cannot reach it. Moving it is a tidying job, not a fix, and an attempt
showed the swap is not behaviour-preserving: `yamldoc.Null()` and
`yamldoc.Map()` do not reproduce `cache: null` or a `- {}` sequence entry.
