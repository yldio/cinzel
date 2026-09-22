---
title: "HCL typos sent to the issue tracker, and a template with one interpolation refused"
module: "internal/hclparser"
problem_type: logic_error
component: "internal/hclparser, provider/github/step"
severity: medium
root_cause: "internal/hclparser/errors.go carried no UserInput marking, and provider/github/step/stepparse.go appended the open-an-issue line by hand"
symptoms:
  - "a misspelled variable reference asks the author to open an issue about their own typo"
  - "the same index prints clean at job level and with the issue line at step level"
  - "\"${variable.list[0]}\" is refused where variable.list[0] resolves"
  - "that refusal prints a raw pointer dump with a broken %!s verb in it"
tags:
  - "errors"
  - "user-input"
  - "hclparser"
status: "fixed"
created_date: "2026-09-20"
updated_date: "2026-09-20"
related:
  - docs/solutions/logic-errors/HANDOFF-github-internal-review.md
---

# HCL typos sent to the issue tracker

## What the handoff asked

`HANDOFF-github-internal-review.md` flagged this as a cheap win:

> note `internal/hclparser/errors.go` has zero `UserInput` markings, where
> every other errors.go marks most of them. Some of those errors are certainly
> the author's fault and are sending people to the issue tracker over their own
> typo, which is exactly what `20f8a92` fixed in `internal/command`.

It was right about the markings. It was not cheap, for two reasons.

## 1. The markings

`cinzelerror.New` appends the open-an-issue line unless the error is marked
`UserInput`. `internal/hclparser/errors.go` held three sentinels and marked
none of them, and most of the package's errors were built inline with
`fmt.Errorf` and so could not be marked at all.

Probed through the CLI, every one of these prints the invitation to file a bug:

| Written | Refused with |
| --- | --- |
| `variable.list_os["prod"]` | a variable can only be indexed with a number |
| `variable.cfg[0]` | only a list or tuple variable can be indexed by position |
| `variable.list_os.foo` | a variable reference cannot reach into an attribute of its value |
| `variable.nope` | variable does not exist |
| `variable.list_os[9]` | index out of range |
| `5 / variable.zero` | division by zero |
| `5 % 2` | unsupported binary operator |

Each names something in a file the author keeps and commits.

The inline errors were given sentinels so they reach the mark. One is
deliberately left unmarked: `errUnsupportedExpressionType` means cinzel met an
expression it has no case for, which is cinzel's to answer for, and that one
keeps the line. A mark that covers the whole package says nothing.

## 2. The mark did not reach the output

Marking the sentinels and re-probing changed nothing at step level. The line
was still there.

`provider/github/step/stepparse.go` wrote it in by hand, at twelve sites:

```go
return Step{}, fmt.Errorf("error in step '%s': %w, %w",
	parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
```

So the hand-written line put back exactly what the mark had just taken off.
The job path wraps the same errors plainly and printed clean throughout, which
is how the two were found to disagree: the same index was the author's fault in
a job and cinzel's fault in a step.

`stepExprErr` puts them on the job path's footing. Two refusals in that file
are not expression errors — a step with no identifier, a step defined twice —
and keep the line they had.

This is the part worth remembering. **A `UserInput` mark is only as good as the
wrapping between it and the printer.** Marking a sentinel and stopping there
looks like a fix and is not one, which is why the step test asserts on the
printed message and not only on `IsUserInput`.

## 3. A template holding one interpolation was refused

Found while probing the above, and a different fault.

hclsyntax builds a `TemplateWrapExpr`, not a `TemplateExpr`, when a template
consists only of a single interpolation sequence. `HCLParser.Parse` had no case
for it, so it fell through to:

```go
default:
	return fmt.Errorf("missing hcl type found, found %s", expType)
```

`%s` on that struct produced a raw pointer dump with a broken verb in it:

```
missing hcl type found, found &{%!s(*hclsyntax.ScopeTraversalExpr=&{[{{} variable ...
```

The effect is that `"${variable.list_os[0]}"` is refused where
`variable.list_os[0]` resolves, a difference nothing in the schema asks for.
What the wrap holds is the whole value, so the wrapped expression is parsed in
its place.

## Tests

`internal/hclparser/user_input_test.go` walks the eight refusals, and asserts
the other way on `"a" * 2`, which must keep the line. The wrap case is there
too.

`provider/github/step/user_input_test.go` covers the second fault, and checks
the printed message as well as the mark, for the reason above.

Each of the three was neutered and the matching test watched to fail. The first
attempt at neutering the marking broke the build instead of the test, which
proves nothing; it was redone so the package still compiled.

## Not covered

`internal/hclparser` still returns bare `fmt.Errorf` in a few places that are
cinzel's own faults and correctly keep the line. The other internal packages
listed in the handoff were not looked at.
