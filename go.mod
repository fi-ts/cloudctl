module git.f-i-ts.de/cloud-native/cloudctl

go 1.13

require (
	github.com/Masterminds/semver v1.5.0
	github.com/gardener/gardener v1.0.4
	github.com/go-openapi/errors v0.19.3
	github.com/go-openapi/loads v0.19.5 // indirect
	github.com/go-openapi/runtime v0.19.11
	github.com/go-openapi/strfmt v0.19.4
	github.com/go-openapi/swag v0.19.7
	github.com/go-openapi/validate v0.19.6
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/metal-stack/metal-go v0.3.1
	github.com/metal-stack/metal-lib v0.3.2
	github.com/metal-stack/security v0.3.0
	github.com/metal-stack/updater v1.0.1
	github.com/metal-stack/v v1.0.1
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.6
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.6.2
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/ini.v1 v1.52.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200121175148-a6ecf24a6d71
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.17.0
