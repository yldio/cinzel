// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package cinzelerror

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
)

// Only a diagnostic's detail used to be collected. A diagnostic carrying a
// summary alone left nothing to join, so the user was shown "%!w(<nil>)" and
// no hint of what was wrong.
func TestProcessHCLDiags(t *testing.T) {
	for _, tc := range []struct {
		name  string
		diags hcl.Diagnostics
		want  []string
	}{
		{
			name:  "a summary with no detail",
			diags: hcl.Diagnostics{{Severity: hcl.DiagError, Summary: "Unsupported block type"}},
			want:  []string{"Unsupported block type"},
		},
		{
			name: "a detail is still preferred",
			diags: hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Unsupported block type",
				Detail:   `Blocks of type "bogus" are not expected here.`,
			}},
			want: []string{`Blocks of type "bogus" are not expected here.`},
		},
		{
			name: "one of each is joined",
			diags: hcl.Diagnostics{
				{Severity: hcl.DiagError, Summary: "first"},
				{Severity: hcl.DiagError, Summary: "ignored", Detail: "second"},
			},
			want: []string{"first", "second"},
		},
		{
			name:  "nothing to report at all",
			diags: hcl.Diagnostics{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ProcessHCLDiags(tc.diags).Error()

			if strings.Contains(got, "%!w") {
				t.Fatalf("a formatting verb reached the message: %q", got)
			}

			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("want %q in %q", want, got)
				}
			}

			// A diagnostic says what is wrong with what the author wrote, so
			// there is nothing here to report to us. The exception is a set of
			// diagnostics carrying no message at all, which leaves the author
			// nothing to act on and is ours to fix.
			wantIssueLine := len(tc.want) == 0

			if strings.Contains(got, OpenIssue) != wantIssueLine {
				t.Errorf("want the issue line present=%v in %q", wantIssueLine, got)
			}

			if !wantIssueLine && !IsUserInput(ProcessHCLDiags(tc.diags)) {
				t.Errorf("want %q marked as caused by the input", got)
			}
		})
	}
}
