module github.com/fi-ts/cloudctl

go 1.16

require (
	github.com/Masterminds/semver v1.5.0
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.12.0
	github.com/fi-ts/cloud-go v0.17.5-0.20210716093750-59e88560d1ec
	github.com/gardener/gardener v1.18.2
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/strfmt v0.20.1
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/jinzhu/now v1.1.2
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/metal-stack/metal-lib v0.8.0
	github.com/metal-stack/updater v1.1.2
	github.com/metal-stack/v v1.0.3
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.19.11
	k8s.io/apimachinery v0.19.11
)

replace k8s.io/client-go => k8s.io/client-go v0.19.11
