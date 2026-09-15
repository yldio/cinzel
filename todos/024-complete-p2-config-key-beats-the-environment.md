---
status: complete
priority: p2
issue_id: "024"
tags: [code-review, correctness, assist, config]
dependencies: []
---

# A key in the config file beats the environment variable

## Problem Statement

The documented order is env var > config file. `ResolveAPIKey` never read the
environment, so a key left in the config file won the resolution and the
exported variable was ignored. Its own doc comment said otherwise.

## Findings

`internal/ai/config.go`, before the fix:

```go
// resolution order: env var > config file.
func (c Config) ResolveAPIKey(providerName string) string {
	pc, ok := c.Providers[providerName]
	if ok && pc.APIKey != "" {
		return pc.APIKey
	}

	return ""
}
```

With `ANTHROPIC_API_KEY=from-env` exported and `from-config-file` in the
config, the provider was built with `from-config-file`.

The environment was read one layer down, in `resolveAPIKey` in
`internal/ai/provider.go`, but only when the key handed in is empty. A
non-empty config key meant that fallback never ran. So the env var only
applied when no config key existed, which is the order inverted.

`cinzel init` writes `api_key: ""` when the answer is left blank, and an
empty string does not win, so a config written that way behaves correctly.
The defect needed a config with a key actually filled in — a stale key in
the file silently overriding a fresh one in the environment.

## Recommended Action

Read the env var first in `ResolveAPIKey`, matching the documented order.

## Technical Details

- `internal/ai/config.go`, `ResolveAPIKey`
- `apiKeyEnvVar` maps a provider name to its variable, with the same
  `"anthropic", ""` default case as `resolveAIProvider`
- An unknown provider maps to no variable, so it reads its config entry only
- `resolveAPIKey` in `provider.go` is unchanged and still covers the case
  where nothing is configured at all

## Acceptance Criteria

- [x] The env var beats a key in the config file
- [x] The config file is still used when the env var is unset
- [x] An empty env var does not shadow the config file
- [x] Each provider reads its own variable, with no leaking between them
- [x] An unknown provider reads no variable and keeps its config entry

## Work Log

- 2026-09-15: Found by probing the resolution order left unfinished in the
  previous sweep. Fixed and closed.
