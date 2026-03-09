// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

import (
	"strings"
	"testing"
)

func TestValidateCron(t *testing.T) {
	valid := []struct {
		name string
		expr string
	}{
		{name: "every minute", expr: "* * * * *"},
		{name: "specific time", expr: "0 12 * * 1"},
		{name: "midnight daily", expr: "0 0 * * *"},
		{name: "step syntax", expr: "*/15 * * * *"},
		{name: "range", expr: "0 9-17 * * *"},
		{name: "list", expr: "0 0 1,15 * *"},
		{name: "range with step", expr: "0 0-23/2 * * *"},
		{name: "complex", expr: "30 4,8 1-15 1,6 0-4"},
	}

	for _, tt := range valid {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCron(tt.expr); err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
		})
	}

	invalid := []struct {
		name    string
		expr    string
		wantErr string
	}{
		{name: "empty", expr: "", wantErr: "must not be empty"},
		{name: "too few fields", expr: "* * *", wantErr: "must have 5 fields"},
		{name: "too many fields", expr: "* * * * * *", wantErr: "must have 5 fields"},
		{name: "minute out of range", expr: "60 * * * *", wantErr: "out of range"},
		{name: "hour out of range", expr: "0 24 * * *", wantErr: "out of range"},
		{name: "day out of range", expr: "0 0 32 * *", wantErr: "out of range"},
		{name: "month out of range", expr: "0 0 * 13 *", wantErr: "out of range"},
		{name: "dow out of range", expr: "0 0 * * 7", wantErr: "out of range"},
		{name: "invalid character", expr: "abc * * * *", wantErr: "invalid value"},
		{name: "inverted range", expr: "0 17-9 * * *", wantErr: "greater than end"},
	}

	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCron(tt.expr)

			if err == nil {
				t.Fatal("expected error")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateSchedule(t *testing.T) {
	t.Run("single cron string", func(t *testing.T) {
		err := ValidateSchedule(map[string]any{"cron": "0 0 * * *"})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("cron list", func(t *testing.T) {
		err := ValidateSchedule(map[string]any{"cron": []any{"0 0 * * *", "0 12 * * 1"}})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid cron in list", func(t *testing.T) {
		err := ValidateSchedule(map[string]any{"cron": []any{"0 0 * * *", "bad"}})

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("no cron key is ok", func(t *testing.T) {
		err := ValidateSchedule(map[string]any{"other": "val"})
		if err != nil {
			t.Fatal(err)
		}
	})
}
