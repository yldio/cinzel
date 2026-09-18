// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/hclcomment"
	"github.com/yldio/cinzel/internal/naming"
	"github.com/yldio/cinzel/internal/unescape"
	"github.com/zclconf/go-cty/cty"
	yamlv3 "gopkg.in/yaml.v3"
)

func parseYAMLDocument(content []byte) (map[string]any, error) {
	if err := checkYAMLSoundness(content); err != nil {
		return nil, err
	}

	// A pipeline carrying a "spec" header is two documents: the header, then
	// the configuration. Merging them keeps the whole file, which a single
	// Unmarshal would silently cut short at the first document.
	dec := yaml.NewDecoder(bytes.NewReader(content))
	merged := make(map[string]any)

	for {
		var doc map[string]any

		err := dec.Decode(&doc)

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		for key, value := range doc {
			if _, taken := merged[key]; taken {
				return nil, fmt.Errorf("'%s' is declared in more than one document", key)
			}
			merged[key] = value
		}
	}

	return merged, nil
}

// checkYAMLSoundness runs the document through yaml.v3 before goccy sees
// it, for two things goccy does not do.
//
// Aliases: goccy resolves one every time it is named and caps nothing, so a
// chain of anchors each referencing the one below ten times grows ten-fold
// per level. A 380-byte file reached 110MB of output and did not stop
// there. yaml.v3's decoder caps both alias expansion and nesting depth.
//
// Keys: goccy renders a non-string key as its text, so "~", "null" and
// "NULL" all arrive as the key "null" and quietly overwrite one another,
// and the emitted HCL then reads back as the string "null", which is a
// different pipeline. Rejecting those keys is the only honest answer,
// which is what the GitHub provider already does.
//
// Encoding: goccy reads a byte that cannot start a UTF-8 sequence as one
// anyway, and the encoder later writes it out as U+FFFD. A job named with
// such a byte comes back under a different name, so the pipeline that
// leaves is not the one that arrived. yaml.v3 refuses the document
// instead, which is what the GitHub provider does today.
//
// This costs a second decode, around 400ms on a 1.4MB pipeline, which is
// worth it against a file a hundred times smaller taking the machine down.
func checkYAMLSoundness(content []byte) error {
	if !utf8.Valid(content) {
		return errInvalidUTF8
	}

	dec := yamlv3.NewDecoder(bytes.NewReader(content))

	for {
		var node yamlv3.Node

		err := dec.Decode(&node)

		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			// Anything else wrong with the document is goccy's to report, so
			// its own decode below says it, with the messages and positions
			// the rest of this package is written against.
			if strings.Contains(err.Error(), "excessive aliasing") ||
				strings.Contains(err.Error(), "exceeded max depth") {
				return fmt.Errorf("%w: %s", errYAMLExhausting, err)
			}

			return nil
		}

		if err := rejectNonStringKeys(&node); err != nil {
			return err
		}

		// Decoding into a Node parses the document without resolving a
		// single alias, so the cap has not been tested yet. Decoding that
		// node into a plain value is what expands them, and what trips it.
		var discard any

		if err := node.Decode(&discard); err != nil {
			if strings.Contains(err.Error(), "excessive aliasing") {
				return fmt.Errorf("%w: %s", errYAMLExhausting, err)
			}

			return nil
		}
	}
}

// rejectNonStringKeys refuses a mapping key that is not a string. A merge
// key is left alone: it is how a pipeline shares a block between jobs, and
// the decoder folds away what it points at.
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

func classifyPipelineDocument(doc map[string]any) bool {
	// A pipeline may be nothing but includes, which is how a project pulls
	// its whole configuration in from elsewhere.
	if _, ok := doc["include"]; ok {
		return true
	}

	// Only a pipeline has a "spec" header.
	if _, ok := doc["spec"]; ok {
		return true
	}

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
		if isReservedTopLevelKey(key) {
			continue
		}

		if isJobKey(doc, key) {
			return true
		}
	}

	return false
}

// isJobKey reports whether a top-level key holds a job. Anything that is a
// mapping and is neither reserved nor hidden is one: a job needs no "script" of
// its own, since a trigger job has none by definition and an extending job
// inherits one.
func isJobKey(doc map[string]any, key string) bool {
	if strings.HasPrefix(key, ".") {
		return false
	}

	_, ok := toStringAnyMap(doc[key])

	return ok
}

func pipelineToHCL(doc map[string]any, filename string, c *comments) ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	body := f.Body()

	if rawStages, ok := doc["stages"]; ok {
		if err := writeCommentedAttribute(body, "stages", escapeGitLabVariables(rawStages), c.at("stages")); err != nil {
			return nil, err
		}
	}

	for _, key := range [...]string{"image", "before_script", "after_script", "cache", "services"} {
		value, ok := doc[key]

		if !ok {
			continue
		}

		if err := writeCommentedAttribute(body, key, escapeGitLabVariables(value), c.at(key)); err != nil {
			return nil, err
		}
	}

	if rawVariables, ok := doc["variables"]; ok {
		variables, mapOK := toStringAnyMap(rawVariables)

		if !mapOK {
			return nil, fmt.Errorf("variables must be an object")
		}

		varComments := c.child("variables")
		usedVarIDs := make(map[string]struct{}, len(variables))

		for _, name := range sortedKeys(variables) {
			if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
				body.AppendNewline()
			}

			// A variable written as a plain scalar carries its comments on the
			// key itself; one written as a mapping carries them inside, on the
			// keys of that mapping. Both end up in the same block, so both are
			// read and whichever was written is the one that is there.
			comment := varComments.at(name)
			nested := varComments.child(name)
			hclcomment.WriteLeading(body, firstNonEmpty(comment.head, nested.above()))

			varID := naming.SanitizeIdentifier(strings.ToLower(name))

			if varID == "" {
				varID = "var"
			}

			// Made unique the way a job and a template label already are.
			// "A-B" and "A_B" both sanitize to "a_b", so the file held the
			// same block label twice, and the parse direction files a block's
			// comments under its label: the first block's comment was
			// overwritten by the second and did not come back.
			varID = naming.UniqueIdentifierInSet(varID, usedVarIDs)
			usedVarIDs[varID] = struct{}{}

			vb := body.AppendNewBlock("variable", []string{varID})
			vbody := vb.Body()
			vbody.SetAttributeValue("name", cty.StringVal(name))

			raw := variables[name]

			if vm, ok := toStringAnyMap(raw); ok {
				// The four keys below are the whole of the block, so anything
				// else was written nowhere and dropped without a word.
				for _, key := range sortedKeys(vm) {
					// "name" is cinzel's, not GitLab's: the writer fills it in
					// from the key this mapping sits under. One written in the
					// body was dropped where that happened, in silence.
					if key == "name" {
						return nil, errVariableKeyReservedName
					}

					if !variableSchema.knows(key) {
						return nil, errUnknownKeyword("variable", key)
					}
				}

				if val, hasVal := vm["value"]; hasVal {
					if err := writeCommentedAttribute(vbody, "value", escapeGitLabVariables(val), nested.at("value")); err != nil {
						return nil, err
					}
				}

				for _, key := range [...]string{"description", "options", "expand"} {
					value, has := vm[key]

					if !has {
						continue
					}

					if err := writeCommentedAttribute(vbody, key, escapeGitLabVariables(value), nested.at(key)); err != nil {
						return nil, err
					}
				}
			} else {
				if err := writeCommentedAttribute(vbody, "value", escapeGitLabVariables(raw), comment.withoutHead()); err != nil {
					return nil, err
				}
			}

			hclcomment.WriteLeading(vbody, nested.below())
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

		// The three keys below are the whole of GitLab's workflow, and the HCL
		// block has nowhere to put anything else, so an unknown key can only be
		// dropped. Say so, the way the top-level loop does, rather than losing
		// it in silence.
		for _, key := range sortedKeys(workflowMap) {
			switch key {
			case "name", "auto_cancel", "rules":
			default:
				fmt.Fprintf(os.Stderr, "warning: unsupported workflow key '%s' dropped\n", key)
			}
		}

		workflowComments := c.child("workflow")
		hclcomment.WriteLeading(body, c.at("workflow").head)

		wb := body.AppendNewBlock("workflow", nil)
		wbody := wb.Body()

		if name, hasName := workflowMap["name"]; hasName {
			if err := writeCommentedAttribute(wbody, "name", escapeGitLabVariables(name), workflowComments.at("name")); err != nil {
				return nil, err
			}
		}

		if autoCancel, hasAutoCancel := workflowMap["auto_cancel"]; hasAutoCancel {
			if err := writeCommentedAttribute(wbody, "auto_cancel", escapeGitLabVariables(autoCancel), workflowComments.at("auto_cancel")); err != nil {
				return nil, err
			}
		}

		if rawRules, hasRules := workflowMap["rules"]; hasRules {
			rules, ok := rawRules.([]any)

			if rawRules == nil {
				if err := writeAttributeAny(wbody, "rules", nil); err != nil {
					return nil, err
				}

				rules = nil
			} else if !ok {
				return nil, fmt.Errorf("workflow.rules must be a list")
			}

			if rawRules != nil && len(rules) == 0 {
				if err := writeAttributeAny(wbody, "rules", []any{}); err != nil {
					return nil, err
				}
			}

			ruleComments := workflowComments.child("rules")
			hclcomment.WriteLeading(wbody, workflowComments.at("rules").head)

			for idx, item := range rules {
				ruleMap, ok := toStringAnyMap(item)

				if !ok {
					return nil, fmt.Errorf("workflow.rules entries must be objects")
				}

				if err := writeRuleBlock(wbody, ruleMap, ruleComments.item(idx)); err != nil {
					return nil, err
				}
			}

			hclcomment.WriteLeading(wbody, ruleComments.below())
		}

		hclcomment.WriteLeading(wbody, workflowComments.below())
	}

	if rawDefault, ok := doc["default"]; ok {
		defaultMap, mapOK := toStringAnyMap(rawDefault)

		if !mapOK {
			return nil, fmt.Errorf("default must be an object")
		}

		if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
			body.AppendNewline()
		}

		hclcomment.WriteLeading(body, c.at("default").head)

		db := body.AppendNewBlock("default", nil)

		if err := writeGenericMap(db.Body(), defaultMap, defaultSchema, c.child("default")); err != nil {
			return nil, err
		}
	}

	if rawSpec, ok := doc["spec"]; ok {
		specMap, mapOK := toStringAnyMap(rawSpec)

		if !mapOK && rawSpec != nil {
			return nil, fmt.Errorf("spec must be an object")
		}

		if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
			body.AppendNewline()
		}

		specComments := c.child("spec")
		hclcomment.WriteLeading(body, c.at("spec").head)

		sb := body.AppendNewBlock("spec", nil)

		for _, key := range sortedKeys(specMap) {
			// Refused here rather than written out for parse to refuse later,
			// the way a job body key is.
			if !specSchema.knows(key) {
				return nil, errUnknownKeyword("spec", key)
			}

			if err := writeCommentedAttribute(sb.Body(), key, escapeGitLabVariables(specMap[key]), specComments.at(key)); err != nil {
				return nil, err
			}
		}

		hclcomment.WriteLeading(sb.Body(), specComments.below())
	}

	if rawInclude, ok := doc["include"]; ok {
		if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
			body.AppendNewline()
		}

		hclcomment.WriteLeading(body, c.at("include").head)

		if err := writeIncludeBlocks(body, rawInclude, c.child("include")); err != nil {
			return nil, err
		}
	}

	jobNames := make([]string, 0)
	jobIDMap := make(map[string]string)
	templateIDMap := make(map[string]string)
	usedIDs := make(map[string]struct{})
	usedTemplateIDs := make(map[string]struct{})

	for _, key := range sortedKeys(doc) {
		if !strings.HasPrefix(key, ".") {
			continue
		}

		_, ok := toStringAnyMap(doc[key])

		if !ok {
			continue
		}

		// A template keyed "." has nothing left once the dot is taken off.
		// It cannot be named by extends and has no key to emit, so emitting
		// one wrote id = "", which the parse direction then refused with
		// this same error. The job loop below already refuses its own
		// unnamed case; this is the matching refusal for a template.
		if strings.TrimPrefix(key, ".") == "" {
			return nil, errBlockIDNotString
		}

		tplID := naming.SanitizeIdentifier(strings.TrimPrefix(key, "."))

		if tplID == "" {
			tplID = "template"
		}

		tplID = naming.UniqueIdentifierInSet(tplID, usedTemplateIDs)
		usedTemplateIDs[tplID] = struct{}{}
		templateIDMap[key] = tplID
	}

	for _, key := range sortedKeys(doc) {
		if isReservedTopLevelKey(key) || !isJobKey(doc, key) {
			continue
		}

		// An unnamed job cannot be referred to by needs or extends and has
		// no key to emit. Emitting one wrote id = "", which the parse
		// direction then refused with this same error, so the pipeline
		// could not come back.
		if key == "" {
			return nil, errBlockIDNotString
		}

		jobNames = append(jobNames, key)
		id := naming.SanitizeIdentifier(key)

		if id == "" {
			id = "job"
		}
		id = naming.UniqueIdentifierInSet(id, usedIDs)
		usedIDs[id] = struct{}{}
		jobIDMap[key] = id
	}
	sort.Strings(jobNames)

	for _, name := range jobNames {
		if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
			body.AppendNewline()
		}
		jobMap, _ := toStringAnyMap(doc[name])
		hclcomment.WriteLeading(body, c.at(name).head)

		jb := body.AppendNewBlock("job", []string{jobIDMap[name]})

		writeBlockKey(jb.Body(), name, jobIDMap[name])

		if err := writeJobBlock(jb.Body(), jobMap, jobIDMap, templateIDMap, c.child(name)); err != nil {
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
			tplID := templateIDMap[key]

			if tplID == "" {
				tplID = naming.SanitizeIdentifier(strings.TrimPrefix(key, "."))

				if tplID == "" {
					tplID = "template"
				}
			}
			hclcomment.WriteLeading(body, c.at(key).head)

			tb := body.AppendNewBlock("template", []string{tplID})

			writeBlockKey(tb.Body(), strings.TrimPrefix(key, "."), tplID)

			// A template body is a job body: it takes the same rule, cache,
			// artifacts and service blocks, which the generic writer would
			// write as attributes.
			if err := writeJobBlock(tb.Body(), hiddenJobMap, jobIDMap, templateIDMap, c.child(key)); err != nil {
				return nil, fmt.Errorf("error in template '%s': %w", key, err)
			}
			continue
		}

		if genericMap, ok := toStringAnyMap(doc[key]); ok {
			if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
				body.AppendNewline()
			}
			hclcomment.WriteLeading(body, c.at(key).head)

			gb := body.AppendNewBlock(key, nil)

			if err := writeGenericMap(gb.Body(), genericMap, passthroughSchema, c.child(key)); err != nil {
				return nil, err
			}
		} else {
			// The key is written as an attribute name, so a key that is not an
			// identifier produced HCL nothing can read back: "my weird key =
			// v" is three block labels and an equals sign. It went out with
			// only the passthrough warning and exit 0.
			if naming.SanitizeIdentifier(key) != key {
				return nil, errKeyNotAnIdentifier(key)
			}

			if err := writeCommentedAttribute(body, key, escapeGitLabVariables(doc[key]), c.at(key)); err != nil {
				return nil, err
			}
		}
	}

	_ = filename

	return unescape.Unicode(hclwrite.Format(f.Bytes())), nil
}

// writeBlockKey records the original YAML key when the block label differs from
// it. Labels are sanitized so the block can be referenced as job.<id> or
// template.<id>, which turns "build-app" into "build_app". Parse prefers this
// attribute when writing the block back out.
func writeBlockKey(body *hclwrite.Body, key string, id string) {
	if key == id {
		return
	}

	body.SetAttributeValue("id", cty.StringVal(key))
}

// jobRefID returns the label a "needs" entry refers to. A name the pipeline
// declares uses that job's label, so a reference follows a renamed job.
//
// A name that sanitizes to nothing is refused rather than written. Emitting it
// produced "job." with no identifier after the dot, which is not HCL: the file
// was written, the command exited 0, and parsing it back failed with "an
// attribute name is required after a dot". "extends" already refuses the same
// input, and this is the matching refusal for "needs".
func jobRefID(name string, jobIDMap map[string]string) (string, error) {
	if refID, exists := jobIDMap[name]; exists {
		return refID, nil
	}

	// A hidden key is a template, which GitLab never runs, so nothing can wait
	// on one. The name went out as a job reference to a job that is not in the
	// file, and parse refused the result with "needs unknown job".
	if strings.HasPrefix(name, ".") {
		return "", fmt.Errorf("%w: '%s'", errNeedsHiddenJob, name)
	}

	refID := naming.SanitizeIdentifier(name)

	if refID == "" {
		return "", errNeedsJobEmpty
	}

	return refID, nil
}

// writeNeedBlock writes the object form of a "needs" entry. Its "job" is
// written as a job reference so it tracks a renamed job like "depends_on"
// does; everything else is copied through.
func writeNeedBlock(body *hclwrite.Body, need map[string]any, jobIDMap map[string]string, c *comments) error {
	// A need naming neither a job nor an upstream pipeline is the same empty
	// reference in block form: it wrote "need {}", which parse then refuses.
	// This is the rule parse already applies, moved to where the file is
	// written rather than left for whoever reads it back.
	if _, hasJob := need["job"]; !hasJob {
		if _, crossPipeline := need["pipeline"]; !crossPipeline {
			return errNeedsJobEmpty
		}
	}

	hclcomment.WriteLeading(body, c.above())

	// A need carrying "project" or "pipeline" waits on a job in another
	// pipeline, which this file does not declare.
	_, crossProject := need["project"]
	_, crossPipeline := need["pipeline"]
	remote := crossProject || crossPipeline

	nb := body.AppendNewBlock("need", nil)

	for _, key := range sortedKeys(need) {
		value := need[key]

		// Refused here rather than written out for parse to refuse later, the
		// way a job body key is.
		if !needSchema.knows(key) {
			return errUnknownKeyword("need", key)
		}

		if key == "job" {
			name, ok := value.(string)

			if !ok {
				return fmt.Errorf("needs job must be a string")
			}

			// A remote job's name was written as a reference to a job here,
			// so it was sanitized like a local label and came back renamed,
			// and one that matched a local job followed that job's renames.
			// It is the other pipeline's name, so it is written as it stands.
			if remote {
				if err := writeCommentedAttribute(nb.Body(), "job", name, c.at(key)); err != nil {
					return err
				}

				continue
			}

			refID, err := jobRefID(name, jobIDMap)
			if err != nil {
				return err
			}

			writeReferenceAttribute(nb.Body(), "job", "job", refID)

			continue
		}

		if err := writeCommentedAttribute(nb.Body(), key, escapeGitLabVariables(value), c.at(key)); err != nil {
			return err
		}
	}

	hclcomment.WriteLeading(nb.Body(), c.below())

	return nil
}

// writeRuleBlock writes one entry of a "rules" list as a rule block, carrying
// whatever comments that entry held.
func writeRuleBlock(body *hclwrite.Body, rule map[string]any, c *comments) error {
	hclcomment.WriteLeading(body, c.above())

	rb := body.AppendNewBlock("rule", nil)

	for _, key := range sortedKeys(rule) {
		// Refused here rather than written out for parse to refuse later, the
		// way a job body key is.
		if !ruleSchema.knows(key) {
			return errUnknownKeyword("rule", key)
		}

		if err := writeCommentedAttribute(rb.Body(), key, escapeGitLabVariables(rule[key]), c.at(key)); err != nil {
			return err
		}
	}

	hclcomment.WriteLeading(rb.Body(), c.below())

	return nil
}

func writeJobBlock(body *hclwrite.Body, job map[string]any, jobIDMap map[string]string, templateIDMap map[string]string, c *comments) error {
	// "id" is cinzel's, not GitLab's: writeBlockKey puts a job's original name
	// there when the label had to be sanitized. A body key of that name was
	// written straight into the block, where parse read it as the name, and the
	// job came back under a name nobody wrote with both directions exiting 0.
	if _, reserved := job["id"]; reserved {
		return errJobKeyReservedID
	}

	for _, key := range sortedKeys(job) {
		value := job[key]

		// A key the schema does not declare was written out as an attribute
		// all the same, so the file went out at exit 0 and the same tool's
		// parse direction refused it with "an argument named ... is not
		// expected here". The refusal belongs here, where the input that
		// caused it is still in hand.
		if !jobSchema.knows(key) && !jobBodyAliases[key] {
			return errUnknownKeyword("job", key)
		}

		// Written here rather than inside each case: every key below becomes
		// either an attribute or one or more blocks, and the comment above it
		// belongs above whichever it becomes.
		comment := c.at(key)
		hclcomment.WriteLeading(body, comment.head)

		switch key {
		case "needs":
			needs, ok := value.([]any)

			if !ok {
				return fmt.Errorf("needs must be a list")
			}
			refs := make([]string, 0, len(needs))

			for idx, n := range needs {
				// A "needs" entry is either a job name or an object
				// carrying that name plus options. The object form
				// becomes a "need" block, since an attribute cannot
				// hold the job reference.
				if nested, isMap := toStringAnyMap(n); isMap {
					if err := writeNeedBlock(body, nested, jobIDMap, c.child("needs").item(idx)); err != nil {
						return err
					}

					continue
				}

				name, ok := n.(string)

				if !ok {
					return fmt.Errorf("needs entries must be strings or objects")
				}

				refID, err := jobRefID(name, jobIDMap)
				if err != nil {
					return err
				}

				refs = append(refs, refID)
			}

			if len(needs) == 0 {
				if err := writeAttributeAny(body, "depends_on", []any{}); err != nil {
					return err
				}

				continue
			}

			if err := writeReferenceListAttribute(body, "depends_on", "job", refs); err != nil {
				return err
			}
		case "rules":
			if value == nil {
				if err := writeAttributeAny(body, "rules", nil); err != nil {
					return err
				}

				continue
			}

			rules, ok := value.([]any)

			if !ok {
				return fmt.Errorf("rules must be a list")
			}

			if len(rules) == 0 {
				if err := writeAttributeAny(body, "rules", []any{}); err != nil {
					return err
				}

				continue
			}

			ruleComments := c.child("rules")

			for idx, item := range rules {
				ruleMap, ok := toStringAnyMap(item)

				if !ok {
					return fmt.Errorf("rules entries must be objects")
				}

				if err := writeRuleBlock(body, ruleMap, ruleComments.item(idx)); err != nil {
					return err
				}
			}

			hclcomment.WriteLeading(body, ruleComments.below())
		case "cache", "artifacts":
			schema := cacheSchema

			if key == "artifacts" {
				schema = artifactsSchema
			}

			// A job may declare several caches as a list. Each one is
			// its own block, which is how the schema already spells
			// repeated caches.
			if value == nil {
				if err := writeAttributeAny(body, key, nil); err != nil {
					return err
				}

				continue
			}

			entries, isList := value.([]any)

			if isList && len(entries) == 0 {
				if err := writeAttributeAny(body, key, []any{}); err != nil {
					return err
				}

				continue
			}

			if !isList {
				entries = []any{value}
			}

			// A cache may repeat; "artifacts" may not. Writing a block per
			// entry produced a file parse refuses with "job can include at
			// most one artifacts block", so the refusal belongs here, where
			// the input that caused it is still in hand.
			if key == "artifacts" && len(entries) > 1 {
				return fmt.Errorf("%w, got %d", errArtifactsNotAList, len(entries))
			}

			// A key given as a mapping carries its comments on that mapping;
			// given as a list, each entry carries its own. Both readings are
			// there, and only the one that was written holds anything.
			entryComments := c.child(key)

			for idx, entry := range entries {
				mapVal, ok := toStringAnyMap(entry)

				if !ok {
					return fmt.Errorf("%s must be an object", key)
				}

				nested := entryComments

				if isList {
					nested = entryComments.item(idx)
					hclcomment.WriteLeading(body, nested.above())
				}

				b := body.AppendNewBlock(key, nil)

				if err := writeGenericMap(b.Body(), mapVal, schema, nested); err != nil {
					return err
				}
			}

			if isList {
				hclcomment.WriteLeading(body, entryComments.below())
			}
		case "services":
			if err := writeServicesBlocks(body, value, c.child("services")); err != nil {
				return err
			}
		case "extends":
			refsAny, isList := value.([]any)

			if isList && len(refsAny) == 0 {
				if err := writeAttributeAny(body, "extends", []any{}); err != nil {
					return err
				}

				continue
			}

			if !isList {
				if single, ok := value.(string); ok {
					refsAny = []any{single}
				} else {
					return fmt.Errorf("extends must be a string or list")
				}
			}

			roots := make([]string, 0, len(refsAny))
			refs := make([]string, 0, len(refsAny))

			for _, item := range refsAny {
				extendsName, ok := item.(string)

				if !ok || extendsName == "" {
					return fmt.Errorf("extends entries must be non-empty strings")
				}

				if strings.HasPrefix(extendsName, ".") {
					roots = append(roots, "template")
					rest := strings.TrimPrefix(extendsName, ".")
					templateID := templateIDMap[extendsName]

					if templateID == "" {
						// A name that sanitizes to nothing was replaced with
						// the bare word "template", so "." went out as
						// template.template and bound to whichever template
						// happened to carry that label, or to none at all,
						// with both directions exiting 0. "needs" refuses the
						// same input.
						templateID = naming.SanitizeIdentifier(rest)

						if templateID == "" {
							return fmt.Errorf("%w: '%s'", errExtendsNameEmpty, extendsName)
						}
					}
					refs = append(refs, templateID)
					continue
				}

				roots = append(roots, "job")
				refID, exists := jobIDMap[extendsName]

				if !exists {
					refID = naming.SanitizeIdentifier(extendsName)

					if refID == "" {
						return fmt.Errorf("%w: '%s'", errExtendsNameEmpty, extendsName)
					}
				}
				refs = append(refs, refID)
			}

			if err := writeScopedReferenceListAttribute(body, "extends", roots, refs); err != nil {
				return err
			}
		default:
			if err := writeCommentedAttribute(body, key, escapeGitLabVariables(value), comment.withoutHead()); err != nil {
				return err
			}
		}
	}

	hclcomment.WriteLeading(body, c.below())

	return nil
}

// bodySchema says which keys of a block body the HCL schema in config.go
// declares as nested blocks. Without it the writer guesses from the value's
// shape and turns every nested map into a block, which the parser then rejects
// for a key the schema declares as an attribute.
type bodySchema struct {
	// any is set for a body declared with `hcl:",remain"`, which takes a
	// block of any name.
	any    bool
	blocks map[string]bodySchema
	attrs  map[string]struct{}
	// owner names the block this schema describes, for the error reporting a
	// key it does not declare.
	owner string
}

// knows reports whether the body takes key at all, as an attribute or a block.
// A body declared with `hcl:",remain"` takes anything.
func (s bodySchema) knows(key string) bool {
	if s.any {
		return true
	}

	if _, ok := s.attrs[key]; ok {
		return true
	}

	_, ok := s.blocks[key]

	return ok
}

// child returns the schema for a nested block named key, and whether the body
// takes one at all.
func (s bodySchema) child(key string) (bodySchema, bool) {
	if child, ok := s.blocks[key]; ok {
		return child, true
	}

	if s.any {
		return s, true
	}

	return bodySchema{}, false
}

// schemaOf reads a bodySchema off the hcl tags of a config struct.
func schemaOf(owner string, v any) bodySchema {
	schema := schemaOfType(reflect.TypeOf(v))
	schema.owner = owner

	return schema
}

func schemaOfType(t reflect.Type) bodySchema {
	schema := bodySchema{blocks: map[string]bodySchema{}, attrs: map[string]struct{}{}}

	for i := range t.NumField() {
		field := t.Field(i)
		tag, ok := field.Tag.Lookup("hcl")

		if !ok {
			continue
		}

		name, kind, _ := strings.Cut(tag, ",")

		switch kind {
		case "remain":
			schema.any = true
		case "optional", "attr":
			schema.attrs[name] = struct{}{}
		case "block":
			elem := field.Type

			for elem.Kind() == reflect.Slice || elem.Kind() == reflect.Pointer {
				elem = elem.Elem()
			}

			schema.blocks[name] = schemaOfType(elem)
		}
	}

	return schema
}

var (
	defaultSchema   = schemaOf("default", hclDefaultBlock{})
	cacheSchema     = schemaOf("cache", hclCacheBlock{})
	artifactsSchema = schemaOf("artifacts", hclArtifactsBlock{})
	serviceSchema   = schemaOf("service", hclServiceBlock{})
	includeSchema   = schemaOf("include", hclIncludeBlock{})
	jobSchema       = schemaOf("job", hclJobBlock{})
	ruleSchema      = schemaOf("rule", hclRuleBlock{})
	needSchema      = schemaOf("need", hclNeedBlock{})
	specSchema      = schemaOf("spec", hclSpecBlock{})
	variableSchema  = schemaOf("variable", hclVariableBlock{})

	// jobBodyAliases are the YAML keys a job body takes that the HCL schema
	// spells differently: "needs" becomes "depends_on" or a "need" block, and
	// each of the rest becomes a block of its own name.
	jobBodyAliases = map[string]bool{
		"needs": true, "rules": true, "cache": true,
		"artifacts": true, "services": true,
	}
	// passthroughSchema is used for a top-level key outside the schema, where
	// there is nothing to check against.
	passthroughSchema = bodySchema{any: true}
)

// singleBlockOnly maps a block key the parse side accepts at most once to the
// error to report when the YAML gives it as a longer list.
var singleBlockOnly = map[string]error{
	"artifacts": errArtifactsNotAList,
	"reports":   errReportsNotAList,
}

// allStringAnyMaps reports whether every entry of a list is an object, and the
// list is not empty.
func allStringAnyMaps(entries []any) bool {
	if len(entries) == 0 {
		return false
	}

	for _, entry := range entries {
		if _, ok := toStringAnyMap(entry); !ok {
			return false
		}
	}

	return true
}

func writeGenericMap(body *hclwrite.Body, mapping map[string]any, schema bodySchema, c *comments) error {
	for _, key := range sortedKeys(mapping) {
		value := mapping[key]

		// Refused here for the same reason as in a job body: an undeclared key
		// was written as an attribute, and the file only failed later, on the
		// way back in.
		if !schema.knows(key) {
			return errUnknownKeyword(schema.owner, key)
		}

		// Written here rather than in each branch: the key becomes either an
		// attribute or one or more blocks, and the comment above it belongs
		// above whichever it becomes.
		comment := c.at(key)
		hclcomment.WriteLeading(body, comment.head)

		if key == "services" {
			if err := writeServicesBlocks(body, value, c.child(key)); err != nil {
				return err
			}
			continue
		}

		if nested, ok := toStringAnyMap(value); ok {
			if child, isBlock := schema.child(key); isBlock {
				b := body.AppendNewBlock(key, nil)

				if err := writeGenericMap(b.Body(), nested, child, c.child(key)); err != nil {
					return err
				}
				continue
			}
		}

		// A declared block that the YAML gives as a list of objects is
		// written once per entry, e.g. the several caches a "default"
		// may declare.
		if entries, ok := value.([]any); ok {
			if _, declared := schema.blocks[key]; declared && len(entries) == 0 {
				if err := writeAttributeAny(body, key, []any{}); err != nil {
					return err
				}

				continue
			}

			if child, declared := schema.blocks[key]; declared && allStringAnyMaps(entries) {
				// A cache or a service may repeat; "artifacts" and "reports"
				// may not. Writing a block per entry produced a file parse
				// refuses with "can include at most one ... block", so the
				// refusal belongs here, where the input that caused it is
				// still in hand. The job-level writer keeps the same guard.
				if err, capped := singleBlockOnly[key]; capped && len(entries) > 1 {
					return fmt.Errorf("%w, got %d", err, len(entries))
				}

				entryComments := c.child(key)

				for idx, entry := range entries {
					nested, _ := toStringAnyMap(entry)
					itemComments := entryComments.item(idx)
					hclcomment.WriteLeading(body, itemComments.above())

					b := body.AppendNewBlock(key, nil)

					if err := writeGenericMap(b.Body(), nested, child, itemComments); err != nil {
						return err
					}
				}

				hclcomment.WriteLeading(body, entryComments.below())

				continue
			}
		}

		if err := writeCommentedAttribute(body, key, escapeGitLabVariables(value), comment.withoutHead()); err != nil {
			return err
		}
	}

	hclcomment.WriteLeading(body, c.below())

	return nil
}

func writeServicesBlocks(body *hclwrite.Body, raw any, c *comments) error {
	// An explicit null clears an inherited "services", which is not the same as
	// leaving the keyword out, so it is written back as one. "cache" already
	// does this a few cases above; "services" refused it and broke the
	// roundtrip.
	if raw == nil {
		return writeAttributeAny(body, "services", nil)
	}

	services, ok := raw.([]any)

	if !ok {
		return fmt.Errorf("services must be a list")
	}

	if len(services) == 0 {
		return writeAttributeAny(body, "services", []any{})
	}

	for idx, item := range services {
		itemComments := c.item(idx)
		hclcomment.WriteLeading(body, itemComments.above())

		sb := body.AppendNewBlock("service", nil)

		switch service := item.(type) {
		case string:
			if err := writeAttributeAny(sb.Body(), "name", escapeGitLabVariables(service)); err != nil {
				return err
			}
		case map[string]any:
			if err := writeGenericMap(sb.Body(), service, serviceSchema, itemComments); err != nil {
				return err
			}
		default:
			return fmt.Errorf("services entries must be strings or objects")
		}
	}

	hclcomment.WriteLeading(body, c.below())

	return nil
}

func writeIncludeBlocks(body *hclwrite.Body, raw any, c *comments) error {
	switch include := raw.(type) {
	case string:
		ib := body.AppendNewBlock("include", nil)

		if err := writeAttributeAny(ib.Body(), "local", escapeGitLabVariables(include)); err != nil {
			return err
		}

		return nil
	case map[string]any:
		ib := body.AppendNewBlock("include", nil)

		return writeGenericMap(ib.Body(), include, includeSchema, c)
	case []any:
		for idx, item := range include {
			itemComments := c.item(idx)
			hclcomment.WriteLeading(body, itemComments.above())

			switch v := item.(type) {
			case string:
				ib := body.AppendNewBlock("include", nil)

				if err := writeAttributeAny(ib.Body(), "local", escapeGitLabVariables(v)); err != nil {
					return err
				}
			case map[string]any:
				ib := body.AppendNewBlock("include", nil)

				if err := writeGenericMap(ib.Body(), v, includeSchema, itemComments); err != nil {
					return err
				}
			default:
				return fmt.Errorf("include entries must be strings or objects")
			}
		}

		hclcomment.WriteLeading(body, c.below())

		return nil
	default:
		return fmt.Errorf("include must be a string, object, or list")
	}
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
	case "include", "spec":
		return true
	default:
		return isGlobalDefaultKey(key)
	}
}

// isGlobalDefaultKey reports whether a top-level key is one of the five GitLab
// still reads outside a "default" block, where it means the same thing. Each is
// written straight back as a top-level attribute rather than being folded into
// a default block, since GitLab does not document which wins when a pipeline
// has both.
func isGlobalDefaultKey(key string) bool {
	switch key {
	case "image", "before_script", "after_script", "cache", "services":
		return true
	default:
		return false
	}
}
