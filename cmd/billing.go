package cmd

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
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
	"github.com/xuri/excelize/v2"
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
	Month       string
	Year        string
	Filename    string
	UUID        string
	Annotations []string
	CSV         bool
}

var (
	billingOpts *BillingOpts
)

func byteSecondsToGiBHours(byteSeconds string) int64 {
	i := new(big.Float)
	i.SetString(byteSeconds)
	gibs,_ := new(big.Float).Quo(i, big.NewFloat(1<<30*3600)).Int64()
	return gibs
}

func newBillingCmd(c *config) *cobra.Command {
	billingCmd := &cobra.Command{
		Use:   "billing",
		Short: "lookup resource consumption of your cloud resources",
	}
	excelBillingCmd := &cobra.Command{
		Use:   "excel",
		Short: "create excel file with monthly billing information for all resources",
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

	billingCmd.AddCommand(excelBillingCmd)
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

	excelBillingCmd.Flags().StringVarP(&billingOpts.Tenant, "tenant", "t", "", "the tenant to account")
	excelBillingCmd.Flags().StringVarP(&billingOpts.ProjectID, "project", "p", "", "the project to account")
	excelBillingCmd.Flags().StringVarP(&billingOpts.Month, "month", "m", "", "requested month")
	excelBillingCmd.Flags().StringVarP(&billingOpts.Year, "year", "y", "", "requested year")
	excelBillingCmd.Flags().StringVarP(&billingOpts.Filename, "file", "f", "", "excel filename")

	genericcli.Must(viper.BindPFlags(excelBillingCmd.Flags()))

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

func (c *config) excel() error {
	now := time.Now()
	month := int(now.Month())
	year := int(now.Year())

	clusters := make(map[string]bool)

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
	if billingOpts.Year == "" && int(now.Month()) < month {
		year--
	}

	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	fromDT := strfmt.DateTime(from)
	to := from.AddDate(0, 1, 0)

	f := excelize.NewFile()
	defer func() { genericcli.Must(f.Close()) }()

	genericcli.Must(f.SetSheetName("Sheet1", "Parameter"))

	datePrefix := fmt.Sprintf("%04d-%02d", year, month)
	clusterSheetName := datePrefix + " Cluster"
	_, err = f.NewSheet(clusterSheetName)
	genericcli.Must(err)
	containerSheetName := datePrefix + " Container"
	_, err = f.NewSheet(containerSheetName)
	genericcli.Must(err)
	volumeSheetName := datePrefix + " Volume"
	_, err = f.NewSheet(volumeSheetName)
	genericcli.Must(err)
	ipSheetName := datePrefix + " IPs"
	_, err = f.NewSheet(ipSheetName)
	genericcli.Must(err)
	networkSheetName := datePrefix + " Network traffic"
	_, err = f.NewSheet(networkSheetName)
	genericcli.Must(err)
	s3SheetName := datePrefix + " S3"
	_, err = f.NewSheet(s3SheetName)
	genericcli.Must(err)
	postgresSheetName := datePrefix + " Postgres"
	_, err = f.NewSheet(postgresSheetName)
	genericcli.Must(err)

	// Parameter
	genericcli.Must(f.SetCellValue("Parameter", "A1", "Finance Cloud Native billing"))
	if billingOpts.Tenant != "" {
		genericcli.Must(f.SetCellValue("Parameter", "A3", "Tenant"))
		genericcli.Must(f.SetCellValue("Parameter", "B3", billingOpts.Tenant))
	}
	if billingOpts.ProjectID != "" {
		genericcli.Must(f.SetCellValue("Parameter", "C3", "ProjectID"))
		genericcli.Must(f.SetCellValue("Parameter", "D3", billingOpts.ProjectID))
	}
	genericcli.Must(f.SetCellValue("Parameter", "A4", "Period Start"))
	genericcli.Must(f.SetCellValue("Parameter", "B4", from))
	genericcli.Must(f.SetCellValue("Parameter", "A5", "Period End"))
	if now.Before(to) {
		genericcli.Must(f.SetCellValue("Parameter", "B5", now))
	} else {
		genericcli.Must(f.SetCellValue("Parameter", "B5", to))
	}
	genericcli.Must(f.SetCellValue("Parameter", "A7", "CPU included"))
	genericcli.Must(f.SetCellValue("Parameter", "B7", 64))
	genericcli.Must(f.SetCellValue("Parameter", "A8", "RAM included"))
	genericcli.Must(f.SetCellValue("Parameter", "B8", 128))
	genericcli.Must(f.SetCellValue("Parameter", "A9", "Local vol included"))
	genericcli.Must(f.SetCellValue("Parameter", "B9", 512))
	genericcli.Must(f.SetColWidth("Parameter", "A", "B", 14))
	genericcli.Must(f.SetCellValue("Parameter", "A11", "* Lifetime[h] is always calculated for the given period"))

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

	genericcli.Must(f.SetCellValue(clusterSheetName, "A1", "Tenant"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "B1", "Project ID"))
	genericcli.Must(f.SetColWidth(clusterSheetName, "B", "B", 0))
	genericcli.Must(f.SetCellValue(clusterSheetName, "C1", "Project"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "D1", "Cluster ID"))
	genericcli.Must(f.SetColWidth(clusterSheetName, "D", "D", 0))
	genericcli.Must(f.SetCellValue(clusterSheetName, "E1", "Cluster"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "F1", "Start"))
	genericcli.Must(f.SetColWidth(clusterSheetName, "F", "F", 15))
	genericcli.Must(f.SetCellValue(clusterSheetName, "G1", "Lifetime[h]"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "H1", "CPUs used [CPU*h]"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "I1", "CPUs on top [CPU*h]"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "J1", "CPUs avg [CPU]"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "K1", "Memory used [GiB*h]"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "L1", "Memory on top [GiB*h]"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "M1", "Memory avg [GiB]"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "N1", "Local vol used [GiB*h]"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "O1", "Local vol on top [GiB*h]"))
	genericcli.Must(f.SetCellValue(clusterSheetName, "P1", "Local vol avg [GiB]"))
	genericcli.Must(f.SetColWidth(clusterSheetName, "H", "P", 15))
	for i, v := range responseCluster.Payload.Usage {
		genericcli.Must(f.SetCellValue(clusterSheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		genericcli.Must(f.SetCellValue(clusterSheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		genericcli.Must(f.SetCellValue(clusterSheetName, "C"+fmt.Sprint(i+2), *v.Projectname))
		genericcli.Must(f.SetCellValue(clusterSheetName, "D"+fmt.Sprint(i+2), *v.Clusterid))
		clusters[*v.Clusterid] = true
		genericcli.Must(f.SetCellValue(clusterSheetName, "E"+fmt.Sprint(i+2), *v.Clustername))
		start := time.Time(*v.Clusterstart)
		genericcli.Must(f.SetCellValue(clusterSheetName, "F"+fmt.Sprint(i+2), start))
		genericcli.Must(f.SetCellValue(clusterSheetName, "G"+fmt.Sprint(i+2), *v.Lifetime/1000000000/3600))
		genericcli.Must(f.SetCellFormula(clusterSheetName, "H"+fmt.Sprint(i+2), "=SUMIF('"+containerSheetName+"'!D$1:D$999999,D"+fmt.Sprint(i+2)+",'"+containerSheetName+"'!M$1:M$999999)"))
		genericcli.Must(f.SetCellFormula(clusterSheetName, "I"+fmt.Sprint(i+2), "=MAX(H"+fmt.Sprint(i+2)+"-Parameter!B7*G"+fmt.Sprint(i+2)+",0)"))
		genericcli.Must(f.SetCellFormula(clusterSheetName, "J"+fmt.Sprint(i+2), "=H"+fmt.Sprint(i+2)+"/G"+fmt.Sprint(i+2)))
		genericcli.Must(f.SetCellFormula(clusterSheetName, "K"+fmt.Sprint(i+2), "=SUMIF('"+containerSheetName+"'!D$1:D$999999,D"+fmt.Sprint(i+2)+",'"+containerSheetName+"'!N$1:N$999999)"))
		genericcli.Must(f.SetCellFormula(clusterSheetName, "L"+fmt.Sprint(i+2), "=MAX(K"+fmt.Sprint(i+2)+"-Parameter!B8*G"+fmt.Sprint(i+2)+",0)"))
		genericcli.Must(f.SetCellFormula(clusterSheetName, "M"+fmt.Sprint(i+2), "=K"+fmt.Sprint(i+2)+"/G"+fmt.Sprint(i+2)))
		genericcli.Must(f.SetCellFormula(clusterSheetName, "N"+fmt.Sprint(i+2), "=SUMIF('"+volumeSheetName+"'!D$1:D$999999,D"+fmt.Sprint(i+2)+",'"+volumeSheetName+"'!M$1:M$999999)"))
		genericcli.Must(f.SetCellFormula(clusterSheetName, "O"+fmt.Sprint(i+2), "=MAX(N"+fmt.Sprint(i+2)+"-Parameter!B9*G"+fmt.Sprint(i+2)+",0)"))
		genericcli.Must(f.SetCellFormula(clusterSheetName, "P"+fmt.Sprint(i+2), "=N"+fmt.Sprint(i+2)+"/G"+fmt.Sprint(i+2)))
	}

	// Container Billing
	genericcli.Must(f.SetCellValue(containerSheetName, "A1", "Tenant"))
	genericcli.Must(f.SetCellValue(containerSheetName, "B1", "Project ID"))
	genericcli.Must(f.SetColWidth(containerSheetName, "B", "B", 0))
	genericcli.Must(f.SetCellValue(containerSheetName, "C1", "Project"))
	genericcli.Must(f.SetCellValue(containerSheetName, "D1", "Cluster ID"))
	genericcli.Must(f.SetColWidth(containerSheetName, "D", "D", 0))
	genericcli.Must(f.SetCellValue(containerSheetName, "E1", "Cluster"))
	genericcli.Must(f.SetCellValue(containerSheetName, "F1", "Namespace"))
	genericcli.Must(f.SetCellValue(containerSheetName, "G1", "Pod ID"))
	genericcli.Must(f.SetColWidth(containerSheetName, "G", "G", 0))
	genericcli.Must(f.SetCellValue(containerSheetName, "H1", "Podname"))
	genericcli.Must(f.SetCellValue(containerSheetName, "I1", "Containername"))
	genericcli.Must(f.SetCellValue(containerSheetName, "J1", "Containerimage"))
	genericcli.Must(f.SetCellValue(containerSheetName, "K1", "Start"))
	genericcli.Must(f.SetColWidth(containerSheetName, "K", "K", 15))
	genericcli.Must(f.SetCellValue(containerSheetName, "L1", "Lifetime[h]"))
	genericcli.Must(f.SetCellValue(containerSheetName, "M1", "CPU[CPU*h]"))
	genericcli.Must(f.SetCellValue(containerSheetName, "N1", "Memory[GiB*h]"))
	genericcli.Must(f.SetCellValue(containerSheetName, "O1", "Annotations"))

	ci := 0

	for cluster := range clusters {
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
		curContainer.Clusterid = cluster

		requestContainer := accounting.NewContainerUsageParams()
		requestContainer.SetBody(&curContainer)

		responseContainer, err := c.cloud.Accounting.ContainerUsage(requestContainer, nil)
		if err != nil {
			return err
		}

		for _, v := range responseContainer.Payload.Usage {
			genericcli.Must(f.SetCellValue(containerSheetName, "A"+fmt.Sprint(ci+2), *v.Tenant))
			genericcli.Must(f.SetCellValue(containerSheetName, "B"+fmt.Sprint(ci+2), *v.Projectid))
			genericcli.Must(f.SetCellValue(containerSheetName, "C"+fmt.Sprint(ci+2), *v.Projectname))
			genericcli.Must(f.SetCellValue(containerSheetName, "D"+fmt.Sprint(ci+2), *v.Clusterid))
			genericcli.Must(f.SetCellValue(containerSheetName, "E"+fmt.Sprint(ci+2), *v.Clustername))
			genericcli.Must(f.SetCellValue(containerSheetName, "F"+fmt.Sprint(ci+2), *v.Namespace))
			genericcli.Must(f.SetCellValue(containerSheetName, "G"+fmt.Sprint(ci+2), *v.Poduuid))
			genericcli.Must(f.SetCellValue(containerSheetName, "H"+fmt.Sprint(ci+2), *v.Podname))
			genericcli.Must(f.SetCellValue(containerSheetName, "I"+fmt.Sprint(ci+2), *v.Containername))
			genericcli.Must(f.SetCellValue(containerSheetName, "J"+fmt.Sprint(ci+2), *v.Containerimage))
			start := time.Time(*v.Podstart)
			genericcli.Must(f.SetCellValue(containerSheetName, "K"+fmt.Sprint(ci+2), start))
			genericcli.Must(f.SetCellValue(containerSheetName, "L"+fmt.Sprint(ci+2), *v.Lifetime/1000000000/3600))
			cpuseconds, _ := strconv.Atoi(*v.Cpuseconds)
			genericcli.Must(f.SetCellValue(containerSheetName, "M"+fmt.Sprint(ci+2), cpuseconds/3600))
			memoryhours := byteSecondsToGiBHours(*v.Memoryseconds)
			genericcli.Must(f.SetCellValue(containerSheetName, "N"+fmt.Sprint(ci+2), memoryhours))
			genericcli.Must(f.SetCellValue(containerSheetName, "O"+fmt.Sprint(ci+2), strings.Join(v.Annotations, "; ")))

			ci++
		}
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
	genericcli.Must(f.SetCellValue(volumeSheetName, "A1", "Tenant"))
	genericcli.Must(f.SetCellValue(volumeSheetName, "B1", "Project ID"))
	genericcli.Must(f.SetColWidth(volumeSheetName, "B", "B", 0))
	genericcli.Must(f.SetCellValue(volumeSheetName, "C1", "Project"))
	genericcli.Must(f.SetCellValue(volumeSheetName, "D1", "Cluster ID"))
	genericcli.Must(f.SetColWidth(volumeSheetName, "D", "D", 0))
	genericcli.Must(f.SetCellValue(volumeSheetName, "E1", "Cluster"))
	genericcli.Must(f.SetCellValue(volumeSheetName, "F1", "Volume ID"))
	genericcli.Must(f.SetColWidth(volumeSheetName, "F", "F", 0))
	genericcli.Must(f.SetCellValue(volumeSheetName, "G1", "Volume"))
	genericcli.Must(f.SetCellValue(volumeSheetName, "H1", "Type"))
	genericcli.Must(f.SetCellValue(volumeSheetName, "I1", "Start"))
	genericcli.Must(f.SetColWidth(volumeSheetName, "I", "I", 15))
	genericcli.Must(f.SetCellValue(volumeSheetName, "J1", "Lifetime[h]"))
	genericcli.Must(f.SetCellValue(volumeSheetName, "K1", "Capacity[GiB*h]"))
	genericcli.Must(f.SetCellValue(volumeSheetName, "L1", "Local[GiB*h]"))
	genericcli.Must(f.SetCellValue(volumeSheetName, "M1", "Block[GiB*h]"))
	for i, v := range responseVolume.Payload.Usage {
		genericcli.Must(f.SetCellValue(volumeSheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		genericcli.Must(f.SetCellValue(volumeSheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		genericcli.Must(f.SetCellValue(volumeSheetName, "C"+fmt.Sprint(i+2), *v.Projectname))
		genericcli.Must(f.SetCellValue(volumeSheetName, "D"+fmt.Sprint(i+2), *v.Clusterid))
		genericcli.Must(f.SetCellValue(volumeSheetName, "E"+fmt.Sprint(i+2), *v.Clustername))
		genericcli.Must(f.SetCellValue(volumeSheetName, "F"+fmt.Sprint(i+2), *v.UUID))
		genericcli.Must(f.SetCellValue(volumeSheetName, "G"+fmt.Sprint(i+2), *v.Name))
		genericcli.Must(f.SetCellValue(volumeSheetName, "H"+fmt.Sprint(i+2), *v.Type))
		start := time.Time(*v.Start)
		genericcli.Must(f.SetCellValue(volumeSheetName, "I"+fmt.Sprint(i+2), start))
		genericcli.Must(f.SetCellValue(volumeSheetName, "J"+fmt.Sprint(i+2), *v.Lifetime/1000000000/3600))
		capacityhours := byteSecondsToGiBHours(*v.Capacityseconds)
		genericcli.Must(f.SetCellValue(volumeSheetName, "K"+fmt.Sprint(i+2), capacityhours))
		genericcli.Must(f.SetCellFormula(volumeSheetName, "L"+fmt.Sprint(i+2), "=IF(LEFT(H"+fmt.Sprint(i+2)+",7)=\"csi-lvm\",K"+fmt.Sprint(i+2)+",0)"))
		genericcli.Must(f.SetCellFormula(volumeSheetName, "M"+fmt.Sprint(i+2), "=IF(LEFT(H"+fmt.Sprint(i+2)+",10)=\"partition-\",K"+fmt.Sprint(i+2)+",0)"))
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
	genericcli.Must(f.SetCellValue(ipSheetName, "A1", "Tenant"))
	genericcli.Must(f.SetCellValue(ipSheetName, "B1", "Project ID"))
	genericcli.Must(f.SetColWidth(ipSheetName, "B", "B", 0))
	genericcli.Must(f.SetCellValue(ipSheetName, "C1", "Project"))
	genericcli.Must(f.SetCellValue(ipSheetName, "D1", "IP"))
	genericcli.Must(f.SetColWidth(ipSheetName, "D", "D", 14))
	genericcli.Must(f.SetCellValue(ipSheetName, "E1", "Lifetime[h]"))
	for i, v := range responseIP.Payload.Usage {
		genericcli.Must(f.SetCellValue(ipSheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		genericcli.Must(f.SetCellValue(ipSheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		genericcli.Must(f.SetCellValue(ipSheetName, "C"+fmt.Sprint(i+2), *v.Projectname))
		genericcli.Must(f.SetCellValue(ipSheetName, "D"+fmt.Sprint(i+2), *v.IP))
		genericcli.Must(f.SetCellValue(ipSheetName, "E"+fmt.Sprint(i+2), *v.Lifetime/1000000000/3600))
	}

	// Network Traffic Billing
	curNetwork := models.V1NetworkUsageRequest{
		From: &fromDT,
		To:   strfmt.DateTime(to),
	}
	if billingOpts.Tenant != "" {
		curNetwork.Tenant = billingOpts.Tenant
	}
	if billingOpts.ProjectID != "" {
		curNetwork.Projectid = billingOpts.ProjectID
	}

	requestNetwork := accounting.NewNetworkUsageParams()
	requestNetwork.SetBody(&curNetwork)

	responseNetwork, err := c.cloud.Accounting.NetworkUsage(requestNetwork, nil)
	if err != nil {
		return err
	}
	genericcli.Must(f.SetCellValue(networkSheetName, "A1", "Tenant"))
	genericcli.Must(f.SetCellValue(networkSheetName, "B1", "Project ID"))
	genericcli.Must(f.SetColWidth(networkSheetName, "B", "B", 0))
	genericcli.Must(f.SetCellValue(networkSheetName, "C1", "Project"))
	genericcli.Must(f.SetCellValue(networkSheetName, "D1", "Cluster ID"))
	genericcli.Must(f.SetColWidth(networkSheetName, "D", "D", 0))
	genericcli.Must(f.SetCellValue(networkSheetName, "E1", "Cluster"))
	genericcli.Must(f.SetCellValue(networkSheetName, "F1", "Device"))
	genericcli.Must(f.SetCellValue(networkSheetName, "G1", "In[GiB]"))
	genericcli.Must(f.SetCellValue(networkSheetName, "H1", "Out[GiB]"))
	genericcli.Must(f.SetCellValue(networkSheetName, "I1", "Total[GiB]"))
	for i, v := range responseNetwork.Payload.Usage {
		genericcli.Must(f.SetCellValue(networkSheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		genericcli.Must(f.SetCellValue(networkSheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		genericcli.Must(f.SetCellValue(networkSheetName, "C"+fmt.Sprint(i+2), *v.Projectname))
		genericcli.Must(f.SetCellValue(networkSheetName, "D"+fmt.Sprint(i+2), *v.Clusterid))
		genericcli.Must(f.SetCellValue(networkSheetName, "E"+fmt.Sprint(i+2), *v.Clustername))
		genericcli.Must(f.SetCellValue(networkSheetName, "F"+fmt.Sprint(i+2), *v.Device))
		in := byteSecondsToGiBHours(*v.In)
		genericcli.Must(f.SetCellValue(networkSheetName, "G"+fmt.Sprint(i+2), in))
		out := byteSecondsToGiBHours(*v.Out)
		genericcli.Must(f.SetCellValue(networkSheetName, "H"+fmt.Sprint(i+2), out))
		total := byteSecondsToGiBHours(*v.Total)
		genericcli.Must(f.SetCellValue(networkSheetName, "I"+fmt.Sprint(i+2), total))
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
	genericcli.Must(f.SetCellValue(s3SheetName, "A1", "Tenant"))
	genericcli.Must(f.SetCellValue(s3SheetName, "B1", "Project ID"))
	genericcli.Must(f.SetColWidth(s3SheetName, "B", "B", 0))
	genericcli.Must(f.SetCellValue(s3SheetName, "C1", "Project"))
	genericcli.Must(f.SetCellValue(s3SheetName, "D1", "Partition"))
	genericcli.Must(f.SetCellValue(s3SheetName, "E1", "User"))
	genericcli.Must(f.SetCellValue(s3SheetName, "F1", "Bucket ID"))
	genericcli.Must(f.SetColWidth(s3SheetName, "F", "F", 0))
	genericcli.Must(f.SetCellValue(s3SheetName, "G1", "Bucket"))
	genericcli.Must(f.SetCellValue(s3SheetName, "H1", "Objects"))
	genericcli.Must(f.SetCellValue(s3SheetName, "I1", "Capacity[GiB*h]"))
	genericcli.Must(f.SetCellValue(s3SheetName, "J1", "Lifetime[h]"))
	for i, v := range responseS3.Payload.Usage {
		genericcli.Must(f.SetCellValue(s3SheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		genericcli.Must(f.SetCellValue(s3SheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		genericcli.Must(f.SetCellValue(s3SheetName, "C"+fmt.Sprint(i+2), *v.Projectname))
		genericcli.Must(f.SetCellValue(s3SheetName, "D"+fmt.Sprint(i+2), *v.Partition))
		genericcli.Must(f.SetCellValue(s3SheetName, "E"+fmt.Sprint(i+2), *v.User))
		genericcli.Must(f.SetCellValue(s3SheetName, "F"+fmt.Sprint(i+2), *v.Bucketid))
		genericcli.Must(f.SetCellValue(s3SheetName, "G"+fmt.Sprint(i+2), *v.Bucketname))
		objects, _ := strconv.Atoi(*v.Currentnumberofobjects)
		genericcli.Must(f.SetCellValue(s3SheetName, "H"+fmt.Sprint(i+2), objects))
		capacityhours := byteSecondsToGiBHours(*v.Storageseconds)
		genericcli.Must(f.SetCellValue(s3SheetName, "I"+fmt.Sprint(i+2), capacityhours))
		genericcli.Must(f.SetCellValue(s3SheetName, "J"+fmt.Sprint(i+2), *v.Lifetime/1000000000/3600))
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
	genericcli.Must(f.SetCellValue(postgresSheetName, "A1", "Tenant"))
	genericcli.Must(f.SetCellValue(postgresSheetName, "B1", "Project ID"))
	genericcli.Must(f.SetColWidth(postgresSheetName, "B", "B", 0))
	genericcli.Must(f.SetCellValue(postgresSheetName, "C1", "Project"))
	genericcli.Must(f.SetCellValue(postgresSheetName, "D1", "Partition"))
	genericcli.Must(f.SetCellValue(postgresSheetName, "E1", "Postgres ID"))
	genericcli.Must(f.SetColWidth(postgresSheetName, "E", "E", 0))
	genericcli.Must(f.SetCellValue(postgresSheetName, "F1", "Description"))
	genericcli.Must(f.SetCellValue(postgresSheetName, "G1", "Start"))
	genericcli.Must(f.SetColWidth(postgresSheetName, "G", "G", 15))
	genericcli.Must(f.SetCellValue(postgresSheetName, "H1", "CPU[CPU*s]"))
	genericcli.Must(f.SetCellValue(postgresSheetName, "I1", "Memory[GiB*h]"))
	genericcli.Must(f.SetCellValue(postgresSheetName, "J1", "Storage[GiB*h]"))
	genericcli.Must(f.SetCellValue(postgresSheetName, "K1", "Lifetime[h]"))
	for i, v := range responsePostgres.Payload.Usage {
		genericcli.Must(f.SetCellValue(postgresSheetName, "A"+fmt.Sprint(i+2), *v.Tenant))
		genericcli.Must(f.SetCellValue(postgresSheetName, "B"+fmt.Sprint(i+2), *v.Projectid))
		genericcli.Must(f.SetCellValue(postgresSheetName, "C"+fmt.Sprint(i+2), ""))
		genericcli.Must(f.SetCellValue(postgresSheetName, "D"+fmt.Sprint(i+2), *v.Partition))
		genericcli.Must(f.SetCellValue(postgresSheetName, "E"+fmt.Sprint(i+2), *v.Postgresid))
		genericcli.Must(f.SetCellValue(postgresSheetName, "F"+fmt.Sprint(i+2), *v.Postgresdescription))
		start := time.Time(*v.Postgresstart)
		genericcli.Must(f.SetCellValue(postgresSheetName, "G"+fmt.Sprint(i+2), start))
		cpuseconds, _ := strconv.Atoi(*v.Cpuseconds)
		genericcli.Must(f.SetCellValue(postgresSheetName, "H"+fmt.Sprint(i+2), cpuseconds))
		memoryhours := byteSecondsToGiBHours(*v.Memoryseconds)
		genericcli.Must(f.SetCellValue(postgresSheetName, "I"+fmt.Sprint(i+2), memoryhours))
		storagehours := byteSecondsToGiBHours(*v.Storageseconds)
		genericcli.Must(f.SetCellValue(postgresSheetName, "J"+fmt.Sprint(i+2), storagehours))
		genericcli.Must(f.SetCellValue(postgresSheetName, "K"+fmt.Sprint(i+2), *v.Lifetime/1000000000/3600))
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
