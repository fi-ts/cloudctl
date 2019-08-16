package pkg

import (
	"time"

	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	garden "github.com/gardener/gardener/pkg/client/garden/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateProject(client *garden.Clientset, owner string) (*gardenv1beta1.Project, error) {
	c := rbacv1.Subject{
		Kind:     rbacv1.UserKind,
		Name:     owner,
		APIGroup: "rbac.authorization.k8s.io",
	}
	o := rbacv1.Subject{
		Kind:     rbacv1.UserKind,
		Name:     owner,
		APIGroup: "rbac.authorization.k8s.io",
	}
	members := []rbacv1.Subject{
		o,
	}

	p := &gardenv1beta1.Project{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "mtl-",
		},
		Spec: gardenv1beta1.ProjectSpec{
			CreatedBy: &c,
			Owner:     &o,
			Members:   members,
		},
	}
	project, err := client.GardenV1beta1().Projects().Create(p)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func CreateSecretBinding(client *garden.Clientset, project *gardenv1beta1.Project) (*gardenv1beta1.SecretBinding, error) {
	// 	apiVersion: garden.sapcloud.io/v1beta1
	// kind: SecretBinding
	// metadata:
	//   labels:
	//     cloudprofile.garden.sapcloud.io/name: metal
	//   name: cloudprovider-binding
	//   namespace: garden-<cluster-id>
	// secretRef:
	//   name: cloudprovider-config # is deployed during seed cluster bootstrapping
	//   namespace: garden

	var namespace string

	// FIXME this must be implemented with a Watcher until Namespace is set in the project.
	for namespace == "" {
		p, err := client.GardenV1beta1().Projects().Get(project.GetName(), metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if p.Spec.Namespace != nil {
			namespace = *p.Spec.Namespace
		}
		time.Sleep(10 * time.Millisecond)
	}

	sb := &gardenv1beta1.SecretBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metal-binding",
			Namespace: namespace,
			Labels:    map[string]string{"cloudprofile.garden.sapcloud.io/name": "metal"},
		},
		SecretRef: corev1.SecretReference{
			Name:      "cloudprovider-config",
			Namespace: "garden",
		},
	}
	secretBinding, err := client.GardenV1beta1().SecretBindings(namespace).Create(sb)
	if err != nil {
		return nil, err
	}
	return secretBinding, nil
}
