// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/maputil"
	"github.com/yldio/cinzel/internal/naming"
	"github.com/yldio/cinzel/internal/unescape"
	"github.com/yldio/cinzel/provider/github/step"
	ghworkflow "github.com/yldio/cinzel/provider/github/workflow"
	yamlv3 "gopkg.in/yaml.v3"
)

// parseYAMLDocument parses YAML content into a document map and extracts
// job names in source order in a single pass using the yaml.v3 Node API.
func parseYAMLDocument(content []byte) (map[string]any, []string, error) {
	// GitHub reads one document per file. A file holding more than one used to
	// be cut short at the first, losing the rest without a word, so the whole
	// file is rejected instead.
	dec := yamlv3.NewDecoder(bytes.NewReader(content))

	var first *yamlv3.Node

	for {
		var node yamlv3.Node

		err := dec.Decode(&node)

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, nil, err
		}

		if len(node.Content) == 0 || isNullNode(node.Content[0]) {
			continue
		}

		if first != nil {
			return nil, nil, errMultipleDocuments
		}

		first = node.Content[0]
	}

	if first == nil {
		return nil, nil, nil
	}

	if err := rejectNonStringKeys(first); err != nil {
		return nil, nil, err
	}

	keepWholeNumbersExact(first)

	var doc map[string]any

	if err := first.Decode(&doc); err != nil {
		return nil, nil, err
	}

	return doc, jobOrderFromNode(first), nil
}

// keepWholeNumbersExact retags a whole number too large for an integer so it
// decodes as the text it was written as.
//
// yaml.v3 resolves a run of digits that does not fit in int64 or uint64 to a
// float, and a float that wide has no room for every digit: an ID written as
// 99999999999999999999 comes back as 1e+20 and is emitted as
// 100000000000000000000, a different number, without a word. Left as text the
// digits survive, which is what goccy does with the same input and what the
// GitLab provider therefore already does.
//
// Only a plain run of digits is retagged. A quoted or explicitly tagged scalar
// carries a style, and .inf and .nan are not whole numbers, so both keep the
// float the file asked for.
func keepWholeNumbersExact(node *yamlv3.Node) {
	if node.Kind == yamlv3.ScalarNode && node.Tag == "!!float" && node.Style == 0 {
		if _, whole := new(big.Int).SetString(node.Value, 10); whole {
			node.Tag = "!!str"
		}
	}

	for _, child := range node.Content {
		keepWholeNumbersExact(child)
	}
}

// rejectNonStringKeys refuses a mapping key that is not a string.
//
// yaml.v3 decodes a mapping holding one into map[any]any rather than
// map[string]any, and a null key lands in that map as a nil, which crashes the
// encoder the validator runs the document through. A number or a boolean does
// not crash but is dropped just as quietly: the job order reads the key as
// written while the jobs map is keyed by a value no lookup here can produce,
// so the job goes missing without a word.
//
// A merge key is a string as far as this matters; the decoder folds it away
// before anything else sees the document.
func rejectNonStringKeys(node *yamlv3.Node) error {
	if node.Kind == yamlv3.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			if tag := node.Content[i].Tag; tag != "" && tag != "!!str" && tag != "!!merge" {
				return fmt.Errorf("%w: %s", errNonStringKey, node.Content[i].Value)
			}
		}
	}

	for _, child := range node.Content {
		if err := rejectNonStringKeys(child); err != nil {
			return err
		}
	}

	return nil
}

// isNullNode reports whether a document holds nothing. A stray "---" marker
// at either end of a file decodes to a null scalar, which is not a second
// document in any sense that matters here.
func isNullNode(node *yamlv3.Node) bool {
	return node.Kind == yamlv3.ScalarNode && node.Tag == "!!null"
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
				// A merge key is not a job. The decoder folds what it
				// points at into the mapping, so those jobs are named by
				// the keys they arrive under, which is more than this pass
				// can see: the order falls back to sorted keys instead.
				if jobs.Content[j].Tag == "!!merge" {
					return nil
				}

				keys = append(keys, jobKeyName(jobs.Content[j]))
			}

			return keys
		}
	}

	return nil
}

// jobKeyName returns the name a job key node carries once it is resolved.
//
// An alias in key position holds the anchor's name in Value and the text it
// stands for in Alias, while Decode records the resolved text. Reading Value
// here would put a name in the order that the jobs map does not hold, and the
// job would be reported as missing under a name that never appears in the
// file.
func jobKeyName(key *yamlv3.Node) string {
	if key.Kind == yamlv3.AliasNode && key.Alias != nil {
		return key.Alias.Value
	}

	return key.Value
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
	usedStepIDs := map[string]struct{}{}

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

	return unescape.Unicode(hclwrite.Format(f.Bytes())), nil
}

func writeJobBody(root *hclwrite.Body, jobBody *hclwrite.Body, jobID string, job map[string]any, jobIDMap map[string]string, generatedVariables map[string]any, stepRegistry map[string]string, usedStepIDs map[string]struct{}) error {
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

		if err := checkHCLKeyRoundtrips(blockType, key); err != nil {
			return err
		}

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

// checkHCLKeyRoundtrips rejects a key these blocks cannot carry back out.
// Their keys become bare HCL identifiers, and parse maps "_" to "-" so that
// hand-written "cancel_in_progress" reads as "cancel-in-progress". That leaves
// two keys unwritable: one already holding "_" would come back holding "-"
// instead, and one holding anything outside an identifier, such as a dot, would
// produce HCL that does not parse. Every key in the GitHub schema for these
// blocks is a plain word or hyphenated, so this rejects only input that was
// silently corrupted before.
func checkHCLKeyRoundtrips(blockType string, key string) error {
	if strings.Contains(key, "_") {
		return fmt.Errorf("%s key '%s' cannot be written to HCL: an underscore would be read back as a dash", blockType, key)
	}

	if !hclsyntax.ValidIdentifier(toHCLKey(key)) {
		return fmt.Errorf("%s key '%s' cannot be written to HCL: it is not a valid identifier", blockType, key)
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

func stepIdentifier(idx int, stepMap map[string]any, used map[string]struct{}) string {
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

	id = naming.UniqueIdentifierInSet(strings.ToLower(id), used)
	used[id] = struct{}{}

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

func uniqueIdentifierInSet(base string, existing map[string]struct{}) string {
	return naming.UniqueIdentifierInSet(base, existing)
}

func toHCLKey(name string) string {
	return naming.ToHCLKey(name)
}
