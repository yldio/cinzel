// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/maputil"
	"github.com/yldio/cinzel/internal/naming"
	"github.com/yldio/cinzel/provider/github/step"
	ghworkflow "github.com/yldio/cinzel/provider/github/workflow"
	yamlv3 "gopkg.in/yaml.v3"
)

// parseYAMLDocument parses YAML content into a document map and extracts
// job names in source order in a single pass using the yaml.v3 Node API.
func parseYAMLDocument(content []byte) (map[string]any, []string, error) {
	var node yamlv3.Node

	if err := yamlv3.Unmarshal(content, &node); err != nil {
		return nil, nil, err
	}

	if len(node.Content) == 0 {
		return nil, nil, nil
	}

	var doc map[string]any

	if err := node.Decode(&doc); err != nil {
		return nil, nil, err
	}

	return doc, jobOrderFromNode(node.Content[0]), nil
}

// jobOrderFromNode extracts job names in source order from a yaml.v3 mapping
// node. It relies on the Node API's preservation of mapping key order, which
// is not available when unmarshaling directly into map[string]any.
func jobOrderFromNode(root *yamlv3.Node) []string {
	if root.Kind != yamlv3.MappingNode {
		return nil
	}

	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == "jobs" {
			jobs := root.Content[i+1]

			if jobs.Kind != yamlv3.MappingNode {
				return nil
			}

			keys := make([]string, 0, len(jobs.Content)/2)

			for j := 0; j+1 < len(jobs.Content); j += 2 {
				if key := jobs.Content[j].Value; key != "" {
					keys = append(keys, key)
				}
			}

			return keys
		}
	}

	return nil
}

func classifyWorkflowDocument(doc map[string]any) (*ghworkflow.YAMLDocument, error) {
	if len(doc) == 0 {
		return nil, nil
	}

	workflowDoc, isWorkflow, err := ghworkflow.NewYAMLDocument(doc, maputil.ToStringAnyMap)
	if err != nil {
		return nil, err
	}

	if isWorkflow {
		return &workflowDoc, nil
	}

	return nil, nil
}

func workflowToHCL(doc ghworkflow.YAMLDocument, filename string, jobOrder []string) ([]byte, error) {
	if err := validateWorkflowYAMLDoc(doc); err != nil {
		return nil, err
	}

	f, root, workflowBody := newWorkflowRoot(filename)
	generatedVariables := map[string]any{}
	stepRegistry := map[string]string{}
	usedStepIDs := map[string]int{}

	if len(doc.Jobs) == 0 {
		return nil, errors.New("workflow must define at least one job in 'jobs'")
	}

	jobEntries, jobRefs, jobIDMap, err := buildWorkflowJobIndex(doc.Jobs, jobOrder)
	if err != nil {
		return nil, err
	}

	if err := writeWorkflowMetadata(workflowBody, doc); err != nil {
		return nil, err
	}

	if len(workflowBody.Attributes()) > 0 || len(workflowBody.Blocks()) > 0 {
		workflowBody.AppendNewline()
	}

	if err := writeReferenceListAttribute(workflowBody, "jobs", "job", jobRefs); err != nil {
		return nil, err
	}

	if err := writeWorkflowJobs(root, jobEntries, jobIDMap, generatedVariables, stepRegistry, usedStepIDs); err != nil {
		return nil, err
	}

	if err := writeGeneratedVariables(root, generatedVariables); err != nil {
		return nil, err
	}

	return unescapeHCLUnicode(hclwrite.Format(f.Bytes())), nil
}

// unescapeHCLUnicode replaces \uXXXX and \UXXXXXXXX escape sequences in HCL
// source with their raw UTF-8 equivalents for characters above U+009F.
// hclwrite escapes any rune where Go's unicode.IsPrint returns false, which
// includes category-Cf characters like U+200D (ZWJ) used in emoji sequences.
// These are valid UTF-8 in HCL strings; keeping them escaped causes downstream
// YAML serialisers to re-escape the surrounding emoji.
func unescapeHCLUnicode(src []byte) []byte {
	return reHCLUnicodeEscape.ReplaceAllFunc(src, func(match []byte) []byte {
		n, err := strconv.ParseInt(string(match[2:]), 16, 32)
		if err != nil || n <= 0x9F || !utf8.ValidRune(rune(n)) {
			return match
		}

		var buf [utf8.UTFMax]byte
		l := utf8.EncodeRune(buf[:], rune(n))

		return append([]byte(nil), buf[:l]...)
	})
}

var reHCLUnicodeEscape = regexp.MustCompile(`\\u[0-9a-fA-F]{4}|\\U[0-9a-fA-F]{8}`)

func writeJobBody(root *hclwrite.Body, jobBody *hclwrite.Body, jobID string, job map[string]any, jobIDMap map[string]string, generatedVariables map[string]any, stepRegistry map[string]string, usedStepIDs map[string]int) error {
	stepRefs := []string{}

	for _, key := range sortedKeys(job) {
		if key == "steps" {
			refs, err := writeJobSteps(root, job[key], stepRegistry, usedStepIDs)
			if err != nil {
				return err
			}

			stepRefs = append(stepRefs, refs...)
			continue
		}

		if len(jobBody.Attributes()) > 0 || len(jobBody.Blocks()) > 0 {
			jobBody.AppendNewline()
		}

		if err := writeJobKey(root, jobBody, jobID, key, job[key], jobIDMap, generatedVariables, stepRegistry, usedStepIDs, &stepRefs); err != nil {
			return err
		}
	}

	if len(stepRefs) > 0 {
		if len(jobBody.Attributes()) > 0 || len(jobBody.Blocks()) > 0 {
			jobBody.AppendNewline()
		}

		if err := writeReferenceListAttribute(jobBody, "steps", "step", stepRefs); err != nil {
			return err
		}
	}

	return nil
}

func writeServicesBlocks(body *hclwrite.Body, raw any) error {
	services, ok := toStringAnyMap(raw)

	if !ok {
		return errors.New("services must be an object")
	}

	for _, serviceName := range sortedKeys(services) {
		svcVal, ok := toStringAnyMap(services[serviceName])

		if !ok {
			return fmt.Errorf("service '%s' must be an object", serviceName)
		}

		serviceBlock := body.AppendNewBlock("service", []string{serviceName})
		serviceBody := serviceBlock.Body()

		for _, key := range sortedKeys(svcVal) {
			value := svcVal[key]
			switch key {
			case "env":
				if err := writeNameValueBlocks(serviceBody, "env", value); err != nil {
					return err
				}
			case "credentials":
				if err := writeNestedMapAsBlock(serviceBody, key, value); err != nil {
					return err
				}
			default:
				if err := writeAttributeAny(serviceBody, toHCLKey(key), value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func writeRunsOn(body *hclwrite.Body, raw any) error {
	block := body.AppendNewBlock("runs_on", nil)
	blockBody := block.Body()

	if str, ok := raw.(string); ok {
		return writeAttributeAny(blockBody, "runners", str)
	}

	if list, ok := raw.([]any); ok {
		return writeAttributeAny(blockBody, "runners", list)
	}

	mapping, ok := toStringAnyMap(raw)

	if !ok {
		return errors.New("runs-on must be a string, list, or an object")
	}

	for _, key := range sortedKeys(mapping) {
		if err := writeAttributeAny(blockBody, toHCLKey(key), mapping[key]); err != nil {
			return err
		}
	}

	return nil
}

func writeNestedMapAsBlock(body *hclwrite.Body, blockType string, raw any) error {
	if blockType == "env" {
		return writeNameValueBlocks(body, "env", raw)
	}

	if blockType == "with" {
		return writeNameValueBlocks(body, "with", raw)
	}

	if blockType == "output" || blockType == "outputs" {
		return writeNameValueBlocks(body, "output", raw)
	}

	if blockType == "secret" || blockType == "secrets" {
		return writeNameValueBlocks(body, "secret", raw)
	}

	mapping, ok := toStringAnyMap(raw)

	if !ok {
		return writeAttributeAny(body, toHCLKey(blockType), raw)
	}

	block := body.AppendNewBlock(toHCLKey(blockType), nil)
	blockBody := block.Body()

	for _, key := range sortedKeys(mapping) {
		value := mapping[key]

		if nestedMap, isMap := toStringAnyMap(value); isMap {
			if err := writeNestedMapAsBlock(blockBody, key, nestedMap); err != nil {
				return err
			}
			continue
		}

		if err := writeAttributeAny(blockBody, toHCLKey(key), value); err != nil {
			return err
		}
	}

	return nil
}

func writeNameValueBlocks(body *hclwrite.Body, blockType string, raw any) error {
	mapping, ok := toStringAnyMap(raw)

	if !ok {
		return fmt.Errorf("%s must be an object", blockType)
	}

	for _, key := range sortedKeys(mapping) {
		block := body.AppendNewBlock(blockType, nil)
		blockBody := block.Body()

		if err := writeAttributeAny(blockBody, "name", key); err != nil {
			return err
		}

		if err := writeAttributeAny(blockBody, "value", mapping[key]); err != nil {
			return err
		}
	}

	return nil
}

func writeAttributeAny(body *hclwrite.Body, attr string, raw any) error {
	ctyValue, err := anyToCty(raw)
	if err != nil {
		return err
	}

	body.SetAttributeValue(attr, ctyValue)

	return nil
}

func writeReferenceListAttribute(body *hclwrite.Body, attr string, root string, refs []string) error {
	if len(refs) == 0 {
		return nil
	}

	tokens := hclwrite.Tokens{{Type: hclsyntax.TokenOBrack, Bytes: []byte("[")}, {Type: hclsyntax.TokenNewline, Bytes: []byte("\n")}}

	for _, ref := range refs {
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("%s.%s", root, ref))})
		tokens = append(tokens,
			&hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(",")},
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
		)
	}

	tokens = append(tokens,
		&hclwrite.Token{Type: hclsyntax.TokenCBrack, Bytes: []byte("]")},
	)

	body.SetAttributeRaw(attr, tokens)

	return nil
}

func traversalTokens(root string, attr string) hclwrite.Tokens {
	return hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("%s.%s", root, attr))},
	}
}

func stepFromMap(value map[string]any) (step.Step, error) {
	ctyValue, err := anyToCty(value)
	if err != nil {
		return step.Step{}, err
	}

	var s step.Step

	if err := s.PreDecode(ctyValue); err != nil {
		return step.Step{}, err
	}

	return s, nil
}

func stepIdentifier(idx int, stepMap map[string]any, used map[string]int) string {
	id := ""

	if raw, ok := stepMap["id"].(string); ok && raw != "" {
		id = sanitizeIdentifier(raw)
	}

	if id == "" {
		if name, ok := stepMap["name"].(string); ok && name != "" {
			id = sanitizeIdentifier(name)
		}
	}

	if id == "" {
		if uses, ok := stepMap["uses"].(string); ok && uses != "" {
			id = sanitizeIdentifier(stepActionName(uses))
		}
	}

	if id == "" {
		if run, ok := stepMap["run"].(string); ok && run != "" {
			id = sanitizeIdentifier(stepFirstWord(run))
		}
	}

	if id == "" {
		id = fmt.Sprintf("step_%d", idx+1)
	}

	id = strings.ToLower(id)

	if count, exists := used[id]; exists {
		used[id] = count + 1

		return fmt.Sprintf("%s_%d", id, count+1)
	}

	used[id] = 0

	return id
}

// stepActionName extracts a short name from a uses string such as
// "actions/checkout@v4" → "checkout".
func stepActionName(uses string) string {
	if at := strings.IndexByte(uses, '@'); at >= 0 {
		uses = uses[:at]
	}

	if slash := strings.LastIndexByte(uses, '/'); slash >= 0 {
		uses = uses[slash+1:]
	}

	return uses
}

// stepFirstWord returns the first whitespace-delimited word of the first
// line of a run script, used as a fallback step identifier.
func stepFirstWord(run string) string {
	line := run

	if nl := strings.IndexByte(run, '\n'); nl >= 0 {
		line = run[:nl]
	}

	line = strings.TrimSpace(line)

	if sp := strings.IndexAny(line, " \t"); sp >= 0 {
		line = line[:sp]
	}

	return line
}

// stepFingerprint returns a canonical string representing the content of a
// step, excluding the step id field which is assigned by the unparser.
// Used to detect duplicate steps across jobs.
func stepFingerprint(stepMap map[string]any) string {
	m := make(map[string]any, len(stepMap))

	for k, v := range stepMap {
		if k != "id" {
			m[k] = v
		}
	}

	b, _ := json.Marshal(m)

	return string(b)
}

func normalizeNeeds(raw any, jobIDMap map[string]string) ([]string, error) {
	list, ok := raw.([]any)

	if !ok {
		if one, ok := raw.(string); ok {
			return []string{jobIDMapOrSanitized(one, jobIDMap)}, nil
		}

		return nil, errors.New("'needs' must be a string or a list")
	}

	refs := make([]string, 0, len(list))

	for _, item := range list {
		name, ok := item.(string)

		if !ok {
			return nil, errors.New("'needs' entries must be strings")
		}

		refs = append(refs, jobIDMapOrSanitized(name, jobIDMap))
	}

	return refs, nil
}

func jobIDMapOrSanitized(name string, ids map[string]string) string {
	if v, ok := ids[name]; ok {
		return v
	}

	return sanitizeIdentifier(name)
}

func toStringAnyMap(value any) (map[string]any, bool) {
	return maputil.ToStringAnyMap(value)
}

func sortedKeys[T any](mapping map[string]T) []string {
	return maputil.SortedKeys(mapping)
}

func sanitizeIdentifier(in string) string {
	return naming.SanitizeIdentifier(in)
}

func uniqueIdentifier(base string, existing []string) string {
	return naming.UniqueIdentifier(base, existing)
}

func uniqueIdentifierInSet(base string, existing map[string]struct{}) string {
	return naming.UniqueIdentifierInSet(base, existing)
}

func toHCLKey(name string) string {
	return naming.ToHCLKey(name)
}
