// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"

	yaml "github.com/goccy/go-yaml"
	"github.com/yldio/cinzel/provider/github/action"
	ghjob "github.com/yldio/cinzel/provider/github/job"
	ghworkflow "github.com/yldio/cinzel/provider/github/workflow"
)

type workflowYAMLShape struct {
	Name        any                     `yaml:"name,omitempty"`
	RunName     any                     `yaml:"run-name,omitempty"`
	On          any                     `yaml:"on,omitempty"`
	Jobs        map[string]jobYAMLShape `yaml:"jobs,omitempty"`
	Permissions any                     `yaml:"permissions,omitempty"`
	Defaults    any                     `yaml:"defaults,omitempty"`
	Concurrency any                     `yaml:"concurrency,omitempty"`
	Env         any                     `yaml:"env,omitempty"`
}

type jobYAMLShape struct {
	Name            any             `yaml:"name,omitempty"`
	If              any             `yaml:"if,omitempty"`
	Uses            any             `yaml:"uses,omitempty"`
	With            any             `yaml:"with,omitempty"`
	Secrets         any             `yaml:"secrets,omitempty"`
	Permissions     any             `yaml:"permissions,omitempty"`
	Defaults        any             `yaml:"defaults,omitempty"`
	Concurrency     any             `yaml:"concurrency,omitempty"`
	Container       any             `yaml:"container,omitempty"`
	Services        any             `yaml:"services,omitempty"`
	Environment     any             `yaml:"environment,omitempty"`
	Strategy        any             `yaml:"strategy,omitempty"`
	RunsOn          any             `yaml:"runs-on,omitempty"`
	Steps           []stepYAMLShape `yaml:"steps,omitempty"`
	Needs           any             `yaml:"needs,omitempty"`
	TimeoutMinutes  any             `yaml:"timeout-minutes,omitempty"`
	ContinueOnError any             `yaml:"continue-on-error,omitempty"`
	Outputs         any             `yaml:"outputs,omitempty"`
	Env             any             `yaml:"env,omitempty"`
}

type stepYAMLShape struct {
	ID              any `yaml:"id,omitempty"`
	Name            any `yaml:"name,omitempty"`
	If              any `yaml:"if,omitempty"`
	Uses            any `yaml:"uses,omitempty"`
	Run             any `yaml:"run,omitempty"`
	Shell           any `yaml:"shell,omitempty"`
	WorkingDir      any `yaml:"working-directory,omitempty"`
	With            any `yaml:"with,omitempty"`
	Env             any `yaml:"env,omitempty"`
	ContinueOnError any `yaml:"continue-on-error,omitempty"`
	TimeoutMinutes  any `yaml:"timeout-minutes,omitempty"`
}

type actionYAMLShape struct {
	Name        any                         `yaml:"name,omitempty"`
	Description any                         `yaml:"description,omitempty"`
	Author      any                         `yaml:"author,omitempty"`
	Inputs      map[string]actionInputYAML  `yaml:"inputs,omitempty"`
	Outputs     map[string]actionOutputYAML `yaml:"outputs,omitempty"`
	Runs        actionRunsYAML              `yaml:"runs"`
	Branding    actionBrandingYAML          `yaml:"branding,omitempty"`
}

type actionInputYAML struct {
	Description        any `yaml:"description,omitempty"`
	Required           any `yaml:"required,omitempty"`
	Default            any `yaml:"default,omitempty"`
	DeprecationMessage any `yaml:"deprecationMessage,omitempty"`
}

type actionOutputYAML struct {
	Description any `yaml:"description,omitempty"`
	Value       any `yaml:"value,omitempty"`
}

type actionRunsYAML struct {
	Using      any `yaml:"using,omitempty"`
	Main       any `yaml:"main,omitempty"`
	Pre        any `yaml:"pre,omitempty"`
	PreIf      any `yaml:"pre-if,omitempty"`
	Post       any `yaml:"post,omitempty"`
	PostIf     any `yaml:"post-if,omitempty"`
	Image      any `yaml:"image,omitempty"`
	Args       any `yaml:"args,omitempty"`
	Entrypoint any `yaml:"entrypoint,omitempty"`
	Steps      any `yaml:"steps,omitempty"`
	Env        any `yaml:"env,omitempty"`
}

type actionBrandingYAML struct {
	Icon  any `yaml:"icon,omitempty"`
	Color any `yaml:"color,omitempty"`
}

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
	if err := strictValidateYAMLShape(doc.Raw, &workflowYAMLShape{}); err != nil {
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

func strictValidateYAMLShape(raw map[string]any, target any) error {
	content, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}

	if err := yaml.UnmarshalWithOptions(content, target, yaml.Strict()); err != nil {
		return err
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
