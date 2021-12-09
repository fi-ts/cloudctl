module github.com/fi-ts/cloudctl

go 1.16

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.13.0
	github.com/fi-ts/cloud-go v0.18.4-0.20211209115648-e0bb90b6f119
	github.com/gardener/gardener v1.22.6
	github.com/gizak/termui/v3 v3.1.0
	github.com/go-openapi/strfmt v0.21.1
	github.com/go-playground/validator/v10 v10.9.0
	github.com/gosimple/slug v1.11.2
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jinzhu/now v1.1.3
	github.com/metal-stack/duros-go v0.2.2
	github.com/metal-stack/metal-lib v0.9.0
	github.com/metal-stack/updater v1.1.3
	github.com/metal-stack/v v1.0.3
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.20.9
	k8s.io/apimachinery v0.20.9
	k8s.io/utils v0.0.0-20211116205334-6203023598ed
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	k8s.io/client-go => k8s.io/client-go v0.20.9
)
