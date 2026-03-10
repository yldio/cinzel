// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package workflow

type triggerRenameRule struct {
	event string
	from  string
	to    string
}

type onEventBlockRule struct {
	event     string
	yamlKey   string
	blockType string
}

var triggerRenameRules = []triggerRenameRule{
	{event: "workflow_call", from: "input", to: "inputs"},
	{event: "workflow_call", from: "output", to: "outputs"},
	{event: "workflow_call", from: "secret", to: "secrets"},
	{event: "workflow_dispatch", from: "input", to: "inputs"},
}

var onEventBlockRules = []onEventBlockRule{
	{event: "workflow_call", yamlKey: "inputs", blockType: "input"},
	{event: "workflow_call", yamlKey: "outputs", blockType: "output"},
	{event: "workflow_call", yamlKey: "secrets", blockType: "secret"},
	{event: "workflow_dispatch", yamlKey: "inputs", blockType: "input"},
}

var triggerRenameIndex = buildTriggerRenameIndex(triggerRenameRules)
var onEventBlockRuleIndex = buildOnEventBlockRuleIndex(onEventBlockRules)

// NormalizeOnEvent applies rename rules to an event's configuration map (e.g. input to inputs).
func NormalizeOnEvent(event string, value map[string]any) map[string]any {
	if len(value) == 0 {
		return value
	}

	for from, to := range triggerRenameIndex[event] {
		if v, ok := value[from]; ok {
			value[to] = v
			delete(value, from)
		}
	}

	return value
}

// TriggerBlockTypeForEventKey returns the HCL block type for a given event and YAML key.
func TriggerBlockTypeForEventKey(event string, key string) (string, bool) {
	rules, ok := onEventBlockRuleIndex[event]

	if !ok {
		return "", false
	}

	blockType, ok := rules[key]

	if ok {
		return blockType, true
	}

	return "", false
}

func buildTriggerRenameIndex(rules []triggerRenameRule) map[string]map[string]string {
	index := make(map[string]map[string]string)

	for _, rule := range rules {
		if _, ok := index[rule.event]; !ok {
			index[rule.event] = make(map[string]string)
		}

		index[rule.event][rule.from] = rule.to
	}

	return index
}

func buildOnEventBlockRuleIndex(rules []onEventBlockRule) map[string]map[string]string {
	index := make(map[string]map[string]string)

	for _, rule := range rules {
		if _, ok := index[rule.event]; !ok {
			index[rule.event] = make(map[string]string)
		}

		index[rule.event][rule.yamlKey] = rule.blockType
	}

	return index
}
