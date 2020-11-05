package output

import (
	"fmt"
	"strings"
	"time"

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
	s.wideHeader = []string{"UID", "Name", "Version", "Partition", "Domain", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose", "Privileged", "Runtime", "Firewall", "Egress Net", "Egress IP"}
	s.shortHeader = []string{"UID", "Tenant", "Project", "Name", "Version", "Partition", "Operation", "Progress", "Api", "Control", "Nodes", "System", "Size", "Age", "Purpose"}
	s.Order(data)
	for i := range data {
		shoot := data[i]

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

		age := ""
		if shoot.CreationTimestamp != nil {
			age = helper.HumanizeDuration(time.Since(time.Time(*shoot.CreationTimestamp)))
		}
		operation := ""
		progress := "0%"
		if shoot.Status.LastOperation != nil {
			operation = *shoot.Status.LastOperation.State
			progress = fmt.Sprintf("%d%% [%s]", *shoot.Status.LastOperation.Progress, *shoot.Status.LastOperation.Type)
		}
		partition := ""
		if shoot.PartitionID != nil {
			partition = *shoot.PartitionID
		}
		dnsdomain := ""
		if shoot.DNSEndpoint != nil {
			dnsdomain = *shoot.DNSEndpoint
		}
		version := ""
		if shoot.Kubernetes.Version != nil {
			version = *shoot.Kubernetes.Version
		}
		purpose := ""
		if shoot.Purpose != nil {
			p := *shoot.Purpose
			purpose = p[:4]
		}

		privileged := ""
		if shoot.Kubernetes.AllowPrivilegedContainers != nil {
			privileged = fmt.Sprintf("%t", *shoot.Kubernetes.AllowPrivilegedContainers)
		}

		runtime := "docker"
		autoScaleMin := int32(0)
		autoScaleMax := int32(0)
		if shoot.Workers != nil && len(shoot.Workers) > 0 {
			workers := shoot.Workers[0]
			autoScaleMin = *workers.Minimum
			autoScaleMax = *workers.Maximum
			if workers.CRI != nil && *workers.CRI != "" {
				runtime = *workers.CRI
			}
		}
		size := fmt.Sprintf("%d/%d", autoScaleMin, autoScaleMax)
		tenant := ""
		if shoot.Tenant != nil {
			tenant = *shoot.Tenant
		}
		project := ""
		if shoot.ProjectID != nil {
			project = *shoot.ProjectID
		}

		firewallImage := ""
		if shoot.FirewallImage != nil {
			firewallImage = *shoot.FirewallImage
		}

		egressNets := []string{}
		egressIPs := []string{}
		for _, e := range shoot.EgressRules {
			if e == nil {
				continue
			}
			for _, i := range e.Ips {
				egressNets = append(egressNets, *e.NetworkID)
				egressIPs = append(egressIPs, i)
			}
		}

		wide := []string{*shoot.ID, *shoot.Name,
			version, partition, dnsdomain,
			operation,
			progress,
			apiserver, controlplane, nodes, system,
			size,
			age,
			purpose,
			privileged,
			runtime,
			firewallImage,
			strings.Join(egressNets, "\n"),
			strings.Join(egressIPs, "\n"),
		}
		short := []string{*shoot.ID,
			tenant,
			project,
			*shoot.Name,
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
