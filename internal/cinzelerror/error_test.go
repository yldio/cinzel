// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.have, tt.expect) {
				t.Fatal(tt.name)
			}
		})
	}
}
