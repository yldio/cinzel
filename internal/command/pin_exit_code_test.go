// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
	githubprovider "github.com/yldio/cinzel/provider/github"
)

// A local action names no repository, so neither command can resolve it and
// neither reaches the network to find that out. That is the whole failure this
// needs: the run prints "1 failed" either way, and the question is only what
// the command does about it.
const unresolvableAction = `step "build" {
  uses {
    action  = "./.github/actions/build"
    version = "v1"
  }
}
`

// An action already on a full commit SHA is left alone by pin without a
// request, so a run holding only this one has nothing to fail on.
const pinnedAction = `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "abcdef1234567890abcdef1234567890abcdef12"
  }
}
`

// Upgrade has no offline success path with a real action in it: it asks for
// the latest release whatever the version already says, so a SHA-pinned action
// still reaches the network and the control would fail on a machine with no
// route to it. A step using nothing has no action to resolve and no request to
// make.
const noActionAtAll = `step "build" {
  run = "make"
}
`

// Both summaries counted a failed result and printed it, and both actions then
// returned nil, so "1 failed" left the process at exit 0. A CI step calling
// "cinzel github pin" went green over a workflow still holding an unpinned
// action, which is the thing pinning is there to prevent.
func TestPinAndUpgradeFailWhenAnActionCannotBeResolved(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "pin",
			args: []string{"cinzel", "github", "pin", "--directory", "actions", "--dry-run"},
			want: "could not be pinned",
		},
		{
			name: "upgrade",
			args: []string{"cinzel", "github", "upgrade", "--directory", "actions", "--dry-run"},
			want: "could not be upgraded",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withTempWorkingDir(t, func() {
				writeActionDir(t, unresolvableAction)

				err := runCommandLine(t, tc.args)
				if err == nil {
					t.Fatal("a run reporting a failed action returned nil, so the process exits 0")
				}

				if !strings.Contains(err.Error(), tc.want) {
					t.Errorf("error = %v, want it to mention %q", err, tc.want)
				}
			})
		})
	}
}

// The control: a run with nothing to fail on still has to succeed, or the
// check above passes for the wrong reason.
func TestPinAndUpgradeSucceedWhenEveryActionResolves(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
		args []string
	}{
		{"pin", pinnedAction, []string{"cinzel", "github", "pin", "--directory", "actions", "--dry-run"}},
		{"upgrade", noActionAtAll, []string{"cinzel", "github", "upgrade", "--directory", "actions", "--dry-run"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withTempWorkingDir(t, func() {
				writeActionDir(t, tc.hcl)

				if err := runCommandLine(t, tc.args); err != nil {
					t.Errorf("a run with no failed action returned %v", err)
				}
			})
		})
	}
}

// writeActionDir puts one HCL file under "actions" in the working directory,
// which is what both commands are pointed at.
func writeActionDir(t *testing.T, content string) {
	t.Helper()

	if err := os.MkdirAll("actions", 0o750); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	writeFile(t, filepath.Join("actions", "steps.hcl"), []byte(content))
}

// runCommandLine drives the real command, so the result is the error the
// process exits on rather than what a summary helper returned.
func runCommandLine(t *testing.T, args []string) error {
	t.Helper()

	app := New(new(bytes.Buffer), "v.test")
	app.Cmd.ErrWriter = new(bytes.Buffer)

	return app.Execute(args, []provider.Provider{githubprovider.New()})
}
