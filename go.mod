module github.com/fi-ts/cloudctl

go 1.16

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.12.0
	github.com/fi-ts/cloud-go v0.17.16-0.20210903170323-84c05bbc2cf7
	github.com/gardener/gardener v1.19.3
	github.com/ghodss/yaml v1.0.0
	github.com/gizak/termui/v3 v3.1.0
	github.com/go-openapi/errors v0.20.1 // indirect
	github.com/go-openapi/runtime v0.19.30 // indirect
	github.com/go-openapi/strfmt v0.20.2
	github.com/go-playground/validator/v10 v10.9.0
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jinzhu/now v1.1.2
	github.com/metal-stack/metal-lib v0.8.0
	github.com/metal-stack/security v0.6.1 // indirect
	github.com/metal-stack/updater v1.1.3
	github.com/metal-stack/v v1.0.3
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/tidwall/pretty v1.2.0 // indirect
	go.uber.org/zap v1.18.1 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.20.9
	k8s.io/apimachinery v0.20.9
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
)

replace k8s.io/client-go => k8s.io/client-go v0.20.9
