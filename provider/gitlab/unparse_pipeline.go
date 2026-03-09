// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package gitlab

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/naming"
	"github.com/zclconf/go-cty/cty"
)

func parseYAMLDocument(content []byte) (map[string]any, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func classifyPipelineDocument(doc map[string]any) bool {
	if rawStages, ok := doc["stages"]; ok {
		if _, isList := rawStages.([]any); isList {
			return true
		}
	}

	if rawWorkflow, ok := doc["workflow"]; ok {
		if workflowMap, isMap := toStringAnyMap(rawWorkflow); isMap {
			if _, hasRules := workflowMap["rules"]; hasRules {
				return true
			}
		}
	}

	for _, key := range sortedKeys(doc) {
		jobMap, ok := toStringAnyMap(doc[key])
		if !ok {
			continue
		}
		if _, hasScript := jobMap["script"]; hasScript {
			return true
		}
	}

	return false
}

func pipelineToHCL(doc map[string]any, filename string) ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	body := f.Body()

	if rawStages, ok := doc["stages"]; ok {
		if err := writeAttributeAny(body, "stages", escapeGitLabVariables(rawStages)); err != nil {
			return nil, err
		}
	}

	if rawVariables, ok := doc["variables"]; ok {
		variables, mapOK := toStringAnyMap(rawVariables)
		if !mapOK {
			return nil, fmt.Errorf("variables must be an object")
		}

		for _, name := range sortedKeys(variables) {
			if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
				body.AppendNewline()
			}

			varID := naming.SanitizeIdentifier(strings.ToLower(name))
			if varID == "" {
				varID = "var"
			}

			vb := body.AppendNewBlock("variable", []string{varID})
			vbody := vb.Body()
			vbody.SetAttributeValue("name", cty.StringVal(name))

			raw := variables[name]
			if vm, ok := toStringAnyMap(raw); ok {
				if val, hasVal := vm["value"]; hasVal {
					if err := writeAttributeAny(vbody, "value", escapeGitLabVariables(val)); err != nil {
						return nil, err
					}
				}
				if desc, hasDesc := vm["description"]; hasDesc {
					if err := writeAttributeAny(vbody, "description", escapeGitLabVariables(desc)); err != nil {
						return nil, err
					}
				}
			} else {
				if err := writeAttributeAny(vbody, "value", escapeGitLabVariables(raw)); err != nil {
					return nil, err
				}
			}
		}
	}

	if rawWorkflow, ok := doc["workflow"]; ok {
		workflowMap, mapOK := toStringAnyMap(rawWorkflow)
		if !mapOK {
			return nil, fmt.Errorf("workflow must be an object")
		}

		if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
			body.AppendNewline()
		}

		wb := body.AppendNewBlock("workflow", nil)
		wbody := wb.Body()

		if name, hasName := workflowMap["name"]; hasName {
			if err := writeAttributeAny(wbody, "name", escapeGitLabVariables(name)); err != nil {
				return nil, err
			}
		}

		if rawRules, hasRules := workflowMap["rules"]; hasRules {
			rules, ok := rawRules.([]any)
			if !ok {
				return nil, fmt.Errorf("workflow.rules must be a list")
			}
			for _, item := range rules {
				ruleMap, ok := toStringAnyMap(item)
				if !ok {
					return nil, fmt.Errorf("workflow.rules entries must be objects")
				}
				rb := wbody.AppendNewBlock("rule", nil)
				for _, key := range sortedKeys(ruleMap) {
					if err := writeAttributeAny(rb.Body(), key, escapeGitLabVariables(ruleMap[key])); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	if rawDefault, ok := doc["default"]; ok {
		defaultMap, mapOK := toStringAnyMap(rawDefault)
		if !mapOK {
			return nil, fmt.Errorf("default must be an object")
		}

		if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
			body.AppendNewline()
		}

		db := body.AppendNewBlock("default", nil)
		if err := writeGenericMap(db.Body(), defaultMap); err != nil {
			return nil, err
		}
	}

	jobNames := make([]string, 0)
	jobIDMap := make(map[string]string)
	usedIDs := make([]string, 0)
	for _, key := range sortedKeys(doc) {
		if isReservedTopLevelKey(key) {
			continue
		}
		if jobMap, ok := toStringAnyMap(doc[key]); ok {
			if _, hasScript := jobMap["script"]; hasScript {
				jobNames = append(jobNames, key)
				id := naming.SanitizeIdentifier(key)
				if id == "" {
					id = "job"
				}
				id = naming.UniqueIdentifier(id, usedIDs)
				usedIDs = append(usedIDs, id)
				jobIDMap[key] = id
			}
		}
	}
	sort.Strings(jobNames)

	for _, name := range jobNames {
		if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
			body.AppendNewline()
		}
		jobMap, _ := toStringAnyMap(doc[name])
		jb := body.AppendNewBlock("job", []string{jobIDMap[name]})
		if err := writeJobBlock(jb.Body(), jobMap, jobIDMap); err != nil {
			return nil, fmt.Errorf("error in job '%s': %w", name, err)
		}
	}

	for _, key := range sortedKeys(doc) {
		if isReservedTopLevelKey(key) {
			continue
		}
		if _, isMappedJob := jobIDMap[key]; isMappedJob {
			continue
		}

		fmt.Fprintf(os.Stderr, "warning: unsupported top-level key '%s' passed through\n", key)

		if hiddenJobMap, ok := toStringAnyMap(doc[key]); ok && strings.HasPrefix(key, ".") {
			if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
				body.AppendNewline()
			}
			tplID := naming.SanitizeIdentifier(strings.TrimPrefix(key, "."))
			if tplID == "" {
				tplID = "template"
			}
			tb := body.AppendNewBlock("template", []string{tplID})
			if err := writeGenericMap(tb.Body(), hiddenJobMap); err != nil {
				return nil, err
			}
			continue
		}

		if genericMap, ok := toStringAnyMap(doc[key]); ok {
			if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
				body.AppendNewline()
			}
			gb := body.AppendNewBlock(key, nil)
			if err := writeGenericMap(gb.Body(), genericMap); err != nil {
				return nil, err
			}
		} else {
			if err := writeAttributeAny(body, key, escapeGitLabVariables(doc[key])); err != nil {
				return nil, err
			}
		}
	}

	_ = filename
	return hclwrite.Format(f.Bytes()), nil
}

func writeJobBlock(body *hclwrite.Body, job map[string]any, jobIDMap map[string]string) error {
	for _, key := range sortedKeys(job) {
		value := job[key]
		switch key {
		case "needs":
			needs, ok := value.([]any)
			if !ok {
				return fmt.Errorf("needs must be a list")
			}
			refs := make([]string, 0, len(needs))
			for _, n := range needs {
				name, ok := n.(string)
				if !ok {
					return fmt.Errorf("needs entries must be strings")
				}
				refID, exists := jobIDMap[name]
				if !exists {
					refID = naming.SanitizeIdentifier(name)
				}
				refs = append(refs, refID)
			}
			if err := writeReferenceListAttribute(body, "depends_on", "job", refs); err != nil {
				return err
			}
		case "rules":
			rules, ok := value.([]any)
			if !ok {
				return fmt.Errorf("rules must be a list")
			}
			for _, item := range rules {
				ruleMap, ok := toStringAnyMap(item)
				if !ok {
					return fmt.Errorf("rules entries must be objects")
				}
				rb := body.AppendNewBlock("rule", nil)
				for _, attr := range sortedKeys(ruleMap) {
					if err := writeAttributeAny(rb.Body(), attr, escapeGitLabVariables(ruleMap[attr])); err != nil {
						return err
					}
				}
			}
		case "cache", "artifacts":
			mapVal, ok := toStringAnyMap(value)
			if !ok {
				return fmt.Errorf("%s must be an object", key)
			}
			b := body.AppendNewBlock(key, nil)
			if err := writeGenericMap(b.Body(), mapVal); err != nil {
				return err
			}
		default:
			if err := writeAttributeAny(body, key, escapeGitLabVariables(value)); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeGenericMap(body *hclwrite.Body, mapping map[string]any) error {
	for _, key := range sortedKeys(mapping) {
		value := mapping[key]
		if nested, ok := toStringAnyMap(value); ok {
			b := body.AppendNewBlock(key, nil)
			if err := writeGenericMap(b.Body(), nested); err != nil {
				return err
			}
			continue
		}
		if err := writeAttributeAny(body, key, escapeGitLabVariables(value)); err != nil {
			return err
		}
	}
	return nil
}

func escapeGitLabVariables(value any) any {
	switch v := value.(type) {
	case string:
		return v
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, escapeGitLabVariables(item))
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			out[key] = escapeGitLabVariables(item)
		}
		return out
	default:
		return value
	}
}

func toStringAnyMap(raw any) (map[string]any, bool) {
	m, ok := raw.(map[string]any)
	return m, ok
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func isReservedTopLevelKey(key string) bool {
	switch key {
	case "stages", "variables", "workflow", "default":
		return true
	default:
		return false
	}
}
