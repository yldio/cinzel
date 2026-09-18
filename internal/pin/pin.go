// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

const (
	cacheTTL      = 24 * time.Hour
	cacheSubdir   = "cinzel/pins"
	githubAPIBase = "https://api.github.com"
	tokenEnvVar   = "GITHUB_TOKEN"
)

// tagPattern matches version strings that look like release tags: v1, v1.2,
// v1.2.3, and the prerelease forms v1.2.3-beta.1 and 1.0.0-rc1. A build suffix
// is left out because "+" is not a character validateGitHubNames lets into a
// request URL.
var tagPattern = regexp.MustCompile(`^v?\d+(\.\d+)*(-[0-9A-Za-z.-]+)?$`)

// safeNamePattern validates GitHub owner, repo, and tag names to prevent
// URL injection. Allows alphanumeric, hyphens, dots, underscores.
var safeNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Resolver resolves action version tags to commit SHAs.
type Resolver interface {
	ResolveTag(ctx context.Context, owner, repo, tag string) (string, error)
}

// GitHubResolver resolves tags via the GitHub API.
type GitHubResolver struct {
	token  string
	client *http.Client
}

// NewGitHubResolver creates a resolver that uses the GitHub API.
// If token is empty, it falls back to GITHUB_TOKEN env var.
// Unauthenticated requests are limited to 60/hr; authenticated to 5000/hr.
func NewGitHubResolver(token string) *GitHubResolver {
	if token == "" {
		token = os.Getenv(tokenEnvVar)
	}

	return &GitHubResolver{
		token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// ResolveTag resolves a tag to a commit SHA via the GitHub API.
func (r *GitHubResolver) ResolveTag(ctx context.Context, owner, repo, tag string) (string, error) {
	if err := validateGitHubNames(owner, repo, tag); err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/repos/%s/%s/git/ref/tags/%s", githubAPIBase, owner, repo, tag)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub API request failed: %w", err)
	}

	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", classifyGitHubError(resp.StatusCode, fmt.Sprintf("%s/%s@%s", owner, repo, tag), r.token == "")
	}

	var ref struct {
		Object struct {
			SHA  string `json:"sha"`
			Type string `json:"type"`
		} `json:"object"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ref); err != nil {
		return "", fmt.Errorf("failed to decode GitHub API response: %w", err)
	}

	// If the ref points to a tag object (annotated tag), dereference to the commit.
	if ref.Object.Type == "tag" {
		return r.dereferenceTag(ctx, owner, repo, ref.Object.SHA)
	}

	return ref.Object.SHA, nil
}

func (r *GitHubResolver) dereferenceTag(ctx context.Context, owner, repo, tagSHA string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/tags/%s", githubAPIBase, owner, repo, tagSHA)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub API request failed: %w", err)
	}

	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", classifyGitHubError(resp.StatusCode, "tag/"+tagSHA, r.token == "")
	}

	var tag struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return "", fmt.Errorf("failed to decode tag response: %w", err)
	}

	return tag.Object.SHA, nil
}

// LatestTag returns the latest semver tag for a repository by listing tags
// sorted by version descending.
func (r *GitHubResolver) LatestTag(ctx context.Context, owner, repo string) (string, error) {
	if !safeNamePattern.MatchString(owner) || !safeNamePattern.MatchString(repo) {
		return "", fmt.Errorf("invalid owner/repo name: %s/%s", owner, repo)
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPIBase, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub API request failed: %w", err)
	}

	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", classifyGitHubError(resp.StatusCode, fmt.Sprintf("%s/%s latest release", owner, repo), r.token == "")
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode release response: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no releases found for %s/%s", owner, repo)
	}

	return release.TagName, nil
}

// CachedResolver wraps a Resolver with a file-based cache.
type CachedResolver struct {
	inner    Resolver
	cacheDir string
}

// NewCachedResolver creates a resolver that caches results for 24 hours.
func NewCachedResolver(inner Resolver) *CachedResolver {
	cacheDir, _ := os.UserCacheDir()

	return &CachedResolver{
		inner:    inner,
		cacheDir: filepath.Join(cacheDir, cacheSubdir),
	}
}

// ResolveTag checks the cache first, then falls back to the inner resolver.
func (r *CachedResolver) ResolveTag(ctx context.Context, owner, repo, tag string) (string, error) {
	key := cacheKey(owner, repo, tag)
	cachePath := filepath.Join(r.cacheDir, key)

	if sha, ok := r.readCache(cachePath); ok {
		return sha, nil
	}

	sha, err := r.inner.ResolveTag(ctx, owner, repo, tag)
	if err != nil {
		return "", err
	}

	r.writeCache(cachePath, sha)

	return sha, nil
}

func cacheKey(owner, repo, tag string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s/%s@%s", owner, repo, tag)))

	return fmt.Sprintf("%x", h[:16])
}

func (r *CachedResolver) readCache(path string) (string, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return "", false
	}

	if time.Since(info.ModTime()) > cacheTTL {
		return "", false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	sha := strings.TrimSpace(string(data))
	if sha == "" {
		return "", false
	}

	return sha, true
}

func (r *CachedResolver) writeCache(path, sha string) {
	_ = os.MkdirAll(r.cacheDir, 0700)
	_ = os.WriteFile(path, []byte(sha), 0600)
}

// ActionRef represents an action reference found in an HCL file.
type ActionRef struct {
	Action  string // e.g., "actions/checkout"
	Version string // e.g., "v4" or "abc123..."
	IsTag   bool   // true if Version looks like a release tag

	// Where the version assignment sits in the file, the trailing comment
	// included. Rewrites go here rather than to the first text that looks
	// the same, which is not always this action's line.
	start int
	end   int
}

// versionEdit is one rewritten version assignment, held back until every
// action has been resolved so the offsets it carries stay valid.
type versionEdit struct {
	start int
	end   int
	text  string
}

// applyVersionEdits splices edits into content back to front, so an earlier
// rewrite cannot shift the offsets of a later one.
func applyVersionEdits(content string, edits []versionEdit) string {
	for i := len(edits) - 1; i >= 0; i-- {
		e := edits[i]
		content = content[:e.start] + e.text + content[e.end:]
	}

	return content
}

// isCommitSHA reports whether s is a full 40-character hex commit SHA, which
// is the only thing a resolve is allowed to return.
func isCommitSHA(s string) bool {
	const shaLength = 40

	if len(s) != shaLength {
		return false
	}

	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}

// shortSHA abbreviates a SHA for a progress line, returning it whole when it is
// already shorter than the abbreviation. The API's own response used to decide
// this: a missing "sha" field decodes to "" with no error, and slicing that to
// twelve panicked in the middle of a run that had already rewritten earlier
// lines.
func shortSHA(sha string) string {
	const abbrev = 12

	if len(sha) <= abbrev {
		return sha
	}

	return sha[:abbrev]
}

// versionLine renders the assignment written in place of the old one.
func versionLine(sha, comment string) string {
	return fmt.Sprintf(`version = %q # %s`, sha, comment)
}

// trailingCommentEnd returns the offset just past a comment sitting at the end
// of the version line, or from if there is none. The replacement carries its
// own comment, so leaving the old one stacked a stale tag beside the new one:
// a line reading "# v5 # v4" names a version the SHA is not.
//
// A "/* */" comment is taken whole, over however many lines it runs. Leaving
// one standing put the new "# tag" comment in front of its "/*", which
// commented out the opening while the closing "*/" stayed on a line of its
// own, and the file no longer parsed. The pin reported success and the
// breakage surfaced on the next parse.
func trailingCommentEnd(content string, from int) int {
	i := from
	for i < len(content) && (content[i] == ' ' || content[i] == '\t') {
		i++
	}

	if i >= len(content) {
		return from
	}

	if strings.HasPrefix(content[i:], "/*") {
		end := strings.Index(content[i+2:], "*/")

		// Unterminated: there is no comment to take, and swallowing the rest
		// of the file would delete every block below this one.
		if end < 0 {
			return from
		}

		return i + 2 + end + 2
	}

	if content[i] != '#' && !strings.HasPrefix(content[i:], "//") {
		return from
	}

	for i < len(content) && content[i] != '\n' {
		i++
	}

	return i
}

// splitAction reads the repository an action lives in out of its reference.
// An action may sit in a subdirectory of its repository, written
// "github/codeql-action/init", and the tag belongs to the repository, so
// everything past the second segment names a path inside it. Splitting into
// two asked the API for a repository called "codeql-action/init", which no
// name pattern lets through, and the action was reported as unpinnable.
//
// A reference starting with "." is a path into the repository being built.
// GitHub takes that version from the checkout, so there is no release to
// resolve and false is returned.
func splitAction(action string) (owner, repo string, ok bool) {
	if strings.HasPrefix(action, ".") {
		return "", "", false
	}

	parts := strings.Split(action, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// isTag returns true if the version looks like a tag rather than a SHA.
func isTag(version string) bool {
	return tagPattern.MatchString(version)
}

// PinResult holds the result of pinning a single action.
type PinResult struct {
	Action     string
	Tag        string
	SHA        string
	Error      error
	WasAlready bool // true if version was already a full commit SHA
}

// PinFile reads an HCL file, resolves all tag-based action versions to SHAs,
// and writes the updated file. When dryRun is true, resolutions are reported
// but the file is not modified.
func PinFile(ctx context.Context, path string, resolver Resolver, w io.Writer, dryRun bool) ([]PinResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	refs, err := findActionRefs(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse action refs in %s: %w", path, err)
	}

	if len(refs) == 0 {
		return nil, nil
	}

	var results []PinResult

	var edits []versionEdit

	for _, ref := range refs {
		// Only a full commit SHA is already pinned. "Not a tag" used to be
		// treated as one, so a branch, a short SHA and any tag the pattern
		// did not recognise were all counted in the "already pinned" total
		// and left exactly as they were.
		if isCommitSHA(ref.Version) {
			results = append(results, PinResult{
				Action:     ref.Action,
				SHA:        ref.Version,
				WasAlready: true,
			})

			continue
		}

		if !ref.IsTag {
			results = append(results, PinResult{
				Action: ref.Action,
				Tag:    ref.Version,
				Error:  errNotPinnable(ref.Version),
			})

			continue
		}

		owner, repo, ok := splitAction(ref.Action)
		if !ok {
			results = append(results, PinResult{
				Action: ref.Action,
				Tag:    ref.Version,
				Error:  errNotRemoteAction(ref.Action),
			})

			continue
		}

		sha, err := resolver.ResolveTag(ctx, owner, repo, ref.Version)

		// A response with no "sha" decodes to "" and no error. Writing that
		// out gave the file a version = "" and reported the action pinned, so
		// refuse it here and take the same path as a failed request.
		if err == nil && !isCommitSHA(sha) {
			err = errShortSHA(sha)
		}

		if err != nil {
			_, _ = fmt.Fprintf(w, "warning: could not pin %s@%s: %v\n", ref.Action, ref.Version, err)

			results = append(results, PinResult{
				Action: ref.Action,
				Tag:    ref.Version,
				Error:  err,
			})

			continue
		}

		// The rewrite goes at this action's own offsets. Searching the file
		// for text matching "version = <tag>" found the first line that read
		// that way, which is this action's only while every earlier one was
		// also rewritten. One failed resolve left an earlier line matchable,
		// and it took this action's SHA: the failed action came out pinned to
		// another action's commit, and this one stayed on a moving tag.
		edits = append(edits, versionEdit{
			start: ref.start,
			end:   ref.end,
			text:  versionLine(sha, ref.Version),
		})

		_, _ = fmt.Fprintf(w, "pinned %s@%s → %s\n", ref.Action, ref.Version, shortSHA(sha))

		results = append(results, PinResult{
			Action: ref.Action,
			Tag:    ref.Version,
			SHA:    sha,
		})
	}

	updated := applyVersionEdits(string(content), edits)

	if !dryRun && updated != string(content) {
		if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
			return results, fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	return results, nil
}

// PinDirectory pins all HCL files in a directory.
func PinDirectory(ctx context.Context, dir string, resolver Resolver, w io.Writer, dryRun bool) ([]PinResult, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var allResults []PinResult

	found := false

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".hcl") {
			continue
		}

		found = true
		path := filepath.Join(dir, entry.Name())

		results, err := PinFile(ctx, path, resolver, w, dryRun)
		if err != nil {
			_, _ = fmt.Fprintf(w, "warning: %s: %v\n", entry.Name(), err)

			continue
		}

		allResults = append(allResults, results...)
	}

	if !found {
		return nil, errNoHCLFiles
	}

	return allResults, nil
}

// findActionRefs extracts action references from the uses blocks of an HCL
// document.
//
// The document is parsed rather than scanned. Matching the text of every
// "action =" and "version =" and pairing them by position counted anything
// that read that way, a comment included, so one line reading
//
//	// TODO: was on version = "v3"
//
// left one more version than actions and the file was refused with a
// mismatched count. Nothing in it was pinned, and because PinDirectory
// reports a failed file as a warning, the run still exited 0.
//
// The parser also gives the byte range of each attribute, which is what the
// rewrite needs to put a SHA on the right line.
func findActionRefs(content string) ([]ActionRef, error) {
	file, diags := hclsyntax.ParseConfig([]byte(content), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %w", diags)
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, errNotHCLSyntax
	}

	var refs []ActionRef

	collectUsesRefs(body, content, &refs)

	sort.Slice(refs, func(i, j int) bool { return refs[i].start < refs[j].start })

	return refs, nil
}

// collectUsesRefs walks every block looking for "uses", at whatever depth it
// sits. A uses block missing either attribute is skipped: there is no action
// to pin without both halves.
func collectUsesRefs(body *hclsyntax.Body, content string, refs *[]ActionRef) {
	for _, block := range body.Blocks {
		if block.Body == nil {
			continue
		}

		if block.Type == "uses" {
			if ref, ok := usesRef(block.Body, content); ok {
				*refs = append(*refs, ref)
			}
		}

		collectUsesRefs(block.Body, content, refs)
	}
}

// usesRef reads the action and version out of one uses block.
func usesRef(body *hclsyntax.Body, content string) (ActionRef, bool) {
	actionAttr, hasAction := body.Attributes["action"]
	versionAttr, hasVersion := body.Attributes["version"]

	if !hasAction || !hasVersion {
		return ActionRef{}, false
	}

	action, ok := literalString(actionAttr.Expr)
	if !ok {
		return ActionRef{}, false
	}

	version, ok := literalString(versionAttr.Expr)
	if !ok {
		return ActionRef{}, false
	}

	rng := versionAttr.SrcRange

	return ActionRef{
		Action:  action,
		Version: version,
		IsTag:   isTag(version),
		start:   rng.Start.Byte,
		end:     trailingCommentEnd(content, rng.End.Byte),
	}, true
}

// literalString reads a plain quoted string. An action built from a variable
// or an interpolation is left alone, since its text is not known here and
// rewriting it would destroy the expression.
func literalString(expr hclsyntax.Expression) (string, bool) {
	value, diags := expr.Value(nil)
	if diags.HasErrors() || value.IsNull() || !value.IsKnown() {
		return "", false
	}

	if value.Type() != cty.String {
		return "", false
	}

	return value.AsString(), true
}

// drainAndClose reads the remaining body to enable HTTP connection reuse,
// then closes it.
func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	body.Close()
}
