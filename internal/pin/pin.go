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
	"strings"
	"time"
)

const (
	cacheTTL      = 24 * time.Hour
	cacheSubdir   = "cinzel/pins"
	githubAPIBase = "https://api.github.com"
	tokenEnvVar   = "GITHUB_TOKEN"
)

// tagPattern matches version strings that look like tags (v1, v1.2, v1.2.3)
// as opposed to SHAs (40+ hex chars).
var tagPattern = regexp.MustCompile(`^v?\d+(\.\d+)*$`)

// safeNamePattern validates GitHub owner, repo, and tag names to prevent
// URL injection. Allows alphanumeric, hyphens, dots, underscores.
var safeNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

var actionPattern = regexp.MustCompile(`action\s*=\s*"([^"]+)"`)
var versionPattern = regexp.MustCompile(`version\s*=\s*"([^"]+)"`)

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
	IsTag   bool   // true if Version looks like a tag, not a SHA

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

// versionLine renders the assignment written in place of the old one.
func versionLine(sha, comment string) string {
	return fmt.Sprintf(`version = %q # %s`, sha, comment)
}

// trailingCommentEnd returns the offset just past a comment sitting at the end
// of the version line, or from if there is none. The replacement carries its
// own comment, so leaving the old one stacked a stale tag beside the new one:
// a line reading "# v5 # v4" names a version the SHA is not.
func trailingCommentEnd(content string, from int) int {
	i := from
	for i < len(content) && (content[i] == ' ' || content[i] == '\t') {
		i++
	}

	if i >= len(content) || (content[i] != '#' && !strings.HasPrefix(content[i:], "//")) {
		return from
	}

	for i < len(content) && content[i] != '\n' {
		i++
	}

	return i
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
	WasAlready bool // true if version was already a SHA
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
		if !ref.IsTag {
			results = append(results, PinResult{
				Action:     ref.Action,
				SHA:        ref.Version,
				WasAlready: true,
			})

			continue
		}

		parts := strings.SplitN(ref.Action, "/", 2)
		if len(parts) != 2 {
			results = append(results, PinResult{
				Action: ref.Action,
				Tag:    ref.Version,
				Error:  fmt.Errorf("invalid action format: %s", ref.Action),
			})

			continue
		}

		sha, err := resolver.ResolveTag(ctx, parts[0], parts[1], ref.Version)
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

		_, _ = fmt.Fprintf(w, "pinned %s@%s → %s\n", ref.Action, ref.Version, sha[:12])

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

// findActionRefs extracts action references from HCL content by looking for
// uses blocks containing action and version attributes.
func findActionRefs(content string) ([]ActionRef, error) {
	actionMatches := actionPattern.FindAllStringSubmatchIndex(content, -1)
	versionMatches := versionPattern.FindAllStringSubmatchIndex(content, -1)

	if len(actionMatches) != len(versionMatches) {
		return nil, fmt.Errorf("mismatched action/version count: %d actions, %d versions", len(actionMatches), len(versionMatches))
	}

	var refs []ActionRef

	for i, am := range actionMatches {
		action := content[am[2]:am[3]]
		vm := versionMatches[i]
		version := content[vm[2]:vm[3]]

		refs = append(refs, ActionRef{
			Action:  action,
			Version: version,
			IsTag:   isTag(version),
			start:   vm[0],
			end:     trailingCommentEnd(content, vm[1]),
		})
	}

	return refs, nil
}

// drainAndClose reads the remaining body to enable HTTP connection reuse,
// then closes it.
func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	body.Close()
}
