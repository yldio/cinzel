# CLAUDE.md

## Project overview

`cinzel` is a bidirectional converter between HCL and CI/CD pipeline YAML. The first (and currently only) provider is GitHub Actions. The architecture is designed for multiple providers (GitLab Pipelines, etc.) via the `provider.Provider` interface.

- **Module**: `github.com/yldio/cinzel`
- **Go version**: 1.26
- **License**: AGPL-3.0-or-later
- **Task runner**: [mise](https://mise.jdx.dev/)

## Quick reference

```sh
mise run test          # run tests with coverage
mise run test-ci       # run tests (CI, no coverage file)
mise run fmt           # format Go + HCL files
mise run build         # build binary to ./bin/cinzel
mise run bench         # run benchmarks
mise run license       # apply license headers
mise run license-check # verify license headers
```

## Architecture

```
cinzel.go                     # CLI entrypoint, wires providers
internal/
  command/                  # CLI framework (urfave/cli)
  filereader/               # reads HCL/YAML from disk
  filewriter/               # writes output files
  hclparser/                # HCL expression evaluator
  yamlwriter/               # YAML marshalling with gopkg.in/yaml.v3 node control
  maputil/                  # sorted map iteration helpers
  naming/                   # HCL<->YAML identifier normalization
  fsutil/                   # filesystem utilities
  cinzelerror/               # shared sentinel errors
  test/                     # test helpers
provider/
  provider.go               # Provider interface (Parse, Unparse)
  github/                   # GitHub Actions provider
    github.go               # main Parse/Unparse entry points
    parse_workflow.go        # HCL -> workflow YAML
    parse_action.go          # HCL -> action YAML (composite, node, docker)
    unparse_workflow.go      # workflow YAML -> HCL
    unparse_action.go        # action YAML -> HCL
    workflow_yaml.go         # YAML node builder, quote style decisions
    models.go                # shared types
    config.go                # HCL block schema (parseConfig)
    validate.go              # cross-direction validation
    expression.go            # ${{ }} expression handling
    errors.go                # sentinel errors
    job/                     # job parsing/validation
    step/                    # step parsing/encoding
    action/                  # action validation (uses refs)
    workflow/                # workflow triggers, permissions, cron
```

## Key conventions

### Code style

- Every `.go` file starts with the copyright header. Run `mise run license` to apply.
- Every package has a `doc.go` with a package-level doc comment.
- Use `maputil.SortedKeys()` for deterministic iteration over maps. Never iterate maps directly when output order matters.
- Sentinel errors live in `errors.go` within each package. Use `errCamelCase` naming.
- No `testify` or assertion libraries — tests use stdlib `testing` only.

### HCL <-> YAML conversion

- **Parse** = HCL to YAML (the "forward" direction).
- **Unparse** = YAML to HCL (the "reverse" direction).
- `$${{ }}` in HCL string literals becomes `${{ }}` in YAML (double-dollar escaping).
- Document type is auto-detected during unparse: workflow (has `on`+`jobs`), action (has `name`+`runs`, no `on`/`jobs`), or step-only fallback.
- Actions write to `<output-dir>/<filename>/action.yml`. Workflows write to `<output-dir>/<filename>.yaml`.
- All `runs.using` types (composite, node20, docker) work in both directions. Only composite needs special handling for `steps` reference resolution; all other attributes flow through generic parsing.

### YAML output

- Uses `gopkg.in/yaml.v3` node-level marshalling for precise control.
- Strings that need quoting use `DoubleQuotedStyle` (not single quotes — Zed editor converts `'` to `"` on save, breaking golden tests).
- `stringNeedsQuoting()` determines when to quote: empty strings, booleans, null, numbers, YAML special characters. The `@` character does NOT trigger quoting (e.g., `actions/checkout@v4` stays unquoted).
- Top-level key order: `name`, `run-name`, `on`, `jobs`, then remaining keys alphabetically.
- Uses `goccy/go-yaml` for test assertions (semantic comparison), but `gopkg.in/yaml.v3` for production marshalling.

### Testing

- **Golden tests**: compare generated output against `.golden.yaml` files. Use `assertYAMLSemanticEqual` (not byte comparison).
- **Roundtrip tests**: HCL -> YAML -> HCL -> YAML, assert semantic equality. Proves bidirectional stability.
- **Fixture matrix**: structured under `testdata/fixtures/matrix/{parse,unparse}/{valid,invalid}/`. Valid cases have `.hcl`+`.golden.yaml`, invalid cases have `.hcl`+`.error.txt`.
- **Action fixtures**: under `testdata/fixtures/actions/` — composite, node, docker action types.
- **Benchmark tests**: `BenchmarkParseWorkflow`, `BenchmarkUnparseWorkflow`, `BenchmarkRoundtripWorkflow`.

### Commits

- Conventional commits enforced via `commitlint` (types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert).
- Header max 50 chars, body/footer max 72 chars per line.
- Changelog managed by `changie`.

### CI/CD

- GitHub Actions workflows in `.github/workflows/`.
- Docker build via `build/Dockerfile`.
- `mise run test` runs the key test suites before release.

## Adding a new provider

1. Create `provider/<name>/` implementing `provider.Provider`.
2. Wire it into `cinzel.go` alongside `github.New()`.
3. Add `provider/<name>/README.md` with HCL schema reference.
4. Add testdata fixtures following the same golden/roundtrip pattern.

## Common pitfalls

- **`parseHCLToWorkflows` return values**: Returns 4 values `([]WorkflowYAMLFile, map[string]any, []ActionYAMLFile, error)`. All error paths must return all 4.
- **Single YAML unmarshal**: `parseYAMLDocument()` unmarshals once, then `classifyWorkflowDocument()` or `classifyActionDocument()` route the result. Never unmarshal the same content twice.
- **Root package `-v` flag**: `go test -v ./...` at the root passes `-v` to the CLI app which rejects it. Run subpackages directly or omit `-v`.
