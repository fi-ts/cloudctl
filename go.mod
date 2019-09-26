module git.f-i-ts.de/cloud-native/cloudctl

go 1.12

require (
	git.f-i-ts.de/cloud-native/metallib v0.0.0-20190926080834-2d8284ff2703
	github.com/gardener/gardener v0.0.0-20190816140908-ed26a3fdf2d6
	github.com/go-openapi/errors v0.19.2
	github.com/go-openapi/runtime v0.19.6
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/swag v0.19.5
	github.com/go-openapi/validate v0.19.3
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/metal-pod/security v0.0.0-20190920091500-ed81ae92725b
	github.com/metal-pod/updater v0.0.0-20190905093442-85b34b88c17c
	github.com/metal-pod/v v0.0.2
	github.com/olekukonko/tablewriter v0.0.1
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	gopkg.in/square/go-jose.v2 v2.3.1
	gopkg.in/yaml.v3 v3.0.0-20190905181640-827449938966
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0 // indirect
	k8s.io/utils v0.0.0-20190809000727-6c36bc71fc4a // indirect
)

replace github.com/gardener/gardener => github.com/metal-pod/gardener v0.0.0-20190827131320-58dad7be7444
