// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package job

import "fmt"

// Parsed holds the intermediate representation of a job after HCL parsing.
type Parsed struct {
	ID       string
	Body     map[string]any
	StepRefs []string
}

// NewParsed creates a Parsed job, extracting step references from the body.
func NewParsed(id string, body map[string]any) Parsed {
	j := Parsed{ID: id, Body: body}

	if refs, ok := body["stepsRefs"].([]string); ok {
		j.StepRefs = refs
		delete(j.Body, "stepsRefs")
	}

	return j
}

// ValidationModel contains the fields needed to validate a job definition.
type ValidationModel struct {
	ID         string
	Uses       string
	HasRunsOn  bool
	StepCount  int
	HasWith    bool
	HasSecrets bool
	Needs      []string
}

// ModelFromParsed builds a ValidationModel from a Parsed job.
func ModelFromParsed(p Parsed) (ValidationModel, error) {
	uses, _ := nonEmptyString(p.Body["uses"])
	needs, err := NeedsFromYAML(p.Body["needs"])
	if err != nil {
		return ValidationModel{}, err
	}

	m := ValidationModel{
		ID:         p.ID,
		Uses:       uses,
		HasRunsOn:  p.Body["runs-on"] != nil,
		HasWith:    p.Body["with"] != nil,
		HasSecrets: p.Body["secrets"] != nil,
		Needs:      needs,
	}

	stepCount, err := stepCountFromRaw(p.Body)
	if err != nil {
		return ValidationModel{}, err
	}
	m.StepCount = stepCount

	return m, nil
}

// ModelFromYAML builds a ValidationModel from raw YAML job data.
func ModelFromYAML(id string, raw map[string]any) (ValidationModel, error) {
	uses, _ := nonEmptyString(raw["uses"])
	needs, err := NeedsFromYAML(raw["needs"])
	if err != nil {
		return ValidationModel{}, err
	}

	m := ValidationModel{
		ID:         id,
		Uses:       uses,
		HasRunsOn:  raw["runs-on"] != nil,
		HasWith:    raw["with"] != nil,
		HasSecrets: raw["secrets"] != nil,
		Needs:      needs,
	}

	stepCount, err := stepCountFromRaw(raw)
	if err != nil {
		return ValidationModel{}, err
	}
	m.StepCount = stepCount

	return m, nil
}

func stepCountFromRaw(raw map[string]any) (int, error) {
	stepsRaw, exists := raw["steps"]

	if !exists || stepsRaw == nil {
		return 0, nil
	}

	steps, ok := stepsRaw.([]any)

	if !ok {
		return 0, fmt.Errorf("'steps' must be a list")
	}

	return len(steps), nil
}
