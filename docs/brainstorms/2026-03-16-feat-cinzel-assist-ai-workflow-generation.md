# Brainstorm: `cinzel assist` — AI-powered workflow generation

**Date**: 2026-03-16
**Status**: brainstorm

## Summary

Add an `assist` subcommand to cinzel that takes a natural language prompt and generates HCL workflow definitions. Works across all providers (GitHub, GitLab, etc.).

## Core approach: YAML-then-unparse

The LLM generates standard CI/CD YAML (which it already knows well), then cinzel's existing unparse engine converts it to HCL. This avoids teaching the LLM cinzel's custom HCL schema.

```
User prompt → LLM → provider YAML → cinzel unparse → HCL files
```

The unparse step acts as both converter and validator — if the YAML is invalid for the provider, unparse will reject it.

## CLI interface

Mirrors the existing parse/unparse pattern:

```sh
cinzel github assist --prompt "golang PR with tests and linting" --output-directory ./cinzel
cinzel gitlab assist --prompt "node.js pipeline with docker build" --output-directory ./cinzel
```

Flags:
- `--prompt` or positional argument — the natural language description
- `--output-directory` — where to write HCL files (default: `./cinzel`)
- `--provider` — AI provider: `anthropic` or `openai` (default: from config/env)
- `--model` — model override (default: from config/env)
- `--dry-run` — output to stdout instead of writing files

After generating HCL, the user runs their normal workflow:

```sh
cinzel github assist --prompt "golang PR workflow"   # → ./cinzel/*.hcl
cinzel github parse --file ./cinzel --output-directory .github/workflows  # → YAML
```

## AI provider support

Two providers at launch: Anthropic and OpenAI.

### Configuration

Two levels of YAML config (consistent with existing `.cinzelrc.yaml`):

**Project-level** (`.cinzelrc.yaml`, committed to git):

```yaml
ai:
  default: anthropic
  providers:
    anthropic:
      model: claude-sonnet-4-5-20250514
    openai:
      model: gpt-4o
```

Non-sensitive settings only. Sets the team's preferred default and models per provider.

**User-level** (`os.UserConfigDir()/cinzel/config.yaml`, never committed):

```yaml
ai:
  default: anthropic
  providers:
    anthropic:
      model: claude-sonnet-4-5-20250514
      api_key: sk-ant-...
    openai:
      model: gpt-4o
      api_key: sk-...
```

Holds API keys and personal overrides. Both providers configured simultaneously — switch between them with `--provider` flag or by changing `default`. Stored as plaintext on the filesystem (same pattern as `~/.config/gh/hosts.yml`, `~/.docker/config.json`).

**Environment variables** (highest precedence, for CI/scripted use):

```sh
export CINZEL_AI_DEFAULT=anthropic       # which provider to use
export ANTHROPIC_API_KEY=sk-ant-...      # provider-specific keys
export OPENAI_API_KEY=sk-...
```

**CLI flag** (highest precedence, per-invocation):

```sh
cinzel github assist --prompt "..."                    # uses default AI provider
cinzel github assist --prompt "..." --ai openai        # override AI provider
cinzel github assist --prompt "..." --ai anthropic --model claude-opus-4-20250514  # override both
```

**Resolution order** (highest wins):

```
CLI flags → env vars → user config.yaml → project .cinzelrc.yaml
```

All config is YAML — no extra parsing layer beyond what cinzel already uses. Multiple providers can be configured simultaneously; `default` selects which one is used when no flag is passed.

### Provider interface

```go
// internal/ai/provider.go
type Provider interface {
    Generate(ctx context.Context, req GenerateRequest) (string, error)
}

type GenerateRequest struct {
    SystemPrompt string
    UserPrompt   string
    Model        string
}
```

Implementations:
- `internal/ai/anthropic.go` — uses `github.com/anthropics/anthropic-sdk-go`
- `internal/ai/openai.go` — uses `github.com/sashabaranov/go-openai`

OpenAI-compatible API also covers local models (Ollama, LM Studio) via custom endpoint, but that's not a launch goal.

## System prompt design

The system prompt is provider-specific (github vs gitlab) and instructs the LLM to:

1. Generate valid CI/CD YAML for the target provider
2. Use SHA-pinned action references where possible
3. Follow security best practices (least-privilege permissions, no hardcoded secrets)
4. Output only YAML, no markdown wrapping

Example system prompt skeleton:

```
You are a CI/CD workflow generator for {provider}.
Generate valid {provider format} YAML based on the user's description.

Rules:
- Output only valid YAML, no markdown code fences
- Use SHA-pinned action versions (not tags)
- Set minimum required permissions
- Use environment variables for secrets, never hardcode
- Include descriptive step names

{provider-specific schema reference if needed}
```

The system prompt can include a curated example YAML for the provider to anchor the output format.

## Flow detail

```
1.  User runs: cinzel github assist --prompt "..."
2.  cinzel reads existing ./cinzel/*.hcl, redacts sensitive patterns
3.  cinzel selects AI provider from config/env
4.  cinzel builds prompt: system prompt + redacted HCL context + user prompt
5.  Show cost confirmation (or skip with --acknowledge)
6.  LLM generates provider YAML (show "Generating workflow..." spinner)
7.  Strip markdown fences if present
8.  cinzel runs unparse (YAML → HCL) internally
9.  If unparse fails (retry ≤ 2): feed error back to LLM, go to step 6
10. If unparse still fails: show error + raw YAML, suggest refining the prompt
11. Pin action tags to SHAs via GitHub API
12. Restore redacted values in output
13. Write HCL to ./cinzel/assist/
```

## Provider-agnostic design

The `assist` subcommand is registered per provider, but the AI layer is shared:

```
cinzel github assist → github system prompt → LLM → github YAML → github.Unparse()
cinzel gitlab assist → gitlab system prompt → LLM → gitlab YAML → gitlab.Unparse()
```

Each provider supplies:
- A system prompt tailored to its YAML format
- Its existing `Unparse()` function for validation and conversion

The AI infrastructure (`internal/ai/`) is completely provider-agnostic.

## Decisions made

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Generation strategy | YAML-then-unparse | LLMs know CI/CD YAML well; unparse validates and converts |
| AI providers | Anthropic + OpenAI | Covers majority of users |
| Subcommand name | `assist` | Neutral, works across providers |
| Output location | `./cinzel/assist/` | Non-destructive, user reviews before merging |
| Templates | No | cinzel is a tool, not a template library; LLM is the template engine |
| Config | Env vars + optional TOML | Simple, secure (no keys in files) |
| Privacy | Redact secrets/org names before sending | Never leak sensitive data to external APIs |
| Action versions | LLM outputs tags, `cinzel pin` resolves SHAs post-generation | LLMs don't know current SHAs; post-processing is reliable |
| Validation retry | Up to 2 retries with error feedback | Balances success rate vs cost |
| Testing | No unit tests for LLM output | Non-deterministic; validate via unparse success |
| Context injection | Redacted existing HCL (steps + variables only) | Gives LLM structural awareness without leaking data |

## Starter workflow grounding

The system prompt instructs the LLM to base its output on official starter workflows when relevant. LLMs already know these templates from training data — no need for cinzel to fetch them at runtime. The one gap (stale action versions) is solved by `cinzel pin` in post-processing.

System prompt includes:

```
When relevant, base your output on official starter workflows:
- GitHub: github.com/actions/starter-workflows
- GitLab: gitlab.com/gitlab-org/gitlab/-/tree/master/lib/gitlab/ci/templates
Use current best practices and action versions from these sources.
```

### Future enhancement: live template fetching

If LLM output quality proves insufficient without real templates, cinzel could fetch them at runtime:

| Provider | Repository |
|----------|-----------|
| GitHub | [actions/starter-workflows](https://github.com/actions/starter-workflows) |
| GitLab | [gitlab-org/gitlab/.../ci/templates](https://gitlab.com/gitlab-org/gitlab/-/tree/master/lib/gitlab/ci/templates) |

This would involve keyword extraction from the prompt, fetching matching templates via API, caching under `os.UserCacheDir()/cinzel/templates/` (OS-agnostic), and injecting them into the prompt. Deferred because it adds HTTP client, caching, keyword matching, and coupling to external repo structures — significant complexity for marginal gain over what the LLM already knows.

## Post-processing pipeline

After the LLM generates YAML and unparse converts to HCL, two post-processing steps run:

### 1. Action version pinning (`cinzel pin`)

The LLM outputs tag-based versions (`actions/checkout@v4`) because it doesn't know current SHAs. cinzel resolves them via GitHub API:

```
1. Walk HCL AST, find all `uses` blocks
2. For each action@tag:
   GET /repos/{owner}/{action}/git/ref/tags/{tag} → SHA
3. Replace: version = "de0fac2e4500dabe..."
4. Add comment: // actions/checkout v4
```

`cinzel pin` is also a standalone command — useful for any HCL file:

```sh
cinzel github pin --file ./cinzel
```

### 2. Redaction restoration

If context injection was used, restore redacted values in the output (see Privacy section below).

### Full assist pipeline

```
existing HCL → redact → build prompt (system prompt + redacted context + user prompt) → LLM → YAML → strip fences → unparse → HCL → pin SHAs → restore redactions → write to ./cinzel/assist/
```

## Privacy: context redaction

Existing HCL files are valuable context but contain sensitive patterns. cinzel redacts before sending and restores after:

**Redaction map (built automatically):**

| Original | Sent to LLM |
|----------|-------------|
| `secrets.RELEASE_APP_ID` | `secrets.SECRET_1` |
| `secrets.RELEASE_PRIVATE_KEY` | `secrets.SECRET_2` |
| `yldio/cinzel` | `org/repo` |
| `homebrew-cinzel` | `secondary-repo` |
| `hello@yld.io` | `team@example.com` |

**What gets redacted:**
- `secrets.*` references → numbered placeholders
- GitHub org/repo names → generic placeholders
- Email addresses → example.com
- URLs containing org names → sanitized

**What stays visible** (the LLM needs this):
- Block structure (workflow, job, step shapes)
- Action references (`actions/checkout`, `jdx/mise-action`)
- Step names and ordering
- Attribute keys and types
- Control flow (`if`, `matrix`, `timeout`)

The LLM sees the structural skeleton without any identifying information.

## Output: non-destructive assist folder

`assist` writes to `./cinzel/assist/` by default, never touching existing files:

```
cinzel/
  workflows.hcl    # existing, untouched
  jobs.hcl         # existing, untouched
  steps.hcl        # existing, untouched
  assist/          # AI-generated output
    pr-workflow.hcl
```

The user reviews the generated HCL, then manually moves or merges into main files. This is non-destructive — running `assist` multiple times overwrites only the `assist/` folder.

## Resolved questions

1. **Context injection** — Yes. `assist` reads existing `./cinzel/*.hcl` files and includes them in the LLM prompt. This lets the LLM reference existing steps (e.g., `step.checkout`, `step.mise_setup`) instead of duplicating them, and avoids conflicting workflow/job names.

2. **Validation feedback loop** — Yes. If unparse fails, cinzel feeds the error back to the LLM for a retry. Cap at 2 retries to avoid runaway costs. Flow: generate → unparse → fail → retry with error context → unparse → succeed or abort.

3. **Multiple files** — Optimize LLM calls. If the prompt implies multiple workflows, generate them in a single LLM call (the YAML can contain multiple documents separated by `---`). Only split into separate calls if the single-call output fails validation.

## Resolved: remaining questions

4. **Refinement flow** — Both `--prompt` and `--refine` supported:
   ```sh
   cinzel github assist --prompt "golang PR with tests"      # fresh generation
   cinzel github assist --refine "add slack on failure"      # iterates on previous
   ```
   `--prompt` starts fresh (injects existing `./cinzel/*.hcl` as context).
   `--refine` builds on previous output (injects both `./cinzel/*.hcl` AND `./cinzel/assist/*.hcl` as context). The LLM sees the full picture and modifies accordingly. `--refine` without a previous `assist/` output errors with "nothing to refine".

5. **Cost transparency** — First run shows a confirmation: "This will call {provider} ({model}). API usage will incur costs. Continue? [y/N]". The `--acknowledge` flag bypasses this confirmation for CI or scripted use:
   ```sh
   cinzel github assist --prompt "..." --acknowledge
   ```

6. **Streaming** — No streaming. Show a spinner/working message ("Generating workflow...") while waiting for the LLM response. Keep the UX simple.

## New command: `cinzel pin`

Standalone command to resolve action tags to SHAs. Works independently of `assist`:

```sh
cinzel github pin --file ./cinzel          # pin all actions in HCL files
cinzel github pin --file ./cinzel/assist   # pin only assist output
```

This is useful for any HCL file, not just AI-generated ones.

## Not in scope

- Hosting an AI API (users bring their own keys)
- Local model support at launch (Ollama could work via OpenAI-compatible endpoint later)
- Template library
- Interactive/conversational mode (can be added later)
- Testing LLM output quality (non-deterministic; unparse validates structure)
