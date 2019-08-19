package pkg

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	garden "github.com/gardener/gardener/pkg/client/garden/clientset/versioned"
)

// Gardener provides gardener functions
type Gardener struct {
	client *garden.Clientset
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

	// create the clientset
	clientset, err := garden.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Gardener{client: clientset}, nil
}
