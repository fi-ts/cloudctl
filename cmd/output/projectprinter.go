package output

import (
	"fmt"
	"strconv"

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
	p.wideHeader = []string{"UID", "Tenant", "Name", "Description", "Clusters", "Machines", "IPs"}
	p.shortHeader = p.wideHeader

	for _, pr := range data.Projects {
		clusterQuota := ""
		machineQuota := ""
		ipQuota := ""
		if pr.Quotas != nil {
			qs := pr.Quotas
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
		wide := []string{pr.Meta.ID, pr.TenantID, pr.Name, pr.Description, clusterQuota, machineQuota, ipQuota}
		p.addWideData(wide, pr)
		p.addShortData(wide, pr)
	}
	p.render()
}
