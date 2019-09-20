package output

import (
	"fmt"
	"os"
	"time"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
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
func (s ShootTablePrinter) Print(data []*models.V1beta1Shoot) {
	s.wideHeader = []string{"UID", "Name", "Version", "Partition", "Domain", "Operation", "Progress", "Apiserver", "Control", "Nodes", "System", "Size", "Age"}
	s.shortHeader = s.wideHeader
	for _, shoot := range data {

		apiserver := ""
		controlplane := ""
		nodes := ""
		system := ""
		for _, condition := range shoot.Status.Conditions {
			status := *condition.Status
			switch *condition.Type {
			case string(v1beta1.ShootControlPlaneHealthy):
				controlplane = status
			case string(v1beta1.ShootEveryNodeReady):
				nodes = status
			case string(v1beta1.ShootSystemComponentsHealthy):
				system = status
			case string(v1beta1.ShootAlertsInactive):
			case string(v1beta1.ShootAPIServerAvailable):
				apiserver = status
			}
		}

		created, err := time.Parse(time.RFC3339, shoot.Metadata.CreationTimestamp)
		if err != nil {
			fmt.Printf("unable to parse creationtime: %v", err)
			os.Exit(1)
		}
		age := helper.HumanizeDuration(time.Since(created))
		operation := ""
		progress := "0%"
		if shoot.Status.LastOperation != nil {
			operation = *shoot.Status.LastOperation.State
			progress = fmt.Sprintf("%d%% [%s]", *shoot.Status.LastOperation.Progress, *shoot.Status.LastOperation.Type)
		}
		partition := shoot.Spec.Cloud.Metal.Zones[0]
		dnsdomain := ""
		if shoot.Spec.DNS.Domain != "" {
			dnsdomain = shoot.Spec.DNS.Domain
		}
		version := ""
		if shoot.Spec.Kubernetes.Version != nil {
			version = *shoot.Spec.Kubernetes.Version
		}

		autoScaleMin := int32(0)
		autoScaleMax := int32(0)
		if shoot.Spec.Cloud.Metal.Workers != nil && len(shoot.Spec.Cloud.Metal.Workers) > 0 {
			autoScaleMin = *shoot.Spec.Cloud.Metal.Workers[0].AutoScalerMin
			autoScaleMax = *shoot.Spec.Cloud.Metal.Workers[0].AutoScalerMax
		}
		size := fmt.Sprintf("%d/%d", autoScaleMin, autoScaleMax)

		wide := []string{shoot.Metadata.UID, shoot.Metadata.Name,
			version, partition, dnsdomain,
			operation,
			progress,
			apiserver, controlplane, nodes, system,
			size,
			age,
		}

		s.addWideData(wide, shoot)
		s.addShortData(wide, shoot)
	}
	s.render()
}
