package cmd

import (
	"fmt"
	"time"

	"git.f-i-ts.de/cloud-native/cloudctl/api/client/billing"
	output "git.f-i-ts.de/cloud-native/cloudctl/cmd/output"
	"github.com/go-openapi/strfmt"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"
)

type BillingOpts struct {
	Tenant     string
	FromString string `validate:"required"`
	ToString   string
	From       time.Time
	To         time.Time
	ProjectID  string
	ClusterID  string
	Namespace  string
	CSV        bool
	Forecast   bool
}

var (
	billingOpts *BillingOpts
	billingCmd  = &cobra.Command{
		Use:   "billing",
		Short: "manage bills",
		Long:  "TODO",
	}
	containerBillingCmd = &cobra.Command{
		Use:   "container",
		Short: "look at container bills",
		Example: `If you want to get the costs in Euro, then set two environment variables with the prices from your contract:

		export CLOUDCTL_COSTS_CPU_HOUR=0.01        # Costs in Euro per CPU Hour
		export CLOUDCTL_COSTS_MEMORY_GI_HOUR=0.01  # Costs in Euro per Gi Memory Hour

		cloudctl billing container --from 2019-01-01
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return containerUsage()
		},
		PreRun: bindPFlags,
	}
	clusterBillingCmd = &cobra.Command{
		Use:   "cluster",
		Short: "look at cluster bills",
		Example: `If you want to get the costs in Euro, then set two environment variables with the prices from your contract:

		export CLOUDCTL_COSTS_CLUSTER_HOUR=0.01        # Costs in Euro per cluster hour

		cloudctl billing cluster --from 2019-01-01
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return clusterUsage()
		},
		PreRun: bindPFlags,
	}
)

func init() {
	rootCmd.AddCommand(billingCmd)

	billingCmd.AddCommand(containerBillingCmd)
	billingCmd.AddCommand(clusterBillingCmd)

	billingOpts = &BillingOpts{}

	containerBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	containerBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	containerBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at")
	containerBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	containerBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project", "p", "", "the project to account")
	containerBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster", "c", "", "the cluster to account")
	containerBillingCmd.Flags().StringVarP(&billingOpts.Namespace, "namespace", "n", "", "the namespace to account")
	containerBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")
	containerBillingCmd.Flags().BoolVarP(&billingOpts.Forecast, "forecast", "", false, "calculates resource usage until end of time window as if accountable data would not change any more (defaults to false)")

	containerBillingCmd.MarkFlagRequired("from")

	clusterBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	clusterBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project", "p", "", "the project to account")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster", "c", "", "the cluster to account")
	clusterBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")
	clusterBillingCmd.Flags().BoolVarP(&billingOpts.Forecast, "forecast", "", false, "calculates resource usage until end of time window as if accountable data would not change any more (defaults to false)")

	clusterBillingCmd.MarkFlagRequired("from")

	viper.BindPFlags(containerBillingCmd.Flags())
}

func initBillingOpts() error {
	validate := validator.New()
	err := validate.Struct(billingOpts)
	if err != nil {
		return err
	}

	from, err := time.Parse(viper.GetString("time-format"), billingOpts.FromString)
	if err != nil {
		return err
	}
	billingOpts.From = from

	if billingOpts.ToString == "" {
		billingOpts.To = time.Now()
	} else {
		to, err := time.Parse(viper.GetString("time-format"), billingOpts.ToString)
		if err != nil {
			return err
		}
		billingOpts.To = to
	}

	return nil
}

func clusterUsage() error {
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
	if billingOpts.Forecast {
		cur.Forecast = billingOpts.Forecast
	}

	if billingOpts.CSV {
		return clusterUsageCSV(&cur)
	} else {
		return clusterUsageJSON(&cur)
	}
}

func clusterUsageJSON(cur *models.V1ClusterUsageRequest) error {
	request := billing.NewClusterUsageParams()
	request.SetBody(cur)

	response, err := cloud.Billing.ClusterUsage(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *billing.ClusterUsageDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	return printer.Print(response.Payload)
}

func clusterUsageCSV(cur *models.V1ClusterUsageRequest) error {
	request := billing.NewClusterUsageCSVParams()
	request.SetBody(cur)

	response, err := cloud.Billing.ClusterUsageCSV(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *billing.ClusterUsageCSVDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	fmt.Println(response.Payload)
	return nil
}

func containerUsage() error {
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
	if billingOpts.Forecast {
		cur.Forecast = billingOpts.Forecast
	}

	if billingOpts.CSV {
		return containerUsageCSV(&cur)
	} else {
		return containerUsageJSON(&cur)
	}
}

func containerUsageJSON(cur *models.V1ContainerUsageRequest) error {
	request := billing.NewContainerUsageParams()
	request.SetBody(cur)

	response, err := cloud.Billing.ContainerUsage(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *billing.ContainerUsageDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	return printer.Print(response.Payload)
}

func containerUsageCSV(cur *models.V1ContainerUsageRequest) error {
	request := billing.NewContainerUsageCSVParams()
	request.SetBody(cur)

	response, err := cloud.Billing.ContainerUsageCSV(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *billing.ContainerUsageCSVDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	fmt.Println(response.Payload)
	return nil
}
