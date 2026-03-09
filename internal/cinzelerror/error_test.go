// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package cinzelerror

import (
	"reflect"
	"testing"
)

func TestErrors(t *testing.T) {
	type Test struct {
		name   string
		have   string
		expect string
	}

	var tests = []Test{
		{"ErrWorkflowFilenameRequired", ErrWorkflowFilenameRequired.Error(), "`workflow` requires a filename"},
		{"ErrOnlyHclFiles", ErrOnlyHclFiles.Error(), "only HCL files are allowed"},
		{"ErrOnRestriction", ErrOnRestriction.Error(), "`on` can only have Events or Event"},
		{"ErrSecretsRestriction", ErrSecretsRestriction.Error(), "only `secrets` blocks or one single `secret` attribute is allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if !reflect.DeepEqual(tt.have, tt.expect) {
				t.Fatal(tt.name)
			}
		})
	}
}
