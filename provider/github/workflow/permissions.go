// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

import "fmt"

// Known GitHub Actions permission scopes.
var knownPermissionScopes = map[string]struct{}{
	"actions":             {},
	"attestations":        {},
	"checks":              {},
	"contents":            {},
	"deployments":         {},
	"discussions":         {},
	"id-token":            {},
	"issues":              {},
	"packages":            {},
	"pages":               {},
	"pull-requests":       {},
	"repository-projects": {},
	"security-events":     {},
	"statuses":            {},
}

// Valid permission levels for individual scopes.
var validPermissionLevels = map[string]struct{}{
	"read":  {},
	"write": {},
	"none":  {},
}

// Valid shorthand string values for the entire permissions block.
var validPermissionShorthands = map[string]struct{}{
	"read-all":  {},
	"write-all": {},
}

// ValidatePermissions checks that a permissions value uses known scopes and valid levels.
// Accepts a string shorthand ("read-all"/"write-all"), an empty map (all none), or
// a map of scope→level.
func ValidatePermissions(raw any) error {
	if raw == nil {
		return nil
	}

	if s, ok := raw.(string); ok {
		if _, valid := validPermissionShorthands[s]; !valid {
			return fmt.Errorf("invalid permissions shorthand %q, expected 'read-all' or 'write-all'", s)
		}
		return nil
	}

	m, ok := toStringMap(raw)
	if !ok {
		return fmt.Errorf("permissions must be a string or an object")
	}

	for scope, levelRaw := range m {
		if _, known := knownPermissionScopes[scope]; !known {
			return fmt.Errorf("unknown permissions scope %q", scope)
		}

		level, ok := levelRaw.(string)
		if !ok {
			return fmt.Errorf("permissions scope %q must have a string value", scope)
		}

		if _, valid := validPermissionLevels[level]; !valid {
			return fmt.Errorf("invalid permission level %q for scope %q, expected 'read', 'write', or 'none'", level, scope)
		}
	}

	return nil
}

func toStringMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
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
	}
	return nil, false
}
