package output

import (
	"github.com/gardener/gardener/pkg/apis/garden/v1beta1"
)

type (
	// ShootTablePrinter print a Shoot Cluster in a Table
	ShootTablePrinter struct {
		TablePrinter
	}
)

// Print a Shoot as table
func (s ShootTablePrinter) Print(data *v1beta1.Shoot) {
	// FIXME implement in a table
	yp := &YAMLPrinter{}
	yp.Print(data)
}
