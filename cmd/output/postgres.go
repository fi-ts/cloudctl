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

	PostgresBackupsTablePrinter struct {
		TablePrinter
	}
)

func (p PostgresTablePrinter) Print(data []*models.V1PostgresResponse) {
	p.shortHeader = []string{"ID", "Description", "Partition", "Tenant", "Project", "CPU", "Buffer", "Storage", "Replica", "Version", "Age", "Status"}
	p.wideHeader = []string{"ID", "Description", "Partition", "Tenant", "Project", "CPU", "Buffer", "Storage", "Replica", "Version", "Address", "Age", "Status", "Labels"}

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
		age := helper.HumanizeDuration(time.Since(time.Time(pg.CreationTimestamp)))
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
		address := ""
		if pg.Status.Socket != nil {
			address = fmt.Sprintf("%s:%d", pg.Status.Socket.IP, pg.Status.Socket.Port)
		}
		labels := []string{}
		for k, v := range pg.Labels {
			labels = append(labels, k+"="+v)
		}
		lbls := strings.Join(labels, "\n")

		replica := fmt.Sprintf("%d", pg.NumberOfInstances)
		short := []string{id, description, partition, tenant, project, cpu, buffer, storage, replica, pg.Version, age, status}
		wide := []string{id, description, partition, tenant, project, cpu, buffer, storage, replica, pg.Version, address, age, status, lbls}

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
func (p PostgresBackupsTablePrinter) Print(data []*models.V1Backup) {
	p.wideHeader = []string{"Project", "Schedule", "Retention", "S3"}
	p.shortHeader = p.wideHeader

	for _, b := range data {
		wide := []string{b.ProjectID, b.Schedule, fmt.Sprintf("%d", b.Retention), b.S3Endpoint + "/" + b.S3BucketName}
		short := wide

		p.addWideData(wide, b)
		p.addShortData(short, b)
	}
	p.render()
}
