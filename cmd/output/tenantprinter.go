package output

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fi-ts/cloud-go/api/models"
)

type (
	// TenantTablePrinter print a Project in a Table
	TenantTablePrinter struct {
		TablePrinter
	}
)

// Print a Project as table
func (p TenantTablePrinter) Print(tenants []*models.V1TenantResponse) {
	p.wideHeader = []string{"ID", "Name", "Description", "Clusters", "Machines", "IPs", "Labels", "Annotations"}
	p.shortHeader = p.wideHeader
	for _, tenantResponse := range tenants {
		tenant := tenantResponse.Tenant
		clusterQuota := ""
		machineQuota := ""
		ipQuota := ""
		// FIXME add actual quotas ?
		if tenant.DefaultQuotas != nil {
			qs := tenant.DefaultQuotas
			if qs.Cluster != nil {
				cq := "∞"
				if qs.Cluster.Quota != 0 {
					cq = strconv.FormatInt(int64(qs.Cluster.Quota), 10)
				}
				clusterQuota = fmt.Sprintf("%d/%s", qs.Cluster.Used, cq)
			}
			if qs.Machine != nil {
				mq := "∞"
				if qs.Cluster.Quota != 0 {
					mq = strconv.FormatInt(int64(qs.Machine.Quota), 10)
				}
				machineQuota = fmt.Sprintf("%d/%s", qs.Machine.Used, mq)
			}
			if qs.IP != nil {
				iq := "∞"
				if qs.IP.Quota != 0 {
					iq = strconv.FormatInt(int64(qs.IP.Quota), 10)
				}
				ipQuota = fmt.Sprintf("%d/%s", qs.IP.Used, iq)
			}
		}
		labels := strings.Join(tenant.Meta.Labels, "\n")
		as := []string{}
		for k, v := range tenant.Meta.Annotations {
			as = append(as, k+"="+v)
		}
		annotations := strings.Join(as, "\n")

		wide := []string{tenant.Meta.ID, tenant.Name, tenant.Description, clusterQuota, machineQuota, ipQuota, labels, annotations}
		p.addWideData(wide, tenant)
		p.addShortData(wide, tenant)
	}
	p.render()
}
