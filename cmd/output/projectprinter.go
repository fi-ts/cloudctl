package output

import (
	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
)

type (
	// ProjectTablePrinter print a Project in a Table
	ProjectTablePrinter struct {
		TablePrinter
	}
)

// Print a Project as table
func (p ProjectTablePrinter) Print(data *models.V1ProjectListResponse) {
	p.wideHeader = []string{"UID", "Tenant", "Name", "Description"}
	p.shortHeader = p.wideHeader

	for _, pr := range data.Projects {
		wide := []string{pr.Meta.ID, pr.TenantID, pr.Name, pr.Description}

		p.addWideData(wide, pr)
		p.addShortData(wide, pr)
	}
	p.render()
}
