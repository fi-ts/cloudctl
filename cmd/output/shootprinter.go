package output

import (
	"fmt"

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
	fmt.Println("Printing Shoot(s):")
	s.wideHeader = []string{"UID", "Name", "Namespace", "Version", "Seed", "Domain", "OPERATION", "PROGRESS"}
	s.shortHeader = s.wideHeader
	for _, shoot := range data {

		wide := []string{string(shoot.UID), shoot.Name, shoot.Namespace,
			shoot.Spec.Kubernetes.Version, shoot.Status.Seed, *shoot.Spec.DNS.Domain,
			string(shoot.Status.LastOperation.State),
			fmt.Sprintf("%d%%", shoot.Status.LastOperation.Progress),
		}

		s.addWideData(wide, shoot)
		s.addShortData(wide, shoot)
	}
	s.render()
}
