package actions

import "strings"

type UsesConfg struct {
	Action  string `hcl:"action,attr"`
	Version string `hcl:"version,attr"`
}

type WithConfg struct {
	Args       *string `hcl:"args,attr"`
	Entrypoint *string `hcl:"entrypoint,attr"`
}

type StepConfig struct {
	Id               string     `hcl:",label"`
	If               *string    `hcl:"if,attr"`
	Name             *string    `hcl:"name,attr"`
	Uses             *UsesConfg `hcl:"uses,block"`
	Run              *string    `hcl:"run,attr"`
	WorkingDirectory *string    `hcl:"working_directory,attr"`
	Shell            *string    `hcl:"shell,attr"`
	With             *WithConfg `hcl:"with,block"`
	Env              *[]string  `hcl:"env,attr"`
	ContinueOnError  *string    `hcl:"continue_on_error,attr"`
	TimeoutMinutes   *int32     `hcl:"timeout_minutes,attr"`
}

type Uses struct {
	Action  string
	Version string
}

func (uses *Uses) ConvertYaml() string {
	var u strings.Builder

	u.WriteString(uses.Action)
	u.WriteString(uses.Version)

	return u.String()
}

type Step struct {
	Id   string
	If   string
	Name string
	Uses Uses `yaml:"uses,omitempty"`
	Run  string
}

type StepYaml struct {
	Name string `yaml:"name,omitempty"`
}

func (s *Step) ConvertToYaml() StepYaml {
	y := StepYaml{
		Name: s.Name,
	}

	return y
}
