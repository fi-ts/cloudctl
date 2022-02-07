package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fi-ts/cloud-go/api/client/accounting"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/go-openapi/strfmt"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/now"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/xuri/excelize/v2"
)

type BillingOpts struct {
	Tenant     string
	FromString string
	ToString   string
	From       time.Time
	To         time.Time
	ProjectID  string
	ClusterID  string
	Device     string
	Namespace  string
	Month      string
	Year       string
	Filename   string
	CSV        bool
}

var (
	billingOpts *BillingOpts
)

func newBillingCmd(c *config) *cobra.Command {
	billingCmd := &cobra.Command{
		Use:   "billing",
		Short: "manage bills",
		Long:  "TODO",
	}
	excelBillingCmd := &cobra.Command{
		Use:   "excel",
		Short: "create excel file with monthly billing information for all ressources",
		Example: `
		cloudctl billing excel
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.excel()
		},
		PreRun: bindPFlags,
	}
	containerBillingCmd := &cobra.Command{
		Use:   "container",
		Short: "look at container bills",
		Example: `If you want to get the costs in Euro, then set two environment variables with the prices from your contract:

		export CLOUDCTL_COSTS_CPU_HOUR=0.01        # Costs in Euro per CPU Hour
		export CLOUDCTL_COSTS_MEMORY_GI_HOUR=0.01  # Costs in Euro per Gi Memory Hour

		cloudctl billing container
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.containerUsage()
		},
		PreRun: bindPFlags,
	}
	clusterBillingCmd := &cobra.Command{
		Use:   "cluster",
		Short: "look at cluster bills",
		Example: `If you want to get the costs in Euro, then set two environment variables with the prices from your contract:

		cloudctl billing cluster
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.clusterUsage()
		},
		PreRun: bindPFlags,
	}
	ipBillingCmd := &cobra.Command{
		Use:   "ip",
		Short: "look at ip bills",
		Example: `If you want to get the costs in Euro, then set two environment variables with the prices from your contract:

		cloudctl billing ip
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.ipUsage()
		},
		PreRun: bindPFlags,
	}
	networkTrafficBillingCmd := &cobra.Command{
		Use:   "network-traffic",
		Short: "look at network traffic bills",
		Example: `If you want to get the costs in Euro, then set two environment variables with the prices from your contract:

		export CLOUDCTL_COSTS_INCOMING_NETWORK_TRAFFIC_GI=0.01        # Costs in Euro per gi
		export CLOUDCTL_COSTS_OUTGOING_NETWORK_TRAFFIC_GI=0.01        # Costs in Euro per gi
		export CLOUDCTL_COSTS_TOTAL_NETWORK_TRAFFIC_GI=0.01           # Costs in Euro per gi

		cloudctl billing network-traffic
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.networkTrafficUsage()
		},
		PreRun: bindPFlags,
	}
	s3BillingCmd := &cobra.Command{
		Use:   "s3",
		Short: "look at s3 bills",
		Example: `If you want to get the costs in Euro, then set two environment variables with the prices from your contract:

		export CLOUDCTL_COSTS_STORAGE_GI_HOUR=0.01        # Costs in Euro per storage Hour

		cloudctl billing s3
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.s3Usage()
		},
		PreRun: bindPFlags,
	}
	volumeBillingCmd := &cobra.Command{
		Use:   "volume",
		Short: "look at volume bills",
		Example: `If you want to get the costs in Euro, then set two environment variables with the prices from your contract:

		export CLOUDCTL_COSTS_CAPACITY_HOUR=0.01        # Costs in Euro per capacity Hour

		cloudctl billing volume
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.volumeUsage()
		},
		PreRun: bindPFlags,
	}
	postgresBillingCmd := &cobra.Command{
		Use:   "postgres",
		Short: "look at postgres bills",
		//TODO set costs via env var?
		Example: `If you want to get the costs in Euro, then set two environment variables with the prices from your contract:

		cloudctl billing postgres
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.postgresUsage()
		},
		PreRun: bindPFlags,
	}

	billingCmd.AddCommand(excelBillingCmd)
	billingCmd.AddCommand(containerBillingCmd)
	billingCmd.AddCommand(clusterBillingCmd)
	billingCmd.AddCommand(ipBillingCmd)
	billingCmd.AddCommand(networkTrafficBillingCmd)
	billingCmd.AddCommand(s3BillingCmd)
	billingCmd.AddCommand(volumeBillingCmd)
	billingCmd.AddCommand(postgresBillingCmd)

	billingOpts = &BillingOpts{}

	excelBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	excelBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	excelBillingCmd.Flags().StringVarP(&billingOpts.Month, "month", "m", "", "requested month")
	excelBillingCmd.Flags().StringVarP(&billingOpts.Year, "year", "y", "", "requested year")
	excelBillingCmd.Flags().StringVarP(&billingOpts.Filename, "file", "f", "", "excel filename")

	must(viper.BindPFlags(excelBillingCmd.Flags()))

	containerBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	containerBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	containerBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	containerBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	containerBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	containerBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster-id", "c", "", "the cluster to account")
	containerBillingCmd.Flags().StringVarP(&billingOpts.Namespace, "namespace", "n", "", "the namespace to account")
	containerBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	must(viper.BindPFlags(containerBillingCmd.Flags()))

	clusterBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	clusterBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster-id", "c", "", "the cluster to account")
	clusterBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	must(clusterBillingCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))

	must(viper.BindPFlags(clusterBillingCmd.Flags()))

	ipBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	ipBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	ipBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	ipBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	ipBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	ipBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	must(viper.BindPFlags(ipBillingCmd.Flags()))

	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	networkTrafficBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster-id", "c", "", "the cluster to account")
	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.Device, "device", "", "", "the device to account")
	networkTrafficBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	must(viper.BindPFlags(networkTrafficBillingCmd.Flags()))

	s3BillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	s3BillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	s3BillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	s3BillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	s3BillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	s3BillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	must(viper.BindPFlags(s3BillingCmd.Flags()))

	volumeBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	volumeBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	volumeBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	volumeBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	volumeBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	volumeBillingCmd.Flags().StringVarP(&billingOpts.Namespace, "namespace", "n", "", "the namespace to account")
	volumeBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster-id", "c", "", "the cluster to account")
	volumeBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	must(viper.BindPFlags(volumeBillingCmd.Flags()))

	postgresBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	postgresBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	postgresBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	postgresBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	postgresBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	postgresBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	must(viper.BindPFlags(postgresBillingCmd.Flags()))

	return billingCmd
}

func initBillingOpts() error {
	validate := validator.New()
	err := validate.Struct(billingOpts)
	if err != nil {
		return err
	}

	from := now.BeginningOfMonth()
	if billingOpts.FromString != "" {
		from, err = time.Parse(viper.GetString("time-format"), billingOpts.FromString)
		if err != nil {
			return err
		}
	}
	billingOpts.From = from

	to := time.Now()
	if billingOpts.ToString != "" {
		to, err = time.Parse(viper.GetString("time-format"), billingOpts.ToString)
		if err != nil {
			return err
		}
	}
	billingOpts.To = to

	return nil
}

func (c *config) clusterUsage() error {
	from := strfmt.DateTime(billingOpts.From)
	cur := models.V1ClusterUsageRequest{
		From: &from,
		To:   strfmt.DateTime(billingOpts.To),
	}
	if billingOpts.Tenant != "" {
		cur.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		cur.Projectid = billingOpts.ProjectID
	}
	if billingOpts.ClusterID != "" {
		cur.Clusterid = billingOpts.ClusterID
	}

	if billingOpts.CSV {
		return c.clusterUsageCSV(&cur)
	}
	return c.clusterUsageJSON(&cur)
}

func (c *config) clusterUsageJSON(cur *models.V1ClusterUsageRequest) error {
	request := accounting.NewClusterUsageParams()
	request.SetBody(cur)

	response, err := c.cloud.Accounting.ClusterUsage(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}

func (c *config) clusterUsageCSV(cur *models.V1ClusterUsageRequest) error {
	request := accounting.NewClusterUsageCSVParams()
	request.SetBody(cur)

	response, err := c.cloud.Accounting.ClusterUsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}

func (c *config) excel() error {
	month := int(time.Now().Month())
	year := int(time.Now().Year())

	var err error
	if billingOpts.Month != "" {
		if month, err = strconv.Atoi(billingOpts.Month); err != nil {
			return err
		}
	}
	if billingOpts.Year != "" {
		if year, err = strconv.Atoi(billingOpts.Year); err != nil {
			return err
		}
	}
	if billingOpts.Year == "" && int(time.Now().Month()) < month {
		year--
	}

	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	fromDT := strfmt.DateTime(from)
	to := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)

	f := excelize.NewFile()

	f.SetSheetName("Sheet1", "Parameter")

	clusterSheetName := fmt.Sprintf("%04d-%02d", year, month) + " Cluster"
	f.NewSheet(clusterSheetName)
	containerSheetName := fmt.Sprintf("%04d-%02d", year, month) + " Container"
	f.NewSheet(containerSheetName)
	volumeSheetName := fmt.Sprintf("%04d-%02d", year, month) + " Volume"
	f.NewSheet(volumeSheetName)
	ipSheetName := fmt.Sprintf("%04d-%02d", year, month) + " IPs"
	f.NewSheet(ipSheetName)
	networkSheetName := fmt.Sprintf("%04d-%02d", year, month) + " Network traffic"
	f.NewSheet(networkSheetName)
	s3SheetName := fmt.Sprintf("%04d-%02d", year, month) + " S3"
	f.NewSheet(s3SheetName)
	postgresSheetName := fmt.Sprintf("%04d-%02d", year, month) + " Postgres"
	f.NewSheet(postgresSheetName)

	// Parameter
	must(f.SetCellValue("Parameter", "A1", "Finance Cloud Native billing"))
	if billingOpts.Tenant != "" {
		must(f.SetCellValue("Parameter", "A3", "Tenant"))
		must(f.SetCellValue("Parameter", "B3", billingOpts.Tenant))
	}
	if billingOpts.ProjectID != "" {
		must(f.SetCellValue("Parameter", "C3", "ProjectID"))
		must(f.SetCellValue("Parameter", "D3", billingOpts.ProjectID))
	}
	must(f.SetCellValue("Parameter", "A4", "Period Start"))
	must(f.SetCellValue("Parameter", "B4", from))
	must(f.SetCellValue("Parameter", "A5", "Period End"))
	if time.Now().Before(to) {
		must(f.SetCellValue("Parameter", "B5", time.Now()))
	} else {
		must(f.SetCellValue("Parameter", "B5", to))
	}
	must(f.SetCellValue("Parameter", "A7", "CPU included"))
	must(f.SetCellValue("Parameter", "B7", 64))
	must(f.SetCellValue("Parameter", "A8", "RAM included"))
	must(f.SetCellValue("Parameter", "B8", 128))
	must(f.SetCellValue("Parameter", "A9", "Local vol included"))
	must(f.SetCellValue("Parameter", "B9", 512))
	must(f.SetColWidth("Parameter", "A", "B", 14))
	must(f.SetCellValue("Parameter", "A11", "* Lifetime[s] is always calculated for the given period"))

	// Cluster Billing
	curCluster := models.V1ClusterUsageRequest{
		From: &fromDT,
		To:   strfmt.DateTime(to),
	}
	if billingOpts.Tenant != "" {
		curCluster.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		curCluster.Projectid = billingOpts.ProjectID
	}

	requestCluster := accounting.NewClusterUsageParams()
	requestCluster.SetBody(&curCluster)

	responseCluster, err := c.cloud.Accounting.ClusterUsage(requestCluster, nil)
	if err != nil {
		return err
	}

	must(f.SetCellValue(clusterSheetName, "A1", "Tenant"))
	must(f.SetCellValue(clusterSheetName, "B1", "Project ID"))
	must(f.SetColWidth(clusterSheetName, "B", "B", 0))
	must(f.SetCellValue(clusterSheetName, "C1", "Project"))
	must(f.SetCellValue(clusterSheetName, "D1", "Cluster ID"))
	must(f.SetColWidth(clusterSheetName, "D", "D", 0))
	must(f.SetCellValue(clusterSheetName, "E1", "Cluster"))
	must(f.SetCellValue(clusterSheetName, "F1", "Start"))
	must(f.SetColWidth(clusterSheetName, "F", "F", 15))
	must(f.SetCellValue(clusterSheetName, "G1", "Lifetime[s]"))
	must(f.SetCellValue(clusterSheetName, "H1", "CPUs used [CPU*h]"))
	must(f.SetCellValue(clusterSheetName, "I1", "CPUs on top [CPU*h]"))
	must(f.SetCellValue(clusterSheetName, "J1", "CPUs avg [CPU]"))
	must(f.SetCellValue(clusterSheetName, "K1", "Memory used [GiB*h]"))
	must(f.SetCellValue(clusterSheetName, "L1", "Memory on top [GiB*h]"))
	must(f.SetCellValue(clusterSheetName, "M1", "Memory avg [GiB]"))
	must(f.SetCellValue(clusterSheetName, "N1", "Local vol used [GiB*h]"))
	must(f.SetCellValue(clusterSheetName, "O1", "Local vol on top [GiB*h]"))
	must(f.SetCellValue(clusterSheetName, "P1", "Local vol avg [GiB]"))
	must(f.SetColWidth(clusterSheetName, "H", "P", 15))
	for i, v := range responseCluster.Payload.Usage {
		must(f.SetCellValue(clusterSheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		must(f.SetCellValue(clusterSheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		must(f.SetCellValue(clusterSheetName, "C"+fmt.Sprint(i+2), *v.Projectname))
		must(f.SetCellValue(clusterSheetName, "D"+fmt.Sprint(i+2), *v.Clusterid))
		must(f.SetCellValue(clusterSheetName, "E"+fmt.Sprint(i+2), *v.Clustername))
		start := time.Time(*v.Clusterstart)
		must(f.SetCellValue(clusterSheetName, "F"+fmt.Sprint(i+2), start))
		must(f.SetCellValue(clusterSheetName, "G"+fmt.Sprint(i+2), *v.Lifetime/1000000000))
		must(f.SetCellFormula(clusterSheetName, "H"+fmt.Sprint(i+2), "=SUMIF('"+containerSheetName+"'!D$1:D$9999,D"+fmt.Sprint(i+2)+",'"+containerSheetName+"'!M$1:M$99999)/3600"))
		must(f.SetCellFormula(clusterSheetName, "I"+fmt.Sprint(i+2), "=MAX(H"+fmt.Sprint(i+2)+"-Parameter!B7*G"+fmt.Sprint(i+2)+"/3600,0)"))
		must(f.SetCellFormula(clusterSheetName, "J"+fmt.Sprint(i+2), "=H"+fmt.Sprint(i+2)+"/G"+fmt.Sprint(i+2)+"*3600"))
		must(f.SetCellFormula(clusterSheetName, "K"+fmt.Sprint(i+2), "=SUMIF('"+containerSheetName+"'!D$1:D$9999,D"+fmt.Sprint(i+2)+",'"+containerSheetName+"'!N$1:N$9999)/3600"))
		must(f.SetCellFormula(clusterSheetName, "L"+fmt.Sprint(i+2), "=MAX(K"+fmt.Sprint(i+2)+"-Parameter!B8*G"+fmt.Sprint(i+2)+"/3600,0)"))
		must(f.SetCellFormula(clusterSheetName, "M"+fmt.Sprint(i+2), "=K"+fmt.Sprint(i+2)+"/G"+fmt.Sprint(i+2)+"*3600"))
		must(f.SetCellFormula(clusterSheetName, "N"+fmt.Sprint(i+2), "=SUMIF('"+volumeSheetName+"'!D$1:D$9999,D"+fmt.Sprint(i+2)+",'"+volumeSheetName+"'!M$1:M$9999)/3600"))
		must(f.SetCellFormula(clusterSheetName, "O"+fmt.Sprint(i+2), "=MAX(N"+fmt.Sprint(i+2)+"-Parameter!B9*G"+fmt.Sprint(i+2)+"/3600,0)"))
		must(f.SetCellFormula(clusterSheetName, "P"+fmt.Sprint(i+2), "=N"+fmt.Sprint(i+2)+"/G"+fmt.Sprint(i+2)+"*3600"))
	}

	// Container Billing
	curContainer := models.V1ContainerUsageRequest{
		From: &fromDT,
		To:   strfmt.DateTime(to),
	}
	if billingOpts.Tenant != "" {
		curContainer.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		curContainer.Projectid = billingOpts.ProjectID
	}

	requestContainer := accounting.NewContainerUsageParams()
	requestContainer.SetBody(&curContainer)

	responseContainer, err := c.cloud.Accounting.ContainerUsage(requestContainer, nil)
	if err != nil {
		return err
	}
	must(f.SetCellValue(containerSheetName, "A1", "Tenant"))
	must(f.SetCellValue(containerSheetName, "B1", "Project ID"))
	must(f.SetColWidth(containerSheetName, "B", "B", 0))
	must(f.SetCellValue(containerSheetName, "C1", "Project"))
	must(f.SetCellValue(containerSheetName, "D1", "Cluster ID"))
	must(f.SetColWidth(containerSheetName, "D", "D", 0))
	must(f.SetCellValue(containerSheetName, "E1", "Cluster"))
	must(f.SetCellValue(containerSheetName, "F1", "Namespace"))
	must(f.SetCellValue(containerSheetName, "G1", "Pod ID"))
	must(f.SetColWidth(containerSheetName, "G", "G", 0))
	must(f.SetCellValue(containerSheetName, "H1", "Podname"))
	must(f.SetCellValue(containerSheetName, "I1", "Containername"))
	must(f.SetCellValue(containerSheetName, "J1", "Containerimage"))
	must(f.SetCellValue(containerSheetName, "K1", "Start"))
	must(f.SetColWidth(containerSheetName, "K", "K", 15))
	must(f.SetCellValue(containerSheetName, "L1", "Lifetime[s]"))
	must(f.SetCellValue(containerSheetName, "M1", "CPU[CPU*s]"))
	must(f.SetCellValue(containerSheetName, "N1", "Memory[GiB*s]"))
	must(f.SetCellValue(containerSheetName, "O1", "Annotations"))
	for i, v := range responseContainer.Payload.Usage {
		must(f.SetCellValue(containerSheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		must(f.SetCellValue(containerSheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		must(f.SetCellValue(containerSheetName, "C"+fmt.Sprint(i+2), *v.Projectname))
		must(f.SetCellValue(containerSheetName, "D"+fmt.Sprint(i+2), *v.Clusterid))
		must(f.SetCellValue(containerSheetName, "E"+fmt.Sprint(i+2), *v.Clustername))
		must(f.SetCellValue(containerSheetName, "F"+fmt.Sprint(i+2), *v.Namespace))
		must(f.SetCellValue(containerSheetName, "G"+fmt.Sprint(i+2), *v.Poduuid))
		must(f.SetCellValue(containerSheetName, "H"+fmt.Sprint(i+2), *v.Podname))
		must(f.SetCellValue(containerSheetName, "I"+fmt.Sprint(i+2), *v.Containername))
		must(f.SetCellValue(containerSheetName, "J"+fmt.Sprint(i+2), *v.Containerimage))
		start := time.Time(*v.Podstart)
		must(f.SetCellValue(containerSheetName, "K"+fmt.Sprint(i+2), start))
		must(f.SetCellValue(containerSheetName, "L"+fmt.Sprint(i+2), *v.Lifetime/1000000000))
		cpuseconds, _ := strconv.Atoi(*v.Cpuseconds)
		must(f.SetCellValue(containerSheetName, "M"+fmt.Sprint(i+2), cpuseconds))
		memoryseconds, _ := strconv.Atoi(*v.Memoryseconds)
		must(f.SetCellValue(containerSheetName, "N"+fmt.Sprint(i+2), memoryseconds/1024/1024/1024))
		must(f.SetCellValue(containerSheetName, "O"+fmt.Sprint(i+2), strings.Join(v.Annotations, "; ")))
	}

	// Volume Billing
	curVolume := models.V1VolumeUsageRequest{
		From: &fromDT,
		To:   strfmt.DateTime(to),
	}
	if billingOpts.Tenant != "" {
		curVolume.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		curVolume.Projectid = billingOpts.ProjectID
	}

	requestVolume := accounting.NewVolumeUsageParams()
	requestVolume.SetBody(&curVolume)

	responseVolume, err := c.cloud.Accounting.VolumeUsage(requestVolume, nil)
	if err != nil {
		return err
	}
	must(f.SetCellValue(volumeSheetName, "A1", "Tenant"))
	must(f.SetCellValue(volumeSheetName, "B1", "Project ID"))
	must(f.SetColWidth(volumeSheetName, "B", "B", 0))
	must(f.SetCellValue(volumeSheetName, "C1", "Project"))
	must(f.SetCellValue(volumeSheetName, "D1", "Cluster ID"))
	must(f.SetColWidth(volumeSheetName, "D", "D", 0))
	must(f.SetCellValue(volumeSheetName, "E1", "Cluster"))
	must(f.SetCellValue(volumeSheetName, "F1", "Volume ID"))
	must(f.SetColWidth(volumeSheetName, "F", "F", 0))
	must(f.SetCellValue(volumeSheetName, "G1", "Volume"))
	must(f.SetCellValue(volumeSheetName, "H1", "Class"))
	must(f.SetCellValue(volumeSheetName, "I1", "Type"))
	must(f.SetCellValue(volumeSheetName, "J1", "Start"))
	must(f.SetColWidth(volumeSheetName, "J", "J", 15))
	must(f.SetCellValue(volumeSheetName, "K1", "Lifetime[s]"))
	must(f.SetCellValue(volumeSheetName, "L1", "Capacity[GiB*s]"))
	must(f.SetCellValue(volumeSheetName, "M1", "Local[GiB*s]"))
	must(f.SetCellValue(volumeSheetName, "N1", "Block[GiB*s]"))
	for i, v := range responseVolume.Payload.Usage {
		must(f.SetCellValue(volumeSheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		must(f.SetCellValue(volumeSheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		must(f.SetCellValue(volumeSheetName, "C"+fmt.Sprint(i+2), *v.Projectname))
		must(f.SetCellValue(volumeSheetName, "D"+fmt.Sprint(i+2), *v.Clusterid))
		must(f.SetCellValue(volumeSheetName, "E"+fmt.Sprint(i+2), *v.Clustername))
		must(f.SetCellValue(volumeSheetName, "F"+fmt.Sprint(i+2), *v.UUID))
		must(f.SetCellValue(volumeSheetName, "G"+fmt.Sprint(i+2), *v.Name))
		must(f.SetCellValue(volumeSheetName, "H"+fmt.Sprint(i+2), *v.Class))
		must(f.SetCellValue(volumeSheetName, "I"+fmt.Sprint(i+2), *v.Type))
		start := time.Time(*v.Start)
		must(f.SetCellValue(volumeSheetName, "J"+fmt.Sprint(i+2), start))
		must(f.SetCellValue(volumeSheetName, "K"+fmt.Sprint(i+2), *v.Lifetime/1000000000))
		capacityseconds, _ := strconv.Atoi(*v.Capacityseconds)
		must(f.SetCellValue(volumeSheetName, "L"+fmt.Sprint(i+2), capacityseconds/1024/1024/1024))
		must(f.SetCellFormula(volumeSheetName, "M"+fmt.Sprint(i+2), "=IF(LEFT(H"+fmt.Sprint(i+2)+",7)=\"csi-lvm\",L"+fmt.Sprint(i+2)+",0)"))
		must(f.SetCellFormula(volumeSheetName, "N"+fmt.Sprint(i+2), "=IF(LEFT(H"+fmt.Sprint(i+2)+",10)=\"partition-\",L"+fmt.Sprint(i+2)+",0)"))
	}

	// IP Billing
	curIP := models.V1IPUsageRequest{
		From: &fromDT,
		To:   strfmt.DateTime(to),
	}
	if billingOpts.Tenant != "" {
		curIP.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		curIP.Projectid = billingOpts.ProjectID
	}

	requestIP := accounting.NewIPUsageParams()
	requestIP.SetBody(&curIP)

	responseIP, err := c.cloud.Accounting.IPUsage(requestIP, nil)
	if err != nil {
		return err
	}
	must(f.SetCellValue(ipSheetName, "A1", "Tenant"))
	must(f.SetCellValue(ipSheetName, "B1", "Project ID"))
	must(f.SetColWidth(ipSheetName, "B", "B", 0))
	must(f.SetCellValue(ipSheetName, "C1", "Project"))
	must(f.SetCellValue(ipSheetName, "D1", "IP"))
	must(f.SetColWidth(ipSheetName, "D", "D", 14))
	must(f.SetCellValue(ipSheetName, "E1", "Lifetime[s]"))
	for i, v := range responseIP.Payload.Usage {
		must(f.SetCellValue(ipSheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		must(f.SetCellValue(ipSheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		must(f.SetCellValue(ipSheetName, "C"+fmt.Sprint(i+2), *v.Projectname))
		must(f.SetCellValue(ipSheetName, "D"+fmt.Sprint(i+2), *v.IP))
		must(f.SetCellValue(ipSheetName, "E"+fmt.Sprint(i+2), *v.Lifetime/1000000000))
	}

	// network-traffic Billing
	curNetwork := models.V1NetworkUsageRequest{
		From: &fromDT,
		To:   strfmt.DateTime(to),
	}
	if billingOpts.Tenant != "" {
		curNetwork.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		curNetwork.Tenant = billingOpts.ProjectID
	}

	requestNetwork := accounting.NewNetworkUsageParams()
	requestNetwork.SetBody(&curNetwork)

	responseNetwork, err := c.cloud.Accounting.NetworkUsage(requestNetwork, nil)
	if err != nil {
		return err
	}
	must(f.SetCellValue(networkSheetName, "A1", "Tenant"))
	must(f.SetCellValue(networkSheetName, "B1", "Project ID"))
	must(f.SetColWidth(networkSheetName, "B", "B", 0))
	must(f.SetCellValue(networkSheetName, "C1", "Project"))
	must(f.SetCellValue(networkSheetName, "D1", "Cluster ID"))
	must(f.SetColWidth(networkSheetName, "D", "D", 0))
	must(f.SetCellValue(networkSheetName, "E1", "Cluster"))
	must(f.SetCellValue(networkSheetName, "F1", "Device"))
	must(f.SetCellValue(networkSheetName, "G1", "In[GiB]"))
	must(f.SetCellValue(networkSheetName, "H1", "Out[GiB]"))
	must(f.SetCellValue(networkSheetName, "I1", "Total[GiB]"))
	for i, v := range responseNetwork.Payload.Usage {
		must(f.SetCellValue(networkSheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		must(f.SetCellValue(networkSheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		must(f.SetCellValue(networkSheetName, "C"+fmt.Sprint(i+2), *v.Projectname))
		must(f.SetCellValue(networkSheetName, "D"+fmt.Sprint(i+2), *v.Clusterid))
		must(f.SetCellValue(networkSheetName, "E"+fmt.Sprint(i+2), *v.Clustername))
		must(f.SetCellValue(networkSheetName, "F"+fmt.Sprint(i+2), *v.Device))
		in, _ := strconv.Atoi(*v.In)
		must(f.SetCellValue(networkSheetName, "G"+fmt.Sprint(i+2), in/1024/1024/1024))
		out, _ := strconv.Atoi(*v.Out)
		must(f.SetCellValue(networkSheetName, "H"+fmt.Sprint(i+2), out/1024/1024/1024))
		total, _ := strconv.Atoi(*v.Total)
		must(f.SetCellValue(networkSheetName, "I"+fmt.Sprint(i+2), total/1024/1024/1024))
	}

	// S3 Billing
	curS3 := models.V1S3UsageRequest{
		From: &fromDT,
		To:   strfmt.DateTime(to),
	}
	if billingOpts.Tenant != "" {
		curS3.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		curS3.Projectid = billingOpts.ProjectID
	}

	requestS3 := accounting.NewS3UsageParams()
	requestS3.SetBody(&curS3)

	responseS3, err := c.cloud.Accounting.S3Usage(requestS3, nil)
	if err != nil {
		return err
	}
	must(f.SetCellValue(s3SheetName, "A1", "Tenant"))
	must(f.SetCellValue(s3SheetName, "B1", "Project ID"))
	must(f.SetColWidth(s3SheetName, "B", "B", 0))
	must(f.SetCellValue(s3SheetName, "C1", "Project"))
	must(f.SetCellValue(s3SheetName, "D1", "Partition"))
	must(f.SetCellValue(s3SheetName, "E1", "User"))
	must(f.SetCellValue(s3SheetName, "F1", "Bucket ID"))
	must(f.SetColWidth(s3SheetName, "F", "F", 0))
	must(f.SetCellValue(s3SheetName, "G1", "Bucket"))
	must(f.SetCellValue(s3SheetName, "H1", "Objects"))
	must(f.SetCellValue(s3SheetName, "I1", "Capacity[GiB*s]"))
	must(f.SetCellValue(s3SheetName, "J1", "Lifetime[s]"))
	for i, v := range responseS3.Payload.Usage {
		must(f.SetCellValue(s3SheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		must(f.SetCellValue(s3SheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		must(f.SetCellValue(s3SheetName, "C"+fmt.Sprint(i+2), *v.Projectname))
		must(f.SetCellValue(s3SheetName, "D"+fmt.Sprint(i+2), *v.Partition))
		must(f.SetCellValue(s3SheetName, "E"+fmt.Sprint(i+2), *v.User))
		must(f.SetCellValue(s3SheetName, "F"+fmt.Sprint(i+2), *v.Bucketid))
		must(f.SetCellValue(s3SheetName, "G"+fmt.Sprint(i+2), *v.Bucketname))
		objects, _ := strconv.Atoi(*v.Currentnumberofobjects)
		must(f.SetCellValue(s3SheetName, "H"+fmt.Sprint(i+2), objects))
		capacityseconds, _ := strconv.Atoi(*v.Storageseconds)
		must(f.SetCellValue(s3SheetName, "I"+fmt.Sprint(i+2), capacityseconds/1024/1024/1024))
		must(f.SetCellValue(s3SheetName, "J"+fmt.Sprint(i+2), *v.Lifetime/1000000000))
	}

	// Postgres Billing
	curPostgres := models.V1PostgresUsageRequest{
		From: &fromDT,
		To:   strfmt.DateTime(to),
	}
	if billingOpts.Tenant != "" {
		curPostgres.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		curPostgres.Projectid = billingOpts.ProjectID
	}

	requestPostgres := accounting.NewPostgresUsageParams()
	requestPostgres.SetBody(&curPostgres)

	responsePostgres, err := c.cloud.Accounting.PostgresUsage(requestPostgres, nil)
	if err != nil {
		return err
	}
	must(f.SetCellValue(postgresSheetName, "A1", "Tenant"))
	must(f.SetCellValue(postgresSheetName, "B1", "Project ID"))
	must(f.SetColWidth(postgresSheetName, "B", "B", 0))
	must(f.SetCellValue(postgresSheetName, "C1", "Project"))
	must(f.SetCellValue(postgresSheetName, "D1", "Postgres ID"))
	must(f.SetColWidth(postgresSheetName, "D", "D", 0))
	must(f.SetCellValue(postgresSheetName, "E1", "Description"))
	must(f.SetCellValue(postgresSheetName, "F1", "Start"))
	must(f.SetColWidth(postgresSheetName, "F", "F", 15))
	must(f.SetCellValue(postgresSheetName, "G1", "CPU[CPU*s]"))
	must(f.SetCellValue(postgresSheetName, "H1", "Memory[GiB*s]"))
	must(f.SetCellValue(postgresSheetName, "I1", "Storage[GiB*s]"))
	must(f.SetCellValue(postgresSheetName, "J1", "Lifetime[s]"))
	for i, v := range responsePostgres.Payload.Usage {
		must(f.SetCellValue(postgresSheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		must(f.SetCellValue(postgresSheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		must(f.SetCellValue(postgresSheetName, "C"+fmt.Sprint(i+2), ""))
		must(f.SetCellValue(postgresSheetName, "D"+fmt.Sprint(i+2), *v.Postgresid))
		must(f.SetCellValue(postgresSheetName, "E"+fmt.Sprint(i+2), *v.Postgresdescription))
		start := time.Time(*v.Postgresstart)
		must(f.SetCellValue(postgresSheetName, "F"+fmt.Sprint(i+2), start))
		cpuseconds, _ := strconv.Atoi(*v.Cpuseconds)
		must(f.SetCellValue(postgresSheetName, "G"+fmt.Sprint(i+2), cpuseconds))
		memoryseconds, _ := strconv.Atoi(*v.Memoryseconds)
		must(f.SetCellValue(postgresSheetName, "H"+fmt.Sprint(i+2), memoryseconds/1024/1024/1024))
		storageseconds, _ := strconv.Atoi(*v.Storageseconds)
		must(f.SetCellValue(postgresSheetName, "I"+fmt.Sprint(i+2), storageseconds/1024/1024/1024))
		must(f.SetCellValue(postgresSheetName, "J"+fmt.Sprint(i+2), *v.Lifetime/1000000000))
	}

	filename := ""
	if billingOpts.Filename != "" {
		filename = billingOpts.Filename
		if !strings.HasSuffix(filename, ".xlsx") {
			filename = filename + ".xlsx"
		}
	} else {
		if billingOpts.Tenant != "" {
			filename = fmt.Sprintf("%04d-%02d-%s-billing.xlsx", year, month, billingOpts.Tenant)
		} else {
			filename = fmt.Sprintf("%04d-%02d-billing.xlsx", year, month)
		}
	}

	if err := f.SaveAs(filename); err != nil {
		return err
	}

	fmt.Println("Created " + filename)

	return nil
}

func (c *config) containerUsage() error {
	from := strfmt.DateTime(billingOpts.From)
	cur := models.V1ContainerUsageRequest{
		From: &from,
		To:   strfmt.DateTime(billingOpts.To),
	}
	if billingOpts.Tenant != "" {
		cur.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		cur.Projectid = billingOpts.ProjectID
	}
	if billingOpts.ClusterID != "" {
		cur.Clusterid = billingOpts.ClusterID
	}
	if billingOpts.Namespace != "" {
		cur.Namespace = billingOpts.Namespace
	}

	if billingOpts.CSV {
		return c.containerUsageCSV(&cur)
	}
	return c.containerUsageJSON(&cur)
}

func (c *config) containerUsageJSON(cur *models.V1ContainerUsageRequest) error {
	request := accounting.NewContainerUsageParams()
	request.SetBody(cur)

	response, err := c.cloud.Accounting.ContainerUsage(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}

func (c *config) containerUsageCSV(cur *models.V1ContainerUsageRequest) error {
	request := accounting.NewContainerUsageCSVParams()
	request.SetBody(cur)

	response, err := c.cloud.Accounting.ContainerUsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}

func (c *config) ipUsage() error {
	from := strfmt.DateTime(billingOpts.From)
	iur := models.V1IPUsageRequest{
		From: &from,
		To:   strfmt.DateTime(billingOpts.To),
	}
	if billingOpts.Tenant != "" {
		iur.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		iur.Projectid = billingOpts.ProjectID
	}

	if billingOpts.CSV {
		return c.ipUsageCSV(&iur)
	}
	return c.ipUsageJSON(&iur)
}

func (c *config) ipUsageJSON(iur *models.V1IPUsageRequest) error {
	request := accounting.NewIPUsageParams()
	request.SetBody(iur)

	response, err := c.cloud.Accounting.IPUsage(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}

func (c *config) ipUsageCSV(iur *models.V1IPUsageRequest) error {
	request := accounting.NewIPUsageCSVParams()
	request.SetBody(iur)

	response, err := c.cloud.Accounting.IPUsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}

func (c *config) networkTrafficUsage() error {
	from := strfmt.DateTime(billingOpts.From)
	cur := models.V1NetworkUsageRequest{
		From: &from,
		To:   strfmt.DateTime(billingOpts.To),
	}
	if billingOpts.Tenant != "" {
		cur.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		cur.Projectid = billingOpts.ProjectID
	}
	if billingOpts.ClusterID != "" {
		cur.Clusterid = billingOpts.ClusterID
	}
	if billingOpts.Device != "" {
		cur.Device = billingOpts.Device
	}

	if billingOpts.CSV {
		return c.networkTrafficUsageCSV(&cur)
	}
	return c.networkTrafficUsageJSON(&cur)
}

func (c *config) networkTrafficUsageJSON(cur *models.V1NetworkUsageRequest) error {
	request := accounting.NewNetworkUsageParams()
	request.SetBody(cur)

	response, err := c.cloud.Accounting.NetworkUsage(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}

func (c *config) networkTrafficUsageCSV(cur *models.V1NetworkUsageRequest) error {
	request := accounting.NewNetworkUsageCSVParams()
	request.SetBody(cur)

	response, err := c.cloud.Accounting.NetworkUsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}

func (c *config) s3Usage() error {
	from := strfmt.DateTime(billingOpts.From)
	sur := models.V1S3UsageRequest{
		From: &from,
		To:   strfmt.DateTime(billingOpts.To),
	}
	if billingOpts.Tenant != "" {
		sur.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		sur.Projectid = billingOpts.ProjectID
	}

	if billingOpts.CSV {
		return c.s3UsageCSV(&sur)
	}
	return c.s3UsageJSON(&sur)
}

func (c *config) s3UsageJSON(sur *models.V1S3UsageRequest) error {
	request := accounting.NewS3UsageParams()
	request.SetBody(sur)

	response, err := c.cloud.Accounting.S3Usage(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}

func (c *config) s3UsageCSV(sur *models.V1S3UsageRequest) error {
	request := accounting.NewS3UsageCSVParams()
	request.SetBody(sur)

	response, err := c.cloud.Accounting.S3UsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}

func (c *config) volumeUsage() error {
	from := strfmt.DateTime(billingOpts.From)
	vur := models.V1VolumeUsageRequest{
		From: &from,
		To:   strfmt.DateTime(billingOpts.To),
	}
	if billingOpts.Tenant != "" {
		vur.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		vur.Projectid = billingOpts.ProjectID
	}
	if billingOpts.ClusterID != "" {
		vur.Clusterid = billingOpts.ClusterID
	}
	if billingOpts.Namespace != "" {
		vur.Clusterid = billingOpts.Namespace
	}

	if billingOpts.CSV {
		return c.volumeUsageCSV(&vur)
	}
	return c.volumeUsageJSON(&vur)
}

func (c *config) volumeUsageJSON(vur *models.V1VolumeUsageRequest) error {
	request := accounting.NewVolumeUsageParams()
	request.SetBody(vur)

	response, err := c.cloud.Accounting.VolumeUsage(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}

func (c *config) volumeUsageCSV(vur *models.V1VolumeUsageRequest) error {
	request := accounting.NewVolumeUsageCSVParams()
	request.SetBody(vur)

	response, err := c.cloud.Accounting.VolumeUsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}

func (c *config) postgresUsage() error {
	from := strfmt.DateTime(billingOpts.From)
	cur := models.V1PostgresUsageRequest{
		From: &from,
		To:   strfmt.DateTime(billingOpts.To),
	}
	if billingOpts.Tenant != "" {
		cur.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		cur.Projectid = billingOpts.ProjectID
	}
	if billingOpts.ClusterID != "" {
		cur.Clusterid = billingOpts.ClusterID
	}

	if billingOpts.CSV {
		return c.postgresUsageCSV(&cur)
	}
	return c.postgresUsageJSON(&cur)
}

func (c *config) postgresUsageJSON(cur *models.V1PostgresUsageRequest) error {
	request := accounting.NewPostgresUsageParams()
	request.SetBody(cur)

	response, err := c.cloud.Accounting.PostgresUsage(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}

func (c *config) postgresUsageCSV(cur *models.V1PostgresUsageRequest) error {
	request := accounting.NewPostgresUsageCSVParams()
	request.SetBody(cur)

	response, err := c.cloud.Accounting.PostgresUsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}
