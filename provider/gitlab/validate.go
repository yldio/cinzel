// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

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
		jobMap, ok := rawJob.(map[string]any)
		if !ok {
			return fmt.Errorf("job '%s' must be an object", jobName)
		}

		script, ok := jobMap["script"]
		if !ok {
			return fmt.Errorf("job '%s' must define 'script'", jobName)
		}

		scriptList, ok := script.([]any)
		if !ok || len(scriptList) == 0 {
			return fmt.Errorf("job '%s' script must be a non-empty list", jobName)
		}

		if len(stagesSet) > 0 {
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
