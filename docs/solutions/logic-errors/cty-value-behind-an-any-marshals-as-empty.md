---
title: "A cty.Value reached through an any marshalled as {}"
module: "internal/yamlwriter"
problem_type: logic_error
component: internal/yamlwriter
severity: high
root_cause: "type check placed on the struct field rather than on the value"
symptoms:
  - "A step field written as {} in the YAML with the value gone"
  - "Only a field typed cty.Value survives; the same value in an any field, a map value or a slice element does not"
  - "No error: the file is written and the command exits 0"
tags:
  - cty
  - reflect
  - yaml
  - marshal
created_date: "2026-09-18"
updated_date: "2026-09-18"
---

## Problem

`convert` in `internal/yamlwriter/marshal.go` walks a value with `reflect` and
builds the plain `any` that `yaml.Marshal` writes. A `cty.Value` cannot go
through that walk: it is a struct whose fields are all unexported, so the walk
reads nothing off it and emits an empty map.

```yaml
direct: hello
any: {}
list:
- {}
map:
  k: {}
```

All four values above are the same `cty.StringVal("hello")`. Only the one in a
field declared `cty.Value` came out.

Nothing errors. The empty map is valid YAML, the file is written, the command
exits 0, and the value is simply missing from the workflow.

---

## Root Cause

The check that diverted a `cty.Value` to the `go-cty-yaml` serializer sat inside
the struct-field loop, on the field's static type:

```go
if field.Type() == reflect.TypeOf(cty.Value{}) {
    // ... ctyyaml.Marshal, then yaml.Unmarshal back to any
}
```

`field.Type()` is what the struct declares, not what the field holds. A field
declared `any` has type `interface{}` however a `cty.Value` is stored in it, so
the check was false and the value fell through to the generic struct walk. The
same applies to a `map[string]any` value and a `[]any` element, neither of which
is a struct field at all — they never reached the check.

There was a second reason the interface case could not work even if the check
had been moved: `convert` dereferenced `reflect.Pointer` but not
`reflect.Interface`, so an `any` arrived at the switch with kind `Interface` and
went straight to `default:`, which returns `val.Interface()` — the `cty.Value`
struct itself, handed to `yaml.Marshal` to render as `{}`.

---

## Solution

Two changes, both in `internal/yamlwriter/marshal.go`.

Unwrap an interface the way a pointer is already unwrapped, so the dynamic value
reaches the switch:

```go
for val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
	if val.IsNil() {
		return nil, nil
	}

	val = val.Elem()
}
```

A loop rather than two `if`s: an `any` holding a `*T` needs both, in either
order.

Then move the check off the field and onto the value, at the top of the struct
case, where every route in passes through:

```go
case reflect.Struct:
	if val.Type() == reflect.TypeOf(cty.Value{}) {
		return convertCty(val.Interface().(cty.Value))
	}
```

The body that was inline in the field loop becomes `convertCty`, which returns
`nil` for an unknown or null value — the same "omit it" signal the field loop
already acts on with `convertedValue != nil`.

```go
func convertCty(val cty.Value) (any, error) {
	if !val.IsKnown() || val.IsNull() {
		return nil, nil
	}

	yamlBytes, err := ctyyaml.Marshal(val)
	if err != nil {
		return nil, err
	}

	var out any

	if err := yaml.Unmarshal(yamlBytes, &out); err != nil {
		return nil, err
	}

	return out, nil
}
```

With the check on the value, the field loop no longer needs a special case and
collapses to the plain recursive call.

---

## Reachability

`step.Step` declares `Id`, `If`, `Name`, `Uses`, `Run`, `WorkingDirectory`,
`Shell`, `With`, `Env`, `ContinueOnError` and `TimeoutMinutes` as `cty.Value`
directly, so the only live caller — `stepsToMap` in
`provider/github/parse_workflow.go` — was hitting the field check and was never
affected. This was a latent bug, fixed because the next `any`-typed field or
`map[string]any` added to a marshalled struct would have silently dropped its
value with no error and no test catching it.

---

## Prevention

A type check on `field.Type()` answers "what does this struct declare", not
"what is in here". When the answer has to hold for a value arriving through an
`any`, a map or a slice, check `val.Type()` on the value itself, at a point the
recursion passes through for every route.

Watch for the pair: a `reflect` walk that dereferences `Pointer` but not
`Interface` will always answer the second question wrongly, because the dynamic
value never reaches the switch at all.

---

## Tests

`TestConvertCtyValueBehindAnAny` in `internal/yamlwriter/marshal_test.go`
marshals one `cty.StringVal("hello")` four ways — a `cty.Value` field, an `any`
field, a `map[string]any` value and a `[]any` element — and asserts no `{}`
appears in the output and all four read `hello`. Three of the four failed before
the fix.

---

## Related

- [`../patterns/critical-patterns.md`](../patterns/critical-patterns.md)
