package step

type WithConfig struct {
	Name  string `hcl:"name,attr"`
	Value string `hcl:"value,attr"`
}

type WithsConfig []*WithConfig

func (config *WithsConfig) Parse() (map[string]any, error) {
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

func (config *WithConfig) Parse() (WithConfig, error) {
	return WithConfig{
		Name:  config.Name,
		Value: config.Value,
	}, nil
}
