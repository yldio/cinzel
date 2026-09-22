---
title: "Each path reads half an env block's comments, and the closing one is written where nothing reads it"
module: "provider/github"
problem_type: logic_error
component: "provider/github, provider/github/step, provider/github/action"
severity: medium
root_cause: "two readers of the same block shape, one reading only the value expression and the other only the block body, and a writer putting the closing comment outside the block parse reads it from"
symptoms:
  - "a comment above an 'env {' on a step is dropped; the same comment on a job's env block is kept"
  - "a comment above the 'value =' inside a job's env block is dropped; the same comment on a step is kept"
  - "a comment closing an env or with block survives one conversion and is gone after the second"
  - "exit code 0 throughout"
tags:
  - "comments"
  - "roundtrip"
  - "reader-split"
status: "fixed"
created_date: "2026-09-20"
updated_date: "2026-09-20"
related:
  - docs/solutions/logic-errors/HANDOFF-github-internal-review.md
  - docs/solutions/logic-errors/hcl-inline-comment-propagation-to-yaml.md
---

# Each path reads half an env block's comments

Found by the sweep recorded in `HANDOFF-github-internal-review.md`, and the
same defect family as the findings before it: two code paths reading one shape,
each blind to what the other sees.

An `env` or `with` entry is written as a block:

```hcl
# above the block
env {
  name = "FOO"

  # above the value
  value = "bar" # beside the value
  # closing the block
}
```

Four places a comment can sit, one YAML key at the other end. Every one of them
is the author's, and all four belong on that key.

## The split

Two paths read these blocks, and they read different halves.

The step path (`provider/github/step/stepparse.go` into `setNested`) went
through `EnvListConfig.ValueRanges`, which returns the range of the `value`
expression and nothing else. It kept the comment above `value =` and the one
sharing its line. The comment above `env {` and the one closing the block were
never looked for.

The job, workflow and action path (`parse_workflow.go` and `parse_action.go`
into `blockComments`) read the block body: `sb.SrcRange` for the head and
`sb.EndRange` for the foot. It kept the comment above `env {` and the closing
one. The value's own comments were never looked for.

Mirror images. Whichever way an author wrote it, one of the two levels dropped
it, and the CLI exited 0.

The repository's own config was losing comments to this. `cinzel/steps.hcl`
explains above two `env {` blocks why credentials live in a step's environment
rather than in `.git/config`; none of it reached `.github/workflows`. The
regenerated files in this change are additions only.

## The fix

Both paths read both. `EnvListConfig.BlockComments` and
`WithListConfig.BlockComments` (`provider/github/action/ranges.go`) give the
step path the block's head and foot, which needs `EnvConfig` and `WithConfig`
to keep their source body — a `hcl:",remain"` field, the way `hclNamedBlock`
already had one. `namedBlockComments` (`parse_workflow.go`) gives the job path
the value's head and trailing comment.

The two head runs are joined in the order they were written, because both end
up above the one key the block becomes.

## The second fault, which the first was hiding

With the step foot reaching YAML for the first time, the two-pass check went
unstable. Comparing against a binary built from before the change showed the
job foot had been lost the same way all along; the fix had only widened an
existing hole.

`writeNameValueBlocks` wrote a mapping's closing comment into the surrounding
body, after all the blocks. Parse reads a block's foot from `sb.EndRange` —
inside the block. So the comment survived one conversion and was gone by the
second. It now goes inside the last block written, which is where the next
parse looks. `lastEntry` picks that block off the value mapping rather than off
the comment tree, because the tree holds only the keys that carried a comment
and the last key may not be one of them.

The step writer in `step/stepdecode.go` had the same gap for a different
reason: it never emitted `comment.Foot` at all.

## Not a validation question

The criterion in the handoff: a defect is cinzel's only if cinzel writes output
its own parser rejects, or a roundtrip loses information. This is the second.
Nothing here asks whether GitHub would accept the file.

## Tests

`provider/github/env_block_comments_test.go`. All four positions at both
levels, the closing comment landing inside the last block, and a two-pass
stability check over an input carrying every position at once.

Each of the five changes was neutered in turn and the matching subtests
confirmed failing before the tests were believed.
