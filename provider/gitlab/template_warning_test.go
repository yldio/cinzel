// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// The top-level loop warns about a key it cannot place, and a hidden key was
// warned about on the way into the "template" block the schema declares for
// it. The warning names the key as unsupported and says it went out untouched,
// which is what the two branches after it do; a template is converted, and
// "extends" refers to it as template.<id>. Every pipeline with a hidden job
// read as one it had lost something on.
func TestAConvertedTemplateDoesNotWarn(t *testing.T) {
	for _, tc := range []struct {
		name string
		yml  string
		want string
	}{
		{
			name: "a hidden key becomes a template",
			yml:  ".tpl:\n  script:\n    - make\nbuild:\n  extends: .tpl\n  script:\n    - echo hi\n",
		},
		{
			name: "a key with nowhere to go still warns",
			yml:  "build:\n  script:\n    - echo hi\nleftover: v\n",
			want: "leftover",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			warnings := captureStderr(t, func() {
				if err := unparseYAMLString(t, tc.yml); err != nil {
					t.Fatalf("Unparse() error = %v", err)
				}
			})

			if tc.want == "" {
				if warnings != "" {
					t.Fatalf("want no warning, got %q", warnings)
				}

				return
			}

			if !strings.Contains(warnings, tc.want) {
				t.Errorf("want %q named in %q", tc.want, warnings)
			}
		})
	}
}
