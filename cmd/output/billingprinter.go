package output

import (
	"strings"
	"time"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
)

type (
	// BillingTablePrinter print bills in a Table
	BillingTablePrinter struct {
		TablePrinter
	}
)

// Print a Shoot as table
func (s BillingTablePrinter) Print(data *models.V1ContainerUsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "Partition", "ClusterID", "ClusterName", "Namespace", "PodUUID", "PodName", "PodStartDate", "PodEndDate", "ContainerName", "Lifetime", "CPUSeconds", "MemorySeconds", "Warnings"}
	s.shortHeader = []string{"Tenant", "ProjectName", "Partition", "ClusterName", "Namespace", "PodName", "ContainerName", "Lifetime", "CPUSeconds", "MemorySeconds"}
	for _, u := range data.Usage {
		var from string
		if data.From != nil {
			from = data.From.String()
		}
		var to string
		if !time.Time(data.To).IsZero() {
			to = data.To.String()
		}
		var tenant string
		if u.Tenant != nil {
			tenant = *u.Tenant
		}
		var projectID string
		if u.Projectid != nil {
			projectID = *u.Projectid
		}
		var projectName string
		if u.Projectname != nil {
			projectName = *u.Projectname
		}
		var partition string
		if u.Partition != nil {
			partition = *u.Partition
		}
		var clusterID string
		if u.Clusterid != nil {
			clusterID = *u.Clusterid
		}
		var clusterName string
		if u.Clustername != nil {
			clusterName = *u.Clustername
		}
		var namespace string
		if u.Namespace != nil {
			namespace = *u.Namespace
		}
		var podUUID string
		if u.Poduuid != nil {
			podUUID = *u.Poduuid
		}
		var podName string
		if u.Podname != nil {
			podName = *u.Podname
		}
		var podStart string
		if u.Podstart != nil {
			podStart = u.Podstart.String()
		}
		var podEnd string
		if u.Podend != nil {
			podEnd = u.Podend.String()
		}
		var containerName string
		if u.Containername != nil {
			containerName = *u.Containername
		}
		var lifetime time.Duration
		if u.Lifetime != nil {
			lifetime = time.Duration(*u.Lifetime)
		}
		var cpuSeconds string
		if u.Cpuseconds != nil {
			cpuSeconds = *u.Cpuseconds
			// i := new(big.Int)
			// i.SetString(*u.CPUSeconds, 10)
			// cpuSeconds = new(big.Int).Quo(i, big.NewInt(3600)).String() + "s*h"
		}
		var memorySeconds string
		if u.Memoryseconds != nil {
			memorySeconds = *u.Memoryseconds
			// i := new(big.Int)
			// i.SetString(*u.MemorySeconds, 10)
			// memorySeconds = new(big.Int).Quo(new(big.Int).Quo(i, big.NewInt(3600)), big.NewInt(1024*1024*1024)).String() + "Gi*h"
		}
		var warnings string
		if u.Warnings != nil {
			warnings = strings.Join(u.Warnings, ", ")
		}
		wide := []string{
			tenant,
			from,
			to,
			projectID,
			projectName,
			partition,
			clusterID,
			clusterName,
			namespace,
			podUUID,
			podName,
			podStart,
			podEnd,
			containerName,
			humanizeDuration(lifetime),
			cpuSeconds,
			memorySeconds,
			warnings,
		}
		short := []string{
			tenant,
			projectName,
			partition,
			clusterName,
			namespace,
			podName,
			containerName,
			humanizeDuration(lifetime),
			cpuSeconds,
			memorySeconds,
		}

		s.addWideData(wide, data)
		s.addShortData(short, data)
	}
	s.render()
}
