module github.com/fi-ts/cloudctl

go 1.17

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.13.0
	github.com/fi-ts/cloud-go v0.18.4
	github.com/gardener/gardener v1.24.3
	github.com/gizak/termui/v3 v3.1.0
	github.com/go-openapi/strfmt v0.21.1
	github.com/go-playground/validator/v10 v10.10.0
	github.com/gosimple/slug v1.12.0
	github.com/jinzhu/now v1.1.4
	github.com/metal-stack/duros-go v0.3.0
	github.com/metal-stack/metal-lib v0.9.0
	github.com/metal-stack/updater v1.1.3
	github.com/metal-stack/v v1.0.3
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.20.14
	k8s.io/apimachinery v0.20.14
	k8s.io/utils v0.0.0-20211208161948-7d6a63dca704
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/emicklei/go-restful-openapi/v2 v2.8.0 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.21.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/goccy/go-json v0.9.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/lestrrat-go/jwx v1.2.17 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/nsf/termbox-go v1.1.1 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.17.0 // indirect
	github.com/spf13/afero v1.8.0 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/xuri/excelize/v2 v2.5.0
	go.mongodb.org/mongo-driver v1.8.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/goleak v1.1.12 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.20.0 // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	golang.org/x/net v0.0.0-20220114011407-0dd24b26b47d // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	google.golang.org/genproto v0.0.0-20220114231437-d2e6a121cae0 // indirect
	k8s.io/klog/v2 v2.40.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220114203427-a0453230fd26 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.20.14
