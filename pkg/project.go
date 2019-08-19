package pkg

import (
	"fmt"
	"time"

	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateProject create a gardener project with the given owner
func (g *Gardener) CreateProject(owner string) (*gardenv1beta1.Project, error) {
	// spec:
	// namespace: garden-<cluster-id>
	// createdBy:
	//     apiGroup: rbac.authorization.k8s.io
	//     kind: User
	//     name: heinz.schenk@f-i-ts.de
	// members:
	// - apiGroup: rbac.authorization.k8s.io
	//     kind: User
	//     name: heinz.schenk@f-i-ts.de
	// owner:
	//     apiGroup: rbac.authorization.k8s.io
	//     kind: User
	//     name: heinz.schenk@f-i-ts.de
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
			GenerateName: "p-",
		},
		Spec: gardenv1beta1.ProjectSpec{
			CreatedBy: &c,
			Owner:     &o,
			Members:   members,
		},
	}
	project, err := g.client.GardenV1beta1().Projects().Create(p)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// CreateSecretBinding creates a secretbinding to a existing secret
func (g *Gardener) CreateSecretBinding(project *gardenv1beta1.Project) (*gardenv1beta1.SecretBinding, error) {
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
		p, err := g.client.GardenV1beta1().Projects().Get(project.GetName(), metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if p.Spec.Namespace != nil {
			namespace = *p.Spec.Namespace
		}
		fmt.Printf("namespace:%s\n", namespace)
		time.Sleep(10 * time.Millisecond)
	}

	sb := &gardenv1beta1.SecretBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metal-binding",
			Namespace: namespace,
			Labels:    map[string]string{"cloudprofile.garden.sapcloud.io/name": "metal"},
		},
		SecretRef: corev1.SecretReference{
			Name:      "seed-nbg-gardener-test-01",
			Namespace: "garden",
		},
	}
	secretBinding, err := g.client.GardenV1beta1().SecretBindings(namespace).Create(sb)
	if err != nil {
		return nil, err
	}
	return secretBinding, nil
}
