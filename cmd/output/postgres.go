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
	p.wideHeader = []string{"ID", "Name", "Partition", "Tenant", "Project", "Replica", "Version", "Status"}
	p.shortHeader = p.wideHeader

	for _, pg := range data {
		id := ""
		if pg.ID != nil {
			id = *pg.ID
		}
		name := ""
		if pg.Name != nil {
			id = *pg.Name
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
		replica := fmt.Sprintf("%d", pg.NumberOfInstances)
		wide := []string{id, name, partition, tenant, project, replica, pg.Version, status}
		short := wide

		p.addWideData(wide, pg)
		p.addShortData(short, pg)
	}
	p.render()
}
