package job

type SecretConfig struct {
	Name  string `hcl:"name,attr"`
	Value string `hcl:"value,attr"`
}

type SecretsConfig []SecretConfig

type SecretsInheritConfig string

type Secrets map[string]any

type SecretsInherit string

func (config *SecretsConfig) Parse() (Secrets, error) {
	secrets := make(Secrets)

	for _, secret := range *config {
		secrets[secret.Name] = secret.Value
	}
	return secrets, nil
}

func (config *SecretsInheritConfig) Parse() (SecretsInherit, error) {
	return SecretsInherit(*config), nil
}
