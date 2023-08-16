package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/spf13/viper"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

type (
	// Printer main Interface for implementations which spits out to specified Writer
	Printer interface {
		Print(data interface{}) error
		Type() string
	}
	tablePrinter struct {
		table       *tablewriter.Table
		format      string
		wide        bool
		order       string
		noHeaders   bool
		template    *template.Template
		shortHeader []string
		wideHeader  []string
		shortData   [][]string
		wideData    [][]string
		outWriter   io.Writer
	}
	// jsonPrinter returns the model in json format
	jsonPrinter struct {
		outWriter io.Writer
	}
	// yamlPrinter returns the model in yaml format
	yamlPrinter struct {
		outWriter io.Writer
	}
)

// render the table shortHeader and shortData are always expected.
func (t *tablePrinter) render() {
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
func (t *tablePrinter) addShortData(row []string, data interface{}) {
	if t.wide {
		return
	}
	t.shortData = append(t.shortData, t.rowOrTemplate(row, data))
}
func (t *tablePrinter) addWideData(row []string, data interface{}) {
	if !t.wide {
		return
	}
	t.wideData = append(t.wideData, t.rowOrTemplate(row, data))
}

// rowOrTemplate return either given row or the data rendered with the given template, depending if template is set.
func (t *tablePrinter) rowOrTemplate(row []string, data interface{}) []string {
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

// New returns a suitable stdout printer for the given format
func New() Printer {
	printer, err := newPrinter(
		viper.GetString("output-format"),
		viper.GetString("order"),
		viper.GetString("template"),
		viper.GetBool("no-headers"),
		os.Stdout,
	)
	if err != nil {
		log.Fatalf("unable to initialize printer:%v", err)
	}
	return printer
}

// newPrinter returns a suitable stdout printer for the given format
func newPrinter(format, order, tpl string, noHeaders bool, writer io.Writer) (Printer, error) {
	if format == "" {
		format = "table"
	}
	var printer Printer
	switch format {
	case "yaml":
		printer = &yamlPrinter{
			outWriter: writer,
		}
	case "json":
		printer = &jsonPrinter{
			outWriter: writer,
		}
	case "table", "wide":
		printer = newTablePrinter(format, order, noHeaders, nil, writer)
	case "template":
		tmpl, err := template.New("t").Funcs(sprig.TxtFuncMap()).Parse(tpl)
		if err != nil {
			return nil, fmt.Errorf("template invalid:%w", err)
		}
		printer = newTablePrinter(format, order, true, tmpl, writer)
	default:
		return nil, fmt.Errorf("unknown format:%s", format)
	}
	return printer, nil
}

func newTablePrinter(format, order string, noHeaders bool, template *template.Template, writer io.Writer) tablePrinter {
	tp := tablePrinter{
		format:    format,
		wide:      false,
		order:     order,
		noHeaders: noHeaders,
		outWriter: writer,
	}
	if format == "wide" {
		tp.wide = true
	}
	table := tablewriter.NewWriter(writer)
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

func (t tablePrinter) Type() string {
	return "table"
}

// Print a model in a human readable table
func (t tablePrinter) Print(data interface{}) error {
	switch d := data.(type) {
	case *models.V1AuditResponse:
		AuditTablePrinter{t}.Print([]*models.V1AuditResponse{d})
	case []*models.V1AuditResponse:
		AuditTablePrinter{t}.Print(d)
	case *models.V1ClusterResponse:
		ShootTablePrinter{t}.Print([]*models.V1ClusterResponse{d})
	case []*models.V1ClusterResponse:
		ShootTablePrinter{t}.Print(d)
	case ShootIssuesResponse:
		ShootIssuesTablePrinter{t}.Print([]*models.V1ClusterResponse{d})
	case ShootIssuesResponses:
		ShootIssuesTablePrinter{t}.Print(d)
	case []*models.V1beta1Condition:
		ShootConditionsTablePrinter{t}.Print(d)
	case []*models.V1beta1LastError:
		ShootLastErrorsTablePrinter{t}.Print(d)
	case *models.V1beta1LastOperation:
		ShootLastOperationTablePrinter{t}.Print(d)
	case *models.V1ProjectResponse:
		ProjectTableDetailPrinter{t}.Print(d)
	case []*models.V1ProjectResponse:
		ProjectTablePrinter{t}.Print(d)
	case []*models.V1TenantResponse:
		TenantTablePrinter{t}.Print(d)
	case *models.V1TenantResponse:
		TenantTablePrinter{t}.Print([]*models.V1TenantResponse{d})
	case *models.RestHealthResponse:
		HealthTablePrinter{t}.Print(d)
	case map[string]models.RestHealthResult:
		HealthTablePrinter{t}.PrintServices(d)
	case []*models.ModelsV1IPResponse:
		IPTablePrinter{t}.Print(d)
	case *models.ModelsV1IPResponse:
		IPTablePrinter{t}.Print([]*models.ModelsV1IPResponse{d})
	case []*models.V1ProjectInfoResponse:
		ProjectBillingTablePrinter{t}.Print(d)
	case *models.V1ContainerUsageResponse:
		ContainerBillingTablePrinter{t}.Print(d)
	case *models.V1ClusterUsageResponse:
		ClusterBillingTablePrinter{t}.Print(d)
	case *models.V1IPUsageResponse:
		IPBillingTablePrinter{t}.Print(d)
	case *models.V1NetworkUsageResponse:
		NetworkTrafficBillingTablePrinter{t}.Print(d)
	case *models.V1S3UsageResponse:
		S3BillingTablePrinter{t}.Print(d)
	case *models.V1VolumeUsageResponse:
		VolumeBillingTablePrinter{t}.Print(d)
	case *models.V1PostgresUsageResponse:
		PostgresBillingTablePrinter{t}.Print(d)
	case []*models.ModelsV1MachineResponse:
		MachineTablePrinter{t}.Print(d)
	case []*models.V1S3Response:
		S3TablePrinter{t}.Print(d)
	case *models.V1VolumeResponse:
		VolumeTablePrinter{t}.Print([]*models.V1VolumeResponse{d})
	case []*models.V1VolumeResponse:
		VolumeTablePrinter{t}.Print(d)
	case []*models.V1SnapshotResponse:
		SnapshotTablePrinter{t}.Print(d)
	case *models.V1SnapshotResponse:
		SnapshotTablePrinter{t}.Print(pointer.WrapInSlice(d))
	case []*models.V1StorageClusterInfo:
		VolumeClusterInfoTablePrinter{t}.Print(d)
	case models.V1PostgresPartitionsResponse:
		PostgresPartitionsTablePrinter{t}.Print(d)
	case []*models.V1PostgresVersion:
		PostgresVersionsTablePrinter{t}.Print(d)
	case *models.V1PostgresResponse:
		PostgresTablePrinter{t}.Print([]*models.V1PostgresResponse{d})
	case []*models.V1PostgresResponse:
		PostgresTablePrinter{t}.Print(d)
	case []*models.V1PostgresBackupConfigResponse:
		PostgresBackupsTablePrinter{t}.Print(d)
	case *models.V1PostgresBackupConfigResponse:
		PostgresBackupsTablePrinter{t}.Print([]*models.V1PostgresBackupConfigResponse{d})
	case []*models.V1PostgresBackupEntry:
		PostgresBackupEntryTablePrinter{t}.Print(d)
	case []*models.V1S3PartitionResponse:
		S3PartitionTablePrinter{t}.Print(d)
	case *models.V1ClusterMonitoringSecretResponse:
		return yamlPrinter{
			outWriter: t.outWriter,
		}.Print(d)
	case *models.V1S3CredentialsResponse, *models.V1S3Response:
		return yamlPrinter{
			outWriter: t.outWriter,
		}.Print(d)
	case *api.Contexts:
		ContextPrinter{t}.Print(d)
	case api.Version:
		return yamlPrinter{
			outWriter: t.outWriter,
		}.Print(d)
	case *cluster.ListConstraintsOK:
		return yamlPrinter{
			outWriter: t.outWriter,
		}.Print(d)
	default:
		return fmt.Errorf("unknown table printer for type: %T", d)
	}
	return nil
}

// Print a model in json format
func (j jsonPrinter) Print(data interface{}) error {
	json, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("unable to marshal to json:%w", err)
	}
	fmt.Fprintf(j.outWriter, "%s\n", string(json))
	return nil
}

func (j jsonPrinter) Type() string {
	return "json"
}

// Print a model in yaml format
func (y yamlPrinter) Print(data interface{}) error {
	yml, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to marshal to yaml:%w", err)
	}
	fmt.Fprintf(y.outWriter, "%s", string(yml))
	return nil
}

func (y yamlPrinter) Type() string {
	return "yaml"
}
