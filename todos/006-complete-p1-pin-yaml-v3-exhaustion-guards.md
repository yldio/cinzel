---
status: complete
priority: p1
issue_id: "006"
tags: [code-review, security, performance]
dependencies: []
---

# No input size limit before YAML alias expansion (billion laughs DoS)

## Problem Statement

`parseYAMLDocument` calls `yamlv3.Unmarshal(content, &node)` with no upper bound on input size or alias depth. `gopkg.in/yaml.v3` expands YAML aliases eagerly during unmarshal. A crafted input using deeply nested anchors (e.g., billion-laughs pattern) causes exponential memory allocation before any application check runs.

While `cinzel` is a CLI tool operating on local files, two attack surfaces exist:
1. `cinzel assist` feeds LLM-generated YAML through the unparse path — the LLM output is untrusted
2. A repository CI workflow could trigger `cinzel unparse` against a YAML file controlled by a third party

## Findings

- `provider/github/unparse_workflow.go` — `parseYAMLDocument` calls `yamlv3.Unmarshal` with no size check
- No max-byte guard at any call site in `Unparse` or `unparseYAMLFile`
- `node.Decode(&doc)` after unmarshal expands the node tree into `map[string]any`, compounding allocation

## Proposed Solutions

### Option A — Byte-length guard at `parseYAMLDocument` call site (Recommended)

Add a length check in `parseYAMLDocument` before the unmarshal:

```go
const maxYAMLBytes = 1 << 20 // 1 MB
if len(content) > maxYAMLBytes {
    return nil, nil, fmt.Errorf("YAML input exceeds maximum size of %d bytes", maxYAMLBytes)
}
```

- **Pros:** simple, fast, zero false positives for real workflow files (largest real workflow is <100KB)
- **Cons:** arbitrary limit; must be documented

### Option B — Cap alias expansion with a wrapper decoder

No built-in alias limit in yaml.v3. Would require forking or wrapping.

- **Pros:** more precise
- **Cons:** significant implementation cost; yaml.v3 is not easily intercepted

## Recommended Action

**Correction: the earlier triage was wrong.** It probed only the GitHub
provider, which parses with `gopkg.in/yaml.v3` and is indeed safe. GitLab
parses with `goccy/go-yaml`, which has no alias cap at all, and it is
vulnerable. Measured at b6221f1, an alias bomb through `Unparse`:

| levels | input | output HCL | allocated | elapsed |
|---|---|---|---|---|
| 3 | 230 B | 110 KB | 69 MB | 26ms |
| 4 | 280 B | 1.1 MB | 683 MB | 150ms |
| 5 | 330 B | 11 MB | 6.5 GB | 1.6s |
| 6 | 380 B | 110 MB | 73 GB | 78s |

Ten-fold per level, from a file that fits on a postcard. A level-8 bomb run
through a built binary reached 2.1GB resident at 97% CPU after 3.5 minutes
and was still climbing when it was killed. GitHub rejects the same shape in
1ms.

goccy is lazy: `parseYAMLDocument` shows no growth, so the cost only lands
downstream in `pipelineToHCL`, which is why the earlier probe of the decode
alone found nothing.

**Fix:** run the document through a yaml.v3 decode first in
`provider/gitlab/unparse_pipeline.go` and refuse it if yaml.v3 reports
excessive aliasing or exceeded depth. Every other complaint is left to
goccy, whose messages and positions the rest of the package is written
against. The pre-pass costs about 400ms on a 1.4MB pipeline.

The byte cap this issue originally proposed is still not the answer: it
would reject a large honest pipeline while letting a 380-byte bomb through.

## Technical Details

- Affected file: `provider/gitlab/unparse_pipeline.go`, functions
  `parseYAMLDocument` and `rejectExhaustingYAML`
- Related path: `provider/gitlab/gitlab.go` → `Unparse` → `pipelineToHCL`
- Safe by its dependency: `provider/github/unparse_workflow.go`,
  `parseYAMLDocument`

## Acceptance Criteria

- [x] Alias-expansion and depth exhaustion shown to be already rejected on
      the GitHub path
- [x] The GitLab path, which uses goccy and has no such cap, is guarded
- [x] A test pins both guarantees in both providers so a dependency upgrade
      cannot silently drop them

## Work Log

- 2026-03-31: Finding created during code review of single-pass yaml.v3 parse refactor
- 2026-09-15: Probed against 3a8ee45. Both described attacks are already
  rejected by yaml.v3 in under 0.03s. Size limit dropped from scope; the real
  exhaustion path was quadratic identifier construction and is now fixed.
  Remaining work is a regression test over the dependency's guarantees.
- 2026-09-15: Reopened. The 2026-09-15 conclusion above was reached by
  probing the GitHub provider only and is wrong for GitLab, which decodes
  with goccy and expands ten-fold per alias level with no limit. Guarded
  the GitLab path behind a yaml.v3 pre-pass and pinned both providers with
  tests. Each gate was proven to fail: the GitLab pair by reverting the
  guard (the alias bomb then ran to 1.8GB resident before it was killed),
  the GitHub pair by building against a local yaml.v3 with its alias cap
  removed and its depth cap raised.
