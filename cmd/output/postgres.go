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
	PostgresBackupEntryTablePrinter struct {
		TablePrinter
	}
)

func (p PostgresTablePrinter) Print(data []*models.V1PostgresResponse) {
	p.shortHeader = []string{"ID", "Description", "Partition", "Tenant", "Project", "CPU", "Buffer", "Storage", "Backup", "Replica", "Version", "Age", "Status"}
	p.wideHeader = []string{"ID", "Description", "Partition", "Tenant", "Project", "CPU", "Buffer", "Storage", "Backup", "Replica", "Version", "Address", "Age", "Status", "Maintenance", "Labels"}

	for _, pg := range data {
		id := ""
		if pg.ID != nil {
			id = *pg.ID
		}
		description := pg.Description
		partition := pg.Partition
		project := pg.Project
		tenant := pg.Tenant
		backup := pg.Backup

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
		maint := strings.Join(pg.Maintenance, "\n")

		replica := fmt.Sprintf("%d", pg.NumberOfInstances)
		short := []string{id, description, partition, tenant, project, cpu, buffer, storage, backup, replica, pg.Version, age, status}
		wide := []string{id, description, partition, tenant, project, cpu, buffer, storage, backup, replica, pg.Version, address, age, status, maint, lbls}

		p.addWideData(wide, pg)
		p.addShortData(short, pg)
	}
	p.render()
}
func (p PostgresVersionsTablePrinter) Print(data []*models.V1PostgresVersion) {
	p.wideHeader = []string{"Version", "ExpirationDate"}
	p.shortHeader = p.wideHeader

	for _, pg := range data {

		exp := "never"
		if !time.Time(pg.ExpirationDate).IsZero() {
			exp = pg.ExpirationDate.String()
		}
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
func (p PostgresBackupsTablePrinter) Print(data []*models.V1BackupResponse) {
	p.wideHeader = []string{"ID", "Name", "Project", "Schedule", "Retention", "S3", "CreatedBy"}
	p.shortHeader = p.wideHeader

	for _, b := range data {
		createdBy := ""
		if b.CreatedBy != nil {
			createdBy = *b.CreatedBy
		}
		wide := []string{*b.ID, b.Name, b.ProjectID, b.Schedule, fmt.Sprintf("%d", b.Retention), b.S3Endpoint + "/" + b.S3BucketName, createdBy}
		short := wide

		p.addWideData(wide, b)
		p.addShortData(short, b)
	}
	p.render()
}
func (p PostgresBackupEntryTablePrinter) Print(data []*models.V1BackupEntry) {
	p.wideHeader = []string{"Date", "Size", "Name"}
	p.shortHeader = p.wideHeader
	p.Order(data)
	for _, b := range data {
		wide := []string{b.Timestamp.String(), helper.HumanizeSize(*b.Size), *b.Name}
		short := wide

		p.addWideData(wide, b)
		p.addShortData(short, b)
	}
	p.render()
}
