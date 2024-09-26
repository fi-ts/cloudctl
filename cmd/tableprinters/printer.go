package tableprinters

import (
	"io"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

type TablePrinter struct {
	t *printers.TablePrinter
	// TODO: we want to slowly migrate to the genericcli table printer
	// after everything was shifted to this package we can remove the "oldPrinter"
	oldPrinter printers.Printer
	out        io.Writer
}

func New(out io.Writer) *TablePrinter {
	return &TablePrinter{
		oldPrinter: output.New(),
		out:        out,
	}
}

func (t *TablePrinter) SetPrinter(printer *printers.TablePrinter) {
	t.t = printer
}

func (t *TablePrinter) ToHeaderAndRows(data any, wide bool) ([]string, [][]string, error) {
	t.t.WithOut(t.out)

	// TODO: migrate old output package code to here
	switch d := data.(type) {

	// project machine reservations
	case *models.V1MachineReservationResponse:
		return t.MachineReservationsTable(pointer.WrapInSlice(d), wide)
	case []*models.V1MachineReservationResponse:
		return t.MachineReservationsTable(d, wide)
	case *models.V1MachineReservationUsageResponse:
		return t.MachineReservationsUsageTable(pointer.WrapInSlice(d), wide)
	case []*models.V1MachineReservationUsageResponse:
		return t.MachineReservationsUsageTable(d, wide)
	case *models.V1MachineReservationBillingUsageResponse:
		return t.MachineReservationsBillingTable(d, wide)

	default:
		// fallback to old printer for as long as the migration takes:
		t.t.WithOut(io.Discard)
		return nil, nil, t.oldPrinter.Print(data)
	}
}
