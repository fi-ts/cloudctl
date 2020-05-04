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
	ClientID     string  `yaml:"client_id"`
	ClientSecret string  `yaml:"client_secret"`
	HMAC         *string `yaml:"hmac"`
}
