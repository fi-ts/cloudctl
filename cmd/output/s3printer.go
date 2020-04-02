package output

import (
	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
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
	p.wideHeader = []string{"Name", "Tenant", "Partition"}
	p.shortHeader = p.wideHeader

	for _, user := range data {
		name := ""
		if user.Name != nil {
			name = *user.Name
		}

		tenant := ""
		if user.Tenant != nil {
			tenant = *user.Tenant
		}

		partition := ""
		if user.Partition != nil {
			partition = *user.Partition
		}

		wide := []string{name, tenant, partition}
		p.addWideData(wide, user)
		p.addShortData(wide, user)
	}
	p.render()
}

// Print a S3 partitions as table
func (p S3PartitionTablePrinter) Print(data []*models.V1S3PartitionResponse) {
	p.wideHeader = []string{"Name", "Endpoint"}
	p.shortHeader = p.wideHeader

	for _, partition := range data {
		name := ""
		if partition.Name != nil {
			name = *partition.Name
		}

		endpoint := ""
		if partition.Endpoint != nil {
			endpoint = *partition.Endpoint
		}

		wide := []string{name, endpoint}
		p.addWideData(wide, partition)
		p.addShortData(wide, partition)
	}
	p.render()
}
