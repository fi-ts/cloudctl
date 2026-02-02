/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// instanceCmd represents the instance command
func newInstanceCmd(c *config) *cobra.Command {
	instanceCmd := &cobra.Command{
		Use:   "instance",
		Short: "manage VM instances",
		Long:  "manage virtual machine instances",
	}
	instanceCreateCmd := &cobra.Command{
		Use:   "create",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Short: "create a new VM instance",
		Long:  "create a new virtual machine instance",
	}

	instanceCreateCmd.Flags().IntP("sockets", "s", 1, "Number of CPUs per socket.")
	instanceCreateCmd.Flags().IntP("CPUs per socket", "c", 1, "Number of CPUs per socket.")
	instanceCreateCmd.Flags().IntP("ram", "r", 2, "RAM for the VM, in GB")

	instanceCmd.AddCommand(instanceCreateCmd)

	return instanceCmd
}
