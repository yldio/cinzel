package workflow

type OnConfig string

func (config *OnConfig) Parse() (string, error) {
	// if config != nil {}
	if *config != "" {
		return string(*config), nil
	}

	return "", nil
}
