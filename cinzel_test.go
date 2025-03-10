// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestActo(t *testing.T) {
	t.Run("shows unknown version", func(t *testing.T) {
		os.Args = []string{"cinzel", "-v"}

		buf := new(bytes.Buffer)
		run(buf, "unknown")

		out := buf.String()

		if !strings.EqualFold(out, "cinzel version unknown\n") {
			t.Fatalf("expected version output, got: %q", out)
		}
	})

	t.Run("shows set version", func(t *testing.T) {
		os.Args = []string{"cinzel", "-v"}

		buf := new(bytes.Buffer)
		run(buf, "v9.9.9")

		out := buf.String()

		if !strings.EqualFold(out, "cinzel version v9.9.9\n") {
			t.Fatalf("expected version output, got: %q", out)
		}
	})

	t.Run("no error running main", func(t *testing.T) {
		main()
	})
}
