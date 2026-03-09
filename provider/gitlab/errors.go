// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package gitlab

import "errors"

var (
	errInputPathRequired     = errors.New("`file` or `directory` must be set")
	errInputPathConflict     = errors.New("`file` and `directory` cannot be set together")
	errParseNotImplemented   = errors.New("gitlab parse is not implemented yet")
	errUnparseNotImplemented = errors.New("gitlab unparse is not implemented yet")
)
