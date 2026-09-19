// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package workflow

import (
	"fmt"
	"strconv"
	"strings"
)

// cronFieldSpec defines the valid range for a single cron field, and the names
// that field accepts in place of a number.
type cronFieldSpec struct {
	name  string
	min   int
	max   int
	names map[string]int
}

var monthNames = map[string]int{
	"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4, "MAY": 5, "JUN": 6,
	"JUL": 7, "AUG": 8, "SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
}

var dayNames = map[string]int{
	"SUN": 0, "MON": 1, "TUE": 2, "WED": 3, "THU": 4, "FRI": 5, "SAT": 6,
}

// day-of-week runs to 7 because both 0 and 7 name Sunday. Every name a field
// accepts resolves to a number in its own range, so the range checks below need
// no separate case for them.
var cronFields = []cronFieldSpec{
	{name: "minute", min: 0, max: 59},
	{name: "hour", min: 0, max: 23},
	{name: "day-of-month", min: 1, max: 31},
	{name: "month", min: 1, max: 12, names: monthNames},
	{name: "day-of-week", min: 0, max: 7, names: dayNames},
}

// cronValue resolves one cron field value to a number, taking either a literal
// or one of the three-letter names the field accepts. GitHub takes "MON-FRI"
// and "JAN", which a plain strconv.Atoi refused.
func cronValue(s string, spec cronFieldSpec) (int, error) {
	if n, found := spec.names[strings.ToUpper(s)]; found {
		return n, nil
	}

	n, err := strconv.Atoi(s)

	if err != nil {
		return 0, fmt.Errorf("invalid value %q", s)
	}

	return n, nil
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
		return validateCronStep(field[2:], field)
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

		if err := validateCronStep(step, field); err != nil {
			return err
		}
	}

	bounds := strings.SplitN(rangePart, "-", 2)

	if len(bounds) != 2 {
		return fmt.Errorf("invalid range %q", field)
	}

	low, err := cronValue(bounds[0], spec)
	if err != nil {
		return fmt.Errorf("invalid range start %q in %q", bounds[0], field)
	}

	high, err := cronValue(bounds[1], spec)
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

// validateCronStep checks the "/n" of a stepped field.
//
// A step is how far to jump, not a point in the field, so it is a positive
// count and nothing else: no name belongs in it, and its own range is not the
// field's. Checking it against the field's range let "*/0" through, which never
// advances and which GitHub rejects, while turning away "*/90", which only ever
// fires on the first value and which GitHub takes. A range step was checked for
// being an integer alone, so "0-23/0" and "0-23/-5" both went out.
func validateCronStep(step, field string) error {
	n, err := strconv.Atoi(step)

	if err != nil {
		return fmt.Errorf("invalid step %q in %q", step, field)
	}

	if n < 1 {
		return fmt.Errorf("step %q in %q must be a positive number", step, field)
	}

	return nil
}

func validateCronNumber(s string, spec cronFieldSpec) error {
	n, err := cronValue(s, spec)
	if err != nil {
		return err
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
