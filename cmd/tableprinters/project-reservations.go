package tableprinters

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
)

func (t *TablePrinter) MachineReservationsTable(data []*models.V1MachineReservationResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Tenant", "Project", "Size", "Amount", "Partitions", "Description"}
		rows   [][]string
	)

	if wide {
		header = append(header, "Labels")
	}

	for _, rv := range data {
		sort.Strings(rv.Partitionids)

		row := []string{
			pointer.SafeDeref(rv.ID),
			pointer.SafeDeref(rv.Tenant),
			pointer.SafeDeref(rv.Projectid),
			pointer.SafeDeref(rv.Sizeid),
			strconv.Itoa(int(pointer.SafeDeref(rv.Amount))),
			strings.Join(rv.Partitionids, ","),
			genericcli.TruncateEnd(rv.Description, 50),
		}

		if wide {
			var labels []string
			for k, v := range rv.Labels {
				labels = append(labels, k+"="+v)
			}
			sort.Strings(labels)

			row = append(row, rv.Description, strings.Join(labels, "\n"))
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
		header = []string{"ID", "Tenant", "Project", "Partition", "Size", "Reservations"}
		rows   [][]string
	)

	if wide {
		header = append(header, "Allocations", "Labels")
	}

	for _, rv := range data {
		reservations := "0"
		if pointer.SafeDeref(rv.Reservations) > 0 {
			unused := pointer.SafeDeref(rv.Reservations) - pointer.SafeDeref(rv.Usedreservations)
			reservations = fmt.Sprintf("%d (%d/%d used)", unused, pointer.SafeDeref(rv.Usedreservations), pointer.SafeDeref(rv.Reservations))
		}

		row := []string{
			pointer.SafeDeref(rv.ID),
			pointer.SafeDeref(rv.Tenant),
			pointer.SafeDeref(rv.Projectid),
			pointer.SafeDeref(rv.Partitionid),
			pointer.SafeDeref(rv.Sizeid),
			reservations,
		}

		if wide {
			row = append(row, fmt.Sprintf("%d", pointer.SafeDeref(rv.Projectallocations)))

			labels := []string{}
			for k, v := range rv.Labels {
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

func (t *TablePrinter) MachineReservationsBillingTable(data *models.V1MachineReservationBillingUsageResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Tenant", "From", "To", "ProjectID", "ProjectName", "Partition", "Size", "ID", "Reservations * Time", "Average"}
		rows   [][]string
	)

	for _, rv := range data.Usage {
		row := []string{
			pointer.SafeDeref(rv.Tenant),
			time.Time(pointer.SafeDeref(data.From)).String(),
			time.Time(data.To).String(),
			pointer.SafeDeref(rv.Projectid),
			pointer.SafeDeref(rv.Projectname),
			pointer.SafeDeref(rv.Partition),
			pointer.SafeDeref(rv.Sizeid),
			pointer.SafeDeref(rv.ID),
			humanizeSeconds(pointer.SafeDeref(rv.Reservationseconds)),
			pointer.SafeDeref(rv.Average),
		}

		rows = append(rows, row)
	}

	rows = append(rows, []string{"Total", "", "", "", "", "", "", "",
		humanizeSeconds(pointer.SafeDeref(data.Accumulatedusage.Reservationseconds)) + secondsCosts(pointer.SafeDeref(data.Accumulatedusage.Reservationseconds)),
		pointer.SafeDeref(data.Accumulatedusage.Average),
	})

	return header, rows, nil
}

func humanizeSeconds(seconds string) string {
	duration, err := strconv.ParseInt(seconds, 10, 64)
	if err == nil {
		return helper.HumanizeDuration(time.Duration(duration) * time.Second)
	}
	return ""
}

func secondsCosts(seconds string) string {
	costsPerHour := viper.GetFloat64("costs-hour")
	if costsPerHour <= 0 {
		return ""
	}
	duration, err := strconv.ParseInt(seconds, 10, 64)
	if err == nil {
		return fmt.Sprintf(" (%.2f €)", float64(duration/3600)*costsPerHour)
	}
	return ""
}
