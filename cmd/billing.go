package cmd

import (
	"fmt"
	"time"

	"github.com/fi-ts/cloud-go/api/client/accounting"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/go-openapi/strfmt"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/now"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	CSV        bool
}

var (
	billingOpts *BillingOpts
)

func newBillingCmd() *cobra.Command {
	billingCmd := &cobra.Command{
		Use:   "billing",
		Short: "manage bills",
		Long:  "TODO",
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
			return containerUsage()
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
			return clusterUsage()
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
			return ipUsage()
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
			return networkTrafficUsage()
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
			return s3Usage()
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
			return volumeUsage()
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
			return postgresUsage()
		},
		PreRun: bindPFlags,
	}

	billingCmd.AddCommand(containerBillingCmd)
	billingCmd.AddCommand(clusterBillingCmd)
	billingCmd.AddCommand(ipBillingCmd)
	billingCmd.AddCommand(networkTrafficBillingCmd)
	billingCmd.AddCommand(s3BillingCmd)
	billingCmd.AddCommand(volumeBillingCmd)
	billingCmd.AddCommand(postgresBillingCmd)

	billingOpts = &BillingOpts{}

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

	if billingOpts.CSV {
		return clusterUsageCSV(&cur)
	}
	return clusterUsageJSON(&cur)
}

func clusterUsageJSON(cur *models.V1ClusterUsageRequest) error {
	request := accounting.NewClusterUsageParams()
	request.SetBody(cur)

	response, err := cloud.Accounting.ClusterUsage(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}

func clusterUsageCSV(cur *models.V1ClusterUsageRequest) error {
	request := accounting.NewClusterUsageCSVParams()
	request.SetBody(cur)

	response, err := cloud.Accounting.ClusterUsageCSV(request, nil)
	if err != nil {
		return err
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

	if billingOpts.CSV {
		return containerUsageCSV(&cur)
	}
	return containerUsageJSON(&cur)
}

func containerUsageJSON(cur *models.V1ContainerUsageRequest) error {
	request := accounting.NewContainerUsageParams()
	request.SetBody(cur)

	response, err := cloud.Accounting.ContainerUsage(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}

func containerUsageCSV(cur *models.V1ContainerUsageRequest) error {
	request := accounting.NewContainerUsageCSVParams()
	request.SetBody(cur)

	response, err := cloud.Accounting.ContainerUsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}

func ipUsage() error {
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
		return ipUsageCSV(&iur)
	}
	return ipUsageJSON(&iur)
}

func ipUsageJSON(iur *models.V1IPUsageRequest) error {
	request := accounting.NewIPUsageParams()
	request.SetBody(iur)

	response, err := cloud.Accounting.IPUsage(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}

func ipUsageCSV(iur *models.V1IPUsageRequest) error {
	request := accounting.NewIPUsageCSVParams()
	request.SetBody(iur)

	response, err := cloud.Accounting.IPUsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}

func networkTrafficUsage() error {
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
		return networkTrafficUsageCSV(&cur)
	}
	return networkTrafficUsageJSON(&cur)
}

func networkTrafficUsageJSON(cur *models.V1NetworkUsageRequest) error {
	request := accounting.NewNetworkUsageParams()
	request.SetBody(cur)

	response, err := cloud.Accounting.NetworkUsage(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}

func networkTrafficUsageCSV(cur *models.V1NetworkUsageRequest) error {
	request := accounting.NewNetworkUsageCSVParams()
	request.SetBody(cur)

	response, err := cloud.Accounting.NetworkUsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}

func s3Usage() error {
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
		return s3UsageCSV(&sur)
	}
	return s3UsageJSON(&sur)
}

func s3UsageJSON(sur *models.V1S3UsageRequest) error {
	request := accounting.NewS3UsageParams()
	request.SetBody(sur)

	response, err := cloud.Accounting.S3Usage(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}

func s3UsageCSV(sur *models.V1S3UsageRequest) error {
	request := accounting.NewS3UsageCSVParams()
	request.SetBody(sur)

	response, err := cloud.Accounting.S3UsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}

func volumeUsage() error {
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
		return volumeUsageCSV(&vur)
	}
	return volumeUsageJSON(&vur)
}

func volumeUsageJSON(vur *models.V1VolumeUsageRequest) error {
	request := accounting.NewVolumeUsageParams()
	request.SetBody(vur)

	response, err := cloud.Accounting.VolumeUsage(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}

func volumeUsageCSV(vur *models.V1VolumeUsageRequest) error {
	request := accounting.NewVolumeUsageCSVParams()
	request.SetBody(vur)

	response, err := cloud.Accounting.VolumeUsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}

func postgresUsage() error {
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
		return postgresUsageCSV(&cur)
	}
	return postgresUsageJSON(&cur)
}

func postgresUsageJSON(cur *models.V1PostgresUsageRequest) error {
	request := accounting.NewPostgresUsageParams()
	request.SetBody(cur)

	response, err := cloud.Accounting.PostgresUsage(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}

func postgresUsageCSV(cur *models.V1PostgresUsageRequest) error {
	request := accounting.NewPostgresUsageCSVParams()
	request.SetBody(cur)

	response, err := cloud.Accounting.PostgresUsageCSV(request, nil)
	if err != nil {
		return err
	}

	fmt.Println(response.Payload)
	return nil
}
