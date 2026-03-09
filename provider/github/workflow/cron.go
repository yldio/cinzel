// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

import (
	"fmt"
	"strconv"
	"strings"
)

// cronFieldSpec defines the valid range for a single cron field.
type cronFieldSpec struct {
	name string
	min  int
	max  int
}

var cronFields = []cronFieldSpec{
	{name: "minute", min: 0, max: 59},
	{name: "hour", min: 0, max: 23},
	{name: "day-of-month", min: 1, max: 31},
	{name: "month", min: 1, max: 12},
	{name: "day-of-week", min: 0, max: 6},
}

// ValidateCron checks that a cron expression has 5 fields with valid ranges.
func ValidateCron(expr string) error {
	expr = strings.TrimSpace(expr)

	if expr == "" {
		return fmt.Errorf("cron expression must not be empty")
	}

	fields := strings.Fields(expr)

	if len(fields) != 5 {
		return fmt.Errorf("cron expression must have 5 fields, got %d: %q", len(fields), expr)
	}

	for i, field := range fields {
		if err := validateCronField(field, cronFields[i]); err != nil {
			return fmt.Errorf("cron %s field: %w", cronFields[i].name, err)
		}
	}

	return nil
}

func validateCronField(field string, spec cronFieldSpec) error {
	// Handle wildcard
	if field == "*" {
		return nil
	}

	// Handle step on wildcard: */n
	if strings.HasPrefix(field, "*/") {
		return validateCronNumber(field[2:], spec)
	}

	// Handle list: a,b,c
	parts := strings.Split(field, ",")

	for _, part := range parts {
		// Handle range: a-b or a-b/n
		if strings.Contains(part, "-") {
			if err := validateCronRange(part, spec); err != nil {
				return err
			}
			continue
		}

		if err := validateCronNumber(part, spec); err != nil {
			return err
		}
	}

	return nil
}

func validateCronRange(field string, spec cronFieldSpec) error {
	// Handle step on range: a-b/n
	rangePart := field

	if idx := strings.Index(field, "/"); idx >= 0 {
		rangePart = field[:idx]
		step := field[idx+1:]

		if _, err := strconv.Atoi(step); err != nil {
			return fmt.Errorf("invalid step %q in %q", step, field)
		}
	}

	bounds := strings.SplitN(rangePart, "-", 2)

	if len(bounds) != 2 {
		return fmt.Errorf("invalid range %q", field)
	}

	low, err := strconv.Atoi(bounds[0])
	if err != nil {
		return fmt.Errorf("invalid range start %q in %q", bounds[0], field)
	}

	high, err := strconv.Atoi(bounds[1])
	if err != nil {
		return fmt.Errorf("invalid range end %q in %q", bounds[1], field)
	}

	if low < spec.min || low > spec.max {
		return fmt.Errorf("value %d out of range [%d-%d]", low, spec.min, spec.max)
	}

	if high < spec.min || high > spec.max {
		return fmt.Errorf("value %d out of range [%d-%d]", high, spec.min, spec.max)
	}

	if low > high {
		return fmt.Errorf("range start %d is greater than end %d", low, high)
	}

	return nil
}

func validateCronNumber(s string, spec cronFieldSpec) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid value %q", s)
	}

	if n < spec.min || n > spec.max {
		return fmt.Errorf("value %d out of range [%d-%d]", n, spec.min, spec.max)
	}

	return nil
}

// ValidateSchedule validates all cron entries in a normalized schedule event value.
func ValidateSchedule(value map[string]any) error {
	cronRaw, ok := value["cron"]

	if !ok {
		return nil
	}

	switch cron := cronRaw.(type) {
	case string:
		return ValidateCron(cron)
	case []any:
		for i, item := range cron {
			s, ok := item.(string)

			if !ok {
				return fmt.Errorf("schedule cron[%d] must be a string", i)
			}

			if err := ValidateCron(s); err != nil {
				return fmt.Errorf("schedule cron[%d]: %w", i, err)
			}
		}

		return nil
	default:
		return fmt.Errorf("schedule cron must be a string or list")
	}
}
