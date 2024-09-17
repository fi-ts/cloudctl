package tableprinters

import (
	"strconv"
	"strings"

	"github.com/fi-ts/cloud-go/api/models"
)

func (t *TablePrinter) MachineReservationsTable(data []*models.V1MachineReservationResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Tenant", "Project", "Size", "Partitions", "Amount"}
		rows   [][]string
	)

	for _, p := range data {

		// labels, description

		row := []string{*p.Tenant, *p.Projectid, *p.Sizeid, strings.Join(p.Partitionids, ","), strconv.Itoa(int(*p.Amount))}

		rows = append(rows, row)
	}

	return header, rows, nil
}
