package action

import (
	"errors"
	"fmt"
)

type UsesConfig struct {
	Action  string `hcl:"action,attr"`
	Version string `hcl:"version,attr"`
}

func (config *UsesConfig) Parse() (string, error) {
	if config == nil {
		return "", nil
	}

	if config.Action == "" || config.Version == "" {
		return "", errors.New("properties Action and Version must be defined")
	}

	return fmt.Sprintf("%s@%s", config.Action, config.Version), nil
}
