package workflow

import (
	"fmt"

	"github.com/yldio/atos/internal/parsers/actions"
)

type OnByFilterListConfig []OnByFilterConfig

type OnInput struct {
	Name        string    `hcl:",label" yaml:"-"`
	Type        string    `hcl:"type,attr"`
	Description *string   `hcl:"description,attr" yaml:"description,omitempty"`
	Default     *string   `hcl:"default,attr" yaml:"default,omitempty"`
	Required    *bool     `hcl:"required,attr"`
	Options     *[]string `hcl:"options,attr" yaml:"options,omitempty"`
}

type OnOutput struct {
	Name        string  `hcl:",label" yaml:"-"`
	Description *string `hcl:"description,attr"`
	Value       *string `hcl:"value,attr"`
}

type OnSecret struct {
	Name        string  `hcl:",label" yaml:"-"`
	Description *string `hcl:"description,attr"`
	Required    *bool   `hcl:"required,attr"`
}

type OnByFilterConfig struct {
	Event          string      `hcl:"event,attr"`
	Filter         *string     `hcl:"filter,attr"`
	Values         *[]string   `hcl:"values,attr"`
	Input          []*OnInput  `hcl:"input,block"`
	Output         []*OnOutput `hcl:"output,block"`
	Secret         []*OnSecret `hcl:"secret,block"`
	Workflows      *[]string   `hcl:"workflows,attr"`
	Types          *[]string   `hcl:"types,attr"`
	Branches       *[]string   `hcl:"branches,attr"`
	BranchesIgnore *[]string   `hcl:"branches-ignore,attr"`
}

func (config *OnByFilterListConfig) Parse() (map[string]map[string]any, error) {
	// var on = make(map[string]any)
	for _, onByFilter := range *config {
		event := onByFilter.Event
		filterName := *onByFilter.Filter
		switch filterName {
		case "branches", "tags":
			var on = make(map[string]map[string]string)
			filter, err := onByFilter.Parse()
			if err != nil {
				return map[string]map[string]any{}, err
			}

			fmt.Println(event, on, filter)

		default:
			return map[string]map[string]any{}, fmt.Errorf("'%s' is not a valid filter", filterName)
		}
	}

	return map[string]map[string]any{}, nil
}

func (config *OnByFilterConfig) ParseSchedule() ([]map[string]string, error) {
	var cron []map[string]string

	for _, cronSchedule := range *config.Values {
		schedule := map[string]string{
			"cron": cronSchedule,
		}
		cron = append(cron, schedule)
	}

	return cron, nil
}

func (config *OnByFilterConfig) ParseWorkflowCall() (map[string]any, error) {
	inputs := make(map[string]any)
	outputs := make(map[string]any)
	secrets := make(map[string]any)

	if config.Input != nil {
		for _, content := range config.Input {
			inputs[content.Name] = content
		}
	}

	if config.Output != nil {
		for _, content := range config.Output {
			outputs[content.Name] = content
		}
	}

	if config.Secret != nil {
		for _, content := range config.Secret {
			secrets[content.Name] = content
		}
	}

	workflowCall := map[string]any{}

	if len(inputs) > 0 {
		workflowCall["inputs"] = inputs
	}

	if len(outputs) > 0 {
		workflowCall["outputs"] = outputs
	}

	if len(secrets) > 0 {
		workflowCall["secrets"] = secrets
	}

	return workflowCall, nil
}

func (config *OnByFilterConfig) ParseWorkflowRun() (map[string][]string, error) {
	workflows := []string{}
	types := []string{}
	branches := []string{}
	branchesIgnore := []string{}

	workflowRun := map[string][]string{}

	if config.Workflows != nil {
		workflows = append(workflows, *config.Workflows...)
	}

	if config.Types != nil {
		types = append(types, *config.Types...)
	}

	if config.Branches != nil {
		branches = append(branches, *config.Branches...)
	} else if config.BranchesIgnore != nil {
		branchesIgnore = append(branchesIgnore, *config.BranchesIgnore...)
	}

	if len(workflows) > 0 {
		workflowRun["workflows"] = workflows
	}

	if len(types) > 0 {
		workflowRun["types"] = types
	}

	if len(branches) > 0 {
		workflowRun["branches"] = branches
	}

	if len(branchesIgnore) > 0 {
		workflowRun["branchesIgnore"] = branchesIgnore
	}

	return workflowRun, nil
}

func (config *OnByFilterConfig) ParseWorkflowDispatch() (map[string]any, error) {
	inputs := make(map[string]any)

	if config.Input != nil {
		for _, content := range config.Input {
			inputs[content.Name] = content
		}
	}

	workflowDispatch := map[string]any{}

	if len(inputs) > 0 {
		workflowDispatch["inputs"] = inputs
	}

	return workflowDispatch, nil
}

func (config *OnByFilterConfig) ParsePullRequestOrPullRequestTarget() (map[string][]string, error) {
	var event = make(map[string][]string)

	if *config.Filter == "branches" || *config.Filter == "branche-ignore" {
		event[*config.Filter] = append(event[*config.Filter], *config.Values...)
	} else if *config.Filter == "paths" || *config.Filter == "paths-ignore" {
		event[*config.Filter] = append(event[*config.Filter], *config.Values...)
	}

	return event, nil
}

func (config *OnByFilterConfig) ParsePush() (map[string][]string, error) {
	var event = make(map[string][]string)

	if *config.Filter == "branches" || *config.Filter == "branche-ignore" {
		event[*config.Filter] = append(event[*config.Filter], *config.Values...)
	} else if *config.Filter == "tags" || *config.Filter == "tags-ignore" {
		event[*config.Filter] = append(event[*config.Filter], *config.Values...)
	} else if *config.Filter == "paths" || *config.Filter == "paths-ignore" {
		event[*config.Filter] = append(event[*config.Filter], *config.Values...)
	}

	return event, nil
}

func (config *OnByFilterConfig) Parse() (map[string]any, error) {
	if config == nil {
		return make(map[string]any), nil
	}

	var on = make(map[string]any)

	if config.Event == actions.TriggerSchedule.ToString() {
		event, err := config.ParseSchedule()
		if err != nil {
			return map[string]any{}, err
		}

		on[config.Event] = event
	} else if config.Event == actions.TriggerWorkflowCall.ToString() {
		event, err := config.ParseWorkflowCall()
		if err != nil {
			return map[string]any{}, err
		}

		on[config.Event] = event
	} else if config.Event == actions.TriggerWorkflowRun.ToString() {
		event, err := config.ParseWorkflowRun()
		if err != nil {
			return map[string]any{}, err
		}

		on[config.Event] = event
	} else if config.Event == actions.TriggerWorkflowDispatch.ToString() {
		event, err := config.ParseWorkflowDispatch()
		if err != nil {
			return map[string]any{}, err
		}

		on[config.Event] = event
	} else if config.Event == actions.TriggerPullRequest.ToString() || config.Event == actions.TriggerPullRequestTarget.ToString() {
		event, err := config.ParsePullRequestOrPullRequestTarget()
		if err != nil {
			return map[string]any{}, err
		}

		on[config.Event] = event
	} else if config.Event == actions.TriggerPush.ToString() {
		event, err := config.ParsePush()
		if err != nil {
			return map[string]any{}, err
		}

		on[config.Event] = event
	} else if config.Filter != nil {
		var event = make(map[string][]string)
		if *config.Filter == actions.ActivityTypes.ToString() {
			for _, activityType := range *config.Values {

				ok := actions.ValidateActivityType(activityType)
				if !ok {
					return map[string]any{}, fmt.Errorf("activity type '%s' is not valid", activityType)
				}
				event[*config.Filter] = append(event[*config.Filter], activityType)
			}
		} else {
			event[*config.Filter] = append(event[*config.Filter], *config.Values...)
		}
		on[config.Event] = event
	}

	return on, nil
}
