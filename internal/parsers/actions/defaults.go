package actions

type RunConfig struct {
	Shell            string `hcl:"shell,attr"`
	WorkingDirectory string `hcl:"working_directory,attr"`
}

type DefaultsConfig struct {
	Run *RunConfig `hcl:"run,block"`
}

func (defaults *DefaultsConfig) ConvertFromHcl() (Defaults, error) {
	if defaults == nil {
		return Defaults{}, nil
	}

	if defaults.Run == nil {
		return Defaults{}, nil
	}

	content := Defaults{
		Run: Run{
			Shell:            defaults.Run.Shell,
			WorkingDirectory: defaults.Run.WorkingDirectory,
		},
	}

	return content, nil
}

type Run struct {
	Shell            string
	WorkingDirectory string
}

type Defaults struct {
	Run Run
}

type RunYaml struct {
	Shell            string `yaml:"shell,omitempty"`
	WorkingDirectory string `yaml:"working-directory,omitempty"`
}

type DefaultsYaml struct {
	Run RunYaml
}

func (Defaults *Defaults) ConvertToYaml() (DefaultsYaml, error) {
	yaml := DefaultsYaml{
		Run: RunYaml{
			Shell:            Defaults.Run.Shell,
			WorkingDirectory: Defaults.Run.WorkingDirectory,
		},
	}

	return yaml, nil
}
