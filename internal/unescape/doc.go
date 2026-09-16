// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0
// Package unescape rewrites the Unicode escape sequences that hclwrite and
// gopkg.in/yaml.v3 emit for characters they wrongly treat as unprintable,
// restoring the raw UTF-8 without touching escapes the source itself contained.
package unescape
