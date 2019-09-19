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
func (p ProjectTablePrinter) Print(data []*models.ModelsV1ProjectResponse) {
	p.wideHeader = []string{"UID", "Name", "Description"}
	p.shortHeader = p.wideHeader

	for _, pr := range data {
		wide := []string{*pr.ID, pr.Name, pr.Description}

		p.addWideData(wide, pr)
		p.addShortData(wide, pr)
	}
	p.render()
}
