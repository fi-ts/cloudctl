package cmd

import (
	"fmt"
	"time"

	"github.com/fi-ts/cloud-go/api/client/accounting"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/sorters"
	"github.com/go-openapi/strfmt"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/now"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type BillingOpts struct {
	Tenant      string
	FromString  string
	ToString    string
	From        time.Time
	To          time.Time
	ProjectID   string
	ClusterID   string
	Device      string
	Namespace   string
	UUID        string
	Annotations []string
	CSV         bool
}

var (
	billingOpts *BillingOpts
)

func newBillingCmd(c *config) *cobra.Command {
	billingCmd := &cobra.Command{
		Use:   "billing",
		Short: "lookup resource consumption of your cloud resources",
	}
	projectBillingCmd := &cobra.Command{
		Use:   "projects",
		Short: "discover projects within a given time period",
		Long:  "can be used to find all projects within a given time period, e.g. to narrow down queries that would become very big otherwise",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.projectsBilling()
		},
	}
	containerBillingCmd := &cobra.Command{
		Use:   "container",
		Short: "look at container bills",
		Long: `
You may want to convert the usage to a price in Euro by using the prices from your contract. You can use the following environment variables:

export CLOUDCTL_COSTS_CPU_HOUR=0.01        # costs per cpu hour
export CLOUDCTL_COSTS_MEMORY_GI_HOUR=0.01  # costs per memory hour

⚠ Please be aware that any costs calculated in this fashion can still be different from the final bill as it does not include contract specific details like minimum purchase, discounts, etc.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.containerUsage()
		},
	}
	clusterBillingCmd := &cobra.Command{
		Use:   "cluster",
		Short: "look at cluster bills",
		Long: `
You may want to convert the usage to a price in Euro by using the prices from your contract. You can use the following environment variables:

export CLOUDCTL_COSTS_HOUR=0.01        # costs per hour

⚠ Please be aware that any costs calculated in this fashion can still be different from the final bill as it does not include contract specific details like minimum purchase, discounts, etc.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.clusterUsage()
		},
	}
	ipBillingCmd := &cobra.Command{
		Use:   "ip",
		Short: "look at ip bills",
		Long: `
You may want to convert the usage to a price in Euro by using the prices from your contract. You can use the following environment variables:

export CLOUDCTL_COSTS_HOUR=0.01        # costs per hour

⚠ Please be aware that any costs calculated in this fashion can still be different from the final bill as it does not include contract specific details like minimum purchase, discounts, etc.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.ipUsage()
		},
	}
	machineBillingCmd := &cobra.Command{
		Use:   "machine",
		Short: "look at machine bills",
		Long: `
You may want to convert the usage to a price in Euro by using the prices from your contract. You can use the following environment variables:

export CLOUDCTL_COSTS_HOUR=0.01        # costs per hour

⚠ Please be aware that any costs calculated in this fashion can still be different from the final bill as it does not include contract specific details like minimum purchase, discounts, etc.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.machineUsage()
		},
	}
	machineReservationBillingCmd := &cobra.Command{
		Use:   "machine-reservation",
		Short: "look at machine reservation bills",
		Long: `
You may want to convert the usage to a price in Euro by using the prices from your contract. You can use the following environment variables:

export CLOUDCTL_COSTS_HOUR=0.01        # costs per hour

⚠ Please be aware that any costs calculated in this fashion can still be different from the final bill as it does not include contract specific details like minimum purchase, discounts, etc.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.machineReservationUsage()
		},
	}
	productOptionBillingCmd := &cobra.Command{
		Use:   "product-option",
		Short: "look at product option bills",
		Long: `
You may want to convert the usage to a price in Euro by using the prices from your contract. You can use the following environment variables:

export CLOUDCTL_COSTS_HOUR=0.01        # costs per hour

⚠ Please be aware that any costs calculated in this fashion can still be different from the final bill as it does not include contract specific details like minimum purchase, discounts, etc.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.productOptionUsage()
		},
	}
	networkTrafficBillingCmd := &cobra.Command{
		Use:   "network-traffic",
		Short: "look at network traffic bills",
		Long: `
You may want to convert the usage to a price in Euro by using the prices from your contract. You can use the following environment variables:

export CLOUDCTL_COSTS_INCOMING_NETWORK_TRAFFIC_GI=0.01        # costs per gi
export CLOUDCTL_COSTS_OUTGOING_NETWORK_TRAFFIC_GI=0.01        # costs per gi
export CLOUDCTL_COSTS_TOTAL_NETWORK_TRAFFIC_GI=0.01           # costs per gi

⚠ Please be aware that any costs calculated in this fashion can still be different from the final bill as it does not include contract specific details like minimum purchase, discounts, etc.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.networkTrafficUsage()
		},
	}
	s3BillingCmd := &cobra.Command{
		Use:   "s3",
		Short: "look at s3 bills",
		Long: `
You may want to convert the usage to a price in Euro by using the prices from your contract. You can use the following environment variables:

export CLOUDCTL_COSTS_STORAGE_GI_HOUR=0.01        # costs per storage hour

⚠ Please be aware that any costs calculated in this fashion can still be different from the final bill as it does not include contract specific details like minimum purchase, discounts, etc.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.s3Usage()
		},
	}
	volumeBillingCmd := &cobra.Command{
		Use:   "volume",
		Short: "look at volume bills",
		Long: `
You may want to convert the usage to a price in Euro by using the prices from your contract. You can use the following environment variables:

export CLOUDCTL_COSTS_STORAGE_GI_HOUR=0.01        # costs per capacity hour

⚠ Please be aware that any costs calculated in this fashion can still be different from the final bill as it does not include contract specific details like minimum purchase, discounts, etc.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.volumeUsage()
		},
	}
	postgresBillingCmd := &cobra.Command{
		Use:   "postgres",
		Short: "look at postgres bills",
		Long: `
You may want to convert the usage to a price in Euro by using the prices from your contract. You can use the following environment variables:

export CLOUDCTL_COSTS_CPU_HOUR=0.01        # costs per cpu hour
export CLOUDCTL_COSTS_MEMORY_GI_HOUR=0.01  # costs per memory hour
export CLOUDCTL_COSTS_STORAGE_GI_HOUR=0.01 # Costs per capacity hour

⚠ Please be aware that any costs calculated in this fashion can still be different from the final bill as it does not include contract specific details like minimum purchase, discounts, etc.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := initBillingOpts()
			if err != nil {
				return err
			}
			return c.postgresUsage()
		},
	}

	billingCmd.AddCommand(projectBillingCmd)
	billingCmd.AddCommand(containerBillingCmd)
	billingCmd.AddCommand(clusterBillingCmd)
	billingCmd.AddCommand(ipBillingCmd)
	billingCmd.AddCommand(networkTrafficBillingCmd)
	billingCmd.AddCommand(s3BillingCmd)
	billingCmd.AddCommand(volumeBillingCmd)
	billingCmd.AddCommand(postgresBillingCmd)
	billingCmd.AddCommand(machineBillingCmd)
	billingCmd.AddCommand(machineReservationBillingCmd)
	billingCmd.AddCommand(productOptionBillingCmd)

	billingOpts = &BillingOpts{}

	projectBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	projectBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")

	containerBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	containerBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	containerBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	containerBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	containerBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	containerBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster-id", "c", "", "the cluster to account")
	containerBillingCmd.Flags().StringVarP(&billingOpts.Namespace, "namespace", "n", "", "the namespace to account")
	containerBillingCmd.Flags().StringSliceVar(&billingOpts.Annotations, "annotations", nil, "annotations filtering")
	containerBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	genericcli.Must(containerBillingCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(containerBillingCmd.RegisterFlagCompletionFunc("project-id", c.comp.ProjectListCompletion))
	genericcli.Must(containerBillingCmd.RegisterFlagCompletionFunc("cluster-id", c.comp.ClusterListCompletion))

	genericcli.Must(viper.BindPFlags(containerBillingCmd.Flags()))

	clusterBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	clusterBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	clusterBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster-id", "c", "", "the cluster to account")
	clusterBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	genericcli.Must(clusterBillingCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(clusterBillingCmd.RegisterFlagCompletionFunc("project-id", c.comp.ProjectListCompletion))
	genericcli.Must(clusterBillingCmd.RegisterFlagCompletionFunc("cluster-id", c.comp.ClusterListCompletion))

	genericcli.Must(viper.BindPFlags(clusterBillingCmd.Flags()))

	machineBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	machineBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	machineBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	machineBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	machineBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	machineBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster-id", "c", "", "the cluster to account")
	machineBillingCmd.Flags().String("machine-id", "", "the machine-id to account")
	machineBillingCmd.Flags().String("size-id", "", "the size-id to account")
	machineBillingCmd.Flags().String("partition-id", "", "the partition-id to account")

	genericcli.Must(machineBillingCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(machineBillingCmd.RegisterFlagCompletionFunc("project-id", c.comp.ProjectListCompletion))
	genericcli.Must(machineBillingCmd.RegisterFlagCompletionFunc("cluster-id", c.comp.ClusterListCompletion))
	genericcli.Must(machineBillingCmd.RegisterFlagCompletionFunc("partition-id", c.comp.PartitionListCompletion))
	genericcli.Must(machineBillingCmd.RegisterFlagCompletionFunc("size-id", c.comp.SizeListCompletion))

	genericcli.Must(viper.BindPFlags(machineBillingCmd.Flags()))

	machineReservationBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	machineReservationBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	machineReservationBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	machineReservationBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	machineReservationBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	machineReservationBillingCmd.Flags().String("id", "", "the id to account")
	machineReservationBillingCmd.Flags().String("size-id", "", "the size-id to account")
	machineReservationBillingCmd.Flags().String("partition-id", "", "the partition-id to account")

	genericcli.Must(machineReservationBillingCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(machineReservationBillingCmd.RegisterFlagCompletionFunc("project-id", c.comp.ProjectListCompletion))
	genericcli.Must(machineReservationBillingCmd.RegisterFlagCompletionFunc("partition-id", c.comp.PartitionListCompletion))
	genericcli.Must(machineReservationBillingCmd.RegisterFlagCompletionFunc("size-id", c.comp.SizeListCompletion))
	genericcli.AddSortFlag(machineReservationBillingCmd, sorters.MachineReservationsBillingUsageSorter())

	genericcli.Must(viper.BindPFlags(machineReservationBillingCmd.Flags()))

	productOptionBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	productOptionBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	productOptionBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	productOptionBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	productOptionBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	productOptionBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster-id", "c", "", "the cluster to account")
	productOptionBillingCmd.Flags().String("id", "", "the id of the product option to account")

	genericcli.Must(productOptionBillingCmd.RegisterFlagCompletionFunc("id", c.comp.ProductOptionsCompletion))
	genericcli.Must(productOptionBillingCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(productOptionBillingCmd.RegisterFlagCompletionFunc("project-id", c.comp.ProjectListCompletion))
	genericcli.Must(productOptionBillingCmd.RegisterFlagCompletionFunc("cluster-id", c.comp.ClusterListCompletion))

	genericcli.Must(viper.BindPFlags(productOptionBillingCmd.Flags()))

	ipBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	ipBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	ipBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	ipBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	ipBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	ipBillingCmd.Flags().StringSliceVar(&billingOpts.Annotations, "annotations", nil, "annotations filtering")
	ipBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	genericcli.Must(ipBillingCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(ipBillingCmd.RegisterFlagCompletionFunc("project-id", c.comp.ProjectListCompletion))

	genericcli.Must(viper.BindPFlags(ipBillingCmd.Flags()))

	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	networkTrafficBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster-id", "c", "", "the cluster to account")
	networkTrafficBillingCmd.Flags().StringVarP(&billingOpts.Device, "device", "", "", "the device to account")
	networkTrafficBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	genericcli.Must(networkTrafficBillingCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(networkTrafficBillingCmd.RegisterFlagCompletionFunc("project-id", c.comp.ProjectListCompletion))
	genericcli.Must(networkTrafficBillingCmd.RegisterFlagCompletionFunc("cluster-id", c.comp.ClusterListCompletion))

	genericcli.Must(viper.BindPFlags(networkTrafficBillingCmd.Flags()))

	s3BillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	s3BillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	s3BillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	s3BillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	s3BillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	s3BillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	genericcli.Must(s3BillingCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(s3BillingCmd.RegisterFlagCompletionFunc("project-id", c.comp.ProjectListCompletion))

	genericcli.Must(viper.BindPFlags(s3BillingCmd.Flags()))

	volumeBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	volumeBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	volumeBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	volumeBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	volumeBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	volumeBillingCmd.Flags().StringVarP(&billingOpts.Namespace, "namespace", "n", "", "the namespace to account")
	volumeBillingCmd.Flags().StringVarP(&billingOpts.ClusterID, "cluster-id", "c", "", "the cluster to account")
	volumeBillingCmd.Flags().StringSliceVar(&billingOpts.Annotations, "annotations", nil, "annotations filtering")
	volumeBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	genericcli.Must(volumeBillingCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(volumeBillingCmd.RegisterFlagCompletionFunc("project-id", c.comp.ProjectListCompletion))
	genericcli.Must(volumeBillingCmd.RegisterFlagCompletionFunc("cluster-id", c.comp.ClusterListCompletion))

	genericcli.Must(viper.BindPFlags(volumeBillingCmd.Flags()))

	postgresBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	postgresBillingCmd.Flags().StringP("time-format", "", "2006-01-02", "the time format used to parse the arguments 'from' and 'to'")
	postgresBillingCmd.Flags().StringVarP(&billingOpts.FromString, "from", "", "", "the start time in the accounting window to look at (optional, defaults to start of the month")
	postgresBillingCmd.Flags().StringVarP(&billingOpts.ToString, "to", "", "", "the end time in the accounting window to look at (optional, defaults to current system time)")
	postgresBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project-id", "p", "", "the project to account")
	postgresBillingCmd.Flags().StringVar(&billingOpts.UUID, "uuid", "", "the uuid to account")
	postgresBillingCmd.Flags().StringSliceVar(&billingOpts.Annotations, "annotations", nil, "annotations filtering")
	postgresBillingCmd.Flags().BoolVarP(&billingOpts.CSV, "csv", "", false, "let the server generate a csv file")

	genericcli.Must(postgresBillingCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(postgresBillingCmd.RegisterFlagCompletionFunc("project-id", c.comp.ProjectListCompletion))
	genericcli.Must(postgresBillingCmd.RegisterFlagCompletionFunc("uuid", c.comp.PostgresListCompletion))

	genericcli.Must(viper.BindPFlags(postgresBillingCmd.Flags()))

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

func (c *config) projectsBilling() error {
	from := strfmt.DateTime(billingOpts.From)

	request := accounting.NewProjectsParams().WithBody(&models.V1ProjectInfoRequest{
		From: &from,
		To:   strfmt.DateTime(billingOpts.To),
	})

	response, err := c.cloud.Accounting.Projects(request, nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(response.Payload)
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

	return c.listPrinter.Print(response.Payload)
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

func (c *config) machineUsage() error {
	from := strfmt.DateTime(billingOpts.From)
	cur := models.V1MachineUsageRequest{
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
	if viper.IsSet("machine-id") {
		cur.ID = viper.GetString("machine-id")
	}
	if viper.IsSet("size-id") {
		cur.Sizeid = viper.GetString("size-id")
	}
	if viper.IsSet("partition-id") {
		cur.Partition = viper.GetString("partition-id")
	}

	response, err := c.cloud.Accounting.MachineUsage(accounting.NewMachineUsageParams().WithBody(&cur), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(response.Payload)
}

func (c *config) machineReservationUsage() error {
	from := strfmt.DateTime(billingOpts.From)
	cur := models.V1MachineReservationUsageRequest{
		From: &from,
		To:   strfmt.DateTime(billingOpts.To),
	}
	if billingOpts.Tenant != "" {
		cur.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		cur.Projectid = billingOpts.ProjectID
	}
	if viper.IsSet("id") {
		cur.ID = viper.GetString("id")
	}
	if viper.IsSet("size-id") {
		cur.Sizeid = viper.GetString("size-id")
	}
	if viper.IsSet("partition-id") {
		cur.Partition = viper.GetString("partition-id")
	}

	response, err := c.cloud.Accounting.MachineReservationUsage(accounting.NewMachineReservationUsageParams().WithBody(&cur), nil)
	if err != nil {
		return err
	}

	keys, err := genericcli.ParseSortFlags()
	if err != nil {
		return err
	}

	err = sorters.MachineReservationsBillingUsageSorter().SortBy(response.Payload.Usage, keys...)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(response.Payload)
}

func (c *config) productOptionUsage() error {
	from := strfmt.DateTime(billingOpts.From)
	cur := models.V1ProductOptionUsageRequest{
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
	if viper.IsSet("id") {
		cur.ID = viper.GetString("id")
	}

	response, err := c.cloud.Accounting.ProductOptionUsage(accounting.NewProductOptionUsageParams().WithBody(&cur), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(response.Payload)
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
	if len(billingOpts.Annotations) > 0 {
		cur.Annotations = billingOpts.Annotations
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

	return c.listPrinter.Print(response.Payload)
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
	if len(billingOpts.Annotations) > 0 {
		iur.Annotations = billingOpts.Annotations
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

	return c.listPrinter.Print(response.Payload)
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

	return c.listPrinter.Print(response.Payload)
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
	req := models.V1S3UsageRequest{
		From: &from,
		To:   strfmt.DateTime(billingOpts.To),
	}
	if billingOpts.Tenant != "" {
		req.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		req.Projectid = billingOpts.ProjectID
	}

	if billingOpts.CSV {
		return c.s3UsageCSV(&req)
	}
	return c.s3UsageJSON(&req)
}

func (c *config) s3UsageJSON(req *models.V1S3UsageRequest) error {
	request := accounting.NewS3UsageParams()
	request.SetBody(req)

	response, err := c.cloud.Accounting.S3Usage(request, nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(response.Payload)
}

func (c *config) s3UsageCSV(req *models.V1S3UsageRequest) error {
	request := accounting.NewS3UsageCSVParams()
	request.SetBody(req)

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
		vur.Namespace = billingOpts.Namespace
	}
	if len(billingOpts.Annotations) > 0 {
		vur.Annotations = billingOpts.Annotations
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

	return c.listPrinter.Print(response.Payload)
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
	if billingOpts.UUID != "" {
		cur.UUID = billingOpts.UUID
	}
	if len(billingOpts.Annotations) > 0 {
		cur.Annotations = billingOpts.Annotations
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

	return c.listPrinter.Print(response.Payload)
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
