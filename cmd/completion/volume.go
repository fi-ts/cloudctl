package completion

import (
	"sort"

	"github.com/fi-ts/cloud-go/api/client/volume"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/spf13/cobra"
)

func (c *Completion) VolumeListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := volume.NewListVolumesParams()
	response, err := c.cloud.Volume.ListVolumes(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var names []string
	for _, v := range response.Payload {
		if v.VolumeID == nil {
			continue
		}
		names = append(names, *v.VolumeID)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveDefault
}

func (c *Completion) SnapshotListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var project *string
	if p, _ := cmd.Flags().GetString("project"); p != "" {
		project = pointer.Pointer(p)
	} else {
		return nil, cobra.ShellCompDirectiveDefault
	}

	response, err := c.cloud.Volume.FindSnapshots(volume.NewFindSnapshotsParams().WithBody(&models.V1SnapshotFindRequest{
		ProjectID: project,
	}), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var names []string
	for _, v := range response.Payload {
		if v.SnapshotID == nil {
			continue
		}
		names = append(names, *v.SnapshotID)
	}
	sort.Strings(names)

	return names, cobra.ShellCompDirectiveDefault
}
