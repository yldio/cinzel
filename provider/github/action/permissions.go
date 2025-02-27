// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
	"github.com/zclconf/go-cty/cty"
)

type PermissionsOption string

const (
	Read  PermissionsOption = "read"
	Write PermissionsOption = "write"
	None  PermissionsOption = "none"
)

func (option PermissionsOption) ToString() string {
	return string(option)
}

func ValidatePermissionsOption(option string) bool {
	switch option {
	case Read.ToString():
		return true
	case Write.ToString():
		return true
	case None.ToString():
		return true
	default:
		return false
	}
}

type Permissions struct {
	Actions            *PermissionsOption `yaml:"actions,omitempty" hcl:"actions"`
	Attestations       *PermissionsOption `yaml:"attestations,omitempty" hcl:"attestations"`
	Checks             *PermissionsOption `yaml:"checks,omitempty" hcl:"checks"`
	Contents           *PermissionsOption `yaml:"contents,omitempty" hcl:"contents"`
	Deployments        *PermissionsOption `yaml:"deployments,omitempty" hcl:"deployments"`
	IdToken            *PermissionsOption `yaml:"id-token,omitempty" hcl:"id_token"`
	Issues             *PermissionsOption `yaml:"issues,omitempty" hcl:"issues"`
	Discussions        *PermissionsOption `yaml:"discussions,omitempty" hcl:"discussions"`
	Packages           *PermissionsOption `yaml:"packages,omitempty" hcl:"packages"`
	Pages              *PermissionsOption `yaml:"pages,omitempty" hcl:"pages"`
	PullRequests       *PermissionsOption `yaml:"pull-requests,omitempty" hcl:"pull_requests"`
	RepositoryProjects *PermissionsOption `yaml:"repository-projects,omitempty" hcl:"repository_projects"`
	SecurityEvents     *PermissionsOption `yaml:"security-events,omitempty" hcl:"security_events"`
	Statuses           *PermissionsOption `yaml:"statuses,omitempty" hcl:"statuses"`
}

type PermissionsConfig struct {
	Actions            hcl.Expression `hcl:"actions,attr"`
	Attestations       hcl.Expression `hcl:"attestations,attr"`
	Checks             hcl.Expression `hcl:"checks,attr"`
	Contents           hcl.Expression `hcl:"contents,attr"`
	Deployments        hcl.Expression `hcl:"deployments,attr"`
	IdToken            hcl.Expression `hcl:"id_token,attr"`
	Issues             hcl.Expression `hcl:"issues,attr"`
	Discussions        hcl.Expression `hcl:"discussions,attr"`
	Packages           hcl.Expression `hcl:"packages,attr"`
	Pages              hcl.Expression `hcl:"pages,attr"`
	PullRequests       hcl.Expression `hcl:"pull_requests,attr"`
	RepositoryProjects hcl.Expression `hcl:"repository_projects,attr"`
	SecurityEvents     hcl.Expression `hcl:"security_events,attr"`
	Statuses           hcl.Expression `hcl:"statuses,attr"`
}

func (config *PermissionsConfig) unwrapStatuses(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapStatuses(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'statuses' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parseStatuses() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.Statuses)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapStatuses(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapSecurityEvents(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapSecurityEvents(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'security_events' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parseSecurityEvents() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.SecurityEvents)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapSecurityEvents(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapRepositoryProjects(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapRepositoryProjects(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'repository_projects' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parseRepositoryProjects() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.RepositoryProjects)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapRepositoryProjects(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapPullRequests(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapPullRequests(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'pull_requests' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parsePullRequests() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.PullRequests)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapPullRequests(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapPages(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapPages(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'pages' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parsePages() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.Pages)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapPages(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapPackages(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapPackages(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'packages' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parsePackages() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.Packages)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapPackages(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapDiscussions(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapDiscussions(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'discussions' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parseDiscussions() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.Discussions)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapDiscussions(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapIssues(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapIssues(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'issues' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parseIssues() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.Issues)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapIssues(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapIdToken(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapIdToken(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'id_token' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parseIdToken() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.IdToken)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapIdToken(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapDeployments(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapDeployments(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'deployments' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parseDeployments() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.Deployments)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapDeployments(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapContents(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapContents(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'contents' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parseContents() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.Contents)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapContents(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapChecks(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapChecks(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'checks' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parseChecks() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.Checks)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapChecks(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapAttestations(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapAttestations(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'attestations' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parseAttestations() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.Attestations)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapAttestations(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) unwrapActions(acto *actoparser.Acto) (*PermissionsOption, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		option := PermissionsOption(resultValue)
		return &option, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapActions(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'actions' must one of the following 'read', 'write' or 'none'")
	}
}

func (config *PermissionsConfig) parseActions() (*PermissionsOption, error) {
	acto := actoparser.NewActo(config.Actions)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapActions(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *PermissionsConfig) Parse() (*Permissions, error) {
	if config == nil {
		return nil, nil
	}

	permissions := Permissions{}

	actions, err := config.parseActions()
	if err != nil {
		return nil, err
	}

	if actions != nil {
		permissions.Actions = actions
	}

	attestations, err := config.parseAttestations()
	if err != nil {
		return nil, err
	}

	if attestations != nil {
		permissions.Attestations = attestations
	}

	checks, err := config.parseChecks()
	if err != nil {
		return nil, err
	}

	if checks != nil {
		permissions.Checks = checks
	}

	contents, err := config.parseContents()
	if err != nil {
		return nil, err
	}

	if contents != nil {
		permissions.Contents = contents
	}

	deployments, err := config.parseDeployments()
	if err != nil {
		return nil, err
	}

	if deployments != nil {
		permissions.Deployments = deployments
	}

	idToken, err := config.parseIdToken()
	if err != nil {
		return nil, err
	}

	if idToken != nil {
		permissions.IdToken = idToken
	}

	issues, err := config.parseIssues()
	if err != nil {
		return nil, err
	}

	if issues != nil {
		permissions.Issues = issues
	}

	discussions, err := config.parseDiscussions()
	if err != nil {
		return nil, err
	}

	if discussions != nil {
		permissions.Discussions = discussions
	}

	packages, err := config.parsePackages()
	if err != nil {
		return nil, err
	}

	if packages != nil {
		permissions.Packages = packages
	}

	pages, err := config.parsePages()
	if err != nil {
		return nil, err
	}

	if pages != nil {
		permissions.Pages = pages
	}

	pullRequests, err := config.parsePullRequests()
	if err != nil {
		return nil, err
	}

	if pullRequests != nil {
		permissions.PullRequests = pullRequests
	}

	repositoryProjects, err := config.parseRepositoryProjects()
	if err != nil {
		return nil, err
	}

	if repositoryProjects != nil {
		permissions.RepositoryProjects = repositoryProjects
	}

	securityEvents, err := config.parseSecurityEvents()
	if err != nil {
		return nil, err
	}

	if securityEvents != nil {
		permissions.SecurityEvents = securityEvents
	}

	statuses, err := config.parseStatuses()
	if err != nil {
		return nil, err
	}

	if statuses != nil {
		permissions.Statuses = statuses
	}

	return &permissions, nil
}

func (permissions *Permissions) Decode(body *hclwrite.Body, attr string) error {
	if len(body.Blocks()) > 0 || len(body.Attributes()) > 0 {
		body.AppendNewline()
	}

	permissionsBlock := body.AppendNewBlock(attr, nil)
	permissionsBody := permissionsBlock.Body()

	values := reflect.ValueOf(*permissions)
	types := values.Type()

	for i := 0; i < values.NumField(); i++ {
		if values.Field(i).IsNil() {
			continue
		}

		option := values.Field(i).Elem().String()

		if !ValidatePermissionsOption(option) {
			return errors.New("unknown permission option")
		}

		attr, err := actoparser.GetHclTag(*permissions, types.Field(i).Name)
		if err != nil {
			return err
		}

		permissionsBody.SetAttributeValue(attr, cty.StringVal(option))
	}

	return nil
}
