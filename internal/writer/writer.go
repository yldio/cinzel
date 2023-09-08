package writer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/yldio/atos/internal/parsers/actions"
)

type Structure []actions.Workflow

type Actionwriter[K Structure] struct {
	Content K
}

const (
	YAML_EXT       = "yaml"
	YAML_EXT_SHORT = "yml"
	GHA_PATH       = ".github/workflows"
)

func NewWriter[K Structure]() *Actionwriter[K] {
	return &Actionwriter[K]{}
}

func (aw *Actionwriter[K]) Do(k K) error {
	aw.Content = k
	return nil
}

func (aw *Actionwriter[K]) Save(directory string, file string) error {
	for _, workflow := range aw.Content {
		content, err := workflow.ConvertToYaml()
		if err != nil {
			return err
		}

		out, err := yaml.Marshal(&content)
		if err != nil {
			return err
		}

		tmpFile := fmt.Sprintf("%s.%s", workflow.Id, YAML_EXT)

		f, err := os.Create(filepath.Join(directory, tmpFile))
		if err != nil {
			return err
		}

		defer f.Close()

		_, err = f.Write(out)
		if err != nil {
			return err
		}
	}

	return nil
}
