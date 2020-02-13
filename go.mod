module git.f-i-ts.de/cloud-native/cloudctl

go 1.12

require (
	git.f-i-ts.de/cloud-native/masterdata-api v0.0.0-20191206140543-452cd01b49fb
	git.f-i-ts.de/cloud-native/metallib v0.2.5
	github.com/Masterminds/semver v1.4.2
	github.com/gardener/gardener v0.0.0-20190816140908-ed26a3fdf2d6
	github.com/go-openapi/errors v0.19.2
	github.com/go-openapi/runtime v0.19.6
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/swag v0.19.5
	github.com/go-openapi/validate v0.19.3
	github.com/go-playground/locales v0.12.1 // indirect
	github.com/go-playground/universal-translator v0.16.0 // indirect
	github.com/leodido/go-urn v1.1.0 // indirect
	github.com/metal-pod/metal-go v0.0.0-20191104125404-3f11e972879a
	github.com/metal-pod/security v0.0.0-20191127130239-3547755283e3
	github.com/metal-pod/updater v1.0.0
	github.com/metal-pod/v v1.0.0
	github.com/miekg/dns v1.1.22 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.5.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.29.1
	gopkg.in/yaml.v3 v3.0.0-20191026110619-0b21df46bc1d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0 // indirect
	k8s.io/utils v0.0.0-20190809000727-6c36bc71fc4a // indirect
)

replace github.com/gardener/gardener => github.com/metal-pod/gardener v0.0.0-20190827131320-58dad7be7444
