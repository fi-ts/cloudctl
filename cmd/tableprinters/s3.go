package tableprinters

import (
	"github.com/fatih/color"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func (t *TablePrinter) S3Table(data []*models.V1S3CredentialsResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Tenant", "Project", "Partition", "Endpoint"}

	for _, user := range data {
		rows = append(rows, []string{
			pointer.SafeDeref(user.ID),
			pointer.SafeDeref(user.Tenant),
			pointer.SafeDeref(user.Project),
			pointer.SafeDeref(user.Partition),
			pointer.SafeDeref(user.Endpoint),
		})
	}

	return header, rows, nil
}

func (t *TablePrinter) S3PartitionsTable(data []*models.V1S3PartitionResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"Name", "Endpoint", "Ready"}

	for _, p := range data {
		ready := color.RedString(circle)
		if wide {
			ready = "false"
		}

		if p.Ready != nil && *p.Ready {
			ready = color.GreenString(circle)
			if wide {
				ready = "true"

			}
		}

		rows = append(rows, []string{
			pointer.SafeDeref(p.ID),
			pointer.SafeDeref(p.Endpoint),
			ready,
		})
	}

	return header, rows, nil
}
