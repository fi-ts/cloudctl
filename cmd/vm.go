package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/fi-ts/cloudctl/pkg/client/vm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newVMCmd(c *config) *cobra.Command {
	cmCmd := &cobra.Command{
		Use:   "vm",
		Short: "manage virtual machines",
		Long:  "manage virtual machines",
	}

	var wideOutput bool
	var order string

	vmListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list virtual machines",
		Long:    "list all virtual machines",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vmList(wideOutput, order)
		},
	}
	vmListCmd.Flags().BoolVarP(&wideOutput, "wide", "w", false, "display with additional columns")
	vmListCmd.Flags().StringVar(&order, "order", "", "order by column(s) (comma separated, prefix with - for descending). Supported: uuid, fqdn, status, os, serviceclass, cpu, ram, storage, ip, service, description")

	vmDescribeCmd := &cobra.Command{
		Use:     "describe <vmid>",
		Short:   "describe a virtual machine",
		Long:    "describe a virtual machine by UUID or FQDN",
		Aliases: []string{"get", "info"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vmDescribe(args)
		},
	}

	vmDeleteCmd := &cobra.Command{
		Use:   "delete <vmid>",
		Short: "delete a virtual machine",
		Long:  "delete a virtual machine by UUID or FQDN",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("Not Implemented")
		},
	}

	vmCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a new virtual machine",
		Long:  "create a new virtual machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vmCreate()
		},
	}

	vmUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "update a virtual machine",
		Long:  "update a virtual machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("Not Implemented")
		},
	}

	cmCmd.AddCommand(vmListCmd)
	cmCmd.AddCommand(vmDescribeCmd)
	cmCmd.AddCommand(vmDeleteCmd)
	cmCmd.AddCommand(vmCreateCmd)
	cmCmd.AddCommand(vmUpdateCmd)

	return cmCmd
}

func (c *config) vmCreate() error {
	return fmt.Errorf("Not Implemented")
}

func (c *config) vmList(wide bool, order string) error {

	ctx := api.MustDefaultContext()
	baseURL := ctx.ApiURL

	authContext, err := api.GetAuthContext(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("token not fount in kubeconfig")
	}
	apiToken := authContext.IDToken

	if baseURL == "" {
		return fmt.Errorf("url is not configured")
	}
	if apiToken == "" {
		return fmt.Errorf("api token is not configured")
	}

	client := vm.NewClient(baseURL, apiToken)

	instances, err := client.ListInstances()
	if err != nil {
		return fmt.Errorf("failed to list VM instances: %w", err)
	}

	if order != "" {
		instances = sortInstances(instances, order)
	}

	if wide {
		return printVMInstancesWide(os.Stdout, instances)
	}
	return printVMInstances(os.Stdout, instances)
}

func (c *config) vmDescribe(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("vmid is required")
	}
	instanceUUID := args[0]

	ctx := api.MustDefaultContext()
	baseURL := ctx.ApiURL

	authContext, err := api.GetAuthContext(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("token not fount in kubeconfig")
	}
	apiToken := authContext.IDToken

	if baseURL == "" {
		return fmt.Errorf("url is not configured")
	}
	if apiToken == "" {
		return fmt.Errorf("api token is not configured")
	}

	client := vm.NewClient(baseURL, apiToken)

	instance, err := client.GetInstanceDetails(instanceUUID)
	if err != nil {
		return fmt.Errorf("failed to get VM instance details: %w", err)
	}

	return c.describePrinter.Print(instance)
}

func sortInstances(instances []vm.VMInstanceLight, order string) []vm.VMInstanceLight {
	sorted := make([]vm.VMInstanceLight, len(instances))
	copy(sorted, instances)

	parts := strings.Split(order, ",")
	sort.SliceStable(sorted, func(i, j int) bool {
		for _, part := range parts {
			desc := false
			col := part
			if strings.HasPrefix(col, "-") {
				desc = true
				col = strings.TrimPrefix(col, "-")
			}
			cmp := compareField(sorted[i], sorted[j], col)
			if cmp != 0 {
				if desc {
					return cmp > 0
				}
				return cmp < 0
			}
		}
		return false
	})
	return sorted
}

func compareField(a, b vm.VMInstanceLight, col string) int {
	switch col {
	case "uuid":
		return strings.Compare(a.VmUUID, b.VmUUID)
	case "fqdn":
		return strings.Compare(a.VmFQDN, b.VmFQDN)
	case "status":
		return strings.Compare(a.Status, b.Status)
	case "os":
		return strings.Compare(a.OSTitle, b.OSTitle)
	case "serviceclass":
		return strings.Compare(a.ServiceClass, b.ServiceClass)
	case "cpu":
		return a.CPU - b.CPU
	case "ram":
		return a.RAM - b.RAM
	case "storage":
		return strings.Compare(a.StorageClass, b.StorageClass)
	case "service":
		return strings.Compare(a.ServiceTitle, b.ServiceTitle)
	default:
		return 0
	}
}

func printVMInstances(w interface{ Write([]byte) (int, error) }, instances []vm.VMInstanceLight) error {
	var sb strings.Builder
	header := fmt.Sprintf("%-36s %-40s %-10s %-20s %-6s %-6s %-30s\n",
		"UUID", "FQDN", "STATUS", "OS", "CPU", "RAM", "SERVICE")
	sb.WriteString(header)
	for _, inst := range instances {
		ramStr := formatRAM(inst.RAM)
		sb.WriteString(fmt.Sprintf("%-36s %-40s %-10s %-20s %-6d %-6s %-30s\n",
			inst.VmUUID,
			inst.VmFQDN,
			inst.Status,
			inst.OSTitle,
			inst.CPU,
			ramStr,
			inst.ServiceTitle,
		))
	}
	_, err := w.Write([]byte(sb.String()))
	return err
}

func printVMInstancesWide(w interface{ Write([]byte) (int, error) }, instances []vm.VMInstanceLight) error {
	var sb strings.Builder
	header := fmt.Sprintf("%-36s %-40s %-10s %-20s %-12s %-15s %-6s %-6s %-15s %-30s\n",
		"UUID", "FQDN", "STATUS", "OS", "AVAILABILITY", "SERVICE CLASS", "CPU", "RAM", "STORAGE CLASS", "SERVICE")
	sb.WriteString(header)
	for _, inst := range instances {
		ramStr := formatRAM(inst.RAM)
		sb.WriteString(fmt.Sprintf("%-36s %-40s %-10s %-20s %-12s %-15s %-6d %-6s %-15s %-30s\n",
			inst.VmUUID,
			inst.VmFQDN,
			inst.Status,
			inst.OSTitle,
			inst.Availability,
			inst.ServiceClass,
			inst.CPU,
			ramStr,
			inst.StorageClass,
			inst.ServiceTitle,
		))
	}
	_, err := w.Write([]byte(sb.String()))
	return err
}

func formatRAM(mb int) string {
	if mb >= 1024 {
		gb := float64(mb) / 1024.0
		return fmt.Sprintf("%.0fGB", gb)
	}
	return fmt.Sprintf("%dMB", mb)
}
