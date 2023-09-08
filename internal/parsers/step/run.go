package step

type RunConfig string

func (config *RunConfig) Parse() (string, error) {
	return string(*config), nil
}
