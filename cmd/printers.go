package cmd

import (
	"io"
	"log"

	"github.com/fatih/color"
	"github.com/fi-ts/cloudctl/cmd/tableprinters"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/spf13/viper"
)

func newPrinterFromCLI(out io.Writer) printers.Printer {
	var printer printers.Printer

	switch format := viper.GetString("output-format"); format {
	case "yaml":
		printer = printers.NewYAMLPrinter().WithOut(out)
	case "json":
		printer = printers.NewJSONPrinter().WithOut(out)
	case "table", "wide", "markdown":
		tp := tableprinters.New(out)

		tablePrinter := printers.NewTablePrinter(&printers.TablePrinterConfig{
			ToHeaderAndRows: tp.ToHeaderAndRows,
			Wide:            format == "wide",
			Markdown:        format == "markdown",
			NoHeaders:       viper.GetBool("no-headers"),
		}).WithOut(out)

		tp.SetPrinter(tablePrinter)

		printer = tablePrinter
	case "template":
		printer = printers.NewTemplatePrinter(viper.GetString("template")).WithOut(out)
	default:
		log.Fatalf("unknown output format: %q", format)
	}

	if viper.IsSet("force-color") {
		enabled := viper.GetBool("force-color")
		if enabled {
			color.NoColor = false
		} else {
			color.NoColor = true
		}
	}

	return printer
}

func defaultToYAMLPrinter(out io.Writer) printers.Printer {
	if viper.IsSet("output-format") {
		return newPrinterFromCLI(out)
	}
	return printers.NewYAMLPrinter().WithOut(out)
}
