// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

// Package hclcomment writes comments into an hclwrite body.
//
// It sits here rather than beside its first caller because both the providers
// and the step package write comments, and the step package is imported by the
// providers.
package hclcomment
