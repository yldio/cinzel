package actions

import "fmt"

func (workflowConfig *WorkflowConfig) ParseOnByFilter() (map[string]any, error) {
	var on = make(map[string]any)

	for _, onByFilter := range workflowConfig.OnByFilter {
		var filter = make(map[string][]string)
		if onByFilter.Filter != nil {
			if *onByFilter.Filter == ActivityTypes.ToString() {
				for _, activityType := range *onByFilter.Values {

					ok := ValidateActivityType(activityType)
					if !ok {
						return map[string]any{}, fmt.Errorf("activity type '%s' is not valid", activityType)
					}
					filter[*onByFilter.Filter] = append(filter[*onByFilter.Filter], activityType)
				}
			} else {
				filter[*onByFilter.Filter] = append(filter[*onByFilter.Filter], *onByFilter.Values...)
			}
		}
		on[onByFilter.Event] = filter
	}

	return on, nil
}
