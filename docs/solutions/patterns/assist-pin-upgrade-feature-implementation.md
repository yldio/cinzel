---
title: "feat: cinzel assist, pin, and upgrade — AI-powered workflow generation"
date: 2026-03-17
category: patterns
tags: [ai, llm, cli, hcl, yaml, github-actions, anthropic, openai, pin, upgrade, architecture]
components:
  - internal/ai
  - internal/command
  - internal/pin
origin: docs/plans/2026-03-16-feat-cinzel-assist-ai-workflow-generation-plan.md
---

# cinzel assist, pin, and upgrade

Three new commands added to cinzel for AI-powered workflow generation and GitHub Actions version management.

## Commands

```sh
cinzel github assist --prompt "golang PR with tests"              # AI generates HCL
cinzel github pin                                                  # lock tags to SHAs
cinzel github upgrade                                              # bump to latest + pin
cinzel github upgrade --parse                                      # bump + regenerate YAML
```

## Architecture: YAML-then-unparse pipeline

The LLM generates standard CI/CD YAML (which it knows from training data), then cinzel's existing `Unparse` converts to HCL. This avoids teaching the LLM cinzel's custom HCL schema.

```
prompt → LLM → YAML → strip fences → split docs → temp files → Unparse → merge/dedup HCL → single output
```

### Key design decisions

1. **Temp file approach** — no `Provider` interface changes. Write LLM YAML to temp files, call existing `Unparse(ProviderOps{Directory: tmpDir})`. ~5 lines of code vs adding `UnparseBytes` to every provider.

2. **HCL block merging** — multi-workflow prompts produce separate YAML documents. Each is unparsed independently, then `mergeHCLFiles` uses `hclwrite.ParseConfig` (AST-based, not brace counting) to split and deduplicate identical blocks. Shared steps appear once.

3. **System prompt for step reuse** — instructs the LLM to use identical step names/IDs across workflows so dedup works.

4. **Top-level vs nested** — `assist` is top-level (provider-agnostic, takes `--provider` flag). `pin` and `upgrade` are under `cinzel github` (GitHub-specific operations).

## Privacy: string literal stripping

Instead of a redaction/restoration engine, `StripHCLContext` walks the HCL AST and replaces ALL string literal values with `"..."`. Comments are dropped by the parser. Block labels preserved (accepted residual risk — comparable to function names).

The LLM sees structural skeleton only:
```hcl
step "checkout" {
  name = "..."
  uses {
    action  = "..."
    version = "..."
  }
}
```

Context capped at 8000 tokens with newline-boundary truncation.

## Pin: tag-to-SHA resolution

`Resolver` interface with `GitHubResolver` (GitHub API) and `CachedResolver` (file-based, 24h TTL in `os.UserCacheDir()/cinzel/pins/`).

- Handles annotated tags (dereferences to commit SHA)
- Adds/updates `// actions/checkout v4` comment above `uses` blocks
- Fallback on failure: warning + keep unpinned tag
- `GITHUB_TOKEN` env var for authenticated requests (5000/hr vs 60/hr)

## Upgrade: latest version lookup

Uses GitHub releases API (`/repos/.../releases/latest`). Compares by tag OR SHA — detects already-current versions even when SHA-pinned. No cache (must check live). Optional `--parse` regenerates YAML after upgrading.

## AI provider interface

```go
type Provider interface {
    Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
    Name() string
}
```

Two implementations: `Anthropic` and `OpenAI`. Shared `resolveAPIKey` helper. Error classification distinguishes auth/quota/rate-limit/timeout with actionable messages.

`GenerateResponse` includes `InputTokens`/`OutputTokens` — displayed after every generation.

## Review findings addressed

Three rounds of parallel reviews (architecture, security, simplicity, performance, pattern recognition) produced these fixes:

| Finding | Fix |
|---------|-----|
| Brace-counting HCL splitter fragile | Replaced with `hclwrite.ParseConfig` AST |
| Regex compiled per call | Hoisted to package-level `var` |
| Sentinel errors scattered | Consolidated into `errors.go` per package |
| Constructor duplication | Shared `resolveAPIKey` helper |
| Attribute ordering non-deterministic | Sorted map keys |
| Byte truncation mid-rune | `truncateAtNewline` cuts at last newline |
| Raw YAML in errors could leak data | Truncated to 500 chars |
| Path traversal on `--context-dir` | `validateRelativePath` blocks absolute + `..` |
| Dead `splitHCLBlocks` wrapper | Removed |
| 6 untested functions | Tests added for all |

## File inventory

| Package | File | Purpose |
|---------|------|---------|
| `internal/ai` | `provider.go` | Interface, error classification, StripFences, SystemPrompt |
| `internal/ai` | `anthropic.go` | Anthropic SDK wrapper |
| `internal/ai` | `openai.go` | OpenAI SDK wrapper |
| `internal/ai` | `strip.go` | HCL string stripping for privacy |
| `internal/ai` | `errors.go` | Sentinel errors |
| `internal/command` | `assist.go` | Assist pipeline, mergeHCLFiles, YAML splitting |
| `internal/command` | `pin.go` | Pin CLI command |
| `internal/command` | `upgrade.go` | Upgrade CLI command with --parse |
| `internal/command` | `errors.go` | CLI sentinel errors |
| `internal/pin` | `pin.go` | Resolver, GitHubResolver, CachedResolver, PinFile |
| `internal/pin` | `upgrade.go` | UpgradeFile, UpgradeDirectory, LatestTag |
| `internal/pin` | `errors.go` | GitHub API error classification with token hints |

## Related

- [Brainstorm](../../brainstorms/2026-03-16-feat-cinzel-assist-ai-workflow-generation.md)
- [Plan](../../plans/2026-03-16-feat-cinzel-assist-ai-workflow-generation-plan.md)
- [git-cliff release notes fix](../integration-issues/git-cliff-release-notes-wrong-changelog.md)
