package job

type WithInputConfig struct {
	Name  string `hcl:"name,attr"`
	Value string `hcl:"value,attr"`
}

type WithConfig struct {
	Input []WithInputConfig `hcl:"input,block"`
}

type With map[string]any

func (config *WithConfig) Parse() (With, error) {
	withs := make(With)

	for _, input := range config.Input {
		withs[input.Name] = input.Value
	}
	return withs, nil
}
