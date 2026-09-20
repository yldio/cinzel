---
title: "the step-only path reads the file through a second reader, and loses digits"
module: "provider/github"
problem_type: logic_error
component: "provider/github"
severity: high
root_cause: "parseStepsFromYAML re-reads the raw bytes with cty's reader, which has none of the retagging pass parseYAMLDocument runs on the workflow path"
symptoms:
  - "a long run of digits in a step-only file comes back with its tail replaced by zeros"
  - "a 180 digit value loses 35 of them, at exit 0"
  - "the same value in a workflow file is correct, so the two paths disagree"
  - "HCL pass 1 and pass 2 differ: the number is written bare, then quoted"
tags:
  - "roundtrip"
  - "yaml"
  - "numbers"
  - "two-readers"
status: "fixed"
created_date: "2026-09-20"
updated_date: "2026-09-20"
related:
  - docs/solutions/logic-errors/HANDOFF-github-internal-review.md
---

# The step-only path reads the file through a second reader

## The criterion

The handoff's test applies: a defect is cinzel's only if cinzel writes output
its own parser rejects, or a roundtrip loses information. This is the second
kind, and the plainest sort of it — digits that went in do not come out.

## What the handoff asked

`HANDOFF-github-internal-review.md` flagged one thing for a deeper look:

> gitlab's hole was that one reader accepted a document the other refused, and
> only the refusal path was checked. Whether github's single goccy-plus-yamlv3
> split has the same asymmetry was not established.

It does not have that asymmetry. `parseYAMLDocument` returns its yaml.v3 error
and `unparseYAMLFile` stops on it, so there is no equivalent of the `return
nil` that `b8d3260` had to remove from gitlab. Probing the three shapes that
commit found — an unknown `%FOO` directive, a `%YAML 1.2` version, an
undefined `!e!foo` tag handle — each is refused at exit 1 on the workflow path.

The split is real, though. It is just in a different place.

## The fault

`unparseYAMLFile` tries three classifications in turn. The first two work from
the document `parseYAMLDocument` already decoded. The third does not:

```go
steps, err := parseStepsFromYAML(yamlBytes)
```

It is handed the raw bytes, and reads them again through `ctyyaml.Unmarshal`.

That matters because of what the first read does on the way past.
`keepWholeNumbersExact` retags a run of digits too long for an integer as a
string: yaml.v3 resolves such a run to a float, and a float that wide has no
room for every digit. The comment on it says as much, and the workflow path has
been correct since.

The second reader has no such pass. cty resolves the run to a `cty.Number`,
which loses the same digits for the same reason:

```yaml
checkout:
  name: Checkout
  run: echo hi
  env:
    BIG: <180 digits>
```

comes back as 145 digits followed by 35 zeros. A different number, at exit 0.
The smaller case in the existing test, `99999999999999999999`, is written bare
on this path where the workflow path quotes it, so the second pass writes
something the first did not and the roundtrip is unstable as well.

## The fix

Run the same pass over the bytes before the second reader sees them:

```go
func keepWholeNumbersExactInYAML(content []byte) ([]byte, error) {
	var node yamlv3.Node

	if err := yamlv3.Unmarshal(content, &node); err != nil {
		return nil, err
	}

	keepWholeNumbersExact(&node)

	return yamlv3.Marshal(&node)
}
```

`parseStepsFromYAML` calls it first and works from what it returns. The two
readers are then looking at the same document, which is what the gitlab fix
was about too, by a different route.

Reusing the existing pass rather than writing a second one is the point: a
number rule that lives in two places drifts.

## Tests

`TestScalarsSurviveTheStepOnlyDecodeBoundary` sits beside the workflow-path
test it mirrors, in `decode_boundary_test.go`, and runs the same shapes through
a step-only document: the twenty digit case, the 180 digit case, the largest
integer that still fits, and a float, which must not be retagged.

Neutered by replacing the `keepWholeNumbersExact(&node)` call with a discard.
The 180 digit case failed with the zeroed tail in the message, which is what it
is there to catch.

## Not covered

Comments are dropped on the step-only path — `collectComments` runs in
`parseYAMLDocument`, and nothing carries the result into `parseStepsFromYAML`.
That predates this change and is unrelated to it; it was confirmed against the
binary built before the fix. Worth a look on its own.
