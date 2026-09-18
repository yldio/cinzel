# cinzel

Converts between HCL and CI/CD YAML.

- Parse = HCL → YAML
- Unparse = YAML → HCL
- Providers: `provider/github`, `provider/gitlab`

---

# Where to read more

Load these only when the task is in that area.

- `docs/architecture/conversion.md` — expressions, document detection, output
  paths, parse defaults, the HCL schema rules, YAML output and key order,
  comment propagation
- `docs/architecture/commands.md` — `assist`, `pin`, `upgrade`
- `docs/solutions/` — notes on problems already solved here: the symptom, the
  cause, and why the fix is shaped the way it is. Read the matching one before
  working on the same area; several of these look trivial and are not. Start
  from `patterns/critical-patterns.md`

The solution notes are records of decisions, so their names can lag the code.
Where one has drifted, a blockquote after the frontmatter says what moved and
where the code is now. Annotate the same way rather than rewriting the
narrative.

---

# Workflow

For non-trivial tasks:

1. Search the repo, and `docs/solutions/` with it.
2. Identify provider and direction.
3. Plan briefly.
4. Implement.
5. Run tests. Fix until stable.

Delegate to a sub-agent only where the boundary is clear: repo research vs
implementation, review vs editing. Not for trivial work.

Use Context7 for external library docs.

---

# Change rules

- Keep changes minimal and localized
- Do not refactor unrelated code
- Do not change public interfaces unless required
- Prefer existing patterns
- Do not introduce unrelated formatting
- The HCL schema lives only in `provider/<name>/config.go`. Adding a field
  means structs, then conversion, then tests

---

# Testing

- stdlib `testing` only
- Golden comparison is semantic, not textual
- Roundtrip must stay stable
- Changing code means updating tests, and golden and roundtrip staying green

---

# Code style

- Every package has `doc.go`; every exported symbol has a doc comment starting
  with its name, attached directly
- Errors in `errors.go`, named `errCamelCase`
- One blank line between logical blocks, and before `return`, `if`, `for` when
  not first
- `switch`/`case` stays compact
- Match the surrounding style

---

# Commits

One intent per commit, reviewable in five minutes. Split when there are several
intents, a refactor mixed with a behaviour change, or unrelated areas touched.
Order: refactor, change, tests, cleanup.

---

# Pitfalls

- `parseHCLToWorkflows` returns 4 values — every error path returns all 4
- Unmarshal YAML once, then classify. Never twice on the same content
- Never hand-escape `$${{ }}`; the parser and `unparse_emit.go` own it
- Avoid `go test -v ./...` at root
