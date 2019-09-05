module git.f-i-ts.de/cloud-native/cloudctl

go 1.12

require (
	git.f-i-ts.de/cloud-native/metallib v0.0.0-20190902112911-ed799ee987fc
	github.com/gardener/gardener v0.0.0-20190816140908-ed26a3fdf2d6
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/json-iterator/go v1.1.7
	github.com/metal-pod/metal-go v0.0.0-20190904133716-d7122fdd20c2
	github.com/metal-pod/updater v0.0.0-20190905093442-85b34b88c17c
	github.com/metal-pod/v v0.0.2
	github.com/olekukonko/tablewriter v0.0.1
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0 // indirect
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf // indirect
	k8s.io/utils v0.0.0-20190809000727-6c36bc71fc4a // indirect
)

replace github.com/gardener/gardener => github.com/metal-pod/gardener v0.0.0-20190827131320-58dad7be7444
