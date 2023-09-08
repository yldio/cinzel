package step

import "fmt"

type UsesConfig struct {
	Action  string `hcl:"action,attr" yaml:"action"`
	Version string `hcl:"version,attr" yaml:"version"`
}

func (config *UsesConfig) Parse() (string, error) {
	return fmt.Sprintf("%s@%s", config.Action, config.Version), nil
}
