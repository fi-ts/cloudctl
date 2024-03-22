package cmd

import (
	"fmt"

	"github.com/fi-ts/cloud-go/api/client/ip"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newIPCmd(c *config) *cobra.Command {
	ipCmd := &cobra.Command{
		Use:   "ip",
		Short: "manage IPs",
		Long:  "manage static IP addresses for your projects, which can be utilized for the service type load balancer of your clusters.",
	}
	ipListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list ips",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ipList()
		},
	}
	ipStaticCmd := &cobra.Command{
		Use:   "static <ip>",
		Short: "make an ephemeral ip static such that it won't be deleted if not used anymore",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ipStatic(args)
		},
	}
	ipAllocateCmd := &cobra.Command{
		Use:   "allocate <ip>",
		Short: "allocate a static IP address for your project that can be used for your cluster's service type load balancer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ipAllocate()
		},
	}
	ipFreeCmd := &cobra.Command{
		Use:     "delete <ip>",
		Aliases: []string{"destroy", "rm", "remove", "free"},
		Short:   "delete an ip",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ipFree(args)
		},
	}

	ipCmd.AddCommand(ipListCmd)
	ipCmd.AddCommand(ipStaticCmd)
	ipCmd.AddCommand(ipFreeCmd)
	ipCmd.AddCommand(ipAllocateCmd)

	ipListCmd.Flags().StringP("ipaddress", "", "", "ipaddress to filter [optional]")
	ipListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	ipListCmd.Flags().StringP("prefix", "", "", "prefix to filter [optional]")
	ipListCmd.Flags().StringP("machineid", "", "", "machineid to filter [optional]")
	ipListCmd.Flags().StringP("network", "", "", "network to filter [optional]")

	genericcli.Must(ipListCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	genericcli.Must(ipListCmd.RegisterFlagCompletionFunc("network", c.comp.NetworkListCompletion))

	ipStaticCmd.Flags().StringP("name", "", "", "set name of the ip address [required]")
	ipStaticCmd.Flags().StringP("description", "", "", "set description of the ip address [required]")
	genericcli.Must(ipStaticCmd.MarkFlagRequired("name"))
	genericcli.Must(ipStaticCmd.MarkFlagRequired("description"))

	ipAllocateCmd.Flags().StringP("name", "", "", "set name of the ip address [required]")
	ipAllocateCmd.Flags().StringP("description", "", "", "set description of the ip address [required]")
	ipAllocateCmd.Flags().StringP("specific-ip", "", "", "try allocating a specific ip address from a network [optional]")
	ipAllocateCmd.Flags().StringP("network", "", "", "the network of the ip address [required]")
	ipAllocateCmd.Flags().StringP("project", "", "", "the project of the ip address [required]")
	ipAllocateCmd.Flags().StringSliceP("tags", "", []string{}, "set tags of the ip address [optional]")
	genericcli.Must(ipAllocateCmd.MarkFlagRequired("name"))
	genericcli.Must(ipAllocateCmd.MarkFlagRequired("description"))
	genericcli.Must(ipAllocateCmd.MarkFlagRequired("network"))
	genericcli.Must(ipAllocateCmd.MarkFlagRequired("project"))
	genericcli.Must(ipAllocateCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))

	return ipCmd
}

func (c *config) ipList() error {
	if helper.AtLeastOneViperStringFlagGiven("ipaddress", "project", "prefix", "machineid", "network") {
		params := ip.NewFindIPsParams()
		ifr := &models.V1IPFindRequest{
			Ipaddress:     pointer.SafeDeref(helper.ViperString("ipaddress")),
			Projectid:     pointer.SafeDeref(helper.ViperString("project")),
			Networkprefix: pointer.SafeDeref(helper.ViperString("prefix")),
			Networkid:     pointer.SafeDeref(helper.ViperString("network")),
			Machineid:     pointer.SafeDeref(helper.ViperString("machineid")),
		}
		params.SetBody(ifr)
		resp, err := c.cloud.IP.FindIPs(params, nil)
		if err != nil {
			return err
		}
		return output.New().Print(resp.Payload)
	}
	resp, err := c.cloud.IP.ListIPs(nil, nil)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Payload)
}

func (c *config) ipStatic(args []string) error {
	ipAddress, err := c.getIPFromArgs(args)
	if err != nil {
		return err
	}

	params := ip.NewUpdateIPParams()
	iur := &models.V1IPUpdateRequest{
		Ipaddress: &ipAddress,
		Type:      pointer.Pointer("static"),
	}
	if helper.ViperString("name") != nil {
		iur.Name = *helper.ViperString("name")
	}
	if helper.ViperString("description") != nil {
		iur.Description = *helper.ViperString("description")
	}

	if !viper.GetBool("yes-i-really-mean-it") {
		fmt.Println("Turning an IP from ephemeral to static is irreversible. The IP address is not cleaned up automatically on cluster deletion. The address will be accounted until the IP address gets freed manually from your side.")
		err = helper.Prompt("Are you sure? (y/n)", "y")
		if err != nil {
			return err
		}
	}

	params.SetBody(iur)
	resp, err := c.cloud.IP.UpdateIP(params, nil)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Payload)
}

func (c *config) ipAllocate() error {
	params := ip.NewAllocateIPParams()
	iar := &models.V1IPAllocateRequest{
		Name:        *helper.ViperString("name"),
		Description: *helper.ViperString("description"),
		Type:        pointer.Pointer("static"),
		Networkid:   helper.ViperString("network"),
		Projectid:   helper.ViperString("project"),
		Tags:        helper.ViperStringSlice("tags"),
	}

	if helper.ViperString("specific-ip") != nil {
		iar.SpecificIP = helper.ViperString("specific-ip")
	}

	if !viper.GetBool("yes-i-really-mean-it") {
		fmt.Println("Allocating a static IP address costs additional money because addresses are limited. The IP address is not cleaned up automatically on cluster deletion. The address will be accounted until the IP address gets freed manually from your side.")
		err := helper.Prompt("Are you sure? (y/n)", "y")
		if err != nil {
			return err
		}
	}

	params.SetBody(iar)
	resp, err := c.cloud.IP.AllocateIP(params, nil)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Payload)
}

func (c *config) ipFree(args []string) error {
	ipAddress, err := c.getIPFromArgs(args)
	if err != nil {
		return err
	}

	params := ip.NewFreeIPParams()
	params.SetIP(ipAddress)
	resp, err := c.cloud.IP.FreeIP(params, nil)
	if err != nil {
		return err
	}

	return output.New().Print(resp.Payload)
}

func (c *config) getIPFromArgs(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("no ip given")
	}

	ipAddress := args[0]
	params := ip.NewGetIPParams()
	params.SetIP(ipAddress)

	_, err := c.cloud.IP.GetIP(params, nil)
	if err != nil {
		return "", err
	}
	return ipAddress, nil
}
