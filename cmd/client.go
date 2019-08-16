package cmd

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	garden "github.com/gardener/gardener/pkg/client/garden/clientset/versioned"
)

func gardenClient(kubeconfig string) (*garden.Clientset, error) {
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
	return clientset, nil
}
