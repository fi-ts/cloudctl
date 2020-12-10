package cmd

import (
	"log"

	"github.com/fi-ts/cloud-go/api/client/volume"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/spf13/cobra"
)

var (
	volumeCmd = &cobra.Command{
		Use:   "volume",
		Short: "manage volume",
		Long:  "list/find/delete pvc volumes",
	}
	volumeListCmd = &cobra.Command{
		Use:     "list",
		Short:   "list volume",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumeFind()
		},
		PreRun: bindPFlags,
	}
	volumeFindCmd = &cobra.Command{
		Use:   "find",
		Short: "find volumes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumeFind()
		},
		PreRun: bindPFlags,
	}
	volumeDeleteCmd = &cobra.Command{
		Use:     "delete <volume>",
		Aliases: []string{"rm", "destroy", "remove", "delete"},
		Short:   "delete a volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumeDelete(args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	rootCmd.AddCommand(volumeCmd)

	volumeCmd.AddCommand(volumeListCmd)
	volumeCmd.AddCommand(volumeFindCmd)
	volumeCmd.AddCommand(volumeDeleteCmd)

	volumeFindCmd.Flags().StringP("volumeid", "", "", "volumeid to filter [optional]")
	volumeFindCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	volumeFindCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")

	volumeDeleteCmd.Flags().StringP("volumeid", "", "", "volumeid to delete [optional]")

	err := volumeFindCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	err = volumeFindCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

func volumeFind() error {
	if helper.AtLeastOneViperStringFlagGiven("volumeid", "project", "partition") {
		params := volume.NewFindVolumesParams()
		ifr := &models.V1VolumeFindRequest{
			VolumeID:    helper.ViperString("volumeid"),
			ProjectID:   helper.ViperString("project"),
			PartitionID: helper.ViperString("partition"),
		}
		params.SetBody(ifr)
		resp, err := cloud.Volume.FindVolumes(params, cloud.Auth)
		if err != nil {
			return err
		}
		return printer.Print(resp.Payload)
	}
	resp, err := cloud.Volume.ListVolumes(nil, cloud.Auth)
	if err != nil {
		return err
	}
	return printer.Print(resp.Payload)
}

func volumeDelete(args []string) error {
	volumeID := args[0]
	params := &volume.DeleteVolumeParams{}
	params.SetVolume(volumeID)
	resp, err := cloud.Volume.DeleteVolume(params, cloud.Auth)
	if err != nil {
		return err
	}

	return printer.Print(resp.Payload)
}
