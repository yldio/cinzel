package actions

import "github.com/hashicorp/hcl/v2"

type JobConfig struct {
	Id    string         `hcl:",label"`
	Name  string         `hcl:"name,attr"`
	Steps hcl.Expression `hcl:"steps,attr"`
}

type Job struct {
	Id    string
	Name  string
	Steps []Step
}

type JobYaml struct {
	Name string `yaml:"name,omitempty"`
}

func (j *Job) ConvertToYaml() JobYaml {
	y := JobYaml{
		Name: j.Name,
	}

	return y
}
