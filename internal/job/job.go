// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package job

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/acto/internal/action"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/step"
	"github.com/yldio/acto/internal/variables"
	"github.com/zclconf/go-cty/cty"
)

type Jobs map[string]*Job

type Job struct {
	Id              string              `yaml:"-"`
	Name            *string             `yaml:"name,omitempty" hcl:"name"`
	Permissions     *action.Permissions `yaml:"permissions,omitempty" hcl:"permissions"`
	Needs           *[]string           `yaml:"needs,omitempty" hcl:"needs"`
	If              *string             `yaml:"if,omitempty" hcl:"if"`
	RunsOn          any                 `yaml:"runs-on,omitempty" hcl:"runs_on"`
	Environment     any                 `yaml:"environment,omitempty" hcl:"environment"`
	Concurrency     *action.Concurrency `yaml:"concurrency,omitempty" hcl:"concurrency"`
	Outputs         *action.Outputs     `yaml:"outputs,omitempty" hcl:"output"`
	Env             *action.Envs        `yaml:"env,omitempty" hcl:"env"`
	Defaults        *action.Defaults    `yaml:"defaults,omitempty" hcl:"defaults"`
	Steps           []*step.Step        `yaml:"steps,omitempty" hcl:"step"`
	StepsIds        []string            `yaml:"-"`
	TimeoutMinutes  *uint64             `yaml:"timeout-minutes,omitempty" hcl:"timeout_minutes"`
	Strategy        *action.Strategy    `yaml:"strategy,omitempty" hcl:"strategy"`
	ContinueOnError any                 `yaml:"continue-on-error,omitempty" hcl:"continue_on_error"`
	Container       *action.Container   `yaml:"container,omitempty" hcl:"container"`
	Services        *action.Services    `yaml:"services,omitempty" hcl:"service"`
	Uses            *string             `yaml:"uses,omitempty" hcl:"uses"`
	With            *map[string]any     `yaml:"with,omitempty" hcl:"with"`
	Secrets         any                 `yaml:"secrets,omitempty" hcl:"secret"`
}

type JobsConfig []JobConfig

type JobConfig struct {
	Identifier      string                    `hcl:"id,label"`
	Name            hcl.Expression            `hcl:"name,attr"`
	Permissions     *action.PermissionsConfig `hcl:"permissions,block"`
	Needs           hcl.Expression            `hcl:"needs,attr"`
	If              hcl.Expression            `hcl:"if,attr"`
	RunsOn          *action.RunsOnConfig      `hcl:"runs_on,block"`
	Environment     *action.EnvironmentConfig `hcl:"environment,block"`
	Concurrency     *action.ConcurrencyConfig `hcl:"concurrency,block"`
	Outputs         action.OutputsConfig      `hcl:"output,block"`
	Env             action.EnvsConfig         `hcl:"env,block"`
	Defaults        *action.DefaultsConfig    `hcl:"defaults,block"`
	Steps           hcl.Expression            `hcl:"steps,attr"`
	TimeoutMinutes  hcl.Expression            `hcl:"timeout_minutes,attr"`
	Strategy        *action.StrategyConfig    `hcl:"strategy,block"`
	ContinueOnError hcl.Expression            `hcl:"continue_on_error,attr"`
	Container       *action.ContainerConfig   `hcl:"container,block"`
	Services        action.ServicesConfig     `hcl:"service,block"`
	Uses            *action.UsesConfig        `hcl:"uses,block"`
	With            action.WithsConfig        `hcl:"with,block"`
	Secrets         action.SecretsConfig      `hcl:"secret,block"`
	SecretsInherit  hcl.Expression            `hcl:"secrets,attr"`
}

const Inherit = "inherit"

func (config *JobConfig) unwrapSecretsInherit(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		if resultValue != Inherit {
			return nil, fmt.Errorf("attribute 'secrets' must be the hardcoded string '%s'", Inherit)
		}
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapSecretsInherit(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, fmt.Errorf("attribute 'secrets' must be the hardcoded string '%s'", Inherit)
	}
}

func (config *JobConfig) parseSecrets() (any, error) {
	acto := actoparser.NewActo(config.SecretsInherit)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	secretsInherit, err := config.unwrapSecretsInherit(acto)
	if err != nil {
		return nil, err
	}

	secrets, err := config.Secrets.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in secrets: %w", err)
	}

	if secrets != nil && len(*secrets) > 0 && secretsInherit != nil {
		return nil, fmt.Errorf("error in secrets: can only have 'secrets' inherit or a set of secrets")
	}

	if secretsInherit != nil {
		return secretsInherit, nil
	}

	if secrets == nil {
		return nil, nil
	}

	return secrets, nil
}

func (config *JobConfig) parseWith() (*map[string]any, error) {
	value, err := config.With.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in with: %w", err)
	}

	return value, nil
}

func (config *JobConfig) parseUses() (*string, error) {
	value, err := config.Uses.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in uses: %w", err)
	}

	return value, nil
}

func (config *JobConfig) parseServices() (*action.Services, error) {
	services := make(action.Services)

	for _, service := range config.Services {
		svc, err := service.Parse()
		if err != nil {
			return nil, fmt.Errorf("error in service: %w", err)
		}

		if services[svc.Name] != nil {
			return nil, fmt.Errorf("error in service: '%s' already defined ", svc.Name)
		}

		services[svc.Name] = svc
	}

	if len(services) == 0 {
		return nil, nil
	}

	return &services, nil
}

func (config *JobConfig) parseContainer() (*action.Container, error) {
	container, err := config.Container.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in container: %w", err)
	}

	return container, nil
}

func (config *JobConfig) unwrapContinueOnError(acto *actoparser.Acto) (any, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return &resultValue, nil
	case bool:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapContinueOnError(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'continue_on_error' must be a string or a boolean")
	}
}

func (config *JobConfig) parseContinueOnError() (any, error) {
	acto := actoparser.NewActo(config.ContinueOnError)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapContinueOnError(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *JobConfig) parseStrategy() (*action.Strategy, error) {
	strategy, err := config.Strategy.Parse()
	if err != nil {
		return nil, err
	}

	return strategy, nil
}

func (config *JobConfig) unwrapTimeoutMinutes(acto *actoparser.Acto) (*uint64, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case int64:
		if resultValue < 0 {
			return nil, errors.New("attribute 'timeout_minutes' must be a positive number")
		}

		val := uint64(resultValue)
		return &val, nil
	case uint64:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapTimeoutMinutes(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'timeout_minutes' must be a positive number")
	}
}

func (config *JobConfig) parseTimeoutMinutes() (*uint64, error) {
	acto := actoparser.NewActo(config.TimeoutMinutes)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapTimeoutMinutes(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *JobConfig) unwrapStepsIds(acto *actoparser.Acto) (*[]string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case []actoparser.ActoVariableRef:
		list := []string{}
		for _, stepRef := range resultValue {
			if stepRef.Name != "step" {
				return nil, errors.New("invalid step reference, should be step.<step-identifier>")
			}

			list = append(list, stepRef.Attr)
		}

		return &list, nil
	default:
		return nil, errors.New("attribute 'Steps' must be a list of steps relation")
	}
}

func (config *JobConfig) parseStepsIds() (*[]string, error) {
	acto := actoparser.NewActo(config.Steps)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	stepsIds, err := config.unwrapStepsIds(acto)
	if err != nil {
		return nil, err
	}

	if stepsIds != nil && len(*stepsIds) == 0 {
		return nil, errors.New("attribute 'steps' cannot be empty")
	}

	return stepsIds, nil
}

func (config *JobConfig) parseDefaults() (*action.Defaults, error) {
	defaults, err := config.Defaults.Parse()
	if err != nil {
		return nil, err
	}

	return defaults, nil
}

func (config *JobConfig) parseEnvs() (*action.Envs, error) {
	env, err := config.Env.Parse()
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (config *JobConfig) parseOutputs() (*action.Outputs, error) {
	outputs, err := config.Outputs.Parse()
	if err != nil {
		return nil, err
	}

	if outputs == nil || len(*outputs) == 0 {
		return nil, nil
	}

	return outputs, nil
}

func (config *JobConfig) parseConcurrency() (*action.Concurrency, error) {
	concurrency, err := config.Concurrency.Parse()
	if err != nil {
		return nil, err
	}

	if concurrency == nil {
		return nil, nil
	}

	return concurrency, nil
}

func (config *JobConfig) parseEnvironment() (any, error) {
	environment, err := config.Environment.Parse()
	if err != nil {
		return nil, err
	}

	return environment, nil
}

func (config *JobConfig) unwrapIf(acto *actoparser.Acto) (*string, error) {
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

		return config.unwrapIf(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'if' must be a string or bool")
	}
}

func (config *JobConfig) unwrapNeeds(acto *actoparser.Acto) (*[]string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case []actoparser.ActoVariableRef:
		list := []string{}
		for _, jobRef := range resultValue {
			if jobRef.Name != "job" {
				return nil, errors.New("invalid job reference, should be job.<job-identifier>")
			}

			list = append(list, jobRef.Attr)
		}

		return &list, nil
	default:
		return nil, errors.New("attribute 'needs' must be a list of jobs relation")
	}
}

func (config *JobConfig) unwrapName(acto *actoparser.Acto) (*string, error) {
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

		return config.unwrapName(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'name' must be a string")
	}
}

func (config *JobConfig) parseName() (*string, error) {
	acto := actoparser.NewActo(config.Name)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapName(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *JobConfig) parsePermissions() (*action.Permissions, error) {
	permissions, err := config.Permissions.Parse()
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (config *JobConfig) parseNeeds() (*[]string, error) {
	acto := actoparser.NewActo(config.Needs)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapNeeds(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *JobConfig) parseIf() (*string, error) {
	acto := actoparser.NewActo(config.If)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapIf(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *JobConfig) parseRunsOn() (any, error) {
	runsOn, err := config.RunsOn.Parse()
	if err != nil {
		return nil, err
	}

	return runsOn, nil
}

func (config *JobConfig) Parse() (*Job, error) {
	if config == nil {
		return nil, nil
	}

	if config.Identifier == "" {
		return nil, fmt.Errorf("error in job: no identifier, %w", actoerrors.ErrOpenIssue)
	}

	job := Job{
		Id: config.Identifier,
	}

	name, err := config.parseName()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if name != nil {
		job.Name = name
	}

	permissions, err := config.parsePermissions()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if permissions != nil {
		job.Permissions = permissions
	}

	needs, err := config.parseNeeds()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if needs != nil {
		job.Needs = needs
	}

	ifVal, err := config.parseIf()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if ifVal != nil {
		job.If = ifVal
	}

	runsOn, err := config.parseRunsOn()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if runsOn != nil {
		job.RunsOn = runsOn
	}

	environment, err := config.parseEnvironment()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if environment != nil {
		job.Environment = environment
	}

	concurrency, err := config.parseConcurrency()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if concurrency != nil {
		job.Concurrency = concurrency
	}

	outputs, err := config.parseOutputs()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if outputs != nil {
		job.Outputs = outputs
	}

	envs, err := config.parseEnvs()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if envs != nil {
		job.Env = envs
	}

	defaults, err := config.parseDefaults()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if defaults != nil {
		job.Defaults = defaults
	}

	stepsIds, err := config.parseStepsIds()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if stepsIds != nil {
		job.StepsIds = *stepsIds
	}

	timeoutMinutes, err := config.parseTimeoutMinutes()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if timeoutMinutes != nil {
		job.TimeoutMinutes = timeoutMinutes
	}

	strategy, err := config.parseStrategy()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if strategy != nil {
		job.Strategy = strategy
	}

	continueOnError, err := config.parseContinueOnError()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if continueOnError != nil {
		job.ContinueOnError = continueOnError
	}

	container, err := config.parseContainer()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if container != nil {
		job.Container = container
	}

	services, err := config.parseServices()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if services != nil {
		job.Services = services
	}

	uses, err := config.parseUses()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if uses != nil {
		job.Uses = uses
	}

	if runsOn != nil && uses != nil {
		return nil, fmt.Errorf("error in job '%s': can only have 'runs_on' or 'uses', not both, %w", job.Id, actoerrors.ErrOpenIssue)
	}

	withs, err := config.parseWith()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if withs != nil {
		job.With = withs
	}

	if withs != nil && uses == nil {
		return nil, fmt.Errorf("error in job '%s': can only have 'with' when 'uses' is set, %w", job.Id, actoerrors.ErrOpenIssue)
	}

	secrets, err := config.parseSecrets()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if secrets != nil {
		job.Secrets = secrets
	}

	if secrets != nil && uses == nil {
		return nil, fmt.Errorf("error in job '%s': can only have 'secret' when 'uses' is set, %w", job.Id, actoerrors.ErrOpenIssue)
	}

	return &job, nil
}

func (config *JobsConfig) Parse() (Jobs, error) {
	jobs := Jobs{}

	for _, job := range *config {
		parsedJob, err := job.Parse()
		if err != nil {
			return Jobs{}, err
		}

		if jobs[parsedJob.Id] != nil {
			return Jobs{}, fmt.Errorf("error in job '%s': already defined, %w", parsedJob.Id, actoerrors.ErrOpenIssue)
		}

		jobs[parsedJob.Id] = parsedJob
	}

	return jobs, nil
}

func (job *Job) Decode(body *hclwrite.Body, attr string) error {
	body.AppendNewline()

	jobBlock := body.AppendNewBlock(attr, []string{job.Id})

	jobBody := jobBlock.Body()

	if job.Name != nil {
		attr, err := actoparser.GetHclTag(*job, "Name")
		if err != nil {
			return err
		}

		jobBody.SetAttributeValue(attr, cty.StringVal(*job.Name))
	}

	if job.Permissions != nil {
		attr, err := actoparser.GetHclTag(*job, "Permissions")
		if err != nil {
			return err
		}

		jobBody.AppendNewline()
		permissionsBlock := jobBody.AppendNewBlock(attr, nil)

		permissionsBody := permissionsBlock.Body()

		values := reflect.ValueOf(*job.Permissions)
		types := values.Type()

		for i := 0; i < values.NumField(); i++ {
			if values.Field(i).IsNil() {
				continue
			}

			option := values.Field(i).Elem().String()

			if !action.ValidatePermissionsOption(option) {
				panic(option)
			}

			attr, err := actoparser.GetHclTag(*job.Permissions, types.Field(i).Name)
			if err != nil {
				return err
			}

			permissionsBody.SetAttributeValue(attr, cty.StringVal(option))
		}
	}

	if job.Needs != nil {
		needsTokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenOBrack,
				Bytes: []byte(`[`),
			},
			{
				Type:  hclsyntax.TokenNewline,
				Bytes: []byte("\n"),
			},
		}

		for i, need := range *job.Needs {
			if i > 0 {
				needsTokens = append(needsTokens, &hclwrite.Token{
					Type:  hclsyntax.TokenComma,
					Bytes: []byte(`,`),
				})
				needsTokens = append(needsTokens, &hclwrite.Token{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				})
			}

			needsTokens = append(needsTokens, &hclwrite.Token{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(fmt.Sprintf("job.%s", need)),
			})
		}

		needsTokens = append(needsTokens, &hclwrite.Token{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte("\n"),
		})

		needsTokens = append(needsTokens, &hclwrite.Token{
			Type:  hclsyntax.TokenCBrack,
			Bytes: []byte(`]`),
		})

		if len(jobBody.Blocks()) > 0 {
			jobBody.AppendNewline()
		}

		attr, err := actoparser.GetHclTag(*job, "Needs")
		if err != nil {
			return err
		}

		jobBody.SetAttributeRaw(attr, needsTokens)
	}

	if job.If != nil {
		attr, err := actoparser.GetHclTag(*job, "If")
		if err != nil {
			return err
		}

		if len(jobBody.Blocks()) > 0 {
			jobBody.AppendNewline()
		}

		jobBody.SetAttributeValue(attr, cty.StringVal(*job.If))
	}

	if job.RunsOn != nil {
		attr, err := actoparser.GetHclTag(*job, "RunsOn")
		if err != nil {
			return err
		}

		if len(jobBody.Blocks()) > 0 {
			jobBody.AppendNewline()
		}

		runsOnBlock := jobBody.AppendNewBlock(attr, nil)

		runsOnBody := runsOnBlock.Body()

		switch runsOn := job.RunsOn.(type) {
		case string:
			runsOnBody.SetAttributeValue("runners", cty.StringVal(runsOn))
		case map[string]any:
			if runsOn["group"] != nil {
				switch group := runsOn["group"].(type) {
				case string:
					runsOnBody.SetAttributeValue("group", cty.StringVal(group))
				default:
					panic("only strings on group")
				}
			}
			if runsOn["labels"] != nil {
				switch labels := runsOn["labels"].(type) {
				case string:
					runsOnBody.SetAttributeValue("labels", cty.StringVal(labels))
				default:
					panic("only strings on labels")
				}
			}
		}
	}

	if job.Environment != nil {
		attr, err := actoparser.GetHclTag(*job, "Environment")
		if err != nil {
			return err
		}

		if len(jobBody.Blocks()) > 0 {
			jobBody.AppendNewline()
		}

		environmentBlock := jobBody.AppendNewBlock(attr, nil)
		environmentBody := environmentBlock.Body()

		switch environment := job.Environment.(type) {
		case map[string]any:
			if environment["name"] != nil {
				switch name := environment["name"].(type) {
				case string:
					environmentBody.SetAttributeValue("name", cty.StringVal(name))
				default:
					panic("only strings on name")
				}
			}
			if environment["url"] != nil {
				switch url := environment["url"].(type) {
				case string:
					environmentBody.SetAttributeValue("url", cty.StringVal(url))
				default:
					panic("only strings on url")
				}
			}
		default:
			panic("only map[string] on environment")
		}
	}

	if job.Concurrency != nil {
		attr, err := actoparser.GetHclTag(*job, "Concurrency")
		if err != nil {
			return err
		}

		if len(jobBody.Blocks()) > 0 {
			jobBody.AppendNewline()
		}

		concurrencyBlock := jobBody.AppendNewBlock(attr, nil)
		concurrencyBody := concurrencyBlock.Body()

		attr, err = actoparser.GetHclTag(*job.Concurrency, "Group")
		if err != nil {
			return err
		}

		if job.Concurrency.Group != nil {
			attr, err := actoparser.GetHclTag(*job.Concurrency, "Group")
			if err != nil {
				return err
			}

			concurrencyBody.SetAttributeValue(attr, cty.StringVal(*job.Concurrency.Group))
		}

		if job.Concurrency.CancelInProgress != nil {
			attr, err := actoparser.GetHclTag(*job.Concurrency, "CancelInProgress")
			if err != nil {
				return err
			}

			concurrencyBody.SetAttributeValue(attr, cty.BoolVal(*job.Concurrency.CancelInProgress))
		}
	}

	if job.Outputs != nil {
		attr, err := actoparser.GetHclTag(*job, "Outputs")
		if err != nil {
			return err
		}

		for key, output := range *job.Outputs {
			if len(jobBody.Blocks()) > 0 {
				jobBody.AppendNewline()
			}

			outputBlock := jobBody.AppendNewBlock(attr, nil)
			outputBody := outputBlock.Body()

			outputBody.SetAttributeValue("name", cty.StringVal(key))
			outputBody.SetAttributeValue("value", cty.StringVal(output))
		}
	}

	if job.Env != nil {
		for name, env := range *job.Env {

			if len(jobBody.Blocks()) > 0 {
				jobBody.AppendNewline()
			}

			envBlock := jobBody.AppendNewBlock("env", nil)

			envBody := envBlock.Body()
			envBody.SetAttributeValue("name", cty.StringVal(name))

			switch e := env.(type) {
			case string:
				envBody.SetAttributeValue("value", cty.StringVal(e))
			}
		}
	}

	if job.Defaults != nil {
		if len(jobBody.Blocks()) > 0 {
			jobBody.AppendNewline()
		}

		attr, err := actoparser.GetHclTag(*job, "Defaults")
		if err != nil {
			return err
		}

		defaultsBlock := jobBody.AppendNewBlock(attr, nil)
		defaultsBody := defaultsBlock.Body()
		attr, err = actoparser.GetHclTag(*job.Defaults, "Run")
		if err != nil {
			return err
		}

		runBlock := defaultsBody.AppendNewBlock(attr, nil)
		runBody := runBlock.Body()

		if job.Defaults.Run.Shell != nil {
			attr, err := actoparser.GetHclTag(*job.Defaults.Run, "Shell")
			if err != nil {
				return err
			}

			runBody.SetAttributeValue(attr, cty.StringVal(*job.Defaults.Run.Shell))
		}

		if job.Defaults.Run.WorkingDirectory != nil {
			attr, err := actoparser.GetHclTag(*job.Defaults.Run, "WorkingDirectory")
			if err != nil {
				return err
			}

			runBody.SetAttributeValue(attr, cty.StringVal(*job.Defaults.Run.WorkingDirectory))
		}
	}

	if job.Steps != nil {
		attr, err := actoparser.GetHclTag(*job, "Steps")
		if err != nil {
			return err
		}

		for i, s := range job.Steps {
			var identifier string
			if s.Id != nil {
				identifier = fmt.Sprintf("%s-%s", job.Id, *s.Id)
			} else {
				identifier = fmt.Sprintf("%s-step-%d", job.Id, i+1)
			}

			s.Identifier = identifier

			s.Decode(body, attr)
		}
	}

	if job.TimeoutMinutes != nil {
		attr, err := actoparser.GetHclTag(*job, "TimeoutMinutes")
		if err != nil {
			return err
		}

		if len(jobBody.Blocks()) > 0 {
			jobBody.AppendNewline()
		}

		jobBody.SetAttributeValue(attr, cty.NumberUIntVal(*job.TimeoutMinutes))
	}

	if job.Strategy != nil {
		attr, err := actoparser.GetHclTag(*job, "Strategy")
		if err != nil {
			return err
		}

		if err := job.Strategy.Decode(jobBody, attr); err != nil {
			return err
		}
	}

	if job.ContinueOnError != nil {
		attr, err := actoparser.GetHclTag(*job, "ContinueOnError")
		if err != nil {
			return err
		}

		if len(jobBody.Blocks()) > 0 {
			jobBody.AppendNewline()
		}

		switch v := job.ContinueOnError.(type) {
		case string:
			jobBody.SetAttributeValue(attr, cty.StringVal(v))
		case bool:
			jobBody.SetAttributeValue(attr, cty.BoolVal(v))
		default:
			return errors.New("unkown dealt type")
		}
	}

	if job.Container != nil {
		attr, err := actoparser.GetHclTag(*job, "Container")
		if err != nil {
			return err
		}

		if err := job.Container.Decode(jobBody, attr); err != nil {
			return err
		}
	}

	if job.Services != nil {
		attr, err := actoparser.GetHclTag(*job, "Services")
		if err != nil {
			return err
		}

		if err := job.Services.Decode(jobBody, attr); err != nil {
			return err
		}
	}

	if job.Uses != nil {
		usesAttr, err := actoparser.GetHclTag(*job, "Uses")
		if err != nil {
			return err
		}

		if len(jobBody.Blocks()) > 0 || len(jobBody.Attributes()) > 0 {
			jobBody.AppendNewline()
		}

		jobBody.SetAttributeValue(usesAttr, cty.StringVal(*job.Uses))
	}

	if job.With != nil {
		withAttr, err := actoparser.GetHclTag(*job, "With")
		if err != nil {
			return err
		}

		for key, value := range *job.With {
			if len(jobBody.Blocks()) > 0 || len(jobBody.Attributes()) > 0 {
				jobBody.AppendNewline()
			}

			withBlock := jobBody.AppendNewBlock(withAttr, nil)
			withBody := withBlock.Body()

			withBody.SetAttributeValue("name", cty.StringVal(key))

			switch v := value.(type) {
			case string:
				withBody.SetAttributeValue("value", cty.StringVal(v))
			case bool:
				withBody.SetAttributeValue("value", cty.BoolVal(v))
			case uint64:
				withBody.SetAttributeValue("value", cty.NumberUIntVal(v))
			case int64:
				withBody.SetAttributeValue("value", cty.NumberIntVal(v))
			case float64:
				withBody.SetAttributeValue("value", cty.NumberFloatVal(v))
			default:
				return errors.New("unkown dealt type")
			}
		}
	}

	if job.Secrets != nil {
		secretAttr, err := actoparser.GetHclTag(*job, "Secrets")
		if err != nil {
			return err
		}

		switch secret := job.Secrets.(type) {
		case string:
			jobBody.SetAttributeValue("secrets", cty.StringVal("inherit"))
		case map[string]any:
			for key, value := range secret {
				secretBlock := jobBody.AppendNewBlock(secretAttr, nil)
				secretBody := secretBlock.Body()

				secretBody.SetAttributeValue("name", cty.StringVal(key))

				switch v := value.(type) {
				case string:
					secretBody.SetAttributeValue("value", cty.StringVal(v))
				case bool:
					secretBody.SetAttributeValue("value", cty.BoolVal(v))
				case uint64:
					secretBody.SetAttributeValue("value", cty.NumberUIntVal(v))
				case int64:
					secretBody.SetAttributeValue("value", cty.NumberIntVal(v))
				case float64:
					secretBody.SetAttributeValue("value", cty.NumberFloatVal(v))
				default:
					return errors.New("unkown dealt type")
				}
			}
		default:
			return errors.New("unkown dealt type")
		}
	}

	return nil
}
