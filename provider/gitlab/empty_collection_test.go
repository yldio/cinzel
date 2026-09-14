package gitlab

import (
	"strings"
	"testing"
)

// GitLab reads an explicitly empty "needs", "cache", "services" or "rules" as
// "override whatever this would inherit", which a missing keyword does not do.
// Each is a block in the schema, and a block has no empty spelling, so all four
// used to vanish on the roundtrip.
func TestEmptyCollectionsSurvive(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "job needs",
			yaml: "job1:\n  script:\n    - make\n  needs: []\n",
			want: "needs: []",
		},
		{
			name: "job cache",
			yaml: "job1:\n  script:\n    - make\n  cache: []\n",
			want: "cache: []",
		},
		{
			name: "job services",
			yaml: "job1:\n  script:\n    - make\n  services: []\n",
			want: "services: []",
		},
		{
			name: "job rules",
			yaml: "job1:\n  script:\n    - make\n  rules: []\n",
			want: "rules: []",
		},
		{
			name: "default cache",
			yaml: "default:\n  cache: []\njob1:\n  script:\n    - make\n",
			want: "cache: []",
		},
		{
			name: "default services",
			yaml: "default:\n  services: []\njob1:\n  script:\n    - make\n",
			want: "services: []",
		},
		{
			name: "workflow rules",
			yaml: "workflow:\n  rules: []\njob1:\n  script:\n    - make\n",
			want: "rules: []",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl, back := roundtripYAML(t, tc.yaml)

			if !strings.Contains(back, tc.want) {
				t.Errorf("roundtrip lost %q\nHCL:\n%s\nYAML:\n%s", tc.want, hcl, back)
			}
		})
	}
}

// The empty list must not be invented where the keyword was absent, nor
// replace a keyword that carries real entries.
func TestAbsentAndPopulatedCollectionsAreUnchanged(t *testing.T) {
	for _, tc := range []struct {
		name    string
		yaml    string
		want    []string
		notWant []string
	}{
		{
			name:    "absent stays absent",
			yaml:    "job1:\n  script:\n    - make\n",
			notWant: []string{"needs:", "cache:", "services:", "rules:"},
		},
		{
			name: "populated needs",
			yaml: "build:\n  script:\n    - make\njob1:\n  script:\n    - make\n  needs:\n    - build\n",
			want: []string{"needs:", "- build"},
		},
		{
			name: "populated cache",
			yaml: "job1:\n  script:\n    - make\n  cache:\n    key: k\n    paths:\n      - vendor\n",
			want: []string{"key: k", "- vendor"},
		},
		{
			name: "populated services",
			yaml: "job1:\n  script:\n    - make\n  services:\n    - postgres:16\n",
			// GitLab reads a bare service name as the "name" of one,
			// which is the spelling that comes back.
			want: []string{"name: \"postgres:16\""},
		},
		{
			name: "populated rules",
			yaml: "job1:\n  script:\n    - make\n  rules:\n    - if: $CI_COMMIT_BRANCH\n",
			want: []string{"if: $CI_COMMIT_BRANCH"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl, back := roundtripYAML(t, tc.yaml)

			for _, want := range tc.want {
				if !strings.Contains(back, want) {
					t.Errorf("roundtrip lost %q\nHCL:\n%s\nYAML:\n%s", want, hcl, back)
				}
			}

			for _, notWant := range tc.notWant {
				if strings.Contains(back, notWant) {
					t.Errorf("roundtrip invented %q\nHCL:\n%s\nYAML:\n%s", notWant, hcl, back)
				}
			}
		})
	}
}

// A non-empty list written as an attribute is not a spelling of the block, so
// it is rejected rather than silently ignored.
func TestNonEmptyCollectionAttributeIsRejected(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
	}{
		{
			name: "cache",
			hcl:  "job \"job1\" {\n  script = [\"make\"]\n  cache  = [\"vendor\"]\n}\n",
		},
		{
			name: "services",
			hcl:  "job \"job1\" {\n  script   = [\"make\"]\n  services = [\"postgres\"]\n}\n",
		},
		{
			name: "rules",
			hcl:  "job \"job1\" {\n  script = [\"make\"]\n  rules  = [\"nope\"]\n}\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseHCLString(t, tc.hcl)

			if err == nil {
				t.Fatal("Parse() accepted a non-empty list where blocks are required")
			}

			if !strings.Contains(err.Error(), "must be written as blocks") {
				t.Errorf("Parse() error = %v, want it to mention blocks", err)
			}
		})
	}
}
