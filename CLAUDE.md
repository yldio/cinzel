# cinzel

Converts between HCL and CI/CD YAML.

- Parse = HCL → YAML
- Unparse = YAML → HCL
- Providers: `provider/github`, `provider/gitlab`

---

# Prior solutions

`docs/solutions/` holds notes on problems already solved here: the symptom, the
cause, and why the fix is shaped the way it is. Read the matching one before
working on the same area — several of these problems look trivial and are not.

Directories by problem kind: `logic-errors/`, `runtime-errors/`,
`build-errors/`, `test-failures/`, `integration-issues/`, `best-practices/`,
`developer-experience/`, `documentation-gaps/`, `patterns/`. Start from
`patterns/critical-patterns.md` — determinism, expression escaping, single
unmarshal, quote style, return arity — which is the short version of the rest.

The notes are records of decisions, so their names can lag the code. Where one
has drifted, a blockquote after the frontmatter says what moved and where the
code is now. Annotate the same way rather than rewriting the narrative.

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

---

# Schema contracts (critical)

- HCL schema lives only in `provider/<name>/config.go`
- No ad-hoc key maps; no schema duplicated in validation
- `hcl:",remain"` only for intentional pass-through
- YAML validation: strict typed decode (`goccy/go-yaml`), no allowlists

Adding a field: structs → conversion → tests.

---

# Conversion rules

- `$${{ }}` ↔ `${{ }}`, handled by the parser and `unparse_emit.go`. Never
  escape by hand
- Detect on unparse: `on` + `jobs` → workflow; `name` + `runs` → action; else
  step-only
- Output: actions → `<dir>/<name>/action.yml`, workflows → `<dir>/<name>.yaml`

Defaults on parse:

- a workflow with no permissions gets `permissions: {}`
- a step with no `id` gets one from its block label, unless `ignore_id` is set

---

# YAML output

Built through `internal/yamldoc`, which carries key order, comments and the
empty-map distinction in the document rather than recovering them from encoded
bytes. Double quotes only — single quotes break golden tests.

What gets quoted is `needsQuoting` (`internal/yamldoc/encode.go`): empty,
bool/null words in any case, numbers, leading or trailing whitespace, YAML
special characters. `@` is not quoted. GitLab keeps its own
`stringNeedsQuoting` (`provider/gitlab/pipeline_yaml.go`) on purpose.

Workflow key order is `workflowKeyOrder` (`provider/github/workflow_yaml.go`):
name, run-name, on, permissions, env, defaults, concurrency, jobs, then the
rest sorted. `jobs` goes last, the way a hand-written workflow reads: the short
top-level keys first, then the long tail.

---

# AI assist (`cinzel <provider> assist`)

- Pipeline: prompt → LLM → YAML → strip fences → split docs → temp files →
  Unparse → merge/dedup HCL → session folder
- Output: `cinzel/assist/{timestamp}/assist.hcl`, one session per prompt
- `--refine` targets the latest session, `--from {timestamp}` a specific one
- Blocks identical to existing `cinzel/*.hcl` become `// reuses:` comments;
  different blocks with the same signature get `// note:`
- Actions are pinned to SHAs after generation
- Privacy: `StripHCLContext` replaces every string value with `"..."`
- Config: `cinzel init` writes `os.UserConfigDir()/cinzel/config.yaml`. It holds
  no API key; keys come from the environment
- Resolution order: CLI flags > env vars > config file > defaults

# Version management (`cinzel github pin/upgrade`)

- `pin`: action tags → SHAs via the GitHub API, cached 24h. The tag stays on
  the line as `version = "<sha>" # <tag>`
- `upgrade`: finds the latest release, compares by tag or SHA, updates both
- No token needed for public actions; `GITHUB_TOKEN` raises the rate limit

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
- Avoid `go test -v ./...` at root
