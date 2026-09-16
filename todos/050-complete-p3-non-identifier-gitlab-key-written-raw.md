---
status: complete
priority: p3
issue_id: "050"
tags: [code-review, gitlab, unparse]
dependencies: []
---

# A top-level key that is not an identifier is written as-is

## Problem Statement

The passthrough branch writes the key unchanged, so `my weird key` becomes
`my weird key = "v"`, which is not valid HCL. Only reachable through the path
that already warns about unsupported keys.

## Findings

`provider/gitlab/unparse_pipeline.go:509`.

## Recommended Action

When `naming.SanitizeIdentifier` changes the key, fail rather than warn — the
file produced cannot be read back either way, and an error says so at the point
it happens.

## Acceptance Criteria

- [x] A non-identifier top-level key is refused with a clear message

## Technical Details

`provider/gitlab/unparse_pipeline.go` — the scalar branch of the passthrough
loop now refuses a key `naming.SanitizeIdentifier` would change:

```go
if naming.SanitizeIdentifier(key) != key {
	return nil, errKeyNotAnIdentifier(key)
}
```

`provider/gitlab/errors.go` — `errKeyNotAnIdentifier` added, so the import
became a block to take `fmt`.

`provider/gitlab/passthrough_key_test.go` — `TestANonIdentifierTopLevelKeyIsRefused`,
four cases: a key with spaces, one with a dash, one with a colon, and an
identifier that still passes through with only the warning.

The map-valued branch above this one does not share the hole. GitLab treats any
top-level mapping as a job, so `my weird map:` never reaches the passthrough
loop — it becomes `job "my_weird_map" { id = "my weird map" }`, a sanitized
label with the real name kept in `id`.

## Work Log

### 2026-09-16

Reproduced through the CLI. `/tmp/r50/.gitlab-ci.yml`:

```yaml
build:
  script:
    - make
my weird key: v
```

`go run . gitlab unparse --file ... --output-directory ...` printed only
`warning: unsupported top-level key 'my weird key' passed through`, exited 0,
and wrote `my weird key = "v"`. Reading that back fails: *The equals sign "="
indicates an argument definition, and must not be used when defining a block.*

With the guard the same run refuses with `top-level key "my weird key" is not a
valid HCL identifier, so it cannot be passed through` and exits 1.

Checked the map-valued branch while there, since it sits directly above. Its
output reparses with `An argument named "a" is not expected here` — arbitrary
passthrough keys landing in a job block whose schema does not accept them. That
is a separate defect, pre-existing, and not what this todo describes, so it was
left alone.

Proved the test fails without the fix by keeping `errKeyNotAnIdentifier` and
reverting only the guard: three of the four cases fail with `want an error, got
nil`. Full sweep clean — `go test -count=1 ./...`, `mise run lint`,
`mise run drift`, `mise run license-check`.
