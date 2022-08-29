package tableprinters

import (
	"strings"

	"github.com/fi-ts/cloud-go/api/models"
)

func (t *TablePrinter) ProjectTable(data []*models.V1ProjectResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"UID", "Tenant", "Name", "Description", "Labels", "Annotations"}

	for _, pr := range data {
		labels := strings.Join(pr.Meta.Labels, "\n")
		as := []string{}
		for k, v := range pr.Meta.Annotations {
			as = append(as, k+"="+v)
		}
		annotations := strings.Join(as, "\n")

		rows = append(rows, []string{pr.Meta.ID, pr.TenantID, pr.Name, pr.Description, labels, annotations})
	}

	return header, rows, nil
}
