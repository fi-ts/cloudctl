package output

import (
	"github.com/fatih/color"
	"github.com/metal-stack/cloud-go/api/models"
)

type (
	// S3TablePrinter print S3 storage in a Table
	S3TablePrinter struct {
		TablePrinter
	}

	S3PartitionTablePrinter struct {
		TablePrinter
	}
)

// Print a S3 storage as table
func (p S3TablePrinter) Print(data []*models.V1S3Response) {
	p.wideHeader = []string{"ID", "Tenant", "Project", "Partition", "Endpoint"}
	p.shortHeader = p.wideHeader

	for _, user := range data {
		name := ""
		if user.ID != nil {
			name = *user.ID
		}

		tenant := ""
		if user.Tenant != nil {
			tenant = *user.Tenant
		}

		project := ""
		if user.Project != nil {
			project = *user.Project
		}

		partition := ""
		if user.Partition != nil {
			partition = *user.Partition
		}

		endpoint := ""
		if user.Endpoint != nil {
			endpoint = *user.Endpoint
		}

		wide := []string{name, tenant, project, partition, endpoint}
		p.addWideData(wide, user)
		p.addShortData(wide, user)
	}
	p.render()
}

// Print a S3 partitions as table
func (p S3PartitionTablePrinter) Print(data []*models.V1S3PartitionResponse) {
	p.wideHeader = []string{"Name", "Endpoint", "Ready"}
	p.shortHeader = p.wideHeader
	p.Order(data)

	for _, partition := range data {
		name := ""
		if partition.ID != nil {
			name = *partition.ID
		}

		endpoint := ""
		if partition.Endpoint != nil {
			endpoint = *partition.Endpoint
		}

		ready := false
		if partition.Ready != nil {
			ready = *partition.Ready
		}

		readyStatus := color.RedString(circle)
		if ready {
			readyStatus = color.GreenString(circle)
		}
		wide := []string{name, endpoint, readyStatus}
		p.addWideData(wide, partition)
		p.addShortData(wide, partition)
	}
	p.render()
}
