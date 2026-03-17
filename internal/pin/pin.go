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
	cacheTTL       = 24 * time.Hour
	cacheSubdir    = "cinzel/pins"
	githubAPIBase  = "https://api.github.com"
	tokenEnvVar    = "GITHUB_TOKEN"
)

// tagPattern matches version strings that look like tags (v1, v1.2, v1.2.3)
// as opposed to SHAs (40+ hex chars).
var tagPattern = regexp.MustCompile(`^v?\d+(\.\d+)*$`)

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

	defer resp.Body.Close()

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

	defer resp.Body.Close()

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

	defer resp.Body.Close()

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

	refs := findActionRefs(string(content))
	if len(refs) == 0 {
		return nil, nil
	}

	var results []PinResult

	updated := string(content)

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

		// Replace version value in the HCL content.
		oldLine := fmt.Sprintf(`version = %q`, ref.Version)
		newLine := fmt.Sprintf(`version = %q`, sha)
		updated = strings.Replace(updated, oldLine, newLine, 1)

		// Add or update the comment above the uses block.
		updated = upsertUsesComment(updated, ref.Action, ref.Version)

		_, _ = fmt.Fprintf(w, "pinned %s@%s → %s\n", ref.Action, ref.Version, sha[:12])

		results = append(results, PinResult{
			Action: ref.Action,
			Tag:    ref.Version,
			SHA:    sha,
		})
	}

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
func findActionRefs(content string) []ActionRef {
	// Match action = "owner/repo" followed by version = "tag-or-sha"
	// within uses blocks. Uses a simple regex approach since the HCL
	// structure is well-defined from cinzel's own output.
	actionPattern := regexp.MustCompile(`action\s*=\s*"([^"]+)"`)
	versionPattern := regexp.MustCompile(`version\s*=\s*"([^"]+)"`)

	actionMatches := actionPattern.FindAllStringSubmatchIndex(content, -1)
	versionMatches := versionPattern.FindAllStringSubmatchIndex(content, -1)

	if len(actionMatches) != len(versionMatches) {
		return nil
	}

	var refs []ActionRef

	for i, am := range actionMatches {
		action := content[am[2]:am[3]]
		version := content[versionMatches[i][2]:versionMatches[i][3]]

		refs = append(refs, ActionRef{
			Action:  action,
			Version: version,
			IsTag:   isTag(version),
		})
	}

	return refs
}

// upsertUsesComment adds or updates the comment line above a uses block
// to document the original action and tag. For example:
//
//	// actions/checkout v4
//	uses {
//	  action  = "actions/checkout"
//	  version = "abc123..."
//	}
func upsertUsesComment(content, action, tag string) string {
	comment := fmt.Sprintf("// %s %s", action, tag)
	actionLine := fmt.Sprintf(`action  = %q`, action)

	idx := strings.Index(content, actionLine)
	if idx <= 0 {
		return content
	}

	beforeAction := content[:idx]
	usesIdx := strings.LastIndex(beforeAction, "uses {")

	if usesIdx < 0 {
		return content
	}

	// Find the indent by looking at what's before "uses {" on its line.
	lineStart := strings.LastIndex(content[:usesIdx], "\n") + 1
	indent := content[lineStart:usesIdx]

	beforeUses := content[:lineStart]
	afterUses := content[lineStart:]
	lines := strings.Split(beforeUses, "\n")

	// Check if the line before uses (skipping blank line) is already a comment.
	lastIdx := len(lines) - 1

	if lastIdx >= 0 && strings.TrimSpace(lines[lastIdx]) == "" {
		lastIdx--
	}

	if lastIdx >= 0 && strings.HasPrefix(strings.TrimSpace(lines[lastIdx]), "//") {
		// Update existing comment, preserving its indent.
		existingIndent := lines[lastIdx][:len(lines[lastIdx])-len(strings.TrimLeft(lines[lastIdx], " \t"))]
		lines[lastIdx] = existingIndent + comment

		return strings.Join(lines, "\n") + afterUses
	}

	// No existing comment — insert one before uses, same indent.
	return beforeUses + indent + comment + "\n" + afterUses
}
