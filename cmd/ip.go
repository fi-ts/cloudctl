package cmd

import (
	"fmt"
	"log"

	"github.com/fi-ts/cloud-go/api/client/ip"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	output "github.com/fi-ts/cloudctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ipCmd = &cobra.Command{
		Use:   "ip",
		Short: "manage ips",
		Long:  "TODO",
	}
	ipListCmd = &cobra.Command{
		Use:     "list",
		Short:   "list ips",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ipList()
		},
		PreRun: bindPFlags,
	}
	ipStaticCmd = &cobra.Command{
		Use:   "static <ip>",
		Short: "make an ephemeral ip static such that it won't be deleted if not used anymore",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ipStatic(args)
		},
		PreRun: bindPFlags,
	}
	ipAllocateCmd = &cobra.Command{
		Use:   "allocate <ip>",
		Short: "allocate a static IP address for your project that can be used for your cluster's service type load balancer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ipAllocate(args)
		},
		PreRun: bindPFlags,
	}
	ipFreeCmd = &cobra.Command{
		Use:     "free <ip>",
		Aliases: []string{"rm", "destroy", "remove", "delete"},
		Short:   "free an ip",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ipFree(args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	rootCmd.AddCommand(ipCmd)

	ipCmd.AddCommand(ipListCmd)
	ipCmd.AddCommand(ipStaticCmd)
	ipCmd.AddCommand(ipFreeCmd)
	ipCmd.AddCommand(ipAllocateCmd)

	ipListCmd.Flags().StringP("ipaddress", "", "", "ipaddress to filter [optional]")
	ipListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	ipListCmd.Flags().StringP("prefix", "", "", "prefx to filter [optional]")
	ipListCmd.Flags().StringP("machineid", "", "", "machineid to filter [optional]")
	ipListCmd.Flags().StringP("network", "", "", "network to filter [optional]")

	ipStaticCmd.Flags().StringP("name", "", "", "set name of the ip address [required]")
	ipStaticCmd.Flags().StringP("description", "", "", "set description of the ip address [required]")
	err := ipStaticCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = ipStaticCmd.MarkFlagRequired("description")
	if err != nil {
		log.Fatal(err.Error())
	}

	ipAllocateCmd.Flags().StringP("name", "", "", "set name of the ip address [required]")
	ipAllocateCmd.Flags().StringP("description", "", "", "set description of the ip address [required]")
	ipAllocateCmd.Flags().StringP("specific-ip", "", "", "try allocating a specific ip address from a network [optional]")
	ipAllocateCmd.Flags().StringP("network", "", "", "the network of the ip address [required]")
	ipAllocateCmd.Flags().StringP("project", "", "", "the project of the ip address [required]")
	ipAllocateCmd.Flags().StringSliceP("tags", "", []string{}, "set tags of the ip address [optional]")
	err = ipAllocateCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = ipAllocateCmd.MarkFlagRequired("description")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = ipAllocateCmd.MarkFlagRequired("network")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = ipAllocateCmd.MarkFlagRequired("project")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = ipAllocateCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	err = ipListCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

func ipList() error {
	if helper.AtLeastOneViperStringFlagGiven("ipaddress", "project", "prefix", "machineid", "network") {
		params := ip.NewFindIpsParams()
		ifr := &models.V1IPFindRequest{
			IPAddress:        helper.ViperString("ipaddress"),
			ProjectID:        helper.ViperString("project"),
			ParentPrefixCidr: helper.ViperString("prefix"),
			NetworkID:        helper.ViperString("network"),
			MachineID:        helper.ViperString("machineid"),
		}
		params.SetBody(ifr)
		resp, err := cloud.IP.FindIps(params, cloud.Auth)
		if err != nil {
			switch e := err.(type) {
			case *ip.FindIpsDefault:
				return output.HTTPError(e.Payload)
			default:
				return output.UnconventionalError(err)
			}
		}
		return printer.Print(resp.Payload)
	}
	resp, err := cloud.IP.ListIps(nil, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *ip.ListIpsDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	return printer.Print(resp.Payload)
}

func ipStatic(args []string) error {
	ipAddress, err := getIPFromArgs(args)
	if err != nil {
		return err
	}

	params := ip.NewUpdateIPParams()
	iur := &models.V1IPUpdateRequest{
		Ipaddress: &ipAddress,
		Type:      "static",
		Tags:      []string{},
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
	resp, err := cloud.IP.UpdateIP(params, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *ip.UpdateIPDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	return printer.Print(resp.Payload)
}

func ipAllocate(args []string) error {
	params := ip.NewAllocateIPParams()
	iar := &models.V1IPAllocateRequest{
		Name:        *helper.ViperString("name"),
		Description: *helper.ViperString("description"),
		Type:        "static",
		Networkid:   helper.ViperString("network"),
		Projectid:   helper.ViperString("project"),
		Tags:        helper.ViperStringSlice("tags"),
	}

	if helper.ViperString("specific-ip") != nil {
		iar.Ipaddress = helper.ViperString("specific-ip")
	}

	if !viper.GetBool("yes-i-really-mean-it") {
		fmt.Println("Allocating a static IP address costs additional money because addresses are limited. The IP address is not cleaned up automatically on cluster deletion. The address will be accounted until the IP address gets freed manually from your side.")
		err := helper.Prompt("Are you sure? (y/n)", "y")
		if err != nil {
			return err
		}
	}

	params.SetBody(iar)
	resp, err := cloud.IP.AllocateIP(params, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *ip.AllocateIPDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	return printer.Print(resp.Payload)
}

func ipFree(args []string) error {
	ipAddress, err := getIPFromArgs(args)
	if err != nil {
		return err
	}

	params := ip.NewFreeIPParams()
	params.SetIP(ipAddress)
	resp, err := cloud.IP.FreeIP(params, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *ip.FreeIPDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	return printer.Print(resp.Payload)
}

func getIPFromArgs(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("no ip given")
	}

	ipAddress := args[0]
	params := ip.NewGetIPParams()
	params.SetIP(ipAddress)

	_, err := cloud.IP.GetIP(params, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *ip.GetIPDefault:
			return "", output.HTTPError(e.Payload)
		default:
			return "", output.UnconventionalError(err)
		}
	}
	return ipAddress, nil
}
