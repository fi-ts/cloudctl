package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
)

type (
	// PostgresTablePrinter prints postgres databases in a table
	PostgresTablePrinter struct {
		TablePrinter
	}

	PostgresVersionsTablePrinter struct {
		TablePrinter
	}
	PostgresPartitionsTablePrinter struct {
		TablePrinter
	}
)

func (p PostgresTablePrinter) Print(data []*models.V1PostgresResponse) {
	p.wideHeader = []string{"ID", "Description", "Partition", "Tenant", "Project", "CPU", "Buffer", "Storage", "Replica", "Version", "Age", "Status"}
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
		age := ""
		if pg.CreationTimestamp != nil {
			age = helper.HumanizeDuration(time.Since(time.Time(*pg.CreationTimestamp)))
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
		wide := []string{id, description, partition, tenant, project, cpu, buffer, storage, replica, pg.Version, age, status}
		short := wide

		p.addWideData(wide, pg)
		p.addShortData(short, pg)
	}
	p.render()
}
func (p PostgresVersionsTablePrinter) Print(data []*models.V1PostgresVersion) {
	p.wideHeader = []string{"Version", "ExpirationDate"}
	p.shortHeader = p.wideHeader

	for _, pg := range data {

		exp := pg.ExpirationDate.String()
		wide := []string{pg.Version, exp}
		short := wide

		p.addWideData(wide, pg)
		p.addShortData(short, pg)
	}
	p.render()
}
func (p PostgresPartitionsTablePrinter) Print(data models.V1PostgresPartitionsResponse) {
	p.wideHeader = []string{"Name", "AllowedTenants"}
	p.shortHeader = p.wideHeader

	for name, pg := range data {
		tenants := []string{}
		if len(pg.AllowedTenants) == 0 {
			tenants = []string{"any"}
		}
		for k := range pg.AllowedTenants {
			tenants = append(tenants, k)
		}
		wide := []string{name, strings.Join(tenants, ",")}
		short := wide

		p.addWideData(wide, pg)
		p.addShortData(short, pg)
	}
	p.render()
}
