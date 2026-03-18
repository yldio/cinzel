# Project

`cinzel` converts between HCL and CI/CD YAML.

- Parse = HCL → YAML
- Unparse = YAML → HCL

---

# Workflow

For non-trivial tasks:

1. Search the repo.
2. Identify provider and direction.
3. Plan briefly.
4. Implement minimal changes.
5. Run tests.
6. Fix until stable.

Keep changes minimal and localized.

---

# Routing (critical)

Do not rely on implicit workflow selection for important tasks.

Use explicit skills when available:

- planning → `$ce:plan`
- plan review → `$ce:review-plan`
- debugging → `$ce:debug`
- code review → `$ce:review`
- verification → `$ce:verify`

If a matching skill exists, use it instead of recreating the workflow.

---

# Sub-agents

Use sub-agents for non-trivial tasks when a clear boundary exists.

Good delegation:

- repo research vs implementation
- plan review vs editing
- verification vs coding

Do not use sub-agents for trivial work.

---

# Tools

Prefer MCP tools when available.

Use:

- Context7 → external docs
- `ctx_search` → repo search
- `ctx_batch_execute` → multiple commands
- `ctx_execute` → dependent commands

Use native tools only if MCP tools are unavailable or failed.

Summarize outputs. Do not return raw logs.

---

# Schema contracts (critical)

- HCL schema lives only in `provider/<name>/config.go`
- Do not use ad-hoc key maps
- Do not duplicate schema in validation

- `hcl:",remain"` only for intentional pass-through

YAML validation:

- use strict typed decode (`goccy/go-yaml`)
- do not use allowlists

When adding fields:

1. update structs
2. update conversion
3. update tests

---

# Change rules

- Keep changes minimal and localized
- Do not refactor unrelated code
- Do not change public interfaces unless required
- Prefer existing patterns
- Do not introduce unrelated formatting

---

# Conversion rules

- `$${{ }}` → `${{ }}`
- Detect on unparse:
  - workflow: `on` + `jobs`
  - action: `name` + `runs`
  - else: step-only

- Output:
  - actions → `<dir>/<name>/action.yml`
  - workflows → `<dir>/<name>.yaml`

---

# YAML output

- Use `yaml.v3` node API
- Use double quotes when required
- Do not rely on single quotes

Quote when needed for:

- empty
- bool/null
- numbers
- YAML special chars

Do not quote `@`

Key order:

1. name
2. run-name
3. on
4. jobs
5. rest sorted

---

# AI assist (`cinzel <provider> assist`)

- **Pipeline**: prompt → LLM → YAML → strip fences → split docs → temp files → Unparse → merge/dedup HCL → session folder
- Output: `cinzel/assist/{timestamp}/assist.hcl` — each prompt gets its own session
- `--refine` targets latest session by default, or `--from {timestamp}` for a specific one
- Blocks identical to existing `cinzel/*.hcl` replaced with `// reuses:` comments
- Different blocks with same signature get `// note:` comments
- Auto-pins GitHub actions to SHAs after generation
- Privacy: `StripHCLContext` replaces all string values with `"..."` via HCL AST walk
- Config: `cinzel init` creates `os.UserConfigDir()/cinzel/config.yaml` with AI provider defaults + API keys
- Resolution order: CLI flags > env vars > config file > hardcoded defaults

# Version management (`cinzel github pin/upgrade`)

- `pin`: resolves action tags → SHAs via GitHub API. Cached 24h. Adds `// action tag` comments
- `upgrade`: finds latest release, compares by tag or SHA, updates version + comment
- No token required for public actions. `GITHUB_TOKEN` for higher rate limits

---

# Testing

- Golden: semantic YAML comparison
- Roundtrip must remain stable

When changing code:

- update tests
- ensure roundtrip passes
- ensure golden passes

---

# Code style

- Every package has `doc.go`
- Every exported symbol has doc comment (starts with name)
- Comments directly attached

- Errors in `errors.go`, `errCamelCase`
- Use stdlib `testing` only

Formatting:

- one blank line between logical blocks
- blank line before `return`, `if`, `for` (if not first)
- keep `switch/case` compact

- match surrounding style
- no unrelated formatting

---

# Commits

- one intent per commit
- reviewable in ≤5 minutes

Split if:

- multiple intents
- refactor mixed with behavior change
- unrelated areas touched

Preferred order:

1. refactor
2. change
3. tests
4. cleanup

---

# Pitfalls

- `parseHCLToWorkflows` returns 4 values — always return all
- do not unmarshal YAML twice
- avoid `go test -v ./...` at root
