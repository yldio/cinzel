// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"context"
	"os"
	"testing"
)

// TestCacheWithNowhereToWriteStaysOutOfTheWorkingDirectory covers a machine
// with no user cache directory, which is how a container without a home
// directory looks. The cache path was joined onto an empty string, so it
// became the relative "cinzel/pins" and the cache was written into whatever
// directory the command was run from — which for this tool is the directory
// holding the HCL it manages.
func TestCacheWithNowhereToWriteStaysOutOfTheWorkingDirectory(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("LocalAppData", "")

	dir := t.TempDir()
	t.Chdir(dir)

	inner := &mockResolver{shas: map[string]string{
		"actions/checkout@v4": "ffffffffffffffffffffffffffffffffffffffff",
	}}

	resolver := NewCachedResolver(inner)

	sha, err := resolver.ResolveTag(context.Background(), "actions", "checkout", "v4")
	if err != nil {
		t.Fatal(err)
	}

	if sha != "ffffffffffffffffffffffffffffffffffffffff" {
		t.Errorf("got %q", sha)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 0 {
		t.Errorf("the cache was written into the working directory: %v", entries[0].Name())
	}
}
