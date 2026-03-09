// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"fmt"

	"github.com/yldio/cinzel/provider/github/action"
	ghjob "github.com/yldio/cinzel/provider/github/job"
	ghworkflow "github.com/yldio/cinzel/provider/github/workflow"
)

// ValidationPathError wraps a validation error with the path to the offending element.
type ValidationPathError struct {
	Path string
	Err  error
}

// Error returns the path-prefixed error message.
func (e ValidationPathError) Error() string {

	return fmt.Sprintf("%s: %v", e.Path, e.Err)
}

// Unwrap returns the underlying error.
func (e ValidationPathError) Unwrap() error {

	return e.Err
}

func withPath(path string, err error) error {

	if err == nil {

		return nil
	}

	return ValidationPathError{Path: path, Err: err}
}

func validateParsedWorkflow(workflow ghworkflow.Parsed) error {
	onRaw, hasOn := workflow.Body["on"]
	onMap, _ := toStringAnyMap(onRaw)

	model := ghworkflow.ValidationModel{
		ID:      workflow.ID,
		HasOn:   hasOn,
		OnCount: len(onMap),
		JobRefs: workflow.JobRefs,
	}

	if err := ghworkflow.ValidateModel(model); err != nil {

		return withPath("workflow."+workflow.ID, err)
	}

	// Validate workflow-level permissions.

	if perms, ok := workflow.Body["permissions"]; ok {

		if err := ghworkflow.ValidatePermissions(perms); err != nil {

			return withPath("workflow."+workflow.ID+".permissions", err)
		}
	}

	// Validate schedule cron expressions.

	if schedule, ok := onMap["schedule"]; ok {

		if scheduleMap, mapOK := toStringAnyMap(schedule); mapOK {

			if err := ghworkflow.ValidateSchedule(scheduleMap); err != nil {

				return withPath("workflow."+workflow.ID+".on.schedule", err)
			}
		}
	}

	// Validate ${{ }} expression syntax.

	if err := validateExpressions(workflow.Body); err != nil {

		return err
	}

	return nil
}

func validateParsedJobs(jobs map[string]ghjob.Parsed) error {
	models := make(map[string]ghjob.ValidationModel, len(jobs))

	for id, job := range jobs {
		model, err := ghjob.ModelFromParsed(job)
		if err != nil {

			return withPath("job."+id, err)
		}

		if err := ghjob.ValidateModel(model, "runs_on"); err != nil {

			return withPath("job."+id, err)
		}

		// Validate job-level permissions.

		if perms, ok := job.Body["permissions"]; ok {

			if err := ghworkflow.ValidatePermissions(perms); err != nil {

				return withPath("job."+id+".permissions", err)
			}
		}

		// Validate job-level uses (reusable workflow reference).

		if model.Uses != "" {

			if err := action.ValidateUsesRef(model.Uses); err != nil {

				return withPath("job."+id+".uses", err)
			}
		}

		models[id] = model
	}

	for id, model := range models {

		if err := ghjob.ValidateNeedsReferences(model.Needs, models); err != nil {

			return withPath("job."+id+".needs", err)
		}
	}

	if err := ghjob.ValidateNeedsCycles(models); err != nil {

		return withPath("jobs.needs", err)
	}

	return nil
}

func validateWorkflowYAMLDoc(doc ghworkflow.YAMLDocument) error {

	if err := validateAllowedYAMLKeys("workflow_yaml", doc.Raw, allowedWorkflowYAMLKeys); err != nil {

		return withPath("workflow_yaml", err)
	}

	hasOn := doc.HasOn
	jobsRaw := doc.Jobs

	if jobsRaw == nil {

		if hasOn {

			return withPath("workflow_yaml", errWorkflowYAMLOnJobs)
		}

		return nil
	}

	workflowModel := ghworkflow.ValidationModel{
		HasOn:   hasOn,
		OnCount: len(doc.On),
		JobRefs: sortedKeys(jobsRaw),
	}

	if err := ghworkflow.ValidateModel(workflowModel); err != nil {

		return withPath("workflow_yaml", err)
	}

	// Validate workflow-level permissions.

	if perms, ok := doc.Raw["permissions"]; ok {

		if err := ghworkflow.ValidatePermissions(perms); err != nil {

			return withPath("workflow_yaml.permissions", err)
		}
	}

	// Validate schedule cron expressions.

	if schedule, ok := doc.On["schedule"]; ok {

		if scheduleMap, mapOK := toStringAnyMap(schedule); mapOK {

			if err := ghworkflow.ValidateSchedule(scheduleMap); err != nil {

				return withPath("workflow_yaml.on.schedule", err)
			}
		}
	}

	// Validate ${{ }} expression syntax across the entire workflow.

	if err := validateExpressions(doc.Raw); err != nil {

		return err
	}

	jobModels := make(map[string]ghjob.ValidationModel, len(jobsRaw))

	for jobID, jobAny := range jobsRaw {
		jobMap, ok := toStringAnyMap(jobAny)

		if !ok {

			return withPath("jobs."+jobID, fmt.Errorf("must be an object"))
		}

		if err := validateAllowedYAMLKeys("jobs."+jobID, jobMap, allowedJobYAMLKeys); err != nil {

			return withPath("jobs."+jobID, err)
		}

		stepsRaw, hasSteps := jobMap["steps"]

		if hasSteps {
			steps, isList := stepsRaw.([]any)

			if isList {

				for i, stepAny := range steps {
					stepMap, isMap := toStringAnyMap(stepAny)

					if !isMap {
						continue
					}

					if err := validateAllowedYAMLKeys(fmt.Sprintf("jobs.%s.steps[%d]", jobID, i), stepMap, allowedStepYAMLKeys); err != nil {

						return withPath(fmt.Sprintf("jobs.%s.steps[%d]", jobID, i), err)
					}
				}
			}
		}

		model, err := ghjob.ModelFromYAML(jobID, jobMap)
		if err != nil {

			return withPath("jobs."+jobID, err)
		}

		if err := ghjob.ValidateModel(model, "runs-on"); err != nil {

			return withPath("jobs."+jobID, err)
		}

		// Validate job-level permissions.

		if perms, ok := jobMap["permissions"]; ok {

			if err := ghworkflow.ValidatePermissions(perms); err != nil {

				return withPath("jobs."+jobID+".permissions", err)
			}
		}

		// Validate step uses references.

		if err := validateJobStepUses(jobID, jobMap); err != nil {

			return err
		}

		jobModels[jobID] = model
	}

	for id, model := range jobModels {

		if err := ghjob.ValidateNeedsReferences(model.Needs, jobModels); err != nil {

			return withPath("jobs."+id+".needs", err)
		}
	}

	if err := ghjob.ValidateNeedsCycles(jobModels); err != nil {

		return withPath("jobs.needs", err)
	}

	return nil
}

func validateJobStepUses(jobID string, jobMap map[string]any) error {
	stepsRaw, ok := jobMap["steps"]

	if !ok {

		return nil
	}

	steps, ok := stepsRaw.([]any)

	if !ok {

		return nil
	}

	for i, stepRaw := range steps {
		stepMap, ok := toStringAnyMap(stepRaw)

		if !ok {
			continue
		}

		usesRaw, ok := stepMap["uses"]

		if !ok {
			continue
		}

		uses, ok := usesRaw.(string)

		if !ok {
			continue
		}

		if err := action.ValidateUsesRef(uses); err != nil {

			return withPath(fmt.Sprintf("jobs.%s.steps[%d].uses", jobID, i), err)
		}
	}

	return nil
}
