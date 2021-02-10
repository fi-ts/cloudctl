package output

import (
	"fmt"

	"github.com/fi-ts/cloud-go/api/models"
)

type (
	// PostgresTablePrinter prints postgres databases in a table
	PostgresTablePrinter struct {
		TablePrinter
	}
)

func (p PostgresTablePrinter) Print(data []*models.V1PostgresResponse) {
	p.wideHeader = []string{"ID", "Description", "Partition", "Tenant", "Project", "CPU", "Buffer", "Storage", "Replica", "Version", "Status"}
	p.shortHeader = p.wideHeader

	for _, pg := range data {
		id := ""
		if pg.ID != nil {
			id = *pg.ID
		}
		description := ""
		if pg.Description != nil {
			description = *pg.Description
		}
		partition := ""
		if pg.PartitionID != nil {
			partition = *pg.PartitionID
		}
		project := ""
		if pg.ProjectID != nil {
			project = *pg.ProjectID
		}
		tenant := ""
		if pg.Tenant != nil {
			tenant = *pg.Tenant
		}
		status := ""
		if pg.Status != nil {
			status = pg.Status.Description
		}
		cpu := ""
		buffer := ""
		storage := ""
		if pg.Size != nil {
			cpu = pg.Size.CPU
			buffer = pg.Size.SharedBuffer
			storage = pg.Size.StorageSize
		}
		replica := fmt.Sprintf("%d", pg.NumberOfInstances)
		wide := []string{id, description, partition, tenant, project, cpu, buffer, storage, replica, pg.Version, status}
		short := wide

		p.addWideData(wide, pg)
		p.addShortData(short, pg)
	}
	p.render()
}
