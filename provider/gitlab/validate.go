// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import "fmt"

func validatePipeline(pipeline map[string]any, jobs map[string]any) error {
	stagesSet := make(map[string]struct{})

	if rawStages, ok := pipeline["stages"]; ok {
		stages, ok := rawStages.([]any)

		if !ok {
			return fmt.Errorf("'stages' must be a list")
		}

		for _, s := range stages {
			name, ok := s.(string)

			if !ok || name == "" {
				return fmt.Errorf("stages must contain non-empty strings")
			}
			stagesSet[name] = struct{}{}
		}
	}

	for jobName, rawJob := range jobs {
		isTemplate := len(jobName) > 0 && jobName[0] == '.'

		jobMap, ok := rawJob.(map[string]any)

		if !ok {
			return fmt.Errorf("job '%s' must be an object", jobName)
		}

		script, ok := jobMap["script"]

		if !ok && !isTemplate {
			return fmt.Errorf("job '%s' must define 'script'", jobName)
		}

		scriptList, ok := script.([]any)

		if !isTemplate && (!ok || len(scriptList) == 0) {
			return fmt.Errorf("job '%s' script must be a non-empty list", jobName)
		}

		if len(stagesSet) > 0 && !isTemplate {
			stageRaw, ok := jobMap["stage"]

			if !ok {
				return fmt.Errorf("job '%s' must define 'stage'", jobName)
			}
			stage, ok := stageRaw.(string)

			if !ok || stage == "" {
				return fmt.Errorf("job '%s' stage must be a non-empty string", jobName)
			}

			if _, exists := stagesSet[stage]; !exists {
				return fmt.Errorf("job '%s' references undeclared stage '%s'", jobName, stage)
			}
		}

		if rawServices, ok := jobMap["services"]; ok {
			if err := validateServices(rawServices, fmt.Sprintf("job '%s'", jobName)); err != nil {
				return err
			}
		}
	}

	if rawDefault, ok := pipeline["default"]; ok {
		defaultMap, ok := rawDefault.(map[string]any)

		if !ok {
			return fmt.Errorf("default must be an object")
		}

		if rawServices, ok := defaultMap["services"]; ok {
			if err := validateServices(rawServices, "default"); err != nil {
				return err
			}
		}
	}

	graph := make(map[string][]string, len(jobs))

	for jobName, rawJob := range jobs {
		jobMap := rawJob.(map[string]any)
		graph[jobName] = []string{}

		if rawNeeds, ok := jobMap["needs"]; ok {
			needs, ok := rawNeeds.([]any)

			if !ok {
				return fmt.Errorf("job '%s' needs must be a list", jobName)
			}
			seen := map[string]struct{}{}

			for _, n := range needs {
				name, ok := n.(string)

				if !ok || name == "" {
					return fmt.Errorf("job '%s' needs must contain non-empty strings", jobName)
				}

				if _, dup := seen[name]; dup {
					return fmt.Errorf("job '%s' has duplicate needs '%s'", jobName, name)
				}
				seen[name] = struct{}{}

				if _, exists := jobs[name]; !exists {
					return fmt.Errorf("job '%s' needs unknown job '%s'", jobName, name)
				}
				graph[jobName] = append(graph[jobName], name)
			}
		}
	}

	visited := map[string]int{}
	var dfs func(string) error
	dfs = func(node string) error {
		state := visited[node]

		if state == 1 {
			return fmt.Errorf("depends_on cycle detected")
		}

		if state == 2 {
			return nil
		}
		visited[node] = 1

		for _, next := range graph[node] {
			if err := dfs(next); err != nil {
				return err
			}
		}
		visited[node] = 2

		return nil
	}

	for name := range graph {
		if err := dfs(name); err != nil {
			return err
		}
	}

	return nil
}

func validateServices(raw any, owner string) error {
	services, ok := raw.([]any)

	if !ok {
		return fmt.Errorf("%s services must be a list", owner)
	}

	if len(services) == 0 {
		return fmt.Errorf("%s services must not be empty", owner)
	}

	for _, item := range services {
		switch service := item.(type) {
		case string:
			if service == "" {
				return fmt.Errorf("%s services must contain non-empty strings", owner)
			}
		case map[string]any:
			nameRaw, ok := service["name"]

			if !ok {
				return fmt.Errorf("%s service entries must include 'name'", owner)
			}
			name, ok := nameRaw.(string)

			if !ok || name == "" {
				return fmt.Errorf("%s service name must be a non-empty string", owner)
			}
		default:
			return fmt.Errorf("%s services entries must be strings or objects", owner)
		}
	}

	return nil
}
