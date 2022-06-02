package output

import (
	"strings"

	"github.com/fi-ts/cloud-go/api/models"
)

type (
	// TenantTablePrinter print a Project in a Table
	TenantTablePrinter struct {
		tablePrinter
	}
)

// Print a Project as table
func (p TenantTablePrinter) Print(tenants []*models.V1TenantResponse) {
	p.wideHeader = []string{"ID", "Name", "Description", "Labels", "Annotations"}
	p.shortHeader = p.wideHeader
	for _, tenantResponse := range tenants {
		tenant := tenantResponse

		labels := strings.Join(tenant.Meta.Labels, "\n")
		as := []string{}
		for k, v := range tenant.Meta.Annotations {
			as = append(as, k+"="+v)
		}
		annotations := strings.Join(as, "\n")

		wide := []string{tenant.Meta.ID, tenant.Name, tenant.Description, labels, annotations}
		p.addWideData(wide, tenant)
		p.addShortData(wide, tenant)
	}
	p.render()
}
