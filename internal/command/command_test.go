// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

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
			errBuf := new(bytes.Buffer)

			app := NewWithErrWriter(buf, errBuf, "v.9.9.9")

			p := test.MockProvider(t, buf)

			if tt.hasError {
				p.HasError = true
			}

			app.Execute(tt.args, []provider.Provider{p})

			// A failure belongs on stderr: on stdout it would land in
			// whatever the converted output was being collected into.
			got := buf.String()
			if tt.hasError {
				got = errBuf.String()

				if buf.Len() != 0 {
					t.Fatalf("want nothing on stdout, got %q", buf.String())
				}
			}

			if got != tt.expect {
				t.Fatalf("got %q, want %q", got, tt.expect)
			}
		})
	}
}
