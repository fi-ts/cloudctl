module git.f-i-ts.de/cloud-native/cloudctl

go 1.12

require (
	git.f-i-ts.de/cloud-native/metallib v0.0.0-20190701111916-6ee4d6ba0799
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/cheggaaa/pb/v3 v3.0.1
	github.com/gardener/gardener v0.0.0-20190816140908-ed26a3fdf2d6
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/json-iterator/go v1.1.7
	github.com/metal-pod/metal-go v0.0.0-20190826111730-79163cab1356 // indirect
	github.com/metal-pod/v v0.0.2
	github.com/olekukonko/tablewriter v0.0.1
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0 // indirect
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf // indirect
	k8s.io/utils v0.0.0-20190809000727-6c36bc71fc4a // indirect
)

replace github.com/gardener/gardener => github.com/metal-pod/gardener v0.0.0-20190718123228-aa96bb627f09
