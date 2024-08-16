package writer

import "os"

type Writer struct {
	// content []byte
}

func New() *Writer {
	return &Writer{}
}

func (writer *Writer) Do(filePath string, content []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return err
	}

	return nil
}

// func convertToYaml(content actions.WorkflowYaml) ([]byte, error) {
// 	out, err := yaml.Marshal(&content)
// 	if err != nil {
// 		return []byte{}, err
// 	}

// 	return out, nil
// }

// func saveToYaml(filePath string, content []byte) error {
// 	f, err := os.Create(filePath)
// 	if err != nil {
// 		return err
// 	}

// 	defer f.Close()

// 	_, err = f.Write(content)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func dirExists(directory string) error {
// 	if _, err := os.Stat(directory); os.IsNotExist(err) {
// 		return err
// 	}

// 	return nil
// }

// func (writer *Writer) Do() error {
// 	for _, content := range writer.content {
// 		filePath := ""

// 		err := dirExists(filePath)
// 		if err != nil {
// 			return err
// 		}

// 		out, err := convertToYaml(content)
// 		if err != nil {
// 			return err
// 		}

// 		err = saveToYaml(filePath, out)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
