// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"testing"

	yamlv3 "gopkg.in/yaml.v3"
)

// A script line holding a backslash went out of a plain YAML scalar with that
// backslash single, which is the parity unescape.Unicode read as an escape it
// had written itself, so the six characters the author typed came back as the
// one they name. Exit 0 on both passes, and the shell command is no longer the
// command.
func TestABackslashInAScriptSurvivesTheRoundtrip(t *testing.T) {
	for _, tc := range []struct {
		name string
		yml  string
		want string
	}{
		{
			name: "a printf escape",
			yml:  "build:\n  script:\n    - 'printf \\u00e9 now'\n",
			want: `printf \u00e9 now`,
		},
		{
			name: "a windows path",
			yml:  "build:\n  script:\n    - 'copy C:\\users\\u00e9dir out'\n",
			want: `copy C:\users\u00e9dir out`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, back := roundtripYAML(t, tc.yml)

			// Read the YAML rather than matching its text: a value that also
			// holds a colon goes out double-quoted, where the same backslash
			// is written doubled and means one.
			var got struct {
				Build struct {
					Script []string `yaml:"script"`
				} `yaml:"build"`
			}

			if err := yamlv3.Unmarshal([]byte(back), &got); err != nil {
				t.Fatalf("Parse() wrote YAML that does not read: %v\n%s", err, back)
			}

			if len(got.Build.Script) != 1 || got.Build.Script[0] != tc.want {
				t.Errorf("script = %q, want [%q]\n%s", got.Build.Script, tc.want, back)
			}
		})
	}
}
