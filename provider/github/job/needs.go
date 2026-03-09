// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package job

import "fmt"

// NeedsFromYAML extracts the needs list from a raw YAML value (string or list).
func NeedsFromYAML(raw any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}

	if name, ok := raw.(string); ok {
		if name == "" {
			return nil, fmt.Errorf("'needs' entries must be non-empty strings")
		}

		return []string{name}, nil
	}

	items, ok := raw.([]any)

	if !ok {
		return nil, fmt.Errorf("'needs' must be a string or list")
	}

	out := make([]string, 0, len(items))

	for _, item := range items {
		name, ok := item.(string)

		if !ok || name == "" {
			return nil, fmt.Errorf("'needs' entries must be non-empty strings")
		}
		out = append(out, name)
	}

	return out, nil
}
