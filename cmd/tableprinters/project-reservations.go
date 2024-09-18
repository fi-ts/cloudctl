package tableprinters

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/olekukonko/tablewriter"
)

func (t *TablePrinter) MachineReservationsTable(data []*models.V1MachineReservationResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Tenant", "Project", "Size", "Amount", "Partitions"}
		rows   [][]string
	)

	if wide {
		header = append(header, "Description", "Labels")
	}

	for _, p := range data {
		row := []string{
			pointer.SafeDeref(p.Tenant),
			pointer.SafeDeref(p.Projectid),
			pointer.SafeDeref(p.Sizeid),
			strconv.Itoa(int(pointer.SafeDeref(p.Amount))),
			strings.Join(p.Partitionids, ","),
		}

		if wide {
			labels := []string{}
			for k, v := range p.Labels {
				labels = append(labels, k+"="+v)
			}
			lbls := strings.Join(labels, "\n")

			row = append(row, genericcli.TruncateEnd(p.Description, 50), lbls)
		}

		rows = append(rows, row)
	}

	t.t.MutateTable(func(table *tablewriter.Table) {
		table.SetAutoWrapText(false)
	})

	return header, rows, nil
}

func (t *TablePrinter) MachineReservationsUsageTable(data []*models.V1MachineReservationUsageResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Tenant", "Project", "Partition", "Size", "Reservations"}
		rows   [][]string
	)

	if wide {
		header = append(header, "Labels")
	}

	for _, p := range data {
		reservations := "0"
		if pointer.SafeDeref(p.Reservations) > 0 {
			unused := pointer.SafeDeref(p.Reservations) - pointer.SafeDeref(p.Usedreservations)
			reservations = fmt.Sprintf("%d (%d/%d used)", unused, pointer.SafeDeref(p.Usedreservations), pointer.SafeDeref(p.Reservations))
		}

		row := []string{
			pointer.SafeDeref(p.Tenant),
			pointer.SafeDeref(p.Projectid),
			pointer.SafeDeref(p.Partitionid),
			pointer.SafeDeref(p.Sizeid),
			reservations,
		}

		if wide {
			labels := []string{}
			for k, v := range p.Labels {
				labels = append(labels, k+"="+v)
			}
			lbls := strings.Join(labels, "\n")

			row = append(row, lbls)
		}

		rows = append(rows, row)
	}

	t.t.MutateTable(func(table *tablewriter.Table) {
		table.SetAutoWrapText(false)
	})

	return header, rows, nil
}
