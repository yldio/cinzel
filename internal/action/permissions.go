// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type PermissionsOption string

const (
	Read  PermissionsOption = "read"
	Write PermissionsOption = "write"
	None  PermissionsOption = "none"
)

type Permissions struct {
	Actions            *PermissionsOption `yaml:"actions,omitempty"`
	Attestations       *PermissionsOption `yaml:"attestations,omitempty"`
	Checks             *PermissionsOption `yaml:"checks,omitempty"`
	Contents           *PermissionsOption `yaml:"contents,omitempty"`
	Deployments        *PermissionsOption `yaml:"deployments,omitempty"`
	IdToken            *PermissionsOption `yaml:"id-token,omitempty"`
	Issues             *PermissionsOption `yaml:"issues,omitempty"`
	Discussions        *PermissionsOption `yaml:"discussions,omitempty"`
	Packages           *PermissionsOption `yaml:"packages,omitempty"`
	Pages              *PermissionsOption `yaml:"pages,omitempty"`
	PullRequests       *PermissionsOption `yaml:"pull-requests,omitempty"`
	RepositoryProjects *PermissionsOption `yaml:"repository-projects,omitempty"`
	SecurityEvents     *PermissionsOption `yaml:"security-events,omitempty"`
	Statuses           *PermissionsOption `yaml:"statuses,omitempty"`
}

type PermissionsConfig struct {
	Actions            *PermissionsOption `hcl:"actions,attr"`
	Attestations       *PermissionsOption `hcl:"attestations,attr"`
	Checks             *PermissionsOption `hcl:"checks,attr"`
	Contents           *PermissionsOption `hcl:"contents,attr"`
	Deployments        *PermissionsOption `hcl:"deployments,attr"`
	IdToken            *PermissionsOption `hcl:"id_token,attr"`
	Issues             *PermissionsOption `hcl:"issues,attr"`
	Discussions        *PermissionsOption `hcl:"discussions,attr"`
	Packages           *PermissionsOption `hcl:"packages,attr"`
	Pages              *PermissionsOption `hcl:"pages,attr"`
	PullRequests       *PermissionsOption `hcl:"pull_requests,attr"`
	RepositoryProjects *PermissionsOption `hcl:"repository_projects,attr"`
	SecurityEvents     *PermissionsOption `hcl:"security_events,attr"`
	Statuses           *PermissionsOption `hcl:"statuses,attr"`
}

func (config *PermissionsConfig) Parse() (Permissions, error) {
	if config == nil {
		return Permissions{}, nil
	}

	permissions := Permissions(*config)

	return permissions, nil
}
