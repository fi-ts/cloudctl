/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/fi-ts/cloud-go/api/client/instance"
	"github.com/fi-ts/cloud-go/api/models"
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
		Long:  "create a new virtual machine instances",
		RunE: func(cmd *cobra.Command, args []string) error {

			return c.instanceCreate(args[0])
		},
	}
	instanceCmd.AddCommand(instanceCreateCmd)

	return instanceCmd
}

func (c *config) instanceCreate(name string) error {
	params := instance.NewCreateInstanceParams()
	body := &models.V1InstanceCreateRequest{
		Name: &name}
	params.SetBody(body)
	resp, err := c.cloud.Instance.CreateInstance(params, nil)
	if err != nil {
		return err
	}
	return c.listPrinter.Print(resp)
}
