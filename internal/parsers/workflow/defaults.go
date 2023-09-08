package workflow

type RunConfig struct {
	Shell            string `hcl:"shell,attr" yaml:"shell"`
	WorkingDirectory string `hcl:"working_directory,attr" yaml:"working-directory"`
}

type DefaultsConfig struct {
	Run RunConfig `hcl:"run,block" yaml:"run"`
}

func (config *DefaultsConfig) Parse() (DefaultsConfig, error) {
	return *config, nil
}
