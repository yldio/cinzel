package actions

type EnvsConfig struct {
	Name  string `hcl:"name,attr"`
	Value string `hcl:"value,attr"`
}

func (env *EnvsConfig) ConvertFromHcl() (Env, error) {
	content := Env{
		Name:  env.Name,
		Value: env.Value,
	}
	return content, nil
}

type Envs []Env

type Env struct {
	Name  string
	Value string
}

type EnvsYaml map[string]string

func (envs *Envs) ConvertToYaml() (EnvsYaml, error) {
	content := EnvsYaml{}

	for _, env := range *envs {
		content[env.Name] = env.Value
	}

	return content, nil
}
