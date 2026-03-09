// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later
// Package hclparser evaluates HCL expressions into cty values. It handles
// literal values, template strings, binary operations, unary operations, scope
// traversals (variable references), and other HCL expression types. Variables
// can be registered and resolved during expression evaluation.
package hclparser
