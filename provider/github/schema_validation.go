// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/maputil"
)

var allowedHCLAttributesByScope = map[string]map[string]struct{}{
	"workflow": {
		"filename":    {},
		"name":        {},
		"run_name":    {},
		"jobs":        {},
		"permissions": {},
		"concurrency": {},
	},
	"job": {
		"name":              {},
		"if":                {},
		"uses":              {},
		"steps":             {},
		"depends_on":        {},
		"secrets":           {},
		"continue_on_error": {},
		"timeout_minutes":   {},
	},
	"strategy": {
		"fail_fast":    {},
		"max_parallel": {},
		"include":      {},
		"exclude":      {},
	},
	"action": {
		"filename":    {},
		"name":        {},
		"description": {},
		"author":      {},
	},
	"action.input": {
		"description":         {},
		"required":            {},
		"default":             {},
		"deprecation_message": {},
	},
	"action.output": {
		"description": {},
		"value":       {},
	},
	"action.branding": {
		"icon":  {},
		"color": {},
	},
	"action.runs": {
		"using":      {},
		"main":       {},
		"pre":        {},
		"pre_if":     {},
		"post":       {},
		"post_if":    {},
		"image":      {},
		"args":       {},
		"entrypoint": {},
		"steps":      {},
	},
	"name_value": {
		"name":  {},
		"value": {},
	},
}

var allowedHCLBlocksByScope = map[string]map[string]struct{}{
	"workflow": {
		"on":          {},
		"env":         {},
		"permissions": {},
		"defaults":    {},
		"concurrency": {},
	},
	"job": {
		"uses":        {},
		"with":        {},
		"env":         {},
		"output":      {},
		"secret":      {},
		"service":     {},
		"runs_on":     {},
		"strategy":    {},
		"permissions": {},
		"defaults":    {},
		"concurrency": {},
		"container":   {},
		"environment": {},
	},
	"strategy": {
		"matrix": {},
	},
	"action": {
		"input":    {},
		"output":   {},
		"runs":     {},
		"branding": {},
	},
	"action.runs": {
		"env": {},
	},
}

var allowedWorkflowYAMLKeys = map[string]struct{}{
	"name": {}, "run-name": {}, "on": {}, "jobs": {}, "permissions": {}, "defaults": {}, "concurrency": {}, "env": {},
}

var allowedJobYAMLKeys = map[string]struct{}{
	"name": {}, "if": {}, "uses": {}, "with": {}, "secrets": {}, "permissions": {}, "defaults": {}, "concurrency": {},
	"container": {}, "services": {}, "environment": {}, "strategy": {}, "runs-on": {}, "steps": {}, "needs": {},
	"timeout-minutes": {}, "continue-on-error": {}, "outputs": {}, "env": {},
}

var allowedStepYAMLKeys = map[string]struct{}{
	"id": {}, "name": {}, "if": {}, "uses": {}, "run": {}, "shell": {}, "working-directory": {}, "with": {}, "env": {},
	"continue-on-error": {}, "timeout-minutes": {},
}

var allowedActionYAMLKeys = map[string]struct{}{
	"name": {}, "description": {}, "author": {}, "inputs": {}, "outputs": {}, "runs": {}, "branding": {},
}

var allowedActionRunsYAMLKeys = map[string]struct{}{
	"using": {}, "main": {}, "pre": {}, "pre-if": {}, "post": {}, "post-if": {}, "image": {}, "args": {}, "entrypoint": {}, "steps": {}, "env": {},
}

func validateHCLSchema(scope string, body *hclsyntax.Body) error {
	if allowedAttrs, ok := allowedHCLAttributesByScope[scope]; ok {
		for _, name := range maputil.SortedKeys(body.Attributes) {
			if _, allowed := allowedAttrs[name]; !allowed {
				return fmt.Errorf("unknown attribute '%s' in %s", name, scope)
			}
		}
	}

	if allowedBlocks, ok := allowedHCLBlocksByScope[scope]; ok {
		for _, block := range body.Blocks {
			if _, allowed := allowedBlocks[block.Type]; !allowed {
				return fmt.Errorf("unknown block '%s' in %s", block.Type, scope)
			}
		}
	}

	return nil
}

func validateAllowedYAMLKeys(path string, input map[string]any, allowed map[string]struct{}) error {
	for _, key := range maputil.SortedKeys(input) {
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("unknown key '%s' in %s", key, path)
		}
	}

	return nil
}
