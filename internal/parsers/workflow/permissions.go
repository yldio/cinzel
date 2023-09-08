package workflow

type PermissionsConfigOptions string

const (
	Read  PermissionsConfigOptions = "read"
	Write PermissionsConfigOptions = "write"
	None  PermissionsConfigOptions = "none"
)

type PermissionsConfig struct {
	Actions            *PermissionsConfigOptions `hcl:"actions,attr" yaml:"actions,omitempty"`
	Attestations       *PermissionsConfigOptions `hcl:"attestations,attr"  yaml:"attestations,omitempty"`
	Checks             *PermissionsConfigOptions `hcl:"checks,attr"  yaml:"checks,omitempty"`
	Contents           *PermissionsConfigOptions `hcl:"contents,attr"  yaml:"contents,omitempty"`
	Deployments        *PermissionsConfigOptions `hcl:"deployments,attr"  yaml:"deployments,omitempty"`
	IdToken            *PermissionsConfigOptions `hcl:"id_token,attr"  yaml:"id-token,omitempty"`
	Issues             *PermissionsConfigOptions `hcl:"issues,attr"  yaml:"issues,omitempty"`
	Discussions        *PermissionsConfigOptions `hcl:"discussions,attr"  yaml:"discussions,omitempty"`
	Packages           *PermissionsConfigOptions `hcl:"packages,attr"  yaml:"packages,omitempty"`
	Pages              *PermissionsConfigOptions `hcl:"pages,attr"  yaml:"pages,omitempty"`
	PullRequests       *PermissionsConfigOptions `hcl:"pull_requests,attr"  yaml:"pull-requests,omitempty"`
	RepositoryProjects *PermissionsConfigOptions `hcl:"repository_projects,attr"  yaml:"repository-projects,omitempty"`
	SecurityEvents     *PermissionsConfigOptions `hcl:"security_events,attr"  yaml:"security-events,omitempty"`
	Statuses           *PermissionsConfigOptions `hcl:"statuses,attr"  yaml:"statuses,omitempty"`
}

type Permissions PermissionsConfig

func (config *PermissionsConfig) Parse() (Permissions, error) {
	return Permissions(*config), nil
}
