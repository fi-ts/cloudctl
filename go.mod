module git.f-i-ts.de/cloud-native/cloudctl

go 1.12

require (
	git.f-i-ts.de/cloud-native/metallib v0.0.0-20190902112911-ed799ee987fc
	github.com/gardener/gardener v0.0.0-20190816140908-ed26a3fdf2d6
	github.com/go-openapi/analysis v0.19.4 // indirect
	github.com/go-openapi/errors v0.19.2
	github.com/go-openapi/runtime v0.19.4
	github.com/go-openapi/strfmt v0.19.2
	github.com/go-openapi/swag v0.19.5
	github.com/go-openapi/validate v0.19.2
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/json-iterator/go v1.1.7 // indirect
	github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e // indirect
	github.com/metal-pod/security v0.0.0-20190605103437-319d1b2eca89
	github.com/metal-pod/updater v0.0.0-20190905093442-85b34b88c17c
	github.com/metal-pod/v v0.0.2
	github.com/olekukonko/tablewriter v0.0.1
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0 // indirect
	go.mongodb.org/mongo-driver v1.1.0 // indirect
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	gopkg.in/yaml.v3 v3.0.0-20190905181640-827449938966
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0 // indirect
	k8s.io/utils v0.0.0-20190809000727-6c36bc71fc4a // indirect
)

replace github.com/gardener/gardener => github.com/metal-pod/gardener v0.0.0-20190827131320-58dad7be7444
