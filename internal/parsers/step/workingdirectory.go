package step

type WorkingDirectoryConfig string

func (config *WorkingDirectoryConfig) Parse() (string, error) {
	return string(*config), nil
}
