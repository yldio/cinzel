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

		// A job needs no script of its own: a trigger job has none by
		// definition, and an extending job inherits one. A script that is
		// present still has to be a non-empty list.
		_, hasTrigger := jobMap["trigger"]
		_, hasExtends := jobMap["extends"]
		inherits := isTemplate || hasTrigger || hasExtends

		// A job written in the steps syntax carries its commands under "run"
		// instead of "script". It still needs a stage, so this is not the
		// same as inheriting.
		_, hasRun := jobMap["run"]

		script, ok := jobMap["script"]

		if !ok && !inherits && !hasRun {
			return fmt.Errorf("job '%s' must define 'script'", jobName)
		}

		if ok && !isNonEmptyScript(script) {
			return fmt.Errorf("job '%s' script must be a non-empty string or list", jobName)
		}

		// A job that inherits can inherit its stage too, and GitLab defaults
		// an unnamed stage to "test" rather than failing.
		if len(stagesSet) > 0 && !inherits {
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

	// config.go accepts a top-level "services" alongside the one under
	// "default", but only the default was checked, so a top-level entry with no
	// name went straight through.
	if rawServices, ok := pipeline["services"]; ok {
		if err := validateServices(rawServices, "pipeline"); err != nil {
			return err
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
				name, ok := needName(n)

				if !ok {
					return fmt.Errorf("job '%s' needs must contain non-empty strings or objects naming a job", jobName)
				}

				// A cross-project need names no job in this pipeline,
				// so there is nothing to check it against.
				if name == "" {
					continue
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
	// An explicit null is how GitLab spells "override whatever this would
	// inherit", which the parse side keeps for exactly that reason. Rejecting
	// it here made "services = null" the one null collection that could be
	// written and not read back, while cache and rules round-tripped.
	if raw == nil {
		return nil
	}

	services, ok := raw.([]any)

	if !ok {
		return fmt.Errorf("%s services must be a list", owner)
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

// needName returns the job a "needs" entry names, and whether the entry is
// well formed. A cross-project or cross-pipeline entry names no job in this
// pipeline and yields an empty name.
func needName(entry any) (string, bool) {
	if name, ok := entry.(string); ok {
		return name, name != ""
	}

	need, ok := entry.(map[string]any)

	if !ok {
		return "", false
	}

	raw, hasJob := need["job"]

	if !hasJob {
		_, crossPipeline := need["pipeline"]

		return "", crossPipeline
	}

	name, ok := raw.(string)

	if !ok || name == "" {
		return "", false
	}

	if _, crossProject := need["project"]; crossProject {
		return "", true
	}

	return name, true
}

// isNonEmptyScript reports whether a script is one GitLab would run. GitLab
// takes a single command as a bare string as well as a list of them.
func isNonEmptyScript(script any) bool {
	if line, isString := script.(string); isString {
		return line != ""
	}

	list, isList := script.([]any)

	return isList && len(list) > 0
}
