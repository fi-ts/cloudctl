package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

type (
	// Printer main Interface for implementations which spits out to stdout
	Printer interface {
		Print(data interface{}) error
	}
	TablePrinter struct {
		table       *tablewriter.Table
		wide        bool
		order       string
		noHeaders   bool
		template    *template.Template
		shortHeader []string
		wideHeader  []string
		shortData   [][]string
		wideData    [][]string
	}
	// JSONPrinter returns the model in json format
	JSONPrinter struct{}
	// YAMLPrinter returns the model in yaml format
	YAMLPrinter struct{}
	// TablePrinter produces a human readable model representation
)

// render the table shortHeader and shortData are always expected.
func (t *TablePrinter) render() {
	if t.template == nil {
		if !t.noHeaders {
			if t.wide {
				t.table.SetHeader(t.wideHeader)
			} else {
				t.table.SetHeader(t.shortHeader)
			}
		}
		if t.wide {
			t.table.AppendBulk(t.wideData)
		} else {
			t.table.AppendBulk(t.shortData)
		}
		t.table.Render()
		t.table.ClearRows()
	} else {
		rows := t.shortData
		if t.wide {
			rows = t.wideData
		}
		for _, row := range rows {
			if len(row) < 1 {
				continue
			}
			fmt.Println(row[0])
		}
		t.shortData = [][]string{}
		t.wideData = [][]string{}
	}
}
func (t *TablePrinter) addShortData(row []string, data interface{}) {
	if t.wide {
		return
	}
	t.shortData = append(t.shortData, t.rowOrTemplate(row, data))
}
func (t *TablePrinter) addWideData(row []string, data interface{}) {
	if !t.wide {
		return
	}
	t.wideData = append(t.wideData, t.rowOrTemplate(row, data))
}

// rowOrTemplate return either given row or the data rendered with the given template, depending if template is set.
func (t *TablePrinter) rowOrTemplate(row []string, data interface{}) []string {
	tpl := t.template
	if tpl != nil {
		var buf bytes.Buffer
		err := tpl.Execute(&buf, genericObject(data))
		if err != nil {
			fmt.Printf("unable to parse template:%v", err)
			os.Exit(1)
		}
		return []string{buf.String()}
	}
	return row
}

// genericObject transforms the input to a struct which has fields with the same name as in the json struct.
// this is handy for template rendering as the output of -o json|yaml can be used as the input for the template
func genericObject(input interface{}) map[string]interface{} {
	b, err := json.Marshal(input)
	if err != nil {
		fmt.Printf("unable to marshall input:%v", err)
		os.Exit(1)
	}
	var result interface{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		fmt.Printf("unable to unmarshal input:%v", err)
		os.Exit(1)
	}
	return result.(map[string]interface{})

}

// NewPrinter returns a suitable stdout printer for the given format
func NewPrinter(format, order, tpl string, noHeaders bool) (Printer, error) {
	var printer Printer
	switch format {
	case "yaml":
		printer = &YAMLPrinter{}
	case "json":
		printer = &JSONPrinter{}
	case "table", "wide":
		printer = newTablePrinter(format, order, noHeaders, nil)
	case "template":
		tmpl, err := template.New("").Parse(tpl)
		if err != nil {
			return nil, fmt.Errorf("template invalid:%v", err)
		}
		printer = newTablePrinter(format, order, true, tmpl)
	default:
		return nil, fmt.Errorf("unknown format:%s", format)
	}
	return printer, nil
}

func newTablePrinter(format, order string, noHeaders bool, template *template.Template) TablePrinter {
	table := tablewriter.NewWriter(os.Stdout)
	wide := false
	if format == "wide" {
		wide = true
	}
	switch format {
	case "markdown":
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
	default:
		table.SetHeaderLine(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetBorder(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetRowLine(false)
	}
	return TablePrinter{
		table:     table,
		wide:      wide,
		order:     order,
		noHeaders: noHeaders,
		template:  template,
	}
}

// Print a model in a human readable table
func (t TablePrinter) Print(data interface{}) error {
	switch d := data.(type) {
	case *models.V1beta1Shoot:
		ShootTablePrinter{t}.Print([]*models.V1beta1Shoot{d})
	case []*models.V1beta1Shoot:
		ShootTablePrinter{t}.Print(d)
	case *models.ModelsV1ProjectResponse:
		ProjectTablePrinter{t}.Print([]*models.ModelsV1ProjectResponse{d})
	case []*models.ModelsV1ProjectResponse:
		ProjectTablePrinter{t}.Print(d)
	case *models.V1ContainerUsageResponse:
		if t.order == "" {
			t.order = "tenant,project,partition,cluster,namespace,pod,container"
		}
		BillingTablePrinter{t}.Print(d)
	default:
		return fmt.Errorf("unknown table printer for type: %T", d)
	}
	return nil
}

// Print a model in json format
func (j JSONPrinter) Print(data interface{}) error {
	json, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("unable to marshal to json:%v", err)
	}
	fmt.Printf("%s\n", string(json))
	return nil
}

// Print a model in yaml format
func (y YAMLPrinter) Print(data interface{}) error {
	yml, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to marshal to yaml:%v", err)
	}
	fmt.Printf("%s\n", string(yml))
	return nil
}
