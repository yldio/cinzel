package step

type NameConfig string

func (config *NameConfig) Parse() (string, error) {
	return string(*config), nil
}
