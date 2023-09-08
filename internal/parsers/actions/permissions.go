package actions

import "errors"

type PermissionsConfigOptions string

const (
	Read  PermissionsConfigOptions = "read"
	Write PermissionsConfigOptions = "write"
	None  PermissionsConfigOptions = "none"
)

type PermissionsConfig struct {
	Actions            *PermissionsConfigOptions `hcl:"actions,attr"`
	Attestations       *PermissionsConfigOptions `hcl:"attestations,attr"`
	Checks             *PermissionsConfigOptions `hcl:"checks,attr"`
	Contents           *PermissionsConfigOptions `hcl:"contents,attr"`
	Deployments        *PermissionsConfigOptions `hcl:"deployments,attr"`
	IdToken            *PermissionsConfigOptions `hcl:"id_token,attr"`
	Issues             *PermissionsConfigOptions `hcl:"issues,attr"`
	Discussions        *PermissionsConfigOptions `hcl:"discussions,attr"`
	Packages           *PermissionsConfigOptions `hcl:"packages,attr"`
	Pages              *PermissionsConfigOptions `hcl:"pages,attr"`
	PullRequests       *PermissionsConfigOptions `hcl:"pull_requests,attr"`
	RepositoryProjects *PermissionsConfigOptions `hcl:"repository_projects,attr"`
	SecurityEvents     *PermissionsConfigOptions `hcl:"security_events,attr"`
	Statuses           *PermissionsConfigOptions `hcl:"statuses,attr"`
}

type Permissions []Permission

type Permission struct {
	Perm  string
	Value string
}

func getPermValue(perm *PermissionsConfigOptions) (string, error) {
	switch *perm {
	case Read:
		return "read", nil
	case Write:
		return "write", nil
	case None:
		return "none", nil
	default:
		return "", errors.New("wrong Permission, only \"read\", \"write\" and \"none\" are valid")
	}
}

func (permissions *PermissionsConfig) ConvertFromHcl() (Permissions, error) {
	content := Permissions{}

	if permissions == nil {
		return Permissions{}, nil
	}

	if permissions.Actions != nil {
		perm := Permission{
			Perm: "Actions",
		}
		value, err := getPermValue(permissions.Actions)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.Attestations != nil {
		perm := Permission{
			Perm: "Attestations",
		}
		value, err := getPermValue(permissions.Attestations)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.Checks != nil {
		perm := Permission{
			Perm: "Checks",
		}
		value, err := getPermValue(permissions.Checks)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.Contents != nil {
		perm := Permission{
			Perm: "Contents",
		}
		value, err := getPermValue(permissions.Contents)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.Deployments != nil {
		perm := Permission{
			Perm: "Deployments",
		}
		value, err := getPermValue(permissions.Deployments)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.IdToken != nil {
		perm := Permission{
			Perm: "IdToken",
		}
		value, err := getPermValue(permissions.IdToken)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.Issues != nil {
		perm := Permission{
			Perm: "Issues",
		}
		value, err := getPermValue(permissions.Issues)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.Discussions != nil {
		perm := Permission{
			Perm: "Discussions",
		}
		value, err := getPermValue(permissions.Discussions)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.Packages != nil {
		perm := Permission{
			Perm: "Packages",
		}
		value, err := getPermValue(permissions.Packages)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.Pages != nil {
		perm := Permission{
			Perm: "Pages",
		}
		value, err := getPermValue(permissions.Pages)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.PullRequests != nil {
		perm := Permission{
			Perm: "PullRequests",
		}
		value, err := getPermValue(permissions.PullRequests)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.RepositoryProjects != nil {
		perm := Permission{
			Perm: "RepositoryProjects",
		}
		value, err := getPermValue(permissions.RepositoryProjects)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.SecurityEvents != nil {
		perm := Permission{
			Perm: "SecurityEvents",
		}
		value, err := getPermValue(permissions.SecurityEvents)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	if permissions.Statuses != nil {
		perm := Permission{
			Perm: "Statuses",
		}
		value, err := getPermValue(permissions.Statuses)
		if err != nil {
			return Permissions{}, err
		}

		perm.Value = value
		content = append(content, perm)
	}
	return content, nil
}

func (permissions *Permissions) ConvertToYaml() (PermissionsYaml, error) {
	yaml := PermissionsYaml{}

	for _, perm := range *permissions {
		switch perm.Perm {
		case "Actions":
			yaml.Actions = perm.Value
		case "Attestations":
			yaml.Attestations = perm.Value
		case "Checks":
			yaml.Checks = perm.Value
		case "Contents":
			yaml.Contents = perm.Value
		case "Deployments":
			yaml.Deployments = perm.Value
		case "IdToken":
			yaml.IdToken = perm.Value
		case "Issues":
			yaml.Issues = perm.Value
		case "Discussions":
			yaml.Discussions = perm.Value
		case "Packages":
			yaml.Packages = perm.Value
		case "Pages":
			yaml.Pages = perm.Value
		case "PullRequests":
			yaml.PullRequests = perm.Value
		case "RepositoryProjects":
			yaml.RepositoryProjects = perm.Value
		case "SecurityEvents":
			yaml.SecurityEvents = perm.Value
		case "Statuses":
			yaml.Statuses = perm.Value
		}
	}

	return yaml, nil
}

type PermissionsYaml struct {
	Actions            string `yaml:"actions,omitempty"`
	Attestations       string `yaml:"attestations,omitempty"`
	Checks             string `yaml:"checks,omitempty"`
	Contents           string `yaml:"contents,omitempty"`
	Deployments        string `yaml:"deployments,omitempty"`
	IdToken            string `yaml:"id-token,omitempty"`
	Issues             string `yaml:"issues,omitempty"`
	Discussions        string `yaml:"discussions,omitempty"`
	Packages           string `yaml:"packages,omitempty"`
	Pages              string `yaml:"pages,omitempty"`
	PullRequests       string `yaml:"pull-requests,omitempty"`
	RepositoryProjects string `yaml:"repository-projects,omitempty"`
	SecurityEvents     string `yaml:"security-events,omitempty"`
	Statuses           string `yaml:"statuses,omitempty"`
}
