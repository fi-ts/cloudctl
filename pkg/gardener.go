package pkg

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	garden "github.com/gardener/gardener/pkg/client/garden/clientset/versioned"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

// Gardener provides gardener functions
type Gardener struct {
	client    *garden.Clientset
	k8sclient *kubernetes.Clientset
}

// NewGardener create a new Gardener struct from a kubeconfig
func NewGardener(kubeconfig string) (*Gardener, error) {
	var config *rest.Config
	var err error

	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		return nil, err
	}
	k8sclientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// create the clientset
	gclientset, err := garden.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Gardener{client: gclientset, k8sclient: k8sclientset}, nil
}
