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
		tablePrinter
	}

	// ProjectTableDetailPrinter print a Project in a Table
	ProjectTableDetailPrinter struct {
		tablePrinter
	}
)

// Print a Project as table
func (p ProjectTablePrinter) Print(data []*models.V1ProjectResponse) {
	p.Order(data)
	p.wideHeader = []string{"UID", "Tenant", "Name", "Description", "Labels", "Annotations"}
	p.shortHeader = p.wideHeader
	if p.order == "" {
		p.order = "tenant,project"
	}
	for _, pr := range data {
		labels := strings.Join(pr.Meta.Labels, "\n")
		as := []string{}
		for k, v := range pr.Meta.Annotations {
			as = append(as, k+"="+v)
		}
		annotations := strings.Join(as, "\n")

		wide := []string{pr.Meta.ID, pr.TenantID, pr.Name, pr.Description, labels, annotations}
		p.addWideData(wide, pr)
		p.addShortData(wide, pr)
	}
	p.render()
}

// Print a Project as table
func (p ProjectTableDetailPrinter) Print(data *models.V1ProjectResponse) {
	p.wideHeader = []string{"UID", "Tenant", "Name", "Description", "Clusters", "Machines", "IPs", "Labels", "Annotations"}
	p.shortHeader = p.wideHeader
	if p.order == "" {
		p.order = "tenant,project"
	}

	clusterQuota := ""
	machineQuota := ""
	ipQuota := ""
	if data.Quotas != nil {
		qs := data.Quotas
		if qs.Cluster != nil {
			cq := "∞"
			if qs.Cluster.Quota != 0 {
				cq = strconv.FormatInt(int64(qs.Cluster.Quota), 10) // nolint:gosec
			}
			clusterQuota = fmt.Sprintf("%d/%s", qs.Cluster.Used, cq)
		}
		if qs.Machine != nil {
			mq := "∞"
			if qs.Machine.Quota != 0 {
				mq = strconv.FormatInt(int64(qs.Machine.Quota), 10) // nolint:gosec
			}
			machineQuota = fmt.Sprintf("%d/%s", qs.Machine.Used, mq)
		}
		if qs.IP != nil {
			iq := "∞"
			if qs.IP.Quota != 0 {
				iq = strconv.FormatInt(int64(qs.IP.Quota), 10) // nolint:gosec
			}
			ipQuota = fmt.Sprintf("%d/%s", qs.IP.Used, iq)
		}
	}
	labels := strings.Join(data.Meta.Labels, "\n")
	as := []string{}
	for k, v := range data.Meta.Annotations {
		as = append(as, k+"="+v)
	}
	annotations := strings.Join(as, "\n")

	wide := []string{data.Meta.ID, data.TenantID, data.Name, data.Description, clusterQuota, machineQuota, ipQuota, labels, annotations}
	p.addWideData(wide, data)
	p.addShortData(wide, data)

	p.render()
}
