// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package command

import (
	"bytes"
	"testing"

	"github.com/yldio/cinzel/internal/test"
	"github.com/yldio/cinzel/provider"
)

func TestCommand(t *testing.T) {
	type Test struct {
		name     string
		args     []string
		hasError bool
		expect   string
	}

	var tests = []Test{
		{
			"should show version message",
			[]string{"cinzel", "-v"},
			false,
			"cinzel version v.9.9.9\n",
		},
		{
			"should show mock-provider parse message",
			[]string{"cinzel", "mock-provider", "parse"},
			false,
			"parse",
		},
		{
			"should show mock-provider unparse message",
			[]string{"cinzel", "mock-provider", "unparse"},
			false,
			"unparse",
		},
		{
			"should show mock-provider parse error message",
			[]string{"cinzel", "mock-provider", "parse"},
			true,
			"parse error, if you think this is incorrect, consider opening an issue in https://www.github.com/yldio/cinzel/issues\n",
		},
		{
			"should show mock-provider unparse error message",
			[]string{"cinzel", "mock-provider", "unparse"},
			true,
			"unparse error, if you think this is incorrect, consider opening an issue in https://www.github.com/yldio/cinzel/issues\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			app := New(buf, "v.9.9.9")

			p := test.MockProvider(t, buf)

			if tt.hasError {
				p.HasError = true
			}

			app.Execute(tt.args, []provider.Provider{p})

			if buf.String() != tt.expect {
				t.FailNow()
			}
		})
	}
}
