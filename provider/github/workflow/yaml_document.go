// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

import "fmt"

// YAMLDocument represents a parsed GitHub Actions workflow YAML document.
type YAMLDocument struct {
	Raw   map[string]any
	HasOn bool
	On    map[string]any
	Jobs  map[string]any
}

// NewYAMLDocument constructs a YAMLDocument from raw YAML data, returning false if neither on nor jobs are present.
func NewYAMLDocument(raw map[string]any, mapper func(any) (map[string]any, bool)) (YAMLDocument, bool, error) {
	rawOn, hasOn := raw["on"]
	jobs, hasJobs := mapper(raw["jobs"])

	if !hasOn && !hasJobs {

		return YAMLDocument{}, false, nil
	}

	on := map[string]any{}

	if hasOn {
		normalized, err := NormalizeOn(rawOn, mapper)
		if err != nil {

			return YAMLDocument{}, true, err
		}

		on = normalized
	}

	return YAMLDocument{Raw: raw, HasOn: hasOn, On: on, Jobs: jobs}, true, nil
}

// NormalizeOn converts the various YAML representations of 'on' into a uniform map.
func NormalizeOn(raw any, mapper func(any) (map[string]any, bool)) (map[string]any, error) {
	switch v := raw.(type) {
	case nil:
		return map[string]any{}, nil
	case string:
		if v == "" {

			return nil, fmt.Errorf("workflow 'on' event name must not be empty")
		}

		return map[string]any{v: map[string]any{}}, nil
	case []any:
		on := make(map[string]any, len(v))

		for _, item := range v {
			eventName, ok := item.(string)

			if !ok || eventName == "" {

				return nil, fmt.Errorf("workflow 'on' list entries must be non-empty strings")
			}

			on[eventName] = map[string]any{}
		}

		return on, nil
	default:
		onMap, ok := mapper(raw)

		if !ok {

			return nil, fmt.Errorf("workflow 'on' must be a string, list of strings, or object")
		}

		on := make(map[string]any, len(onMap))

		for eventName, eventRaw := range onMap {

			if eventName == "" {

				return nil, fmt.Errorf("workflow 'on' event name must not be empty")
			}

			if eventRaw == nil {
				on[eventName] = map[string]any{}
				continue
			}

			if eventName == "schedule" {
				normalizedSchedule, err := normalizeScheduleEvent(eventRaw, mapper)
				if err != nil {

					return nil, err
				}

				on[eventName] = normalizedSchedule
				continue
			}

			eventMap, mapOK := mapper(eventRaw)

			if !mapOK {

				if b, boolOK := eventRaw.(bool); boolOK && b {
					on[eventName] = map[string]any{}
					continue
				}

				return nil, fmt.Errorf("workflow 'on.%s' must be an object", eventName)
			}

			on[eventName] = NormalizeOnEvent(eventName, eventMap)
		}

		return on, nil
	}
}

// DenormalizeScheduleEvent is the inverse of normalizeScheduleEvent.
// It converts {cron: ["A", "B"]} back to [{cron: "A"}, {cron: "B"}]
// as required by GitHub Actions.
func DenormalizeScheduleEvent(normalized map[string]any) []any {
	cronVals, ok := normalized["cron"]

	if !ok {

		return []any{normalized}
	}

	list, ok := cronVals.([]any)

	if !ok {

		return []any{map[string]any{"cron": cronVals}}
	}

	items := make([]any, 0, len(list))

	for _, cron := range list {
		items = append(items, map[string]any{"cron": cron})
	}

	return items
}

func normalizeScheduleEvent(raw any, mapper func(any) (map[string]any, bool)) (map[string]any, error) {
	items, ok := raw.([]any)

	if !ok {
		eventMap, mapOK := mapper(raw)

		if !mapOK {

			return nil, fmt.Errorf("workflow 'on.schedule' must be a list")
		}

		return eventMap, nil
	}

	cronValues := make([]any, 0, len(items))

	for _, item := range items {
		entry, mapOK := mapper(item)

		if !mapOK {

			return nil, fmt.Errorf("workflow 'on.schedule' entries must be objects")
		}

		cron, ok := entry["cron"].(string)

		if !ok || cron == "" {

			return nil, fmt.Errorf("workflow 'on.schedule' entries must define non-empty 'cron'")
		}

		cronValues = append(cronValues, cron)
	}

	return map[string]any{"cron": cronValues}, nil
}
