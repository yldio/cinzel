// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package maputil

import "sort"

// ToStringAnyMap converts a map[string]any or map[any]any value to map[string]any.
func ToStringAnyMap(value any) (map[string]any, bool) {
	switch m := value.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		out := make(map[string]any, len(m))

		for k, v := range m {
			ks, ok := k.(string)

			if !ok {

				return nil, false
			}

			out[ks] = v
		}

		return out, true
	default:
		return nil, false
	}
}

// SortedKeys returns the keys of mapping in sorted order.
func SortedKeys[T any](mapping map[string]T) []string {
	keys := make([]string, 0, len(mapping))

	for key := range mapping {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}
