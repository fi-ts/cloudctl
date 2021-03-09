module github.com/fi-ts/cloudctl

go 1.15

require (
	github.com/Masterminds/semver v1.5.0
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.10.0
	github.com/fi-ts/cloud-go v0.12.2-0.20210218082347-2b55501e3f23
	github.com/gardener/gardener v1.8.2
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/runtime v0.19.26 // indirect
	github.com/go-openapi/strfmt v0.20.0
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/jinzhu/now v1.1.1
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/metal-stack/metal-lib v0.6.9
	github.com/metal-stack/updater v1.1.1
	github.com/metal-stack/v v1.0.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/ini.v1 v1.60.2 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.17.12
	k8s.io/apimachinery v0.17.12
)

replace k8s.io/client-go => k8s.io/client-go v0.17.12
