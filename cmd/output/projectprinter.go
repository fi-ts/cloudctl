package output

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fi-ts/cloud-go/api/models"
)

type (
	// ProjectTablePrinter print a Project in a Table
	ProjectTablePrinter struct {
		TablePrinter
	}
)

// Print a Project as table
func (p ProjectTablePrinter) Print(data []*models.V1Project) {
	p.wideHeader = []string{"UID", "Tenant", "Name", "Description", "Clusters", "Machines", "IPs", "Labels", "Annotations"}
	p.shortHeader = p.wideHeader
	p.Order(data)
	for _, pr := range data {
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
		labels := strings.Join(pr.Meta.Labels, "\n")
		as := []string{}
		for k, v := range pr.Meta.Annotations {
			as = append(as, k+"="+v)
		}
		annotations := strings.Join(as, "\n")

		wide := []string{pr.Meta.ID, pr.TenantID, pr.Name, pr.Description, clusterQuota, machineQuota, ipQuota, labels, annotations}
		p.addWideData(wide, pr)
		p.addShortData(wide, pr)
	}
	p.render()
}
