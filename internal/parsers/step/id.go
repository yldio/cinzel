package step

type IdConfig string

func (config *IdConfig) Parse() (string, error) {
	return string(*config), nil
}
