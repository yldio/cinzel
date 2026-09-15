---
status: complete
priority: p2
issue_id: "016"
tags: [code-review, correctness]
dependencies: ["015"]
---

# An unparsed action comes back under a name nobody chose

## Problem Statement

Unparse named its output after the file it read. Every action file is called
`action.yml`, so the name it carried was thrown away: the action came back
called "action", and a second one in the same run wrote over the first.

## Findings

Probed at ce8c0f9. Two actions, in their own directories:

```
in/alpha/action.yml
in/beta/action.yml
```

Unparsed together, one file is left, RC=0:

```
out/action.hcl
  action "action" { filename = "action" ... name = "Beta" }
```

Alpha is gone and beta is called something else. A single action does not
round-trip either: `alpha/action.yml` unparsed and parsed again lands at
`action/action.yml`, so a directory the caller wrote by hand is replaced by
one cinzel picked.

The same probe on workflows in separate directories collides too. `x/ci.yaml`
and `y/ci.yaml` both write `out/ci.hcl`. That is a different question, since
a workflow really is named by its file and two of them sharing a name have
no other identity to fall back on. Recorded, not fixed here.

## Recommended Action

Take the name from the directory when the file is one of the fixed names
GitHub reads, `action.yml` or `action.yaml`. That is where an action's
identity lives. Anything else keeps the basename, which is all there is to
go on.

The name has to reach the output path and the HCL together, so
`unparseYAMLFile` returns it rather than the caller deriving it twice.

## Technical Details

- `provider/github/github.go`, `Unparse`, `unparseYAMLFile`, `actionNameFor`

Three existing round-trip tests asserted `action.hcl` and `parse2/action/`,
which is the old behaviour written down. They now assert the fixture name in
both places, which is what a stable round-trip means.

## Acceptance Criteria

- [x] An action keeps its directory name through unparse
- [x] Two actions unparse to two files
- [x] An action round-trips back to the directory it came from
- [x] A workflow, and an action under some other filename, keep their basename

## Work Log

- 2026-09-15: Created and closed. Found by probing the unparse direction
  after 015, which is where the action output path was last in question.
