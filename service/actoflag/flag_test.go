// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package actoflag

import (
	"testing"
)

func TestFlags(t *testing.T) {
	flags := New()

	t.Run("validate flags", func(t *testing.T) {
		if flags == nil {
			t.Fatal("flags should not be nil")
		}

		if flags.Directory != "" || flags.File != "" || flags.Recursive != false {
			t.Fatal("flags should have their 0 value")
		}
	})

	t.Run("reset flag dir", func(t *testing.T) {
		flags.SetDirectory("dummy-directory")

		if flags.Directory != "dummy-directory" {
			t.Fatal("flag dir should be dummy-directory")
		}
	})

	t.Run("reset flag file", func(t *testing.T) {
		flags.SetFile("dummy-file.hcl")

		if flags.File != "dummy-file.hcl" {
			t.Fatal("flag file should be dummy-file.hcl")
		}
	})

	t.Run("reset flag file", func(t *testing.T) {
		flags.SetRecursive(true)

		if flags.Recursive != true {
			t.Fatal("flag r should be true")
		}
	})
}
