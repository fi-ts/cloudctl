package cmd

import (
	"fmt"

	"github.com/metal-stack/cloud-go/api/client/ip"

	"git.f-i-ts.de/cloud-native/cloudctl/cmd/helper"
	output "git.f-i-ts.de/cloud-native/cloudctl/cmd/output"
	"github.com/metal-stack/cloud-go/api/models"
	"github.com/spf13/cobra"
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
	ipDeleteCmd = &cobra.Command{
		Use:     "delete <ip>",
		Aliases: []string{"rm", "destroy", "remove", "free"},
		Short:   "delete an ip",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ipDelete(args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	rootCmd.AddCommand(ipCmd)

	ipCmd.AddCommand(ipListCmd)
	ipCmd.AddCommand(ipStaticCmd)
	ipCmd.AddCommand(ipDeleteCmd)

	ipListCmd.Flags().StringP("ipaddress", "", "", "ipaddress to filter [optional]")
	ipListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	ipListCmd.Flags().StringP("prefix", "", "", "prefx to filter [optional]")
	ipListCmd.Flags().StringP("machineid", "", "", "machineid to filter [optional]")
	ipListCmd.Flags().StringP("network", "", "", "network to filter [optional]")
	ipListCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion()
	})

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

func ipDelete(args []string) error {
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
