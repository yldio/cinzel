// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/yldio/cinzel/internal/fsutil"
	"github.com/yldio/cinzel/internal/unescape"
	"github.com/yldio/cinzel/provider"
)

const (
	defaultParseOutputDirectory   = ".github/workflows"
	defaultUnparseOutputDirectory = "./cinzel"

	providerName = "github"
	providerDesc = "GitHub Actions https://github.com/features/actions"
	parseDesc    = "Convert HCL definitions to GitHub Actions YAML"
	unparseDesc  = "Convert GitHub Actions YAML to HCL definitions"
)

// GitHub is the provider implementation for GitHub Actions workflows.
type GitHub struct{}

// New returns a new GitHub provider instance.
func New() *GitHub { return &GitHub{} }

// GetProviderName returns the provider identifier.
func (p *GitHub) GetProviderName() string { return providerName }

// GetDescription returns a human-readable description of the provider.
func (p *GitHub) GetDescription() string { return providerDesc }

// GetParseDescription returns a description of the parse (HCL to YAML) operation.
func (p *GitHub) GetParseDescription() string { return parseDesc }

// GetUnparseDescription returns a description of the unparse (YAML to HCL) operation.
func (p *GitHub) GetUnparseDescription() string { return unparseDesc }

// Parse converts HCL workflow definitions into GitHub Actions YAML files.
func (p *GitHub) Parse(opts provider.ProviderOps) error {
	plan, err := p.planParse(opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		plan.Print()

		return nil
	}

	for _, out := range plan.Outputs {
		if err := fsutil.WriteFile(out.Path, out.Content); err != nil {
			return err
		}
	}

	for _, dir := range plan.PruneTargets() {
		if err := fsutil.PruneStaleGeneratedYAML(dir, plan.CurrentOutputs, providerName, plan.ReadSources); err != nil {
			return err
		}
	}

	return nil
}

// planParse works out every file a parse would write, without writing one.
//
// Parse is this followed by a write and a prune, and Diff is this followed by
// a read and a comparison, so the two cannot disagree about what a run
// produces. The dry run prints it.
//
// The prune's inputs are carried rather than its answer. Whether a file is
// stale depends on what is on disk, and under Parse that means after the
// writes have happened: sameAsCurrentOutput compares by file identity, and a
// path written moments ago is what makes a case-only rename come out right.
// Asking here would be asking too early.
func (p *GitHub) planParse(opts provider.ProviderOps) (provider.ParsePlan, error) {
	inputPath, err := resolveInputPath(opts)
	if err != nil {
		return provider.ParsePlan{}, err
	}

	body, sources, err := fsutil.ParseHCLInput(inputPath, opts.Recursive)
	if err != nil {
		return provider.ParsePlan{}, err
	}

	workflows, stepMap, actions, err := parseHCLToWorkflows(body, sources)
	if err != nil {
		return provider.ParsePlan{}, err
	}

	// What a run is allowed to delete. A generated file naming a source
	// outside this set was written from an HCL file the run never looked at,
	// so it is not this run's to call stale.
	readSources := fsutil.ReadSourceSet(sources)

	outputDir := resolveParseOutputDirectory(opts)

	plan := provider.ParsePlan{
		OutputDirectory: outputDir,
		ReadSources:     readSources,
	}

	if len(workflows) == 0 && len(actions) == 0 {
		// Nothing was declared. Writing here would leave a "{}" document
		// behind and, worse, make it the only current output, so the prune
		// would delete every other generated file in the directory.
		if len(stepMap) == 0 {
			return provider.ParsePlan{}, errNoDefinitions
		}

		outputBytes, err := marshalStepsYAML(stepMap)
		if err != nil {
			return provider.ParsePlan{}, err
		}

		// No source: steps are collected from every file read, so there is no
		// single block, and no single file, to name.
		outputBytes = fsutil.PrependGeneratedMarker(outputBytes, providerName, "")

		outputPath := filepath.Join(outputDir, resolveParseFilename(opts))

		plan.Outputs = append(plan.Outputs, provider.PlannedOutput{Path: outputPath, Content: outputBytes})
		plan.CurrentOutputs = map[string]struct{}{filepath.Clean(outputPath): {}}

		return plan, nil
	}

	// Every file this run writes has to be recorded, actions included: the
	// prune walks the whole tree, and anything it does not find here is read
	// as stale and removed.
	plan.CurrentOutputs = make(map[string]struct{}, len(workflows)+len(actions))

	// The directories actions land in, deduplicated: several actions sharing
	// a folder is one directory to walk, not one per action.
	pruneDirs := make(map[string]struct{})

	// Checked before anything is written, so a collision leaves the directory
	// as it was rather than partway through the run.
	if err := checkOutputPaths(workflows, actions, outputDir, workflowExt(opts)); err != nil {
		return provider.ParsePlan{}, err
	}

	for _, workflowFile := range workflows {
		outputBytes, err := marshalWorkflowYAML(workflowFile.Content, workflowFile.JobOrder, workflowFile.FootComment)
		if err != nil {
			return provider.ParsePlan{}, err
		}

		outputBytes = fsutil.PrependGeneratedMarker(outputBytes, providerName, workflowFile.Source)

		outputPath := filepath.Join(outputDir, workflowFile.Filename+workflowExt(opts))

		plan.Outputs = append(plan.Outputs, provider.PlannedOutput{Path: outputPath, Content: outputBytes})
		plan.CurrentOutputs[filepath.Clean(outputPath)] = struct{}{}
	}

	for _, actionFile := range actions {
		outputBytes, err := marshalWorkflowYAML(actionFile.Content, nil, actionFile.FootComment)
		if err != nil {
			return provider.ParsePlan{}, err
		}

		// An action is marked like a workflow. Without it the prune has no way
		// to tell an action cinzel wrote from one written by hand, so it can
		// own neither, and renaming an action left the old directory sitting
		// there with a live action.yml in it for good.
		outputBytes = fsutil.PrependGeneratedMarker(outputBytes, providerName, actionFile.Source)

		// Always action.yml, whatever --yml says: that is the name GitHub
		// looks for, and it is not the workflow extension to choose.
		//
		// An action carrying a path is placed against the root rather than
		// the output directory. The path was recorded relative to the root
		// on the way out — "actions-a" means the repository's actions-a —
		// and joining it onto a parse output directory of .github/workflows
		// put the action at .github/workflows/actions-a/setup/action.yml,
		// which converts cleanly, prunes correctly, and is never read by
		// GitHub. An action without a path keeps going where it always did.
		outputPath := filepath.Join(actionOutputBase(opts, outputDir, actionFile.Filename), actionFile.Filename, "action.yml")

		plan.Outputs = append(plan.Outputs, provider.PlannedOutput{Path: outputPath, Content: outputBytes})
		plan.CurrentOutputs[filepath.Clean(outputPath)] = struct{}{}

		// The folder the action's own directory sits in, which is where a
		// rename leaves the one it replaced. Only recorded when the action
		// carries a path: without one it is already under the output
		// directory the prune walks anyway.
		if parent := filepath.Dir(filepath.Dir(outputPath)); !sameDirectory(parent, outputDir) {
			pruneDirs[filepath.Clean(parent)] = struct{}{}
		}
	}

	for dir := range pruneDirs {
		plan.PruneDirectories = append(plan.PruneDirectories, dir)
	}

	// Sorted so a run walks them the same way twice, which keeps what a
	// failure reports from depending on map order.
	sort.Strings(plan.PruneDirectories)

	return plan, nil
}

// sameDirectory reports whether two paths name one directory, so the prune is
// not asked to walk the output directory twice.
func sameDirectory(a string, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

// Unparse converts GitHub Actions YAML files into HCL definitions.
func (p *GitHub) Unparse(opts provider.ProviderOps) error {
	for _, dir := range opts.InputDirectories() {
		if namesTheWorkflowsDirectory(dir) {
			warnf("--directory %s names the directory unparse reads anyway", dir)
		}
	}

	inputPaths, err := unparseInputPaths(opts)
	if err != nil {
		return err
	}

	files := make([]string, 0, len(inputPaths))

	for _, inputPath := range inputPaths {
		found, err := fsutil.ListFilesWithExtensions(inputPath, opts.Recursive, ".yaml", ".yml")
		if err != nil {
			return err
		}

		files = append(files, found...)
	}

	if len(files) == 0 {
		return errNoYAMLFiles
	}

	outputDir := resolveUnparseOutputDirectory(opts)

	// The output name comes from the input's basename, so a recursive run over
	// two directories each holding a "ci.yaml" aimed both at one path.
	takenNames := make(map[string]struct{}, len(files))

	// Step block labels are the address space parse reads, and parse merges
	// every file in a directory into one body. Held per run rather than per
	// file, because two workflows each holding a "mise setup" step used to
	// produce two "step \"mise_setup\"" blocks, and cinzel then refused to read
	// its own output back.
	//
	// Conversion claims an ID as it writes the block, so a file that fails
	// halfway leaves its IDs taken and a later file's "checkout" comes out as
	// "checkout_2". Nothing reads those bytes: a failure anywhere means the
	// run writes nothing at all. Clearing the set per file is not the answer
	// either, since sharing it is the point.
	usedStepIDs := map[string]struct{}{}

	// Job block labels share the same address space for the same reason.
	usedJobIDs := map[string]struct{}{}

	// Every file is converted before any of them is written. Returning at the
	// first failure left the files converted before it on disk and never read
	// the ones after, so which half of a directory survived was decided by
	// where the failing file sorted. A half-converted directory reads as a
	// finished one, and the usual next step is deleting the YAML it came from.
	converted := make([]unparsedFile, 0, len(files))

	// Collected rather than returned at the first, so a directory of broken
	// workflows is read through in one run instead of one key at a time.
	failures := make([]error, 0)

	for _, file := range files {
		yamlBytes, err := os.ReadFile(file)
		if err != nil {
			// Named, like the conversion failure below. A directory run that
			// returned this bare said a file could not be read without saying
			// which one.
			failures = append(failures, fmt.Errorf("error in file '%s': %w", file, err))

			continue
		}

		baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		hclBytes, name, kind, err := unparseYAMLFile(yamlBytes, file, baseName, actionNameFor(file, opts.ResolveRoot()), usedStepIDs, usedJobIDs)
		if err != nil {
			failures = append(failures, fmt.Errorf("error in file '%s': %w", file, err))

			continue
		}

		if hclBytes == nil {
			// Said nothing before. A file skipped in the middle of a directory
			// run left no trace at all, so pointing unparse at a ".github"
			// holding one workflow and four other YAML files reported the same
			// success as a run that converted every one of them.
			//
			// Skipping is not failing: cinzel never had this file, so passing
			// over it loses nothing and the run goes on.
			warnf("skipping '%s': not a workflow, an action or a set of steps", file)

			continue
		}

		converted = append(converted, unparsedFile{
			name:     name,
			source:   file,
			location: locationOf(file, opts.ResolveRoot()),
			kind:     kind,
			hcl:      hclBytes,
		})
	}

	// Before the no-definitions check below, so a run where every file failed
	// reports the failures rather than claiming it found nothing.
	if len(failures) > 0 {
		return errors.Join(failures...)
	}

	// Files were read but none of them held anything this tool has a
	// definition for, so nothing was written. Returning nil there said the
	// input was converted when it was not: pointing at the wrong directory, or
	// at one holding only issue templates and a dependabot config, both land
	// exactly here. A dry run counts: it found the document and only skipped
	// the write it was told to skip.
	if len(converted) == 0 {
		return errNoDefinitions
	}

	if err := checkUnparseNameClash(converted); err != nil {
		return err
	}

	for _, out := range converted {
		outputPath := filepath.Join(outputDir, mirroredName(opts, out, takenNames)+".hcl")

		if opts.DryRun {
			fmt.Printf("# file: %s\n", outputPath)
			fmt.Println(string(out.hcl))

			continue
		}

		if err := fsutil.WriteFile(outputPath, out.hcl); err != nil {
			return err
		}
	}

	return nil
}

// documentKind is which of the three things a converted file turned out to
// be, matching the detection chain in unparseYAMLFile: "on" and "jobs" make
// a workflow, "name" and "runs" an action, and neither a file of steps.
//
// Jobs are not among them. A job is a block inside a workflow document, and
// GitHub has no file that holds bare jobs, so unparse never produces one.
//
// Two kinds can share a name without colliding, because they are written to
// different places: "ci.yaml" and "ci/action.yml".
type documentKind int

const (
	kindWorkflow documentKind = iota
	kindAction
	kindSteps
)

// String names the kind the way an error reporting a clash should.
//
// Every kind is listed rather than defaulting to one: a kind added later and
// forgotten here would be reported as whichever the default was, and an
// error naming the wrong type is worse than one naming none.
func (k documentKind) String() string {
	switch k {
	case kindWorkflow:
		return "workflow"
	case kindAction:
		return "action"
	case kindSteps:
		return "set of steps"
	}

	return "document"
}

// unparsedFile is one converted document waiting to be written: the output
// basename it was given, the file it was read from, and the HCL it became.
type unparsedFile struct {
	name   string
	source string

	// location is the folders the file was read from, relative to the run's
	// root, and empty for one that records none. Only --mirror uses it: it
	// is where the HCL file goes, not where the conversion goes, which the
	// file's own attributes carry.
	location string

	// kind is what the document turned out to be. A workflow and an action
	// of the same name are two files in two places, so the clash check has
	// to know which is which.
	kind documentKind

	hcl []byte
}

// unparseYAMLFile converts a YAML file to HCL bytes, detecting whether
// the document is a workflow, action, or step-only file. Returns nil if
// the document is empty or unrecognized.
//
// source is the file as it was named on the command line, carried so a comment
// the conversion cannot write can be reported where its author will find it.
func unparseYAMLFile(yamlBytes []byte, source, baseName, actionName string, usedStepIDs, usedJobIDs map[string]struct{}) ([]byte, string, documentKind, error) {
	doc, order, comments, err := parseYAMLDocument(yamlBytes)
	if err != nil {
		return nil, "", kindWorkflow, err
	}

	// A file holding no document at all is skipped rather than rejected, the
	// same as one holding an empty document. Otherwise a single stray empty
	// file aborts a whole directory run before the real files are reached.
	if doc == nil {
		return nil, "", kindWorkflow, nil
	}

	workflowDoc, err := classifyWorkflowDocument(doc)
	if err != nil {
		return nil, "", kindWorkflow, err
	}

	if workflowDoc != nil {
		out, err := workflowToHCL(*workflowDoc, baseName, order, comments, commentSite{file: source}, usedStepIDs, usedJobIDs)

		return out, baseName, kindWorkflow, err
	}

	if actionDoc := classifyActionDocument(doc); actionDoc != nil {
		out, err := actionToHCL(actionDoc, actionName, comments, commentSite{file: source}, usedStepIDs)

		// The HCL file is named after the action, not after where the action
		// goes. The folders above it are written inside the file as "path",
		// so naming the file for the whole location would put them in two
		// places and spell "actions-a/setup.hcl" as one flat name.
		return out, path.Base(filepath.ToSlash(actionName)), kindAction, err
	}

	steps, err := parseStepsFromYAML(yamlBytes)
	if err != nil {
		return nil, "", kindWorkflow, err
	}

	if len(steps) == 0 {
		return nil, "", kindWorkflow, nil
	}

	f := hclwrite.NewEmptyFile()
	body := f.Body()

	for _, s := range steps {
		if err := s.Decode(body, "step"); err != nil {
			return nil, "", kindWorkflow, err
		}
	}

	return unescape.Unicode(hclwrite.Format(f.Bytes())), baseName, kindSteps, nil
}

// actionNameFor picks where an action is known to live, relative to the
// working directory: "actions-a/setup" for "actions-a/setup/action.yml".
//
// An action lives at <name>/action.yml, so the directory holding the file is
// its identity and the basename is the same word for every action there is.
// Naming the output after the basename dropped that identity: the action came
// back called "action", and a second one in the same run wrote over the first.
//
// The folders above it are kept too, because an action can sit anywhere the
// repository puts it and that is the only record of where. Taking the last
// directory alone sent every action to one place on the way back, and two
// named "setup" in different folders could not both convert.
//
// Only the fixed filenames GitHub reads are treated this way. A document that
// happens to be an action but sits under some other name is left with that
// name, which is the only thing left to go on.
func actionNameFor(file string, root string) string {
	base := strings.ToLower(filepath.Base(file))

	if base != "action.yml" && base != "action.yaml" {
		return strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	}

	dir := filepath.Dir(file)

	// A file directly under "." or the filesystem root has no directory name
	// to take. Nothing is better than the basename there.
	if dir == "." || dir == string(os.PathSeparator) || dir == "" {
		return strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	}

	if located := relativeToRoot(dir, root); located != "" {
		return located
	}

	// Outside the root, so there is nothing portable to record. The action's
	// own folder is the identity it has always had.
	located := filepath.Base(dir)
	if located == "." || located == string(os.PathSeparator) || located == "" {
		return strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	}

	return located
}

// relativeToRoot spells dir the way the HCL will carry it: relative to the
// root the run was given, with forward slashes.
//
// Relative to the root rather than to the working directory, because what
// goes into the HCL is committed and read again on another machine. A run
// given an absolute --directory would otherwise write that machine's own
// path into the file, and parse would refuse it on the way back as a write
// outside the output directory — which it is.
//
// A directory the root does not contain has no relative spelling that means
// anything to a later reader, so nothing is recorded and the action keeps the
// behaviour it had before it could say where it lived.
func relativeToRoot(dir string, root string) string {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return ""
	}

	rel, err := filepath.Rel(absRoot, absDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ""
	}

	if rel == "." {
		return ""
	}

	return filepath.ToSlash(rel)
}

// warnf writes one warning to stderr, with the control characters in it
// escaped.
//
// A warning quotes a path read off the command line, and a path is free to
// carry an ANSI escape sequence. Written to a terminal as it stands, the
// sequence is acted on rather than shown: a crafted name erases the warning
// that names it and leaves a line of its own in its place, so a run that
// skipped a file reads as a run that skipped nothing. The errors the tool ends
// on are escaped for the same reason.
func warnf(format string, args ...any) {
	fmt.Fprintln(os.Stderr, cinzelerror.SafeForTerminal("warning: "+fmt.Sprintf(format, args...)))
}

// Diff reports what a parse would change in the output directory, writing
// nothing.
func (p *GitHub) Diff(opts provider.ProviderOps) ([]provider.DiffResult, error) {
	plan, err := p.planParse(opts)
	if err != nil {
		return nil, err
	}

	return plan.Diff(providerName)
}

// checkUnparseNameClash refuses a run where two documents of one kind would
// be written under a single name.
//
// The HCL files themselves do not collide: UniqueOutputName gives the second
// one "ci_2.hcl", and the block labels inside it are suffixed the same way.
// What is not suffixed is the name the document is known by on GitHub, so
// both come out claiming one workflow and parse refuses to read them back:
//
//	two definitions write to the same file: 'ci' and 'ci' both write to 'ci'
//
// That error arrives one command later than the mistake and names neither
// YAML file. Unparse knows here, and says which two files.
//
// Kept per kind. A workflow and an action sharing a name are written to
// "ci.yaml" and "ci/action.yml", two files in two places, and refusing that
// pair turned an ordinary repository — a check workflow beside the action it
// calls — into one cinzel would not convert.
func checkUnparseNameClash(converted []unparsedFile) error {
	type claim struct {
		kind documentKind
		name string
	}

	seen := make(map[claim]string, len(converted))

	for _, out := range converted {
		// Folded, the way UniqueOutputName folds: two names differing only in
		// case are one file on macOS and Windows.
		key := claim{kind: out.kind, name: strings.ToLower(out.name)}

		if other, clash := seen[key]; clash {
			return fmt.Errorf(
				"%w: '%s' and '%s' are both the %s '%s', so one of them has to be renamed",
				errUnparseNameClash, other, out.source, out.kind, out.name,
			)
		}

		seen[key] = out.source
	}

	return nil
}

// actionOutputBase is the directory an action's own path is placed against.
//
// The output directory when the action names no path, so nothing written
// before the attribute existed moves. The run's root when it does, because
// that is what the path was recorded against and the only base that puts the
// action back where it was read from.
//
// An explicit --output-directory wins either way: a caller naming a
// destination is asking for everything to go there, which is what a test or
// a scratch conversion wants.
func actionOutputBase(opts provider.ProviderOps, outputDir string, filename string) string {
	if opts.OutputDirectory != "" {
		return outputDir
	}

	if filepath.Dir(filename) == "." {
		return outputDir
	}

	return opts.ResolveRoot()
}

// mirroredName is where one converted file goes under the output directory.
//
// Flat by default: every file in one folder, each saying where it belongs
// through its own attributes. With --mirror the repository's own layout is
// reproduced, which reads better in a tree and costs --recursive on the way
// back.
//
// Only a file that knows where it came from can be mirrored. A workflow has
// one home and says nothing about it, so it stays where it was.
func mirroredName(opts provider.ProviderOps, out unparsedFile, taken map[string]struct{}) string {
	// The location check is belt and braces: two files with none would both
	// be written flat, and checkUnparseNameClash refuses that pair before
	// this is reached. It stays because the alternative is Join("", name)
	// quietly skipping the uniqueness pass if that ever stops being true.
	if !opts.Mirror || out.location == "" {
		return fsutil.UniqueOutputName(taken, out.name)
	}

	// Already unique: two files cannot share one path in the tree they were
	// read from, so the name needs no suffix and gets none.
	return filepath.Join(filepath.FromSlash(out.location), out.name)
}

// locationOf is the folders a converted file was read from, relative to the
// run's root.
//
// An action is one directory deeper than the folder holding it — the file is
// at "actions/setup/action.yml" and belongs under "actions" — so the action's
// own directory is dropped. Anything else is taken as it is.
func locationOf(file string, root string) string {
	dir := filepath.Dir(file)

	base := strings.ToLower(filepath.Base(file))
	if base == "action.yml" || base == "action.yaml" {
		dir = filepath.Dir(dir)
	}

	return relativeToRoot(dir, root)
}
