package step

type ContinueOnErrorConfig bool

func (config *ContinueOnErrorConfig) Parse() (bool, error) {
	return bool(*config), nil
}
