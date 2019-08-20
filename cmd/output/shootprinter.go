package output

import (
	"fmt"
	"time"

	"git.f-i-ts.de/cloud-native/cloudctl/cmd/helper"
	"github.com/gardener/gardener/pkg/apis/garden/v1beta1"
)

type (
	// ShootTablePrinter print a Shoot Cluster in a Table
	ShootTablePrinter struct {
		TablePrinter
	}
)

// Print a Shoot as table
func (s ShootTablePrinter) Print(data []v1beta1.Shoot) {
	s.wideHeader = []string{"UID", "Name", "Version", "Seed", "Domain", "Operation", "Progress", "Apiserver", "Control", "Nodes", "System", "Age"}
	s.shortHeader = s.wideHeader
	for _, shoot := range data {

		apiserver := ""
		controlplane := ""
		nodes := ""
		system := ""
		for _, condition := range shoot.Status.Conditions {
			status := string(condition.Status)
			switch condition.Type {
			case v1beta1.ShootControlPlaneHealthy:
				controlplane = status
			case v1beta1.ShootEveryNodeReady:
				nodes = status
			case v1beta1.ShootSystemComponentsHealthy:
				system = status
			case v1beta1.ShootAlertsInactive:
			case v1beta1.ShootAPIServerAvailable:
				apiserver = status
			}
		}

		created := shoot.ObjectMeta.CreationTimestamp.Time
		age := helper.HumanizeDuration(time.Now().Sub(created))
		operation := ""
		progress := "0%"
		if shoot.Status.LastOperation != nil {
			operation = string(shoot.Status.LastOperation.State)
			progress = fmt.Sprintf("%d%%", shoot.Status.LastOperation.Progress)
		}
		seed := ""
		if shoot.Spec.Cloud.Seed != nil {
			seed = *shoot.Spec.Cloud.Seed
		}
		dnsdomain := ""
		if shoot.Spec.DNS.Domain != nil {
			dnsdomain = *shoot.Spec.DNS.Domain
		}
		wide := []string{string(shoot.UID), shoot.Name,
			shoot.Spec.Kubernetes.Version, seed, dnsdomain,
			operation,
			progress,
			apiserver, controlplane, nodes, system,
			age,
		}

		s.addWideData(wide, shoot)
		s.addShortData(wide, shoot)
	}
	s.render()
}
