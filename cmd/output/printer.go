package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"text/template"

	"github.com/fatih/color"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/pkg/api"
	sprig "github.com/go-task/slim-sprig/v3"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/spf13/viper"

	"github.com/olekukonko/tablewriter"
)

type (
	// Printer main Interface for implementations which spits out to specified Writer
	Printer interface {
		Print(data interface{}) error
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
)

// render the table shortHeader and shortData are always expected.
func (t *tablePrinter) render() {
	if t.template == nil {
		if !t.noHeaders {
			if t.wide {
				t.table.Header(t.wideHeader)
			} else {
				t.table.Header(t.shortHeader)
			}
		}
		if t.wide {
			t.table.Append(t.wideData)
		} else {
			t.table.Append(t.shortData)
		}
		t.table.Render()
		t.table.Reset()
	} else {
		rows := t.shortData
		if t.wide {
			rows = t.wideData
		}
		for _, row := range rows {
			if len(row) < 1 {
				continue
			}
			if len(row[0]) == 0 {
				continue
			}
			fmt.Println(row[0])
		}
		t.shortData = [][]string{}
		t.wideData = [][]string{}
	}
	t.table.Reset()
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
		printer = printers.NewYAMLPrinter().WithOut(writer)
	case "json":
		printer = printers.NewJSONPrinter().WithOut(writer)
	case "table", "wide", "markdown":
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

	if viper.IsSet("force-color") {
		enabled := viper.GetBool("force-color")
		if enabled {
			color.NoColor = false
		} else {
			color.NoColor = true
		}
	}

	return printer, nil
}

func newTablePrinter(format, order string, noHeaders bool, template *template.Template, writer io.Writer) *tablePrinter {
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

	tp.table = table
	return &tp
}

func (t *tablePrinter) Type() string {
	return "table"
}

// Print a model in a human readable table
func (t *tablePrinter) Print(data interface{}) error {
	tp := *t
	switch d := data.(type) {
	case *models.V1AuditResponse:
		AuditTablePrinter{tp}.Print([]*models.V1AuditResponse{d})
	case []*models.V1AuditResponse:
		AuditTablePrinter{tp}.Print(d)
	case *models.V1ClusterResponse:
		ShootTablePrinter{tp}.Print([]*models.V1ClusterResponse{d})
	case []*models.V1ClusterResponse:
		ShootTablePrinter{tp}.Print(d)
	case ShootIssuesResponse:
		ShootIssuesTablePrinter{tp}.Print([]*models.V1ClusterResponse{d})
	case ShootIssuesResponses:
		ShootIssuesTablePrinter{tp}.Print(d)
	case []*models.V1beta1Condition:
		ShootConditionsTablePrinter{tp}.Print(d)
	case []*models.V1beta1LastError:
		ShootLastErrorsTablePrinter{tp}.Print(d)
	case *models.V1beta1LastOperation:
		ShootLastOperationTablePrinter{tp}.Print(d)
	case *models.V1ProjectResponse:
		ProjectTableDetailPrinter{tp}.Print(d)
	case []*models.V1ProjectResponse:
		ProjectTablePrinter{tp}.Print(d)
	case []*models.V1TenantResponse:
		TenantTablePrinter{tp}.Print(d)
	case *models.V1TenantResponse:
		TenantTablePrinter{tp}.Print([]*models.V1TenantResponse{d})
	case *models.RestHealthResponse:
		HealthTablePrinter{tp}.Print(d)
	case map[string]models.RestHealthResponse:
		HealthTablePrinter{tp}.PrintServices(d)
	case []*models.ModelsV1IPResponse:
		IPTablePrinter{tp}.Print(d)
	case *models.ModelsV1IPResponse:
		IPTablePrinter{tp}.Print([]*models.ModelsV1IPResponse{d})
	case []*models.V1ProjectInfoResponse:
		ProjectBillingTablePrinter{tp}.Print(d)
	case *models.V1ContainerUsageResponse:
		ContainerBillingTablePrinter{tp}.Print(d)
	case *models.V1ClusterUsageResponse:
		ClusterBillingTablePrinter{tp}.Print(d)
	case *models.V1MachineUsageResponse:
		MachineBillingTablePrinter{tp}.Print(d)
	case *models.V1ProductOptionUsageResponse:
		ProductOptionBillingTablePrinter{tp}.Print(d)
	case *models.V1IPUsageResponse:
		IPBillingTablePrinter{tp}.Print(d)
	case *models.V1NetworkUsageResponse:
		NetworkTrafficBillingTablePrinter{tp}.Print(d)
	case *models.V1S3UsageResponse:
		S3BillingTablePrinter{tp}.Print(d)
	case *models.V1VolumeUsageResponse:
		VolumeBillingTablePrinter{tp}.Print(d)
	case *models.V1PostgresUsageResponse:
		PostgresBillingTablePrinter{tp}.Print(d)
	case []*models.ModelsV1MachineResponse:
		MachineTablePrinter{tp}.Print(d)
	case []*models.V1S3Response:
		S3TablePrinter{tp}.Print(d)
	case *models.V1VolumeResponse:
		VolumeTablePrinter{tp}.Print([]*models.V1VolumeResponse{d})
	case []*models.V1VolumeResponse:
		VolumeTablePrinter{tp}.Print(d)
	case []*models.V1SnapshotResponse:
		SnapshotTablePrinter{tp}.Print(d)
	case *models.V1SnapshotResponse:
		SnapshotTablePrinter{tp}.Print(pointer.WrapInSlice(d))
	case []*models.V1QoSPolicyResponse:
		QoSPolicyTablePrinter{tp}.Print(d)
	case *models.V1QoSPolicyResponse:
		QoSPolicyTablePrinter{tp}.Print(pointer.WrapInSlice(d))
	case []*models.V1StorageClusterInfo:
		VolumeClusterInfoTablePrinter{tp}.Print(d)
	case models.V1PostgresPartitionsResponse:
		PostgresPartitionsTablePrinter{tp}.Print(d)
	case []*models.V1PostgresVersion:
		PostgresVersionsTablePrinter{tp}.Print(d)
	case *models.V1PostgresResponse:
		PostgresTablePrinter{tp}.Print([]*models.V1PostgresResponse{d})
	case []*models.V1PostgresResponse:
		PostgresTablePrinter{tp}.Print(d)
	case []*models.V1PostgresBackupConfigResponse:
		PostgresBackupsTablePrinter{tp}.Print(d)
	case *models.V1PostgresBackupConfigResponse:
		PostgresBackupsTablePrinter{tp}.Print([]*models.V1PostgresBackupConfigResponse{d})
	case []*models.V1PostgresBackupEntry:
		PostgresBackupEntryTablePrinter{tp}.Print(d)
	case []*models.V1S3PartitionResponse:
		S3PartitionTablePrinter{tp}.Print(d)
	case *api.Contexts:
		ContextPrinter{tp}.Print(d)
	default:
		return fmt.Errorf("unknown table printer for type: %T", d)
	}
	return nil
}
