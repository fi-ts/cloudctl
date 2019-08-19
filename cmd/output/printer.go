package cmd

type (
	// Printer main Interface for implementations which spits out to stdout
	Printer interface {
		Print(data interface{}) error
	}
)

// NewPrinter returns a suitable stdout printer for the given format
func NewPrinter(format, order, tpl string, noHeaders bool) (Printer, error) {

	return nil, nil
}
