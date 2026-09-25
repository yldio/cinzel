---
title: "The empty top-level key sanitizes to itself, so the guard on it passed"
module: "GitLabProvider"
problem_type: "logic_error"
component: "unparse_pipeline"
severity: "medium"
root_cause: "logic_error"
symptoms:
  - "unparse exits 0 and writes HCL with a line reading `= \"v\"`"
  - "parsing that file back fails with 'An argument or block definition is required here.'"
  - "the warning says the key was passed through, and it was, namelessly"
tags:
  - "gitlab"
  - "unparse"
  - "identifiers"
  - "roundtrip"
created_date: "2026-09-22"
updated_date: "2026-09-22"
---

## Problem Description

A top-level key GitLab does not define, whose value is not a mapping, is passed
through as an HCL attribute. Its name has to be an identifier, so there is a
guard that refuses the ones that are not: `my weird key`, `a-b`, `x:y`. The
empty key walked past it.

`"": v` at the top level of a pipeline produced

```hcl
job "build" {
  script = ["make"]
}
= "v"
```

at exit 0, with only the usual passthrough warning. Reading that file back with
`cinzel gitlab parse` fails: `An argument or block definition is required
here.`

## Root Cause

The guard asked whether sanitizing changed the key:

```go
if naming.SanitizeIdentifier(key) != key {
	return nil, errKeyNotAnIdentifier(key)
}
```

which is a good test for every key with a bad character in it, because
sanitizing replaces that character with an underscore and the two strings
differ. `SanitizeIdentifier("")` returns `""`. The empty key is the one input
that is not an identifier and is its own sanitized form, so the comparison said
nothing was wrong with it.

`SanitizeIdentifier` returning the empty string is right, and its callers
elsewhere already treat it as a value to decide about rather than a name: the
template loop substitutes `"template"`, the job loop substitutes `"job"`, and
`jobRefID` refuses. Only this guard read it as an answer about identity.

## Solution Implemented

Ask about the empty key separately, before the comparison it cannot fail.

```go
if key == "" || naming.SanitizeIdentifier(key) != key {
	return nil, errKeyNotAnIdentifier(key)
}
```

The error already in place says the key cannot be written as an identifier,
which is true of the empty key too, so it is reused rather than joined by a
second one.

## Verification

`TestAnEmptyTopLevelKeyIsRefused` covers both values the branch can be reached
with, a scalar and a list. Reverting the `key == ""` clause alone still
compiles and fails both.

`TestANonIdentifierTopLevelKeyIsRefused` next to it is the control, and it
includes a key that does pass through. Both keep passing.

## Prevention Guidance

A guard written as "the sanitized form differs" is a guard with a fixpoint. Ask
what the sanitizer returns for input it can make nothing of, and whether that
return is itself valid input. Here it is the empty string, and the empty string
is not a name.
