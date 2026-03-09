// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package job

func nonEmptyString(v any) (string, bool) {
	s, ok := v.(string)

	if !ok || s == "" {

		return "", false
	}

	return s, true
}
