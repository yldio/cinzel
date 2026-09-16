---
status: pending
priority: p3
issue_id: "045"
tags: [code-review, assist, validation]
dependencies: []
---

# A directory named ..hidden is read as traversal

## Problem Statement

`validateRelativePath` refuses any cleaned path starting with the two
characters `..`, so `..hidden` and `...x` are rejected although neither
escapes the working directory.

## Findings

`internal/command/assist.go:585`.

## Recommended Action

Compare the first path element against `..` exactly, or check for a `../`
prefix and the bare `..` case.

## Acceptance Criteria

- [ ] `..hidden` is accepted
- [ ] `../x`, `..` and absolute paths are still refused
