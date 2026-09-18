// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import "errors"

// errNonNumericIndex is returned when a traversal is indexed with something
// other than a known number, which is the only thing a list index can be.
var errNonNumericIndex = errors.New("a variable can only be indexed with a number")

// errIndexUnsupportedType is returned when a traversal indexes a variable
// whose type cannot be indexed by position.
var errIndexUnsupportedType = errors.New("only a list or tuple variable can be indexed by position")
