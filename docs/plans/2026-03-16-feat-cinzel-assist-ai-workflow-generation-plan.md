---
title: "feat: cinzel assist — AI-powered workflow generation"
type: feat
status: implemented
date: 2026-03-16
origin: docs/brainstorms/2026-03-16-feat-cinzel-assist-ai-workflow-generation.md
---

# cinzel assist — AI-powered workflow generation

## Overview

Add an `assist` subcommand to cinzel that takes a natural language prompt and generates HCL workflow definitions. The LLM generates standard CI/CD YAML, then cinzel's existing `Unparse` engine converts to HCL and validates the output. No template library — the LLM is the template engine; cinzel stays a pure conversion tool with an AI entry point.

Works across all providers (`cinzel github assist`, `cinzel gitlab assist`).

## Problem Statement / Motivation

Writing CI/CD workflows from scratch requires deep knowledge of provider-specific YAML syntax, available actions, and best practices. cinzel already handles HCL-YAML conversion — adding an AI entry point lets users describe what they want in natural language and get valid, idiomatic HCL output.

## Proposed Solution

**Current pipeline (implemented):**

```
existing HCL → strip string literals → build prompt → LLM → YAML → strip fences → split YAML docs → unparse each → merge/dedup HCL blocks → single timestamped file → ./cinzel/assist/
```

**Future pipeline (deferred):**

```
... → (optional) pin SHAs → (optional) retry with error feedback → ./cinzel/assist/
```

## Technical Approach

### Architecture

`assist` lives in the **command layer** — it does NOT modify the `Provider` interface. It orchestrates: call AI → write YAML to temp file → call provider's existing `Unparse` → write HCL output. This avoids interface changes and uses the proven unparse path.

**Why temp files instead of `UnparseBytes`**: Adding `UnparseBytes` to the `Provider` interface forces every provider to implement it (breaking change). Writing to a temp file and calling the existing `Unparse(ProviderOps)` is ~5 lines, requires zero interface changes, and the overhead is negligible for a command that already waits 10+ seconds for an LLM response. (Architecture review finding #1)

### Implementation Phases

#### Phase 1: End-to-end assist (Anthropic only) — COMPLETE

**Goal**: `cinzel github assist --prompt "..."` generates HCL files. Minimal viable pipeline.

**Scope decisions** (from simplicity review):
- Anthropic only — defer OpenAI and provider interface until demand exists
- Env vars + CLI flags only — defer config files
- No context injection, no redaction — defer to Phase 2
- No retry loop — show error + raw YAML, user refines prompt
- Temp file approach — no `Provider` interface changes
- Inline fence stripping and prompt constant — no separate files

**Tasks**:
- Create `internal/ai/doc.go` (package doc comment)
- Create `internal/ai/anthropic.go`:
  - Direct Anthropic SDK client (no provider interface abstraction yet)
  - `Generate(ctx context.Context, systemPrompt, userPrompt, model string) (string, error)`
  - Read `ANTHROPIC_API_KEY` from env var
  - 120-second timeout via `context.WithTimeout`
  - Spinner ("Generating workflow...") while waiting
- Register `assist` subcommand in `internal/command/command.go:addProvider()`:
  - Flags: `--prompt`, `--output-directory` (default `./cinzel/assist/`), `--acknowledge`, `--dry-run`
  - Cost confirmation: "This will call Anthropic (claude-sonnet-4-5-20250514). Continue? [y/N]"
  - `--acknowledge` bypasses confirmation
- Build system prompt inline (constant string per CI provider):
  - Instructs LLM to output only valid YAML, no markdown fences
  - References official starter workflows (GitHub/GitLab)
  - Requests step names, minimum permissions, no hardcoded secrets
- Strip markdown fences from LLM output (inline function, ~20 lines):
  - Handle ` ```yaml `, ` ```yml `, bare ` ``` `
  - Handle multiple YAML documents (split on `---`)
- For multi-document YAML: split on `---`, write each document to a separate temp file, call `Unparse` in a loop
- Temp files must use `.yaml` extension (unparse uses extension for format detection)
- Write YAML to temp file → call `provider.Unparse(ProviderOps{File: tempFile, OutputDirectory: outputDir})`
- Clean up temp files via `defer os.Remove()`
- If unparse fails: show error + raw YAML, suggest refining prompt
- If LLM returns empty/whitespace response: "LLM returned empty response. Try a more specific prompt."
- `--dry-run`: still calls the LLM (costs money) but skips file writing, prints HCL to stdout. The cost confirmation prompt already covers this.
- Output to `--output-directory` (default `./cinzel/assist/`), create dir if needed

**Acceptance criteria**:
- [ ] `cinzel github assist --prompt "golang PR with tests"` generates HCL in `./cinzel/assist/`
- [ ] `cinzel github assist --prompt "..." --output-directory ./custom` respects flag
- [ ] `cinzel github assist --prompt "..." --dry-run` prints HCL to stdout
- [ ] Markdown fences stripped reliably from LLM output
- [ ] `ANTHROPIC_API_KEY` not set → clear error with setup instructions
- [ ] Cost confirmation shown; `--acknowledge` bypasses
- [ ] Spinner displayed during LLM call
- [ ] Unparse failure → error + raw YAML shown (no retry in v1)
- [ ] API timeout at 120 seconds → clear timeout error
- [ ] Multi-document YAML (multiple workflows) generates multiple HCL files

**Error handling**:
- Auth failure → "Invalid API key. Set ANTHROPIC_API_KEY env var."
- Rate limit → "API rate limited. Try again in a moment."
- Timeout → "LLM request timed out after 120s. Try a simpler prompt."
- Unparse failure → "Generated YAML could not be converted to HCL:\n{error}\n\nRaw YAML:\n{yaml}\n\nTry refining your prompt."
- Raw YAML error output must NOT contain any user HCL content (no context injection in v1, so this is inherently safe)

**Files (actual)**:
- `internal/ai/doc.go` (NEW)
- `internal/ai/anthropic.go` (NEW)
- `internal/ai/provider.go` (NEW — interface, GenerateResponse, StripFences, SystemPrompt, resolveAPIKey)
- `internal/ai/openai.go` (NEW)
- `internal/ai/errors.go` (NEW — consolidated sentinel errors)
- `internal/ai/strip.go` (NEW — HCL string stripping for context injection)
- `internal/ai/strip_test.go` (NEW — 12 tests including real fixture)
- `internal/ai/provider_test.go` (NEW — StripFences tests)
- `internal/command/assist.go` (NEW — extracted assist logic, ~280 lines)
- `internal/command/assist_test.go` (NEW — splitYAMLDocuments tests)
- `internal/command/errors.go` (NEW — errCancelled, errPromptRequired)
- `internal/command/command.go` (MODIFY — slimmed down, assist extracted)
- `go.mod` (MODIFY — anthropic-sdk-go + openai-go)

**File count**: 11 new + 2 modified.

#### Phase 2: Context injection + OpenAI — COMPLETE (config files deferred)

**Goal**: Existing HCL context improves output quality. OpenAI as second provider. Config file support.

**Prerequisite**: Phase 1 shipped and validated with real usage.

**Tasks**:

**AI provider interface + OpenAI**:
- Extract `internal/ai/provider.go` interface from the concrete Anthropic client:
  ```go
  type Provider interface {
      Generate(ctx context.Context, req GenerateRequest) (string, error)
  }
  ```
- Create `internal/ai/openai.go` using `github.com/openai/openai-go`
- Add `--provider` and `--model` CLI flags
- Consider build tags (`//go:build ai`) to make AI deps optional — now two SDK deps

**Config file support**:
- Create `internal/ai/config.go`:
  - Load `ai:` section from `.cinzelrc.yaml` (project-level, non-sensitive only)
  - Load `os.UserConfigDir()/cinzel/config.yaml` (user-level, API keys)
  - Create user config with `0600` permissions (security review H1)
  - Warn if permissions are more permissive than `0600`
  - Hard error if `.cinzelrc.yaml` contains `*_API_KEY` or `*_SECRET` patterns (security review H1)
  - Resolution: CLI flags > env vars > user config > project config

**Context injection (string-stripped)**:
- Strip all string literal values from existing HCL before injecting as context. The LLM needs block structure (step/job/workflow shapes, attribute names, nesting) — not actual values. This eliminates secrets, org names, tokens, emails, and all other sensitive data without any redaction/restoration engine.
  - Replace all `= "..."` values with `= "..."` (literal ellipsis)
  - Replace heredoc content with `...` (use HCL AST traversal, not regex — heredocs have complex syntax)
  - Strip all HCL comments (may contain internal URLs, team names, proprietary notes)
  - Preserve: block types, block labels, attribute names, block nesting
  - Block labels (e.g., `step "checkout_release_with_credentials"`) are preserved — accepted residual risk. They reveal naming conventions but not secrets; comparable to function names in public code.
  - ~30 lines of code using `hclsyntax` AST walk, no bidirectional mapping, no restoration step
- Stripping must have unit tests:
  - String literal values replaced with `"..."`
  - Heredoc content replaced with `...`
  - Comments stripped entirely
  - Block labels preserved (`step "checkout"` stays)
  - Attribute names preserved (`name`, `value`, `action`)
  - Block nesting preserved
  - Real-world fixture: strip `cinzel/steps.hcl` and verify no secret patterns, org names, or emails survive
  - Golden test: stripped output compared against `.golden` file for stability
- No restoration needed — the LLM output is fresh YAML that goes through unparse. Nothing from the stripped context needs to be injected back.
- Context injection:
  - Read `./cinzel/*.hcl` (steps + variables primarily)
  - Cap context at 8000 tokens to prevent oversized payloads (security review H3)
  - Warn if context truncated
  - `--no-context` flag to skip injection entirely
- Move system prompt to `internal/ai/prompt.go`:
  - Build per CI provider (github/gitlab)
  - Include stripped HCL context

**`--refine` flag**:
- Reads `./cinzel/assist/*.hcl` as additional context alongside `./cinzel/*.hcl`
- Can coexist with `--prompt` (architecture review #5): `--refine` adds previous assist output as context, `--prompt` provides the new instruction
- Without `--prompt`: error "provide a prompt describing the refinement"
- Without previous `assist/` output: error "nothing to refine — run assist --prompt first"

**Acceptance criteria**:
- [ ] `--provider openai` works with `OPENAI_API_KEY`
- [ ] Multiple providers configured in user config with `default:` selector
- [ ] User config created with `0600` permissions
- [ ] API key in `.cinzelrc.yaml` → hard error
- [ ] Existing steps referenced by LLM instead of duplicated
- [ ] String literals stripped from HCL context (no sensitive data sent)
- [ ] `--no-context` skips injection
- [ ] Context capped at 8000 tokens with truncation warning
- [ ] `--refine "add caching"` includes previous assist output as context

**Files**:
- `internal/ai/provider.go` (NEW)
- `internal/ai/openai.go` (NEW)
- `internal/ai/config.go` (NEW)
- `internal/ai/prompt.go` (NEW — extracted from inline)
- `internal/command/command.go` (MODIFY)
- `internal/command/config.go` (MODIFY — accept ai: section)
- `go.mod` (MODIFY — add `github.com/openai/openai-go`)

#### Phase 3: `cinzel pin` (standalone command) — DEFERRED

**Goal**: Resolve action tags to SHAs. Standalone command, independently useful.

**Prerequisite**: Shipped separately from assist. Does not depend on Phase 2.

**Tasks**:
- Create `internal/pin/doc.go` and `internal/pin/pin.go`:
  - Define `Resolver` interface for testable GitHub API calls (architecture review #3):
    ```go
    type Resolver interface {
        ResolveTag(ctx context.Context, owner, repo, tag string) (string, error)
    }
    ```
  - HTTP implementation: `GET /repos/{owner}/{action}/git/ref/tags/{tag}` → SHA
  - Support `GITHUB_TOKEN` env var for authenticated requests (5000/hr vs 60/hr) (security review M4)
  - Cache resolved SHAs in `os.UserCacheDir()/cinzel/pins/` with 24h TTL
  - Walk HCL AST, find `uses` blocks with `action` + `version` attributes
  - Replace `version` value, add comment `// {action} {tag}`
  - Failures: fall back to unpinned tag with warning (private actions, nonexistent tags, rate limits)
- Register `cinzel <provider> pin --file <path>` command
- Optionally integrate into assist pipeline (runs after unparse, before write) — opt-in via `--pin` flag on assist

**Acceptance criteria**:
- [ ] `cinzel github pin --file ./cinzel` resolves all action tags to SHAs
- [ ] `GITHUB_TOKEN` used for authenticated requests when available
- [ ] Pin failures → warning + fallback to tag (not fatal)
- [ ] Resolved SHAs cached for 24h
- [ ] `cinzel github assist --prompt "..." --pin` pins after generation

**Files**:
- `internal/pin/doc.go` (NEW)
- `internal/pin/pin.go` (NEW)
- `internal/command/command.go` (MODIFY — register pin command)

#### Phase 4: Retry loop + hardening — DEFERRED

**Goal**: Improve success rate for complex prompts. Ship only if failure rates warrant it.

**Prerequisite**: Phase 1 shipped, real failure data collected.

**Tasks**:
- Implement retry loop in assist:
  - If unparse fails, sanitize error message (remove any HCL content, security review M2)
  - Feed sanitized error back to LLM as context
  - Max 2 retries, show "Unparse failed, retrying (1/2)..."
  - Cost confirmation updated to mention potential 3x cost on retry
- Partial failure for multi-workflow: write successes, report failures, exit non-zero
- Typed errors: distinguish retryable (invalid YAML) from non-retryable (auth, timeout)

**Acceptance criteria**:
- [ ] Retry with error context improves success rate (measured)
- [ ] Retry error context sanitized (no HCL content leakage)
- [ ] Partial multi-workflow failure handled correctly
- [ ] Exit codes: 0 = success, 1 = all failed, 2 = partial success

## System-Wide Impact

### Interaction Graph

Current: `assist` calls: `ai.Provider.Generate()` → `StripFences()` → split YAML docs → write temp files → `provider.Unparse()` → `mergeHCLFiles()` → `splitHCLBlocksAST()` (dedup) → single output file.

Future: adds optional `pin.Pin()` and retry loop.

### Error Propagation

- AI API errors classified by type: auth (401), quota exceeded (insufficient_quota), rate limit (429), timeout (DeadlineExceeded)
- Unparse errors → show truncated raw YAML (max 500 chars) + suggestion to refine prompt
- Empty LLM response → clear error message
- Token usage displayed after every successful generation

### State Lifecycle Risks

- `./cinzel/assist/` receives timestamped files (`assist-20260316-193045.hcl`) — no overwrite conflicts
- `--refine` reads from `assist/` — clear error if deleted between runs
- Two temp directories cleaned up via `defer os.RemoveAll()` — no leak on error paths
- No database, no persistent state beyond file output

### API Surface Parity

- `assist` registered alongside `parse`/`unparse` in `addProvider()` — consistent flag patterns
- `--output-directory` supported (architecture review #6)

## Implementation Notes (post-implementation)

Items implemented during development that were not in the original plan:

### HCL block merging and deduplication

Multi-workflow prompts generate separate YAML documents. Each is unparsed independently, producing separate HCL files. `mergeHCLFiles` reads all generated HCL, splits into blocks using `hclwrite.ParseConfig` (AST-based, not brace counting), deduplicates identical blocks, and writes a single output file. This ensures shared steps (checkout, setup) appear once.

### Single timestamped output file

Output is `assist-{timestamp}.hcl` (e.g., `assist-20260316-193045.hcl`) instead of one file per workflow. Avoids overwrite conflicts on repeated runs.

### Token usage display

After generation, prints: `Tokens used: 1247 (input: 892, output: 355)`. Uses `GenerateResponse` struct with `InputTokens`, `OutputTokens`, `TotalTokens()`.

### System prompt for step reuse

Instructs the LLM to use identical step names across workflows so dedup works: "Use consistent names: checkout, setup_go, install_deps — not step_1, step_2."

### Code review findings addressed

From 3 parallel reviews (pattern recognition, performance, security):
- `splitHCLBlocks` replaced with AST-based `splitHCLBlocksAST` using `hclwrite`
- Regex `fencePattern` hoisted to package-level `var`
- Sentinel errors consolidated into `errors.go` per package
- `resolveAPIKey` shared helper eliminates constructor duplication
- Attribute ordering made deterministic via `sort.Strings`
- `truncateAtNewline` for safe context truncation (no mid-rune cuts)
- Raw YAML in error output truncated to 500 chars
- `errCancelled` and `errPromptRequired` as sentinel errors
- Assist logic extracted to `internal/command/assist.go` (~280 lines)

## Acceptance Criteria

### Functional Requirements (v1)

- [ ] `cinzel github assist --prompt "golang PR with tests"` generates valid HCL in `./cinzel/assist/`
- [ ] `--output-directory` customizes output location
- [ ] `--dry-run` prints HCL to stdout without writing files
- [ ] Cost confirmation shown before API call; `--acknowledge` bypasses
- [ ] `ANTHROPIC_API_KEY` missing → clear setup instructions
- [ ] Multi-document YAML from LLM split and unparsed individually
- [ ] Empty LLM response → clear error message
- [ ] `--dry-run` calls LLM but skips file write (cost confirmation still applies)
- [ ] Markdown fences stripped from LLM output

### Functional Requirements (v2+)

- [ ] `--provider openai` with `OPENAI_API_KEY`
- [ ] Config file support with `default:` provider selector
- [ ] Context injection with string literal stripping
- [ ] `--refine` with `--prompt` for iterative generation
- [ ] `--no-context` to skip context injection
- [ ] `cinzel github pin --file ./cinzel` resolves action tags to SHAs

### Non-Functional Requirements

- [ ] API timeout: 120 seconds
- [ ] User config file created with `0600` permissions
- [ ] Config paths use `os.UserConfigDir()` / `os.UserCacheDir()` (OS-agnostic)
- [ ] Every new `.go` file has copyright header and `doc.go`
- [ ] No unit tests for LLM output (non-deterministic); validate via unparse success
- [ ] Spinner displayed during LLM call (no streaming)

### Security Requirements

- [ ] API keys never in `.cinzelrc.yaml` — hard error if detected
- [ ] User config file `0600` permissions enforced
- [ ] v2+: string literals stripped from HCL context before LLM call (no sensitive values sent)
- [ ] v2+: retry error context sanitized (no HCL content)

## Dependencies & Prerequisites

### Phase 1

| Package | Purpose |
|---------|---------|
| `github.com/anthropics/anthropic-sdk-go` | Anthropic API client |

### Phase 2 (deferred)

| Package | Purpose |
|---------|---------|
| `github.com/openai/openai-go` | OpenAI API client |

### No Prerequisites

Phase 1 requires zero interface changes, zero config parser changes. Only adds new files + modifies command registration.

## Risk Analysis & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| LLM generates invalid YAML | High | Medium | v1: show error + raw YAML. v2+: retry with error context |
| Sensitive data in context | Medium | High | v2: strip all string literals from HCL. v1: no context sent |
| API key in project config | Medium | High | Hard error on detection, not just warning |
| GitHub API rate limits (pin) | Medium | Low | `GITHUB_TOKEN` support + SHA caching + graceful fallback |
| AI SDK dependency bloat | Low | Medium | Future: build tags for optional AI deps |
| Raw error output leaks data | Medium | Medium | v1: no context = no leakage. v2: only stripped structure sent |

## Security Considerations

- **v1 is inherently safe**: no context injection means no user data sent to LLM. Only the user's prompt is sent.
- **API keys**: env vars only in v1. v2 adds user config with `0600` enforcement + hard error on project config.
- **Context privacy (v2)**: all string literal values stripped from HCL before injection. LLM sees block structure only — no secrets, tokens, org names, or any values. No restoration needed since output is fresh YAML.
- **TLS**: SDK defaults (system CA bundle). v2+ with custom endpoints: enforce HTTPS scheme.
- **No logging**: prompts and responses never written to disk.
- **Error output**: v1 safe (no context). v2+ only sends stripped structure — no values to leak.

## Sources & References

### Origin

- **Brainstorm document**: [docs/brainstorms/2026-03-16-feat-cinzel-assist-ai-workflow-generation.md](../brainstorms/2026-03-16-feat-cinzel-assist-ai-workflow-generation.md) — Key decisions: YAML-then-unparse strategy, non-destructive `./cinzel/assist/` output, env var config, no templates.

### Review Findings Addressed

- **Simplicity review**: v1 reduced from 10 new files to 2. Deferred OpenAI, config files, context injection, retry, pin to later phases.
- **Architecture review**: temp files instead of `UnparseBytes` (no interface change), `--output-directory` on assist, `--refine` coexists with `--prompt`, `Resolver` interface for pin testability, `.yaml` extension on temp files, multi-doc split-then-loop.
- **Security review**: string literal stripping replaces entire redaction engine (C1, C2), `0600` config permissions (H1), context size cap (H3), `GITHUB_TOKEN` for pin (M4), retry context sanitization (M2).

### Internal References

- CLI command registration: `internal/command/command.go:79` (`addProvider()`)
- Config loading: `internal/command/config.go:16` (`.cinzelrc.yaml`)
- Unparse entry point: `provider/github/github.go:144` (`Unparse()`)
- File writing: `internal/fsutil/` (`WriteFile()`)
- Provider interface: `provider/provider.go`

### Learnings Applied

- Config input precedence: `docs/solutions/logic-errors/config-input-precedence-ignored-for-parse.md`
- YAML string quoting rules: `docs/solutions/developer-experience/yaml-string-quoting-rules.md`
- Single YAML unmarshal pattern: `docs/solutions/patterns/critical-patterns.md`
