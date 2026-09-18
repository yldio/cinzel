// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCinzel(t *testing.T) {
	// Every subtest writes os.Args, so the original is put back once rather
	// than each of them leaning on whichever ran before it. Run on its own,
	// "no error running main" used to read the test binary's own flags and
	// fail on -test.testlogfile.
	original := os.Args
	t.Cleanup(func() { os.Args = original })

	t.Run("shows unknown version", func(t *testing.T) {
		os.Args = []string{"cinzel", "-v"}

		buf := new(bytes.Buffer)

		if err := run(buf, "unknown"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out := buf.String()

		if !strings.EqualFold(out, "cinzel version unknown\n") {
			t.Fatalf("expected version output, got: %q", out)
		}
	})

	t.Run("shows set version", func(t *testing.T) {
		os.Args = []string{"cinzel", "-v"}

		buf := new(bytes.Buffer)

		if err := run(buf, "v9.9.9"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out := buf.String()

		if !strings.EqualFold(out, "cinzel version v9.9.9\n") {
			t.Fatalf("expected version output, got: %q", out)
		}
	})

	t.Run("no error running main", func(t *testing.T) {
		// main() exits the process on error, which would take the test binary
		// with it and report nothing, so the arguments it runs under are set
		// here rather than inherited.
		os.Args = []string{"cinzel", "-v"}

		main()
	})
}
