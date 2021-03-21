package api

// Contexts contains all configuration contexts of cloudctl
type Contexts struct {
	CurrentContext  string `yaml:"current"`
	PreviousContext string `yaml:"previous"`
	Contexts        map[string]Context
}

// Context configure cloudctl behaviour
type Context struct {
	ApiURL       string  `yaml:"url"`
	IssuerURL    string  `yaml:"issuer_url"`
	IssuerType   string  `yaml:"issuer_type"`
	CustomScopes string  `yaml:"custom_scopes"`
	ClientID     string  `yaml:"client_id"`
	ClientSecret string  `yaml:"client_secret"`
	HMAC         *string `yaml:"hmac"`
}
