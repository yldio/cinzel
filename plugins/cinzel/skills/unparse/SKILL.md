---
name: unparse
description: Convert CI/CD YAML back to HCL definitions using cinzel
disable-model-invocation: true
argument-hint: [provider] [flags]
---

# Unparse YAML to HCL

Convert YAML back to HCL definitions for the `$0` provider.

1. Run `go run ./... $0 unparse --file .github/workflows --output-directory ./cinzel $ARGUMENTS`
2. Show the git diff of generated HCL files
3. If there are changes, summarize what changed
4. If there are no changes, confirm output is up to date
