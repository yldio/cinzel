package job

type UsesConfig string

type Uses string

func (config *UsesConfig) Parse() (Uses, error) {
	return Uses(*config), nil
}
