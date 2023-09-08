package step

type EnvConfig struct {
	Name  string `hcl:"name,attr"`
	Value string `hcl:"value,attr"`
}

type EnvsConfig []*EnvConfig

func (config *EnvsConfig) Parse() (map[string]any, error) {
	envs := make(map[string]any)

	for _, env := range *config {
		content, err := env.Parse()
		if err != nil {
			return map[string]any{}, err
		}

		envs[content.Name] = content.Value
	}
	return envs, nil
}

func (config *EnvConfig) Parse() (EnvConfig, error) {
	return EnvConfig{
		Name:  config.Name,
		Value: config.Value,
	}, nil
}
