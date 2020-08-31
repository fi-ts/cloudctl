module github.com/fi-ts/cloudctl

go 1.15

require (
	github.com/Masterminds/semver v1.5.0
	github.com/fatih/color v1.9.0
	github.com/fi-ts/cloud-go v0.7.16-0.20200831112603-f57bc639d446
	github.com/gardener/gardener v1.8.2
	github.com/go-openapi/runtime v0.19.21
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/jinzhu/now v1.1.1
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/metal-stack/metal-lib v0.5.0
	github.com/metal-stack/security v0.3.0
	github.com/metal-stack/updater v1.1.1
	github.com/metal-stack/v v1.0.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de // indirect
	google.golang.org/protobuf v1.21.0 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/ini.v1 v1.60.2 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.17.6
