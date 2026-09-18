# Commands

## AI assist (`cinzel <provider> assist`)

Pipeline: prompt → LLM → YAML → strip fences → split docs → temp files →
Unparse → merge and dedup HCL → session folder.

Output is `cinzel/assist/{timestamp}/assist.hcl`, one session per prompt.
`--refine` targets the latest session; `--from {timestamp}` picks an older one.

Merging compares against the existing `cinzel/*.hcl`. A block identical to one
already there becomes a `// reuses:` comment; a different block with the same
signature gets a `// note:` saying where the other one is. Generated actions
are pinned to SHAs afterwards.

`StripHCLContext` replaces every string value in the context with `"..."` by
walking the HCL AST, so the prompt carries structure and not content.

Config is `os.UserConfigDir()/cinzel/config.yaml`, written by `cinzel init`. It
holds no API key: those come from the environment. Resolution order is CLI
flags, then env vars, then the config file, then the built-in defaults.

A config that still holds an `api_key` is read, because one written before
`cinzel init` stopped asking for keys may have one. That is a migration path,
not a way to configure a key, and `configWarnings` says so on every load.

## Version management (`cinzel github pin/upgrade`)

`pin` resolves action tags to SHAs through the GitHub API, cached for 24 hours.
The tag stays on the line it pinned: `version = "<sha>" # <tag>`.

`upgrade` finds the latest release, compares it against what is there by tag or
by SHA, and updates both the version and its comment.

No token is needed for public actions. `GITHUB_TOKEN` raises the rate limit.
