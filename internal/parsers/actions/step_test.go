package actions

import (
	"testing"
)

func TestStep(t *testing.T) {
	t.Run("converts Uses to yaml", func(t *testing.T) {
		uses := Uses{
			Action:  "some-action",
			Version: "@v1",
		}

		u := uses.ConvertYaml()

		if u != "some-action@v1" {
			t.Errorf("wrong format of use's action and version")
		}
	})
}
