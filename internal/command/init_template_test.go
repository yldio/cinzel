// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yldio/cinzel/internal/ai"
)

// The template used to restate the model names, so a new config started on
// whatever was current when the string was written rather than on what the
// code defaults to.
func TestInitTemplateUsesTheCodeDefaults(t *testing.T) {
	rendered := fmt.Sprintf(configTemplate, "anthropic", providerDefaults())

	models := ai.DefaultModels()

	if len(models) == 0 {
		t.Fatal("no default models to check")
	}

	for name, model := range models {
		want := fmt.Sprintf("    %s:\n      model: %s\n", name, model)

		if !strings.Contains(rendered, want) {
			t.Errorf("template missing %q:\n%s", want, rendered)
		}
	}

	// A name with no model under it means the block was written by hand
	// somewhere and the two fell out of step.
	for _, line := range strings.Split(rendered, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "model:") {
			value := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "model:"))

			if !modelIsADefault(models, value) {
				t.Errorf("template names %q, which no provider defaults to", value)
			}
		}
	}
}

func modelIsADefault(models map[string]string, value string) bool {
	for _, model := range models {
		if model == value {
			return true
		}
	}

	return false
}
