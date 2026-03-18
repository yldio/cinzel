// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import "errors"

var (
	errCancelled      = errors.New("cancelled")
	errPromptRequired = errors.New("--prompt is required (or use --refine to iterate on previous output)")
	errAbsolutePath   = errors.New("path must be relative to the project directory")
	errPathTraversal  = errors.New("path must not escape the project directory (no .. traversal)")
)
