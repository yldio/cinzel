// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package job

func nonEmptyString(v any) (string, bool) {
	s, ok := v.(string)

	if !ok || s == "" {
		return "", false
	}

	return s, true
}
