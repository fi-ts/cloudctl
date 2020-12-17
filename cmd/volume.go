package cmd

import (
	"log"

	"github.com/fi-ts/cloud-go/api/client/volume"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/cmd/output"
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
	volumeDeleteCmd = &cobra.Command{
		Use:     "delete <volume>",
		Aliases: []string{"rm", "destroy", "remove", "delete"},
		Short:   "delete a volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumeDelete(args)
		},
		PreRun: bindPFlags,
	}
	volumePVCmd = &cobra.Command{
		Use:   "pv <volume>",
		Short: "create a static persistenvolumeclaim for a volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumePV(args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	rootCmd.AddCommand(volumeCmd)

	volumeCmd.AddCommand(volumeListCmd)
	volumeCmd.AddCommand(volumeDeleteCmd)
	volumeCmd.AddCommand(volumePVCmd)

	volumeListCmd.Flags().StringP("volumeid", "", "", "volumeid to filter [optional]")
	volumeListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	volumeListCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")

	err := volumeListCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	err = volumeListCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
	params.SetID(volumeID)
	resp, err := cloud.Volume.DeleteVolume(params, cloud.Auth)
	if err != nil {
		return err
	}

	return printer.Print(resp.Payload)
}

func volumePV(args []string) error {
	volumeID := args[0]
	params := volume.NewFindVolumesParams()
	ifr := &models.V1VolumeFindRequest{
		VolumeID: &volumeID,
	}
	params.SetBody(ifr)
	resp, err := cloud.Volume.FindVolumes(params, cloud.Auth)
	if err != nil {
		return err
	}

	return output.PersistenVolume(*resp.Payload[0])
}
