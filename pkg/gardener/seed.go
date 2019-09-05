package gardener

import (
	"git.f-i-ts.de/cloud-native/cloudctl/pkg/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ShootConstraints returns all constraints from our cloudprofile
func (g *Gardener) ShootConstraints() (*api.ShootConstraints, error) {
	cloudProfile, err := g.gclient.GardenV1beta1().CloudProfiles().Get("metal", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	pmap := make(map[string]bool)
	zones := cloudProfile.Spec.Metal.Constraints.Zones
	for _, z := range zones {
		for _, n := range z.Names {
			pmap[n] = true
		}
	}
	partitions := []string{}
	for k := range pmap {
		partitions = append(partitions, k)
	}

	sc := &api.ShootConstraints{
		KubernetesVersions: cloudProfile.Spec.Metal.Constraints.Kubernetes.Versions,
		Partitions:         partitions,
	}

	return sc, nil
}
