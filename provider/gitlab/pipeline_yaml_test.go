// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// yaml.v3 quotes a leading reserved indicator on its own, but reaches for
// single quotes. The project writes double quotes or none.
func TestMarshalQuotesLeadingAtWithDoubleQuotes(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"@daily", "  K: \"@daily\"\n"},
		{"a@b.com", "  K: a@b.com\n"},
	} {
		got, err := marshalPipelineYAML(map[string]any{
			"build": map[string]any{"variables": map[string]any{"K": tc.in}},
		})
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(got), tc.want) {
			t.Errorf("marshal(%q) missing %q:\n%s", tc.in, tc.want, got)
		}
	}
}
