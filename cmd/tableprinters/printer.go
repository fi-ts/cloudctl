package tableprinters

import (
	"io"

	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
)

type TablePrinter struct {
	t *printers.TablePrinter
	// TODO: we want to slowly migrate to the genericcli table printer
	// after everything was shifted to this package we can remove the "oldPrinter"
	oldPrinter printers.Printer
}

func New() *TablePrinter {
	return &TablePrinter{
		oldPrinter: output.New(),
	}
}

func (t *TablePrinter) SetPrinter(printer *printers.TablePrinter) {
	t.t = printer
}

func (t *TablePrinter) ToHeaderAndRows(data any, wide bool) ([]string, [][]string, error) {
	// TODO: migrate old output package code to here
	// switch d := data.(type) {
	// default:
	// 	return nil, nil,  t.oldPrinter.Print(data)
	// }
	//
	// fallback to old printer for as long as the migration takes:
	t.t.WithOut(io.Discard)
	return nil, nil, t.oldPrinter.Print(data)
}
