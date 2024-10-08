// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
)

type EventType string

const (
	EventTypeBoolean     EventType = "boolean"
	EventTypeChoice      EventType = "choice"
	EventTypeNumber      EventType = "number"
	EventTypeEnvironment EventType = "environment"
	EventTypeString      EventType = "string"
)

func (option EventType) ToString() string {
	return string(option)
}

func ConvertToWorkflowDispatchType(typeValue string) (EventType, error) {
	switch typeValue {
	case EventTypeBoolean.ToString():
		return EventTypeBoolean, nil
	case EventTypeChoice.ToString():
		return EventTypeChoice, nil
	case EventTypeNumber.ToString():
		return EventTypeNumber, nil
	case EventTypeEnvironment.ToString():
		return EventTypeEnvironment, nil
	case EventTypeString.ToString():
		return EventTypeString, nil
	default:
		return "", fmt.Errorf("invalid type '%s'", typeValue)
	}
}

func ConvertToWorkflowCallType(typeValue string) (EventType, error) {
	switch typeValue {
	case EventTypeBoolean.ToString():
		return EventTypeBoolean, nil
	case EventTypeNumber.ToString():
		return EventTypeNumber, nil
	case EventTypeString.ToString():
		return EventTypeString, nil
	default:
		return "", fmt.Errorf("invalid type '%s'", typeValue)
	}
}

type EventsConfig []*EventConfig

type EventInputConfig struct {
	Identifier  string         `hcl:"_,label"`
	Description hcl.Expression `hcl:"description,attr"`
	Default     hcl.Expression `hcl:"default,attr"`
	Required    hcl.Expression `hcl:"required,attr"`
	Type        hcl.Expression `hcl:"type,attr"`
	Options     hcl.Expression `hcl:"options,attr"`
}

type EventOutputConfig struct {
	Identifier  string         `hcl:"_,label"`
	Value       hcl.Expression `hcl:"value,attr"`
	Description hcl.Expression `hcl:"description,attr"`
}

type EventSecretConfig struct {
	Identifier  string         `hcl:"_,label"`
	Required    hcl.Expression `hcl:"required,attr"`
	Description hcl.Expression `hcl:"description,attr"`
}

type Input struct {
	Name        string    `yaml:"-"`
	Type        EventType `yaml:"type"`
	Required    *bool     `yaml:"required,omitempty"`
	Description *string   `yaml:"description,omitempty"`
	Default     any       `yaml:"default,omitempty"`
	Options     []*string `yaml:"options,omitempty"`
}

type Output struct {
	Name        string  `yaml:"-"`
	Value       any     `yaml:"value"`
	Description *string `yaml:"description,omitempty"`
}

type Secret struct {
	Name        string  `yaml:"-"`
	Required    bool    `yaml:"required"`
	Description *string `yaml:"description,omitempty"`
}

type EventConfig struct {
	Identifier     string               `hcl:"_,label"`
	Types          hcl.Expression       `hcl:"types,attr"`
	Branches       hcl.Expression       `hcl:"branches,attr"`
	BranchesIgnore hcl.Expression       `hcl:"branches_ignore,attr"`
	Tags           hcl.Expression       `hcl:"tags,attr"`
	TagsIgnore     hcl.Expression       `hcl:"tags_ignore,attr"`
	Paths          hcl.Expression       `hcl:"paths,attr"`
	PathsIgnore    hcl.Expression       `hcl:"paths_ignore,attr"`
	Cron           hcl.Expression       `hcl:"cron,attr"`
	Inputs         []*EventInputConfig  `hcl:"input,block"`
	Outputs        []*EventOutputConfig `hcl:"output,block"`
	Secrets        []*EventSecretConfig `hcl:"secret,block"`
	Workflows      hcl.Expression       `hcl:"workflows,attr"`
}

type Evt interface{}

type EventSchedule []map[string]string

type EventWorkflowDispatch struct {
	Name   string            `yaml:"-"`
	Inputs *map[string]Input `yaml:"inputs,omitempty"`
}

type EventWorkflowCall struct {
	Name    string             `yaml:"-"`
	Types   []*string          `yaml:"types,omitempty"`
	Inputs  *map[string]Input  `yaml:"inputs,omitempty"`
	Outputs *map[string]Output `yaml:"outputs,omitempty"`
	Secrets *map[string]Secret `yaml:"secrets,omitempty"`
}

type EventWorkflowRun struct {
	Name           string    `yaml:"-"`
	Types          []*string `yaml:"types,omitempty"`
	Workflows      []*string `yaml:"workflows,omitempty"`
	Branches       []*string `yaml:"branches,omitempty"`
	BranchesIgnore []*string `yaml:"branches-ignore,omitempty"`
}

type Event struct {
	Name           string    `yaml:"-"`
	Types          []*string `yaml:"types,omitempty"`
	Branches       []*string `yaml:"branches,omitempty"`
	BranchesIgnore []*string `yaml:"branches-ignore,omitempty"`
	Tags           []*string `yaml:"tags,omitempty"`
	TagsIgnore     []*string `yaml:"tags-ignore,omitempty"`
	Paths          []*string `yaml:"paths,omitempty"`
	PathsIgnore    []*string `yaml:"paths-ignore,omitempty"`
}

type On map[EventTrigger]Evt

var (
	ErrEventTriggerNoMoreThanOne = errEventTriggerNoMoreThanOne()
	ErrEventTriggerUnknown       = errEventTriggerUnknown()
)

func errEventTriggerNoMoreThanOne() error {
	return errors.New("can't define more than one event trigger")
}

func errEventTriggerUnknown() error {
	return errors.New("unkown event trigger")
}

func (config *EventConfig) hasInputs() (bool, error) {
	return len(config.Inputs) > 0, nil
}

func (config *EventConfig) hasOutputs() (bool, error) {
	return len(config.Outputs) > 0, nil
}

func (config *EventConfig) hasSecrets() (bool, error) {
	return len(config.Secrets) > 0, nil
}

func (config *EventConfig) hasPaths() (bool, error) {
	val, diags := config.Paths.Value(nil)

	if diags.HasErrors() {
		return false, diags
	}

	return !val.IsNull(), nil
}

func (config *EventConfig) hasPathsIgnore() (bool, error) {
	val, diags := config.PathsIgnore.Value(nil)

	if diags.HasErrors() {
	}

	return !val.IsNull(), nil
}

func (config *EventConfig) hasTags() (bool, error) {
	val, diags := config.Tags.Value(nil)

	if diags.HasErrors() {
	}

	return !val.IsNull(), nil
}

func (config *EventConfig) hasTagsIgnore() (bool, error) {
	val, diags := config.TagsIgnore.Value(nil)

	if diags.HasErrors() {
	}

	return !val.IsNull(), nil
}

func (config *EventConfig) hasBranches() (bool, error) {
	val, diags := config.Branches.Value(nil)

	if diags.HasErrors() {
	}

	return !val.IsNull(), nil
}

func (config *EventConfig) hasBranchesIgnore() (bool, error) {
	val, diags := config.BranchesIgnore.Value(nil)

	if diags.HasErrors() {
	}

	return !val.IsNull(), nil
}

func (config *EventConfig) hasCron() (bool, error) {
	val, diags := config.Cron.Value(nil)

	if diags.HasErrors() {
	}

	return !val.IsNull(), nil
}

func (config *EventConfig) hasWorkflows() (bool, error) {
	val, diags := config.Workflows.Value(nil)

	if diags.HasErrors() {
	}

	return !val.IsNull(), nil
}

func (config *EventConfig) unwrapWorkflows(acto *actoparser.Acto) ([]*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return []*string{&resultValue}, nil
	case []string:
		var types []*string

		for _, valType := range resultValue {
			types = append(types, &valType)
		}

		return types, nil
	case []actoparser.ActoVariableRef:
		list := []*string{}
		for _, val := range resultValue {
			v, err := config.unwrapWorkflows(actoparser.NewActoFromResult(val))
			if err != nil {
				return nil, err
			}

			list = append(list, v...)
		}

		return list, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapWorkflows(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'workflows' must be a list of strings")
	}
}

func (config *EventConfig) parseWorkflows() ([]*string, error) {
	acto := actoparser.NewActo(config.Workflows)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapWorkflows(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventSecretConfig) unwrapRequired(acto *actoparser.Acto) (*bool, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'required' must be a boolean")
	case bool:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapRequired(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'required' must be a boolean")
	}
}

func (config *EventSecretConfig) parseRequired() (*bool, error) {
	acto := actoparser.NewActo(config.Required)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapRequired(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventSecretConfig) unwrapDescription(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapDescription(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'description' must to be a string")
	}
}

func (config *EventSecretConfig) parseDescription() (*string, error) {
	acto := actoparser.NewActo(config.Description)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapDescription(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventOutputConfig) unwrapValue(acto *actoparser.Acto) (any, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'value' must be value a number, string or boolean")
	case string:
		return &resultValue, nil
	case uint64:
		return &resultValue, nil
	case int64:
		return &resultValue, nil
	case float64:
		return &resultValue, nil
	case bool:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapValue(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'value' must be value a number, string or boolean")
	}
}

func (config *EventOutputConfig) parseValue() (any, error) {
	acto := actoparser.NewActo(config.Value)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapValue(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventOutputConfig) unwrapDescription(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapDescription(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'description' must to be a string")
	}
}

func (config *EventOutputConfig) parseDescription() (*string, error) {
	acto := actoparser.NewActo(config.Description)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapDescription(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventInputConfig) unwrapDefault(acto *actoparser.Acto) (any, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return &resultValue, nil
	case uint64:
		return &resultValue, nil
	case int64:
		return &resultValue, nil
	case float64:
		return &resultValue, nil
	case bool:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapDefault(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'default' must be value a number, string or boolean")
	}
}

func (config *EventInputConfig) parseDefault() (any, error) {
	acto := actoparser.NewActo(config.Default)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapDefault(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventInputConfig) unwrapDescription(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapDescription(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'description' must to be a string")
	}
}

func (config *EventInputConfig) parseDescription() (*string, error) {
	acto := actoparser.NewActo(config.Description)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapDescription(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventInputConfig) unwrapWorkflowDispatchType(acto *actoparser.Acto) (*EventType, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'type' must to be a string ('boolean', 'choice', 'number', 'environment' or 'string')")
	case string:
		typeValue, err := ConvertToWorkflowDispatchType(resultValue)
		if err != nil {
			return nil, err
		}

		return &typeValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapWorkflowDispatchType(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'type' must to be a string ('boolean', 'choice', 'number', 'environment' or 'string')")
	}
}

func (config *EventInputConfig) parseWorkflowDispatchType() (*EventType, error) {
	acto := actoparser.NewActo(config.Type)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapWorkflowDispatchType(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventInputConfig) unwrapWorkflowCallType(acto *actoparser.Acto) (*EventType, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'type' must to be a string ('boolean', 'number', or 'string')")
	case string:
		typeValue, err := ConvertToWorkflowCallType(resultValue)
		if err != nil {
			return nil, err
		}

		return &typeValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapWorkflowCallType(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'type' must to be a string ('boolean', 'number', or 'string')")
	}
}

func (config *EventInputConfig) parseWorkflowCallType() (*EventType, error) {
	acto := actoparser.NewActo(config.Type)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapWorkflowCallType(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventInputConfig) unwrapRequired(acto *actoparser.Acto) (*bool, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case bool:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapRequired(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'required' must be a boolean")
	}
}

func (config *EventInputConfig) parseRequired() (*bool, error) {
	acto := actoparser.NewActo(config.Required)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapRequired(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventInputConfig) parseWorkflowDispatch() (Input, error) {
	input := Input{
		Name: config.Identifier,
	}

	valueType, err := config.parseWorkflowCallType()
	if err != nil {
		return Input{}, err
	}

	input.Type = *valueType

	valueRequired, err := config.parseRequired()
	if err != nil {
		return Input{}, err
	}

	if valueRequired != nil {
		input.Required = valueRequired
	}

	valueDescription, err := config.parseDescription()
	if err != nil {
		return Input{}, err
	}

	if valueDescription != nil {
		input.Description = valueDescription
	}

	valueDefault, err := config.parseDefault()
	if err != nil {
		return Input{}, err
	}

	if valueDefault != nil {
		input.Default = valueDefault
	}

	return input, nil
}

func (config *EventInputConfig) parseWorkflowCall() (Input, error) {
	input := Input{
		Name: config.Identifier,
	}

	valueType, err := config.parseWorkflowCallType()
	if err != nil {
		return Input{}, err
	}

	input.Type = *valueType

	valueRequired, err := config.parseRequired()
	if err != nil {
		return Input{}, err
	}

	if valueRequired != nil {
		input.Required = valueRequired
	}

	valueDescription, err := config.parseDescription()
	if err != nil {
		return Input{}, err
	}

	if valueDescription != nil {
		input.Description = valueDescription
	}

	valueDefault, err := config.parseDefault()
	if err != nil {
		return Input{}, err
	}

	if valueDefault != nil {
		input.Default = valueDefault
	}

	return input, nil
}

func (config *EventOutputConfig) parse() (Output, error) {
	output := Output{
		Name: config.Identifier,
	}

	valueValue, err := config.parseValue()
	if err != nil {
		return Output{}, err
	}

	output.Value = valueValue

	valueDescription, err := config.parseDescription()
	if err != nil {
		return Output{}, err
	}

	if valueDescription != nil {
		output.Description = valueDescription
	}

	return output, nil
}

func (config *EventSecretConfig) parse() (Secret, error) {
	secret := Secret{
		Name: config.Identifier,
	}

	valueRequired, err := config.parseRequired()
	if err != nil {
		return Secret{}, err
	}

	secret.Required = *valueRequired

	valueDescription, err := config.parseDescription()
	if err != nil {
		return Secret{}, err
	}

	if valueDescription != nil {
		secret.Description = valueDescription
	}

	return secret, nil
}

func (config *EventConfig) parseWorkflowDispatchInputs() (*map[string]Input, error) {
	if config.Inputs == nil {
		return nil, nil
	}

	inputs := make(map[string]Input)

	for _, i := range config.Inputs {
		input, err := i.parseWorkflowDispatch()
		if err != nil {
			return nil, nil
		}

		if !reflect.DeepEqual(inputs[input.Name], Input{}) {
			return nil, fmt.Errorf("block 'input' '%s' must not be empty", input.Name)
		}

		inputs[input.Name] = input
	}

	return &inputs, nil
}

func (config *EventConfig) parseWorkflowCallInputs() (*map[string]Input, error) {
	if config.Inputs == nil {
		return nil, nil
	}

	inputs := make(map[string]Input)

	for _, i := range config.Inputs {
		input, err := i.parseWorkflowCall()
		if err != nil {
			return nil, nil
		}

		if !reflect.DeepEqual(inputs[input.Name], Input{}) {
			return nil, fmt.Errorf("block 'input' '%s' must not be empty", input.Name)
		}

		inputs[input.Name] = input
	}

	return &inputs, nil
}

func (config *EventConfig) parseOutputs() (*map[string]Output, error) {
	if config.Outputs == nil {
		return nil, nil
	}

	outputs := make(map[string]Output)

	for _, o := range config.Outputs {
		output, err := o.parse()
		if err != nil {
			return nil, nil
		}

		if !reflect.DeepEqual(outputs[output.Name], Output{}) {
			return nil, fmt.Errorf("block 'ouput' '%s' must not be empty", output.Name)
		}

		outputs[output.Name] = output
	}

	return &outputs, nil
}

func (config *EventConfig) parseSecrets() (*map[string]Secret, error) {
	if config.Secrets == nil {
		return nil, nil
	}

	secrets := make(map[string]Secret)

	for _, s := range config.Secrets {
		secret, err := s.parse()
		if err != nil {
			return nil, nil
		}

		if !reflect.DeepEqual(secrets[secret.Name], Secret{}) {
			return nil, fmt.Errorf("block 'secret' '%s' must not be empty", secret.Name)
		}

		secrets[secret.Name] = secret
	}

	return &secrets, nil
}

func (config *EventConfig) unwrapCron(acto *actoparser.Acto) (*[]map[string]string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		crons := []map[string]string{}
		cron := make(map[string]string)
		cron["cron"] = resultValue
		crons = append(crons, cron)

		return &crons, nil
	case []string:
		crons := []map[string]string{}

		for _, valType := range resultValue {
			cron := make(map[string]string)
			cron["cron"] = valType
			crons = append(crons, cron)
		}

		return &crons, nil
	case []actoparser.ActoVariableRef:
		crons := []map[string]string{}
		for _, val := range resultValue {
			v, err := config.unwrapBranches(actoparser.NewActoFromResult(val))
			if err != nil {
				return nil, err
			}

			for _, valType := range v {
				cron := make(map[string]string)
				cron["cron"] = *valType
				crons = append(crons, cron)
			}
		}

		return &crons, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapCron(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'cron' must be a list of strings")
	}
}

func (config *EventConfig) parseCron() (*[]map[string]string, error) {
	acto := actoparser.NewActo(config.Cron)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapCron(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventConfig) unwrapPathsIgnore(acto *actoparser.Acto) ([]*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return []*string{&resultValue}, nil
	case []string:
		var types []*string

		for _, valType := range resultValue {
			types = append(types, &valType)
		}

		return types, nil
	case []actoparser.ActoVariableRef:
		list := []*string{}
		for _, val := range resultValue {
			v, err := config.unwrapPathsIgnore(actoparser.NewActoFromResult(val))
			if err != nil {
				return nil, err
			}

			list = append(list, v...)
		}

		return list, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapPathsIgnore(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'paths_ignore' must be a list of strings")
	}
}

func (config *EventConfig) parsePathsIgnore() ([]*string, error) {
	acto := actoparser.NewActo(config.PathsIgnore)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapPathsIgnore(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventConfig) unwrapPaths(acto *actoparser.Acto) ([]*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return []*string{&resultValue}, nil
	case []string:
		var types []*string

		for _, valType := range resultValue {
			types = append(types, &valType)
		}

		return types, nil
	case []actoparser.ActoVariableRef:
		list := []*string{}
		for _, val := range resultValue {
			v, err := config.unwrapPaths(actoparser.NewActoFromResult(val))
			if err != nil {
				return nil, err
			}

			list = append(list, v...)
		}

		return list, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapPaths(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'paths' must be a list of strings")
	}
}

func (config *EventConfig) parsePaths() ([]*string, error) {
	acto := actoparser.NewActo(config.Paths)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapPaths(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventConfig) unwrapTagsIgnore(acto *actoparser.Acto) ([]*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return []*string{&resultValue}, nil
	case []string:
		var types []*string

		for _, valType := range resultValue {
			types = append(types, &valType)
		}

		return types, nil
	case []actoparser.ActoVariableRef:
		list := []*string{}
		for _, val := range resultValue {
			v, err := config.unwrapTagsIgnore(actoparser.NewActoFromResult(val))
			if err != nil {
				return nil, err
			}

			list = append(list, v...)
		}

		return list, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapTagsIgnore(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'tags_ignore' must be a list of strings")
	}
}

func (config *EventConfig) parseTagsIgnore() ([]*string, error) {
	acto := actoparser.NewActo(config.TagsIgnore)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapTagsIgnore(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventConfig) unwrapTags(acto *actoparser.Acto) ([]*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return []*string{&resultValue}, nil
	case []string:
		var types []*string

		for _, valType := range resultValue {
			types = append(types, &valType)
		}

		return types, nil
	case []actoparser.ActoVariableRef:
		list := []*string{}
		for _, val := range resultValue {
			v, err := config.unwrapTags(actoparser.NewActoFromResult(val))
			if err != nil {
				return nil, err
			}

			list = append(list, v...)
		}

		return list, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapTags(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'tags' must be a list of strings")
	}
}

func (config *EventConfig) parseTags() ([]*string, error) {
	acto := actoparser.NewActo(config.Tags)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapTags(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventConfig) unwrapBranchesIgnore(acto *actoparser.Acto) ([]*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return []*string{&resultValue}, nil
	case []string:
		var types []*string

		for _, valType := range resultValue {
			types = append(types, &valType)
		}

		return types, nil
	case []actoparser.ActoVariableRef:
		list := []*string{}
		for _, val := range resultValue {
			v, err := config.unwrapBranchesIgnore(actoparser.NewActoFromResult(val))
			if err != nil {
				return nil, err
			}

			list = append(list, v...)
		}

		return list, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapBranchesIgnore(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'branches_ignore' must be a list of strings")
	}
}

func (config *EventConfig) parseBranchesIgnore() ([]*string, error) {
	acto := actoparser.NewActo(config.BranchesIgnore)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapBranchesIgnore(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventConfig) unwrapBranches(acto *actoparser.Acto) ([]*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return []*string{&resultValue}, nil
	case []string:
		var types []*string

		for _, valType := range resultValue {
			types = append(types, &valType)
		}

		return types, nil
	case []actoparser.ActoVariableRef:
		list := []*string{}
		for _, val := range resultValue {
			v, err := config.unwrapBranches(actoparser.NewActoFromResult(val))
			if err != nil {
				return nil, err
			}

			list = append(list, v...)
		}

		return list, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapBranches(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'branches' must be a list of strings")
	}
}

func (config *EventConfig) parseBranches() ([]*string, error) {
	acto := actoparser.NewActo(config.Branches)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapBranches(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EventConfig) unwrapTypes(acto *actoparser.Acto) ([]*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case []string:
		var types []*string

		for _, valType := range resultValue {
			types = append(types, &valType)
		}
		return types, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapTypes(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'types' must be a list of strings")
	}
}

func (config *EventConfig) parseTypes() ([]*string, error) {
	acto := actoparser.NewActo(config.Types)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapTypes(acto)
	if err != nil {
		return nil, err
	}

	for _, at := range value {
		eventTrigger := EventTrigger(config.Identifier)
		activityType := ActivityType(*at)

		if !ValidateActivityType(*at) {
			return nil, fmt.Errorf("in event trigger '%s': unknown activity type '%s'", eventTrigger, activityType)
		}

		if !ValidateActivityTypeForEventTrigger(eventTrigger, activityType) {
			return nil, fmt.Errorf("in event trigger '%s': not a valid activity type '%s'", eventTrigger, activityType)
		}
	}

	return value, nil
}

func (config *EventConfig) parsePullRequest() (*Event, error) {
	hasTags, err := config.hasTags()
	if err != nil {
		return nil, err
	}

	if hasTags {
		return nil, fmt.Errorf("attribute 'tags' is not allowed")
	}

	hasTagsIgnore, err := config.hasTagsIgnore()
	if err != nil {
		return nil, err
	}

	if hasTagsIgnore {
		return nil, fmt.Errorf("attribute 'tags_ignore' is not allowed")
	}

	hasCron, err := config.hasCron()
	if err != nil {
		return nil, err
	}

	if hasCron {
		return nil, fmt.Errorf("attribute 'cron' is not allowed")
	}

	hasWorkflows, err := config.hasWorkflows()
	if err != nil {
		return nil, err
	}

	if hasWorkflows {
		return nil, fmt.Errorf("attribute 'workflows' is not allowed")
	}

	hasInputs, err := config.hasInputs()
	if err != nil {
		return nil, err
	}

	if hasInputs {
		return nil, fmt.Errorf("attribute 'inputs' is not allowed")
	}

	hasOutputs, err := config.hasOutputs()
	if err != nil {
		return nil, err
	}

	if hasOutputs {
		return nil, fmt.Errorf("attribute 'outputs' is not allowed")
	}

	hasSecrets, err := config.hasSecrets()
	if err != nil {
		return nil, err
	}

	if hasSecrets {
		return nil, fmt.Errorf("attribute 'secrets' is not allowed")
	}

	event := Event{
		Name: config.Identifier,
	}

	types, err := config.parseTypes()
	if err != nil {
		return nil, err
	}

	if types != nil {
		event.Types = types
	}

	branches, err := config.parseBranches()
	if err != nil {
		return nil, err
	}

	if branches != nil {
		event.Branches = branches
	}

	branchesIgnore, err := config.parseBranchesIgnore()
	if err != nil {
		return nil, err
	}

	if branchesIgnore != nil {
		event.BranchesIgnore = branchesIgnore
	}

	paths, err := config.parsePaths()
	if err != nil {
		return nil, err
	}

	if paths != nil {
		event.Paths = paths
	}

	pathsIgnore, err := config.parsePathsIgnore()
	if err != nil {
		return nil, err
	}

	if pathsIgnore != nil {
		event.PathsIgnore = pathsIgnore
	}

	return &event, nil
}

func (config *EventConfig) parsePullRequestTarget() (*Event, error) {
	hasTags, err := config.hasTags()
	if err != nil {
		return nil, err
	}

	if hasTags {
		return nil, fmt.Errorf("attribute 'tags' is not allowed")
	}

	hasTagsIgnore, err := config.hasTagsIgnore()
	if err != nil {
		return nil, err
	}

	if hasTagsIgnore {
		return nil, fmt.Errorf("attribute 'tags_ignore' is not allowed")
	}

	hasCron, err := config.hasCron()
	if err != nil {
		return nil, err
	}

	if hasCron {
		return nil, fmt.Errorf("attribute 'cron' is not allowed")
	}

	hasWorkflows, err := config.hasWorkflows()
	if err != nil {
		return nil, err
	}

	if hasWorkflows {
		return nil, fmt.Errorf("attribute 'workflows' is not allowed")
	}

	hasInputs, err := config.hasInputs()
	if err != nil {
		return nil, err
	}

	if hasInputs {
		return nil, fmt.Errorf("attribute 'inputs' is not allowed")
	}

	hasOutputs, err := config.hasOutputs()
	if err != nil {
		return nil, err
	}

	if hasOutputs {
		return nil, fmt.Errorf("attribute 'outputs' is not allowed")
	}

	hasSecrets, err := config.hasSecrets()
	if err != nil {
		return nil, err
	}

	if hasSecrets {
		return nil, fmt.Errorf("attribute 'secrets' is not allowed")
	}

	event := Event{
		Name: config.Identifier,
	}

	types, err := config.parseTypes()
	if err != nil {
		return nil, fmt.Errorf("in event 'pull_request_target': %w", err)
	}

	if types != nil {
		event.Types = types
	}

	branches, err := config.parseBranches()
	if err != nil {
		return nil, fmt.Errorf("in event 'pull_request_target': %w", err)
	}

	if branches != nil {
		event.Branches = branches
	}

	branchesIgnore, err := config.parseBranchesIgnore()
	if err != nil {
		return nil, fmt.Errorf("in event 'pull_request_target': %w", err)
	}

	if branchesIgnore != nil {
		event.BranchesIgnore = branchesIgnore
	}

	paths, err := config.parsePaths()
	if err != nil {
		return nil, fmt.Errorf("in event 'pull_request_target': %w", err)
	}

	if paths != nil {
		event.Paths = paths
	}

	pathsIgnore, err := config.parsePathsIgnore()
	if err != nil {
		return nil, fmt.Errorf("in event 'pull_request_target': %w", err)
	}

	if pathsIgnore != nil {
		event.PathsIgnore = pathsIgnore
	}

	return &event, nil
}

func (config *EventConfig) parsePush() (*Event, error) {
	hasCron, err := config.hasCron()
	if err != nil {
		return nil, err
	}

	if hasCron {
		return nil, fmt.Errorf("attribute 'cron' is not allowed")
	}

	hasWorkflows, err := config.hasWorkflows()
	if err != nil {
		return nil, err
	}

	if hasWorkflows {
		return nil, fmt.Errorf("attribute 'workflows' is not allowed")
	}

	if config.Inputs != nil {
		return nil, fmt.Errorf("attribute 'inputs' is not allowed")
	}

	if config.Outputs != nil {
		return nil, fmt.Errorf("attribute 'outputs' is not allowed")
	}

	if config.Secrets != nil {
		return nil, fmt.Errorf("attribute 'secrets' is not allowed")
	}

	event := Event{
		Name: config.Identifier,
	}

	types, err := config.parseTypes()
	if err != nil {
		return nil, err
	}

	if types != nil {
		event.Types = types
	}

	branches, err := config.parseBranches()
	if err != nil {
		return nil, err
	}

	if branches != nil {
		event.Branches = branches
	}

	branchesIgnore, err := config.parseBranchesIgnore()
	if err != nil {
		return nil, err
	}

	if branchesIgnore != nil {
		event.BranchesIgnore = branchesIgnore
	}

	tags, err := config.parseTags()
	if err != nil {
		return nil, err
	}

	if tags != nil {
		event.Tags = tags
	}

	tagsIgnore, err := config.parseTagsIgnore()
	if err != nil {
		return nil, err
	}

	if tagsIgnore != nil {
		event.TagsIgnore = tagsIgnore
	}

	paths, err := config.parsePaths()
	if err != nil {
		return nil, err
	}

	if paths != nil {
		event.Paths = paths
	}

	pathsIgnore, err := config.parsePathsIgnore()
	if err != nil {
		return nil, err
	}

	if pathsIgnore != nil {
		event.PathsIgnore = pathsIgnore
	}

	return &event, nil
}

func (config *EventConfig) parseSchedule() (*EventSchedule, error) {
	hasBranches, err := config.hasBranches()
	if err != nil {
		return nil, err
	}

	if hasBranches {
		return nil, fmt.Errorf("attribute 'branches' is not allowed")
	}

	hasBranchesIgnore, err := config.hasBranchesIgnore()
	if err != nil {
		return nil, err
	}

	if hasBranchesIgnore {
		return nil, fmt.Errorf("attribute 'branches_ignore' is not allowed")
	}

	hasTags, err := config.hasTags()
	if err != nil {
		return nil, err
	}

	if hasTags {
		return nil, fmt.Errorf("attribute 'tags' is not allowed")
	}

	hasTagsIgnore, err := config.hasTagsIgnore()
	if err != nil {
		return nil, err
	}

	if hasTagsIgnore {
		return nil, fmt.Errorf("attribute 'tags_ignore' is not allowed")
	}

	hasPaths, err := config.hasPaths()
	if err != nil {
		return nil, err
	}

	if hasPaths {
		return nil, fmt.Errorf("attribute 'paths' is not allowed")
	}

	hasPathsIgnore, err := config.hasPathsIgnore()
	if err != nil {
		return nil, err
	}

	if hasPathsIgnore {
		return nil, fmt.Errorf("attribute 'paths_ignore' is not allowed")
	}

	hasWorkflows, err := config.hasWorkflows()
	if err != nil {
		return nil, err
	}

	if hasWorkflows {
		return nil, fmt.Errorf("attribute 'workflows' is not allowed")
	}

	event := EventSchedule{}

	cron, err := config.parseCron()
	if err != nil {
		return nil, err
	}

	if cron != nil {
		event = append(event, *cron...)
	}

	return &event, nil
}

func (config *EventConfig) parseWorkflowDispatch() (*EventWorkflowDispatch, error) {
	hasBranches, err := config.hasBranches()
	if err != nil {
		return nil, err
	}

	if hasBranches {
		return nil, fmt.Errorf("attribute 'branches' is not allowed")
	}

	hasBranchesIgnore, err := config.hasBranchesIgnore()
	if err != nil {
		return nil, err
	}

	if hasBranchesIgnore {
		return nil, fmt.Errorf("attribute 'branches_ignore' is not allowed")
	}

	hasTags, err := config.hasTags()
	if err != nil {
		return nil, err
	}

	if hasTags {
		return nil, fmt.Errorf("attribute 'tags' is not allowed")
	}

	hasTagsIgnore, err := config.hasTagsIgnore()
	if err != nil {
		return nil, err
	}

	if hasTagsIgnore {
		return nil, fmt.Errorf("attribute 'tags_ignore' is not allowed")
	}

	hasPaths, err := config.hasPaths()
	if err != nil {
		return nil, err
	}

	if hasPaths {
		return nil, fmt.Errorf("attribute 'paths' is not allowed")
	}

	hasPathsIgnore, err := config.hasPathsIgnore()
	if err != nil {
		return nil, err
	}

	if hasPathsIgnore {
		return nil, fmt.Errorf("attribute 'paths_ignore' is not allowed")
	}

	hasCron, err := config.hasCron()
	if err != nil {
		return nil, err
	}

	if hasCron {
		return nil, fmt.Errorf("attribute 'cron' is not allowed")
	}

	hasWorkflows, err := config.hasWorkflows()
	if err != nil {
		return nil, err
	}

	if hasWorkflows {
		return nil, fmt.Errorf("attribute 'workflows' is not allowed")
	}

	event := EventWorkflowDispatch{
		Name: config.Identifier,
	}

	inputs, err := config.parseWorkflowDispatchInputs()
	if err != nil {
		return nil, err
	}

	if inputs != nil {
		event.Inputs = inputs
	}

	return &event, nil
}

func (config *EventConfig) parseWorkflowRun() (*EventWorkflowRun, error) {
	hasBranches, err := config.hasBranches()
	if err != nil {
		return nil, err
	}

	if hasBranches {
		return nil, fmt.Errorf("attribute 'branches' is not allowed")
	}

	hasBranchesIgnore, err := config.hasBranchesIgnore()
	if err != nil {
		return nil, err
	}

	if hasBranchesIgnore {
		return nil, fmt.Errorf("attribute 'branches_ignore' is not allowed")
	}

	hasTags, err := config.hasTags()
	if err != nil {
		return nil, err
	}

	if hasTags {
		return nil, fmt.Errorf("attribute 'tags' is not allowed")
	}

	hasTagsIgnore, err := config.hasTagsIgnore()
	if err != nil {
		return nil, err
	}

	if hasTagsIgnore {
		return nil, fmt.Errorf("attribute 'tags_ignore' is not allowed")
	}

	hasPaths, err := config.hasPaths()
	if err != nil {
		return nil, err
	}

	if hasPaths {
		return nil, fmt.Errorf("attribute 'paths' is not allowed")
	}

	hasPathsIgnore, err := config.hasPathsIgnore()
	if err != nil {
		return nil, err
	}

	if hasPathsIgnore {
		return nil, fmt.Errorf("attribute 'paths_ignore' is not allowed")
	}

	hasCron, err := config.hasCron()
	if err != nil {
		return nil, err
	}

	if hasCron {
		return nil, fmt.Errorf("attribute 'cron' is not allowed")
	}

	hasInputs, err := config.hasInputs()
	if err != nil {
		return nil, err
	}

	if hasInputs {
		return nil, fmt.Errorf("attribute 'inputs' is not allowed")
	}

	hasOutputs, err := config.hasOutputs()
	if err != nil {
		return nil, err
	}

	if hasOutputs {
		return nil, fmt.Errorf("attribute 'outputs' is not allowed")
	}

	hasSecrets, err := config.hasSecrets()
	if err != nil {
		return nil, err
	}

	if hasSecrets {
		return nil, fmt.Errorf("attribute 'secrets' is not allowed")
	}

	event := EventWorkflowRun{
		Name: config.Identifier,
	}

	workflows, err := config.parseWorkflows()
	if err != nil {
		return nil, err
	}

	if workflows != nil {
		event.Workflows = workflows
	}

	types, err := config.parseTypes()
	if err != nil {
		return nil, err
	}

	if types != nil {
		event.Types = types
	}

	branches, err := config.parseBranches()
	if err != nil {
		return nil, err
	}

	if branches != nil {
		event.Branches = branches
	}

	branchesIgnore, err := config.parseBranchesIgnore()
	if err != nil {
		return nil, err
	}

	if branchesIgnore != nil {
		event.BranchesIgnore = branchesIgnore
	}

	return &event, nil
}

func (config *EventConfig) parseWorkflowCall() (*EventWorkflowCall, error) {
	hasBranches, err := config.hasBranches()
	if err != nil {
		return nil, err
	}

	if hasBranches {
		return nil, fmt.Errorf("attribute 'branches' is not allowed")
	}

	hasBranchesIgnore, err := config.hasBranchesIgnore()
	if err != nil {
		return nil, err
	}

	if hasBranchesIgnore {
		return nil, fmt.Errorf("attribute 'branches_ignore' is not allowed")
	}

	hasTags, err := config.hasTags()
	if err != nil {
		return nil, err
	}

	if hasTags {
		return nil, fmt.Errorf("attribute 'tags' is not allowed")
	}

	hasTagsIgnore, err := config.hasTagsIgnore()
	if err != nil {
		return nil, err
	}

	if hasTagsIgnore {
		return nil, fmt.Errorf("attribute 'tags_ignore' is not allowed")
	}

	hasPaths, err := config.hasPaths()
	if err != nil {
		return nil, err
	}

	if hasPaths {
		return nil, fmt.Errorf("attribute 'paths' is not allowed")
	}

	hasPathsIgnore, err := config.hasPathsIgnore()
	if err != nil {
		return nil, err
	}

	if hasPathsIgnore {
		return nil, fmt.Errorf("attribute 'paths_ignore' is not allowed")
	}

	hasCron, err := config.hasCron()
	if err != nil {
		return nil, err
	}

	if hasCron {
		return nil, fmt.Errorf("attribute 'cron' is not allowed")
	}

	hasWorkflows, err := config.hasWorkflows()
	if err != nil {
		return nil, err
	}

	if hasWorkflows {
		return nil, fmt.Errorf("attribute 'workflows' is not allowed")
	}

	event := EventWorkflowCall{
		Name: config.Identifier,
	}

	inputs, err := config.parseWorkflowCallInputs()
	if err != nil {
		return nil, fmt.Errorf("in event 'workflow_call': %w", err)
	}

	if inputs != nil {
		event.Inputs = inputs
	}

	outputs, err := config.parseOutputs()
	if err != nil {
		return nil, fmt.Errorf("in event 'workflow_call': %w", err)
	}

	if outputs != nil {
		event.Outputs = outputs
	}

	secrets, err := config.parseSecrets()
	if err != nil {
		return nil, fmt.Errorf("in event 'workflow_call': %w", err)
	}

	if secrets != nil {
		event.Secrets = secrets
	}

	return &event, nil
}

func (config *EventConfig) parseEventTrigger() (*Event, error) {
	hasBranches, err := config.hasBranches()
	if err != nil {
		return nil, err
	}

	if hasBranches {
		return nil, fmt.Errorf("in event '%s': attribute 'branches' is not allowed", config.Identifier)
	}

	hasBranchesIgnore, err := config.hasBranchesIgnore()
	if err != nil {
		return nil, err
	}

	if hasBranchesIgnore {
		return nil, fmt.Errorf("in event '%s': attribute 'branches_ignore' is not allowed", config.Identifier)
	}

	hasTags, err := config.hasTags()
	if err != nil {
		return nil, err
	}

	if hasTags {
		return nil, fmt.Errorf("in event '%s': attribute 'tags' is not allowed", config.Identifier)
	}

	hasTagsIgnore, err := config.hasTagsIgnore()
	if err != nil {
		return nil, err
	}

	if hasTagsIgnore {
		return nil, fmt.Errorf("in event '%s': attribute 'tags_ignore' is not allowed", config.Identifier)
	}

	hasPaths, err := config.hasPaths()
	if err != nil {
		return nil, err
	}

	if hasPaths {
		return nil, fmt.Errorf("in event '%s': attribute 'paths' is not allowed", config.Identifier)
	}

	hasPathsIgnore, err := config.hasPathsIgnore()
	if err != nil {
		return nil, err
	}

	if hasPathsIgnore {
		return nil, fmt.Errorf("in event '%s': attribute 'paths_ignore' is not allowed", config.Identifier)
	}

	hasCron, err := config.hasCron()
	if err != nil {
		return nil, err
	}

	if hasCron {
		return nil, fmt.Errorf("in event '%s': attribute 'cron' is not allowed", config.Identifier)
	}

	hasWorkflows, err := config.hasWorkflows()
	if err != nil {
		return nil, err
	}

	if hasWorkflows {
		return nil, fmt.Errorf("in event '%s': attribute 'workflows' is not allowed", config.Identifier)
	}

	event := Event{
		Name: config.Identifier,
	}

	types, err := config.parseTypes()
	if err != nil {
		return nil, fmt.Errorf("in event '%s': %w", config.Identifier, err)
	}

	if types != nil {
		event.Types = types
	}

	return &event, nil
}

func (config *EventsConfig) Parse() (*On, error) {
	if len(*config) == 0 {
		return nil, actoerrors.ErrWorkflowEmptyOn
	}

	if reflect.DeepEqual(*config, EventsConfig{}) {
		return nil, actoerrors.ErrWorkflowEmptyOn
	}

	on := On{}

	for _, o := range *config {
		if o.Identifier == "" {
			return nil, errors.New("in block 'on': missing identifier")
		}

		switch o.Identifier {
		case TriggerPullRequest.ToString():
			if on[TriggerPullRequest] != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerPullRequest, ErrEventTriggerNoMoreThanOne)
			}

			event, err := o.parsePullRequest()
			if err != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerPullRequest, err)
			}

			on[TriggerPullRequest] = event
		case TriggerPullRequestTarget.ToString():
			if on[TriggerPullRequestTarget] != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerPullRequestTarget, ErrEventTriggerNoMoreThanOne)
			}

			event, err := o.parsePullRequestTarget()
			if err != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerPullRequestTarget, err)
			}

			on[TriggerPullRequestTarget] = event
		case TriggerPush.ToString():
			if on[TriggerPush] != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerPush, ErrEventTriggerNoMoreThanOne)
			}

			event, err := o.parsePush()
			if err != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerPush, err)
			}

			on[TriggerPush] = event
		case TriggerSchedule.ToString():
			if on[TriggerSchedule] != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerSchedule, ErrEventTriggerNoMoreThanOne)
			}

			event, err := o.parseSchedule()
			if err != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerSchedule, err)
			}

			on[TriggerSchedule] = event
		case TriggerWorkflowCall.ToString():
			if on[TriggerWorkflowCall] != nil {
				return nil, fmt.Errorf("in block 'on23': event '%s': %w", TriggerWorkflowCall, ErrEventTriggerNoMoreThanOne)
			}

			event, err := o.parseWorkflowCall()
			if err != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerWorkflowCall, err)
			}

			on[TriggerWorkflowCall] = event
		case TriggerWorkflowRun.ToString():
			if on[TriggerWorkflowRun] != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerWorkflowRun, ErrEventTriggerNoMoreThanOne)
			}

			event, err := o.parseWorkflowRun()
			if err != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerWorkflowRun, err)
			}

			on[TriggerWorkflowRun] = event
		case TriggerWorkflowDispatch.ToString():
			if on[TriggerWorkflowDispatch] != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerWorkflowDispatch, ErrEventTriggerNoMoreThanOne)
			}

			event, err := o.parseWorkflowDispatch()
			if err != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", TriggerWorkflowDispatch, err)
			}

			on[TriggerWorkflowDispatch] = event
		default:
			if !ValidateEventTrigger(o.Identifier) {
				return nil, fmt.Errorf("in block 'on': event '%s', %w", o.Identifier, ErrEventTriggerUnknown)
			}

			eventTrigger := EventTrigger(o.Identifier)

			if on[eventTrigger] != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", eventTrigger, ErrEventTriggerNoMoreThanOne)
			}

			event, err := o.parseEventTrigger()
			if err != nil {
				return nil, fmt.Errorf("in block 'on': event '%s': %w", eventTrigger, err)
			}

			on[eventTrigger] = event
		}
	}

	return &on, nil
}
