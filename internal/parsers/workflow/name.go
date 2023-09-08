package workflow

type NameConfig string

func (config *NameConfig) Parse() (string, error) {
	// if config != nil {}
	if *config != "" {
		return string(*config), nil
	}

	return "", nil
}
