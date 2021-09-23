package cmd

import (
	"fmt"

	"github.com/fi-ts/cloud-go/api/client/volume"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newVolumeCmd() *cobra.Command {
	volumeCmd := &cobra.Command{
		Use:   "volume",
		Short: "manage volume",
		Long:  "list/find/delete pvc volumes",
	}
	volumeListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list volume",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumeFind()
		},
		PreRun: bindPFlags,
	}
	volumeDescribeCmd := &cobra.Command{
		Use:   "describe <volume>",
		Short: "describes a volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumeDescribe(args)
		},
		ValidArgsFunction: comp.VolumeListCompletion,
		PreRun:            bindPFlags,
	}
	volumeDeleteCmd := &cobra.Command{
		Use:     "delete <volume>",
		Aliases: []string{"rm", "destroy", "remove", "delete"},
		Short:   "delete a volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumeDelete(args)
		},
		ValidArgsFunction: comp.VolumeListCompletion,
		PreRun:            bindPFlags,
	}
	volumeManifestCmd := &cobra.Command{
		Use:   "manifest <volume>",
		Short: "print a manifest for a volume",
		Long:  "this is only useful for volumes which are not used in any k8s cluster. With the PersistenVolumeClaim given you can reuse it in a new cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumeManifest(args)
		},
		ValidArgsFunction: comp.VolumeListCompletion,
		PreRun:            bindPFlags,
	}
	volumeClusterInfoCmd := &cobra.Command{
		Use:   "clusterinfo",
		Short: "show storage cluster infos",
		RunE: func(cmd *cobra.Command, args []string) error {
			return volumeClusterInfo()
		},
		PreRun: bindPFlags,
	}

	volumeCmd.AddCommand(volumeListCmd)
	volumeCmd.AddCommand(volumeDeleteCmd)
	volumeCmd.AddCommand(volumeDescribeCmd)
	volumeCmd.AddCommand(volumeManifestCmd)
	volumeCmd.AddCommand(volumeClusterInfoCmd)

	volumeListCmd.Flags().StringP("volumeid", "", "", "volumeid to filter [optional]")
	volumeListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	volumeListCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")

	volumeManifestCmd.Flags().StringP("name", "", "restored-pv", "name of the PersistentVolume")
	volumeManifestCmd.Flags().StringP("namespace", "", "default", "namespace for the PersistentVolume")

	must(volumeListCmd.RegisterFlagCompletionFunc("project", comp.ProjectListCompletion))
	must(volumeListCmd.RegisterFlagCompletionFunc("partition", comp.PartitionListCompletion))

	return volumeCmd
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
		resp, err := cloud.Volume.FindVolumes(params, nil)
		if err != nil {
			return err
		}
		return printer.Print(resp.Payload)
	}
	resp, err := cloud.Volume.ListVolumes(nil, nil)
	if err != nil {
		return err
	}
	return printer.Print(resp.Payload)
}

func volumeDescribe(args []string) error {
	vol, err := getVolumeFromArgs(args)
	if err != nil {
		return err
	}

	resp, err := cloud.Volume.GetVolume(volume.NewGetVolumeParams().WithID(*vol.VolumeID), nil)
	if err != nil {
		return err
	}

	return output.YAMLPrinter{}.Print(resp.Payload)
}

func volumeDelete(args []string) error {
	vol, err := getVolumeFromArgs(args)
	if err != nil {
		return err
	}

	resp, err := cloud.Volume.DeleteVolume(volume.NewDeleteVolumeParams().WithID(*vol.VolumeID), nil)
	if err != nil {
		return err
	}

	return printer.Print(resp.Payload)
}

func volumeClusterInfo() error {
	params := volume.NewClusterInfoParams()
	resp, err := cloud.Volume.ClusterInfo(params, nil)
	if err != nil {
		return err
	}
	return printer.Print(resp.Payload)
}

func volumeManifest(args []string) error {
	volume, err := getVolumeFromArgs(args)
	if err != nil {
		return err
	}
	name := viper.GetString("name")
	namespace := viper.GetString("namespace")

	return output.VolumeManifest(*volume, name, namespace)
}
func getVolumeFromArgs(args []string) (*models.V1VolumeResponse, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no volume given")
	}

	volumeID := args[0]
	params := volume.NewFindVolumesParams()
	ifr := &models.V1VolumeFindRequest{
		VolumeID: &volumeID,
	}
	params.SetBody(ifr)
	resp, err := cloud.Volume.FindVolumes(params, nil)
	if err != nil {
		return nil, err
	}
	if len(resp.Payload) < 1 {
		return nil, fmt.Errorf("no volume for id:%s found", volumeID)
	}
	if len(resp.Payload) > 1 {
		return nil, fmt.Errorf("more than one volume for id:%s found", volumeID)
	}
	return resp.Payload[0], nil
}
