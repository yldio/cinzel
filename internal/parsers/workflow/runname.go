package workflow

type RunNameConfig string

func (config *RunNameConfig) Parse() (string, error) {
	// if config != nil {}
	if *config != "" {
		return string(*config), nil
	}

	return "", nil
}
