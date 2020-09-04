package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/pkg/api"

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
	t.table.ClearRows()
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
	tp := TablePrinter{
		wide:      false,
		order:     order,
		noHeaders: noHeaders,
	}
	table := tablewriter.NewWriter(os.Stdout)
	if format == "wide" {
		tp.wide = true
	}
	switch format {
	case "template":
		tp.template = template
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
		table.SetTablePadding("\t") // pad with tabs
		table.SetNoWhiteSpace(true) // no whitespace in front of every line
	}
	tp.table = table
	return tp
}

// Print a model in a human readable table
func (t TablePrinter) Print(data interface{}) error {
	switch d := data.(type) {
	case *models.V1ClusterResponse:
		ShootTablePrinter{t}.Print([]*models.V1ClusterResponse{d})
	case []*models.V1ClusterResponse:
		if t.order == "" {
			t.order = "tenant,project,name"
		}
		ShootTablePrinter{t}.Print(d)
	case []*models.V1beta1Condition:
		ShootConditionsTablePrinter{t}.Print(d)
	case *models.V1ProjectResponse:
		ProjectTablePrinter{t}.Print([]*models.V1ProjectResponse{d})
	case []*models.V1ProjectResponse:
		if t.order == "" {
			t.order = "tenant,project"
		}
		ProjectTablePrinter{t}.Print(d)
	case []*models.V1Tenant:
		TenantTablePrinter{t}.Print(d)
	case *models.V1Tenant:
		TenantTablePrinter{t}.Print([]*models.V1Tenant{d})
	case []*models.ModelsV1IPResponse:
		IPTablePrinter{t}.Print(d)
	case *models.ModelsV1IPResponse:
		IPTablePrinter{t}.Print([]*models.ModelsV1IPResponse{d})
	case *models.V1ContainerUsageResponse:
		if t.order == "" {
			t.order = "tenant,project,partition,cluster,namespace,pod,container"
		}
		ContainerBillingTablePrinter{t}.Print(d)
	case *models.V1ClusterUsageResponse:
		ClusterBillingTablePrinter{t}.Print(d)
	case *models.V1IPUsageResponse:
		if t.order == "" {
			t.order = "tenant,project,ip"
		}
		IPBillingTablePrinter{t}.Print(d)
	case *models.V1NetworkUsageResponse:
		if t.order == "" {
			t.order = "tenant,project,partition,cluster,device"
		}
		NetworkTrafficBillingTablePrinter{t}.Print(d)
	case *models.V1S3UsageResponse:
		S3BillingTablePrinter{t}.Print(d)
	case *models.V1VolumeUsageResponse:
		VolumeBillingTablePrinter{t}.Print(d)
	case []*models.ModelsV1MachineResponse:
		MachineTablePrinter{t}.Print(d)
	case []*models.V1S3Response:
		S3TablePrinter{t}.Print(d)
	case []*models.V1S3PartitionResponse:
		if t.order == "" {
			t.order = "id"
		}
		S3PartitionTablePrinter{t}.Print(d)
	case *api.Contexts:
		ContextPrinter{t}.Print(d)
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
