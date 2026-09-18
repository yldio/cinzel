---
title: "Panics reaching the user: AsBigFloat and NumberFloatVal on values they do not accept"
problem_type: runtime-errors
component: internal/hclparser, provider/github
severity: "high"
root_cause: "unchecked_input"
symptoms:
  - "panic: not a number, with a stack trace, instead of an error naming the file"
  - "panic: Float.SetFloat64(NaN) on a workflow holding .nan"
  - "The command exits with a Go stack trace and no file, job or value named"
tags:
  - panic
  - cty
  - user-input
  - hclparser
  - conversion
affected_files:
  - internal/hclparser/binaryopexpr.go
  - provider/github/conversion.go
created_date: "2026-09-18"
updated_date: "2026-09-18"
---

## Problem Description

Two separate values a user can write in an ordinary input file took the whole
command down with a Go stack trace rather than an error naming the file.

```hcl
variable "v" { default = true + 1 }
```
```
panic: not a number
  github.com/zclconf/go-cty/cty.Value.AsBigFloat(...)
      .../go-cty@v1.18.0/cty/value_ops.go:1478
  github.com/yldio/cinzel/internal/hclparser.(*BinaryOpExpr).Parse(...)
```

```yaml
jobs:
  a:
    timeout-minutes: .nan
```
```
panic: Float.SetFloat64(NaN)
```

Neither names the file, the job or the value, so there is nothing to act on.
Both are reachable from a file a user wrote by hand.

## Root Cause

Two `cty` entry points signal a rejected value by panicking rather than by
returning an error, and both were called on a value that came straight from
user input.

| Call | Rejects | How it signals |
|---|---|---|
| `cty.Value.AsBigFloat()` | anything whose type is not `cty.Number` | `panic("not a number")` (go-cty v1.18.0, `cty/value_ops.go:1478`) |
| `cty.NumberFloatVal(f)` | `NaN` | `panic: Float.SetFloat64(NaN)` from `math/big` |

`BinaryOpExpr.Parse` reached `AsBigFloat` for every operator except `==` and
`!=` without checking either operand, so `true + 1`, `"a" * 2` and `null - 1`
all panicked. `anyToCtyDirect` passed any `float32`/`float64` to
`NumberFloatVal`, and `.nan` is a float that YAML is allowed to carry.

## Solution

Check before the call, in both places.

`binaryopexpr.go` — one guard covering every operator that reads its operands
as numbers:

```go
if boe.expression.Op != hclsyntax.OpEqual && boe.expression.Op != hclsyntax.OpNotEqual {
	if err := requireNumbers(lhs, rhs); err != nil {
		return cty.NilVal, err
	}
}
```

`requireNumbers` refuses a nil, a null, an unknown and any non-`Number` type,
naming which side was wrong. `==` and `!=` are left out because `RawEquals`
compares any two values without reading either as a number.

`conversion.go` — decline the NaN rather than converting it:

```go
func ctyFloat(v float64) (cty.Value, bool) {
	if math.IsNaN(v) {
		return cty.NilVal, false
	}

	return cty.NumberFloatVal(v), true
}
```

Returning `false` is what the direct path already means by "not convertible",
so the value falls through to `anyToCtyViaYAML`. That path hands the value to
`go-cty-yaml`, which returns `floating point NaN is not supported`
(`resolve.go:147`) rather than panicking, and cinzel's own wrapping adds the
file and the job.

## Prevention Guidance

- **Rule**: a `cty` constructor or accessor that can panic must not be reached
  from user input without a check first. The ones that bite here are
  `AsBigFloat` (non-number) and `NumberFloatVal` (NaN); `AsString` and
  `AsBigFloat` share the shape.
- Where the surrounding code already has a "cannot convert" path, decline into
  it rather than adding a new error. `ctyFloat` returns `false` and the existing
  YAML fallback produces the message.
- A panic reaching the user is a bug even when the input is nonsense. The test
  for one asserts an error, not a particular message, so the check is what is
  pinned rather than the wording: `TestBinaryOpExprRefusesNonNumberOperands`
  (`internal/hclparser/hclparser_test.go`) and `TestANaNIsReportedNotPanicked`
  (`provider/github/nan_value_test.go`).
