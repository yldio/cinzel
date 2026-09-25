---
title: "Add, subtract and multiply truncated their operands"
module: "internal/hclparser"
problem_type: logic_error
component: "internal/hclparser/binaryopexpr.go"
severity: medium
root_cause: "each side was read with big.Float.Int64, which truncates, before the operation ran"
symptoms:
  - "1.5 + 2.5 comes back as 3"
  - "0.4 * 10 comes back as 0"
  - "0.5 + 0.5 comes back as 0, and is then refused as 'must be greater than zero'"
  - "divide is unaffected"
tags:
  - expressions
  - arithmetic
status: fixed
created_date: "2026-09-25"
updated_date: "2026-09-25"
---

# Add, subtract and multiply truncated their operands

## The symptom

Arithmetic in an HCL attribute gave the wrong answer whenever either side had a
fraction:

```hcl
timeout_minutes = 1.5 + 2.5
```

wrote `timeout-minutes: 4` only after the fix. Before it, the same line wrote
`3`. `0.4 * 10` wrote `0`. `3.7 - 0.7` wrote `3`, which is right by accident,
since both sides truncated the same way.

The worst shape reports an error the author cannot act on:

```hcl
timeout_minutes = 0.5 + 0.5
```

```
jobs.b.timeout-minutes: must be greater than zero, found 0
```

The arithmetic is correct and the message is about a value cinzel computed,
not one the author wrote. It points at the issue tracker.

## The cause

`BinaryOpExpr.Parse` read both sides with `big.Float.Int64` before operating:

```go
case hclsyntax.OpAdd:
	lVal, _ := lhs.AsBigFloat().Int64()
	rVal, _ := rhs.AsBigFloat().Int64()

	return cty.NumberIntVal(lVal + rVal), nil
```

`Int64` truncates, so `1.5 + 2.5` was computed as `1 + 2`. The same three lines
appear for subtract and multiply. Divide never had the fault: it reads both
sides with `Float64` instead, which is why `7 / 2` was always `3.5`.

The discarded second return value is the tell. `Int64` reports its accuracy
there, and every one of the six calls dropped it.

## The fix

The operation runs on the `big.Float` values themselves, and `numberVal` turns
the result back into a cty number:

```go
case hclsyntax.OpAdd:
	return numberVal(new(big.Float).Add(lhs.AsBigFloat(), rhs.AsBigFloat())), nil
```

`numberVal` keeps a whole result whole, returning `cty.NumberIntVal` when the
float is an exact integer. Without that branch an integer result would start
writing as `6.0` where every golden holds `6`, which is a second defect rather
than a fix for the first.

The similar-looking `Int64` in `scopetraversalexpr.go` is correct and was left
alone: it reads a list index, where truncation is the right answer and a
fractional index is refused before it.

## The test

`internal/hclparser/arithmetic_keeps_the_fraction_test.go`. Six fractional
cases, three integer cases and one divide. The six were confirmed to fail with
the `Int64` reads restored; the integer and divide cases pass either way, which
is what makes them controls rather than a restatement of the fix.
