module github.com/fi-ts/cloudctl

go 1.15

require (
	github.com/Masterminds/semver v1.5.0
	github.com/fatih/color v1.9.0
	github.com/fi-ts/cloud-go v0.8.7-0.20201105095015-24be3411a655
	github.com/gardener/gardener v1.8.2
	github.com/go-openapi/runtime v0.19.21
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/jinzhu/now v1.1.1
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/metal-stack/metal-lib v0.6.2
	github.com/metal-stack/security v0.4.0
	github.com/metal-stack/updater v1.1.1
	github.com/metal-stack/v v1.0.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/ini.v1 v1.60.2 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

replace k8s.io/client-go => k8s.io/client-go v0.17.6
