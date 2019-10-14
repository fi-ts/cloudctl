package output

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
	"github.com/spf13/viper"
)

type (
	// ClusterBillingTablePrinter print bills in a Table
	ClusterBillingTablePrinter struct {
		TablePrinter
	}
	// ContainerBillingTablePrinter print bills in a Table
	ContainerBillingTablePrinter struct {
		TablePrinter
	}
)

// Print a cluster usage as table
func (s ClusterBillingTablePrinter) Print(data *models.V1ClusterUsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "Partition", "ClusterID", "ClusterName", "ClusterStart", "ClusterEnd", "Lifetime", "Warnings"}
	s.shortHeader = []string{"Tenant", "ProjectName", "Partition", "ClusterName", "ClusterStart", "ClusterEnd", "Lifetime"}
	s.Order(data.Usage)
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
		var clusterStart string
		if u.Clusterstart != nil {
			clusterStart = u.Clusterstart.String()
		}
		var clusterEnd string
		if u.Clusterend != nil {
			clusterEnd = u.Clusterend.String()
		}
		var lifetime time.Duration
		if u.Lifetime != nil {
			lifetime = time.Duration(*u.Lifetime)
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
			clusterStart,
			clusterEnd,
			humanizeDuration(lifetime),
			warnings,
		}
		short := []string{
			tenant,
			projectName,
			partition,
			clusterName,
			clusterStart,
			clusterEnd,
			humanizeDuration(lifetime),
		}

		s.addWideData(wide, data)
		s.addShortData(short, data)
	}

	footer := []string{"Total",
		humanizeDuration(time.Duration(*data.Accumulatedusage.Lifetime)),
	}
	shortFooter := make([]string, len(s.shortHeader)-len(footer))
	wideFooter := make([]string, len(s.wideHeader)-len(footer))
	s.addWideData(append(wideFooter, footer...), data)

	s.addShortData(append(shortFooter, footer...), data)
	s.render()
}

// Print a container usage as table
func (s ContainerBillingTablePrinter) Print(data *models.V1ContainerUsageResponse) {
	s.wideHeader = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "Partition", "ClusterID", "ClusterName", "Namespace", "PodUUID", "PodName", "PodStartDate", "PodEndDate", "ContainerName", "Lifetime", "CPUSeconds", "MemorySeconds", "Warnings"}
	s.shortHeader = []string{"Tenant", "ProjectName", "Partition", "ClusterName", "Namespace", "PodName", "ContainerName", "Lifetime", "CPU (1 * s)", "Memory (Gi * h)"}
	s.Order(data.Usage)
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
		var cpuUsage string
		if u.Cpuseconds != nil {
			cpuUsage = humanizeCPU(*u.Cpuseconds)
		}
		var memoryUsage string
		if u.Memoryseconds != nil {
			memoryUsage = humanizeMemory(*u.Memoryseconds)
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
			cpuUsage,
			memoryUsage,
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
			cpuUsage,
			memoryUsage,
		}

		s.addWideData(wide, data)
		s.addShortData(short, data)
	}

	footer := []string{"Total",
		humanizeDuration(time.Duration(*data.Accumulatedusage.Lifetime)),
		humanizeCPU(*data.Accumulatedusage.Cpuseconds) + cpuCosts(*data.Accumulatedusage.Cpuseconds),
		humanizeMemory(*data.Accumulatedusage.Memoryseconds) + memoryCosts(*data.Accumulatedusage.Memoryseconds),
	}
	shortFooter := make([]string, len(s.shortHeader)-len(footer))
	wideFooter := make([]string, len(s.wideHeader)-len(footer))
	s.addWideData(append(wideFooter, footer...), data)

	s.addShortData(append(shortFooter, footer...), data)
	s.render()
}

func humanizeMemory(memorySeconds string) string {
	i := new(big.Float)
	i.SetString(memorySeconds)
	ms := new(big.Float).Quo(i, big.NewFloat(1<<30))
	memoryHours := new(big.Float).Quo(ms, big.NewFloat(3600))
	return fmt.Sprintf("%.2f", memoryHours)
}

func humanizeCPU(cpuSeconds string) string {
	duration, err := strconv.ParseInt(cpuSeconds, 10, 64)
	if err == nil {
		return humanizeDuration(time.Duration(duration) * time.Second)
	}
	return ""
}

func cpuCosts(cpuSeconds string) string {
	cpuPerCoreAndHour := viper.GetFloat64("costs-cpu-hour")
	if cpuPerCoreAndHour <= 0 {
		return ""
	}
	duration, err := strconv.ParseInt(cpuSeconds, 10, 64)
	if err == nil {
		return fmt.Sprintf(" (%.2f €)", float64(duration/3600)*cpuPerCoreAndHour)
	}
	return ""
}

func memoryCosts(memorySeconds string) string {
	memoryPerGiAndHour := viper.GetFloat64("costs-memory-gi-hour")
	if memoryPerGiAndHour <= 0 {
		return ""
	}
	i := new(big.Float)
	i.SetString(memorySeconds)
	ms := new(big.Float).Quo(i, big.NewFloat(1<<30))
	memoryHours := new(big.Float).Quo(ms, big.NewFloat(3600))
	memoryCosts := new(big.Float).Mul(memoryHours, big.NewFloat(memoryPerGiAndHour))
	return fmt.Sprintf(" (%.2f €)", memoryCosts)
}
