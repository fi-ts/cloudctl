package output

import (
	"fmt"
	"os"
	"time"

	"github.com/metal-stack/metal-lib/pkg/tag"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
)

type (
	// ShootTablePrinter print a Shoot Cluster in a Table
	ShootTablePrinter struct {
		TablePrinter
	}
	// ShootConditionsTablePrinter print the Conditions of a Shoot Cluster in a Table
	ShootConditionsTablePrinter struct {
		TablePrinter
	}
)

func (s ShootConditionsTablePrinter) Print(data []*models.V1beta1Condition) {
	s.wideHeader = []string{"LastTransition", "LastUpdate", "Message", "Reason", "Status", "Type"}
	s.shortHeader = []string{"LastTransition", "LastUpdate", "Message", "Reason", "Status", "Type"}
	for _, condition := range data {
		wide := []string{
			strValue(condition.LastTransitionTime),
			strValue(condition.LastUpdateTime),
			strValue(condition.Message),
			strValue(condition.Reason),
			strValue(condition.Status),
			strValue(condition.Type),
		}
		short := wide
		s.addWideData(wide, data)
		s.addShortData(short, data)
	}
	s.render()
}

// Print a Shoot as table
func (s ShootTablePrinter) Print(data []*models.V1ClusterResponse) {
	s.wideHeader = []string{"UID", "Name", "Version", "Partition", "Domain", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose", "Privileged", "Firewall"}
	s.shortHeader = []string{"UID", "Tenant", "Project", "Name", "Version", "Partition", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose"}
	s.Order(data)
	for _, cluster := range data {
		shoot := cluster.Shoot
		infrastructure := cluster.Infrastructure

		apiserver := ""
		controlplane := ""
		nodes := ""
		system := ""
		if shoot.Status != nil {
			for _, condition := range shoot.Status.Conditions {
				status := *condition.Status
				switch *condition.Type {
				case string(v1beta1.ShootControlPlaneHealthy):
					controlplane = status
				case string(v1beta1.ShootEveryNodeReady):
					nodes = status
				case string(v1beta1.ShootSystemComponentsHealthy):
					system = status
				case string(v1beta1.ShootAPIServerAvailable):
					apiserver = status
				}
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
		partition := ""
		if infrastructure != nil && infrastructure.PartitionID != nil {
			partition = *infrastructure.PartitionID
		}
		dnsdomain := ""
		if shoot.Spec.DNS != nil && shoot.Spec.DNS.Domain != "" {
			dnsdomain = shoot.Spec.DNS.Domain
		}
		version := ""
		if shoot.Spec.Kubernetes.Version != nil {
			version = *shoot.Spec.Kubernetes.Version
		}
		purpose := ""
		if len(shoot.Spec.Purpose) > 0 {
			purpose = shoot.Spec.Purpose[:4]
		}

		privileged := shoot.Spec.Kubernetes.AllowPrivilegedContainers

		autoScaleMin := int32(0)
		autoScaleMax := int32(0)
		if shoot.Spec.Provider.Workers != nil && len(shoot.Spec.Provider.Workers) > 0 {
			workers := shoot.Spec.Provider.Workers[0]
			autoScaleMin = *workers.Minimum
			autoScaleMax = *workers.Maximum
		}
		size := fmt.Sprintf("%d/%d", autoScaleMin, autoScaleMax)
		tenant := shoot.Metadata.Annotations[tag.ClusterTenant]
		project := shoot.Metadata.Annotations[tag.ClusterProject]

		firewallImage := ""
		if infrastructure.Firewall != nil && infrastructure.Firewall.Image != nil {
			firewallImage = *infrastructure.Firewall.Image
		}
		wide := []string{shoot.Metadata.UID, shoot.Metadata.Name,
			version, partition, dnsdomain,
			operation,
			progress,
			apiserver, controlplane, nodes, system,
			size,
			age,
			purpose,
			fmt.Sprintf("%t", privileged),
			firewallImage,
		}
		short := []string{shoot.Metadata.UID,
			tenant,
			project,
			shoot.Metadata.Name,
			version, partition,
			operation,
			progress,
			apiserver, controlplane, nodes, system,
			size,
			age,
			purpose,
		}
		s.addWideData(wide, shoot)
		s.addShortData(short, shoot)
	}
	s.render()
}
