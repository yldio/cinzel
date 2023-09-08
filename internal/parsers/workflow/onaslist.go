package workflow

import (
	"fmt"

	"github.com/yldio/atos/internal/parsers/actions"
)

type OnAsListConfig []string

func (config *OnAsListConfig) Parse() ([]string, error) {
	if config != nil {
		for _, eventTrigger := range *config {
			ok := actions.ValidateEventTrigger(eventTrigger)
			if !ok {
				return []string{}, fmt.Errorf("event trigger '%s' is not valid", eventTrigger)
			}
		}
	}

	return *config, nil
}
