module github.com/fi-ts/cloudctl

go 1.16

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.12.0
	github.com/fi-ts/cloud-go v0.17.12-0.20210824173002-5e2eb389eba1
	github.com/gardener/gardener v1.19.3
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/strfmt v0.20.1
	github.com/go-playground/validator/v10 v10.9.0
	github.com/jinzhu/now v1.1.2
	github.com/metal-stack/metal-lib v0.8.0
	github.com/metal-stack/updater v1.1.3
	github.com/metal-stack/v v1.0.3
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.20.9
	k8s.io/apimachinery v0.20.9
)

replace k8s.io/client-go => k8s.io/client-go v0.20.9
