package output

import (
	"github.com/fi-ts/cloudctl/pkg/api"
)

type (
	// ContextPrinter print a Context in a Table
	ContextPrinter struct {
		tablePrinter
	}
)

// Print a model in yaml format
func (p ContextPrinter) Print(data *api.Contexts) {
	for name, c := range data.Contexts {
		if name == data.CurrentContext {
			name = name + " [*]"
		}
		row := []string{name, c.ApiURL, c.IssuerURL}
		p.addShortData(row, c)
	}
	p.shortHeader = []string{"Name", "URL", "DEX"}
	p.render()
}
