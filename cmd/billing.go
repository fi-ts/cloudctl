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
}

var (
	billingOpts *BillingOpts
	billingCmd  = &cobra.Command{
		Use:   "billing",
		Short: "look at the bills",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return containerUsage()
		},
		PreRun: bindPFlags,
	}
)

func init() {
	rootCmd.AddCommand(billingCmd)

	billingOpts = &BillingOpts{}

	billingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	billingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	billingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at")
	billingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	billingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project", "p", "", "the project to account")
	billingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster", "c", "", "the cluster to account")
	billingCmd.Flags().StringVarP(&billingOpts.Namespace, "namespace", "n", "", "the namespace to account")
	billingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	billingCmd.MarkFlagRequired("from")

	viper.BindPFlags(billingCmd.Flags())
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
