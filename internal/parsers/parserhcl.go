package parsers

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/yldio/atos/internal/parsers/actions"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type HclConfig struct {
	Workflows []actions.WorkflowConfig `hcl:"workflow,block"`
	Jobs      []actions.JobConfig      `hcl:"job,block"`
	Steps     []actions.StepConfig     `hcl:"step,block"`
}

type HclParser struct {
	hclConfig HclConfig
	steps     []actions.Step
	jobs      []actions.Job
	workflows []actions.Workflow
}

func NewHclParser() *HclParser {
	return &HclParser{}
}

func (parse *HclParser) ParseFiles(bodies []hcl.Body) error {
	ctx := createContext()

	for _, body := range bodies {
		diags := gohcl.DecodeBody(body, ctx, &parse.hclConfig)
		if diags.HasErrors() {
			return errors.New(diags.Error())
		}
	}

	return nil
}

func (parse *HclParser) ParseSteps() error {
	for _, stepConfig := range parse.hclConfig.Steps {
		step := actions.Step{
			Id:   stepConfig.Id,
			Name: *stepConfig.Name,
		}

		if stepConfig.Uses != nil {
			step.Uses = actions.Uses{
				Action:  stepConfig.Uses.Action,
				Version: stepConfig.Uses.Version,
			}
		}

		if stepConfig.Run != nil {
			step.Run = *stepConfig.Run
		}

		parse.steps = append(parse.steps, step)
	}

	return nil
}

func (parse *HclParser) ParseJobs() error {
	for _, jobConfig := range parse.hclConfig.Jobs {
		job := actions.Job{
			Id:   jobConfig.Id,
			Name: jobConfig.Name,
		}

		exprs, diags := hcl.ExprList(jobConfig.Steps)
		if diags.HasErrors() {
			return errors.New(diags.Error())
		}

		for _, expr := range exprs {
			traversal, diags := hcl.AbsTraversalForExpr(expr)
			if diags.HasErrors() {
				return errors.New(diags.Error())
			}

			var nameType string
			var nameId string

			for _, step := range traversal {
				switch tStep := step.(type) {
				case hcl.TraverseRoot:
					nameType = tStep.Name
				case hcl.TraverseAttr:
					nameId = tStep.Name
				}
			}

			if nameType == "step" {
				var stepFound actions.Step
				for _, step := range parse.steps {
					if step.Id == nameId {
						stepFound = step
					}
				}

				if reflect.DeepEqual(stepFound, actions.Step{}) {
					return errors.New("step reference not found")
				}

				job.Steps = append(job.Steps, stepFound)
			} else {
				return errors.New("should contain only step references")
			}

		}

		parse.jobs = append(parse.jobs, job)
	}
	return nil
}

func (parse *HclParser) ParseWorkflows() error {
	for _, workflowConfig := range parse.hclConfig.Workflows {
		workflow := actions.Workflow{
			Id: workflowConfig.Id,
		}

		if workflowConfig.On != nil {
			workflow.On = workflowConfig.On
		} else if workflowConfig.OnAsList != nil {
			for _, eventTrigger := range *workflowConfig.OnAsList {
				ok := actions.ValidateEventTrigger(eventTrigger)
				if !ok {
					return fmt.Errorf("event trigger '%s' is not valid", eventTrigger)
				}
			}
			workflow.On = workflowConfig.OnAsList
		} else if workflowConfig.OnByFilter != nil {
			var on = make(map[string]any)

			for _, onByFilter := range workflowConfig.OnByFilter {
				var filter = make(map[string][]string)
				if onByFilter.Filter != nil {
					if *onByFilter.Filter == actions.ActivityTypes.ToString() {
						for _, activityType := range *onByFilter.Values {

							ok := actions.ValidateActivityType(activityType)
							if !ok {
								return fmt.Errorf("activity type '%s' is not valid", activityType)
							}
							filter[*onByFilter.Filter] = append(filter[*onByFilter.Filter], activityType)
						}
					} else {
						filter[*onByFilter.Filter] = append(filter[*onByFilter.Filter], *onByFilter.Values...)
					}
				}
				on[onByFilter.Event] = filter
			}

			workflow.On = on
		}

		if workflowConfig.Name != nil {
			workflow.Name = *workflowConfig.Name
		}

		if workflowConfig.RunName != nil {
			workflow.RunName = *workflowConfig.RunName
		}

		for _, envConfig := range workflowConfig.Envs {
			env, err := envConfig.ConvertFromHcl()
			if err != nil {
				return err
			}
			workflow.Envs = append(workflow.Envs, env)
		}

		permissions, err := workflowConfig.Permissions.ConvertFromHcl()
		if err != nil {
			return err
		}

		workflow.Permissions = permissions

		defaults, err := workflowConfig.Defaults.ConvertFromHcl()
		if err != nil {
			return err
		}
		workflow.Defaults = defaults

		concurrency, err := workflowConfig.Concurrency.ConvertFromHcl()
		if err != nil {
			return err
		}
		workflow.Concurrency = concurrency

		exprs, diags := hcl.ExprList(workflowConfig.Jobs)
		if diags.HasErrors() {
			return errors.New(diags.Error())
		}

		for _, expr := range exprs {
			traversal, diags := hcl.AbsTraversalForExpr(expr)
			if diags.HasErrors() {
				return errors.New(diags.Error())
			}

			var nameType string
			var nameId string
			for _, step := range traversal {
				switch tStep := step.(type) {
				case hcl.TraverseRoot:
					nameType = tStep.Name
				case hcl.TraverseAttr:
					nameId = tStep.Name
				}
			}

			if nameType == "job" {
				var jobFound actions.Job
				for _, job := range parse.jobs {
					if job.Id == nameId {
						jobFound = job
					}
				}

				if reflect.DeepEqual(jobFound, actions.Job{}) {
					return errors.New("job reference not found")
				}

				workflow.Jobs = append(workflow.Jobs, jobFound)
			} else {
				return errors.New("should contain only job references")
			}
		}

		parse.workflows = append(parse.workflows, workflow)
	}

	return nil
}

func (parse *HclParser) Do() error {
	if err := parse.ParseSteps(); err != nil {
		return err
	}

	if err := parse.ParseJobs(); err != nil {
		return err
	}

	if err := parse.ParseWorkflows(); err != nil {
		return err
	}

	return nil
}

func (parse *HclParser) GetContent() []actions.Workflow {
	return parse.workflows
}

func createContext() *hcl.EvalContext {
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: map[string]function.Function{
			"import_script": ImportScript,
		},
	}
	return ctx
}
