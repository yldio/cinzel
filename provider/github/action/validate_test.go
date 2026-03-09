// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package action

import (
	"strings"
	"testing"
)

func TestValidateUsesRef(t *testing.T) {
	valid := []struct {
		name string
		uses string
	}{
		{name: "standard action", uses: "actions/checkout@v4"},
		{name: "SHA pinned", uses: "actions/checkout@a81bbbf8298c0fa03ea29cdc473d45769f953675"},
		{name: "with path", uses: "actions/aws/ec2@main"},
		{name: "local action", uses: "./actions/my-action"},
		{name: "parent local", uses: "../shared/action"},
		{name: "docker action", uses: "docker://alpine:3.18"},
		{name: "docker hub", uses: "docker://ghcr.io/owner/image:latest"},
		{name: "branch ref", uses: "owner/repo@feature/branch"},
	}

	for _, tt := range valid {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateUsesRef(tt.uses); err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
		})
	}

	invalid := []struct {
		name    string
		uses    string
		wantErr string
	}{
		{name: "empty", uses: "", wantErr: "must not be empty"},
		{name: "no version", uses: "actions/checkout", wantErr: "must include a version reference"},
		{name: "empty ref", uses: "actions/checkout@", wantErr: "empty version reference"},
		{name: "no slash", uses: "checkout@v4", wantErr: "owner/repo@ref format"},
		{name: "empty owner", uses: "/repo@v4", wantErr: "empty owner or repo"},
		{name: "empty repo", uses: "owner/@v4", wantErr: "empty owner or repo"},
		{name: "docker no image", uses: "docker://", wantErr: "must specify an image"},
	}

	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsesRef(tt.uses)

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}
