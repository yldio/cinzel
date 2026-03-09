// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package step

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/hclparser"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
)

func TestStepDecodeEmptyStep(t *testing.T) {
	yaml := []byte("my-step: {}\n")

	tmp, err := ctyyaml.ImpliedType(yaml)
	if err != nil {
		t.Fatal(err)
	}

	val, err := ctyyaml.Unmarshal(yaml, tmp)
	if err != nil {
		t.Fatal(err)
	}

	var s Step

	if err := s.PreDecode(val.AsValueMap()["my-step"]); err != nil {
		t.Fatal(err)
	}
	s.Update("my-step")

	f := hclwrite.NewEmptyFile()

	if err := s.Decode(f.Body(), "step"); err != nil {
		t.Fatal(err)
	}

	got := string(hclwrite.Format(f.Bytes()))
	want := "step \"my-step\" {\n}\n"

	if got != want {
		t.Fatalf("unexpected output:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestStepDecodeWithEnvOrdering(t *testing.T) {
	yaml := []byte(`my-step:
  with:
    zebra: z
    alpha: a
    mango: m
  env:
    ZEBRA: z
    ALPHA: a
    MANGO: m
`)

	tmp, err := ctyyaml.ImpliedType(yaml)
	if err != nil {
		t.Fatal(err)
	}

	val, err := ctyyaml.Unmarshal(yaml, tmp)
	if err != nil {
		t.Fatal(err)
	}

	stepVal := val.AsValueMap()["my-step"]

	var s Step

	if err := s.PreDecode(stepVal); err != nil {
		t.Fatal(err)
	}
	s.Update("my-step")

	f := hclwrite.NewEmptyFile()

	if err := s.Decode(f.Body(), "step"); err != nil {
		t.Fatal(err)
	}

	out := string(hclwrite.Format(f.Bytes()))

	// Verify with blocks appear in alphabetical key order.
	alphaWith := strings.Index(out, `name  = "alpha"`)
	mangoWith := strings.Index(out, `name  = "mango"`)
	zebraWith := strings.Index(out, `name  = "zebra"`)

	if alphaWith < 0 || mangoWith < 0 || zebraWith < 0 {
		t.Fatalf("missing expected with block names in output:\n%s", out)
	}

	if !(alphaWith < mangoWith && mangoWith < zebraWith) {
		t.Fatalf("with blocks not in alphabetical order in output:\n%s", out)
	}

	// Verify env blocks appear in alphabetical key order.
	alphaEnv := strings.Index(out, `name  = "ALPHA"`)
	mangoEnv := strings.Index(out, `name  = "MANGO"`)
	zebraEnv := strings.Index(out, `name  = "ZEBRA"`)

	if alphaEnv < 0 || mangoEnv < 0 || zebraEnv < 0 {
		t.Fatalf("missing expected env block names in output:\n%s", out)
	}

	if !(alphaEnv < mangoEnv && mangoEnv < zebraEnv) {
		t.Fatalf("env blocks not in alphabetical order in output:\n%s", out)
	}
}

func TestStepDecodeLocalAction(t *testing.T) {
	yaml := []byte(`my-step:
  uses: ./.github/actions/my-local-action
`)

	tmp, err := ctyyaml.ImpliedType(yaml)
	if err != nil {
		t.Fatal(err)
	}

	val, err := ctyyaml.Unmarshal(yaml, tmp)
	if err != nil {
		t.Fatal(err)
	}

	stepVal := val.AsValueMap()["my-step"]

	var s Step

	if err := s.PreDecode(stepVal); err != nil {
		t.Fatal(err)
	}
	s.Update("my-step")

	f := hclwrite.NewEmptyFile()

	if err := s.Decode(f.Body(), "step"); err != nil {
		t.Fatal(err)
	}

	out := string(hclwrite.Format(f.Bytes()))

	if !strings.Contains(out, `action = "./.github/actions/my-local-action"`) {
		t.Fatalf("expected local action in uses block, got:\n%s", out)
	}

	if strings.Contains(out, "version") {
		t.Fatalf("expected no version for local action, got:\n%s", out)
	}
}

func TestStepDecodeFailure(t *testing.T) {
	type Test struct {
		name string
		have []byte
	}

	var tests = []Test{
		{
			"test 1",
			[]byte(`1`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, diags := hclsyntax.ParseExpression(tt.have, "example.hcl", hcl.Pos{})

			if diags.HasErrors() {
				t.FailNow()
			}

			hv := hclparser.NewHCLVars()
			hv.Add("test", cty.StringVal("one"))

			hp := hclparser.New(expr, hv)

			if err := hp.Parse(); err != nil {
				t.FailNow()
			}

		})
	}
}
