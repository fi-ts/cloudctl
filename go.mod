module git.f-i-ts.de/cloud-native/cloudctl

go 1.12

require (
	github.com/gardener/gardener v0.0.0-20190816102845-0abd5dfb9e00
	github.com/metal-pod/v v0.0.2
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

// replace github.com/gardener/gardener => ../metal/metal-pod/gardener
