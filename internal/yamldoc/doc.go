// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0
// Package yamldoc provides an ordered, comment-carrying document used as the
// intermediate representation between a provider's source model and its YAML
// output. Key order, inline comments and the distinction between a collapsed
// empty map and an explicit one are properties of the document itself, so the
// encoder never has to recover them by rewriting encoded bytes.
package yamldoc
