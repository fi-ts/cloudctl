package cmd

import (
	"fmt"

	"github.com/fi-ts/cloud-go/api/client/volume"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newVolumeCmd(c *config) *cobra.Command {
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
			return c.volumeFind()
		},
	}
	volumeDescribeCmd := &cobra.Command{
		Use:   "describe <volume>",
		Short: "describes a volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.volumeDescribe(args)
		},
		ValidArgsFunction: c.comp.VolumeListCompletion,
	}
	volumeDeleteCmd := &cobra.Command{
		Use:     "delete <volume>",
		Aliases: []string{"destroy", "rm", "remove"},
		Short:   "delete a volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.volumeDelete(args)
		},
		ValidArgsFunction: c.comp.VolumeListCompletion,
	}
	volumeSetQoSCmd := &cobra.Command{
		Use:     "set-qos <volume>",
		Aliases: []string{"set-qos"},
		Short:   "sets the qos policy of the volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.volumeSetQoS(args)
		},
		ValidArgsFunction: c.comp.VolumeListCompletion,
	}
	volumeManifestCmd := &cobra.Command{
		Use:   "manifest <volume>",
		Short: "print a manifest for a volume",
		Long:  "this is only useful for volumes which are not used in any k8s cluster. With the PersistentVolumeClaim given you can reuse it in a new cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.volumeManifest(args)
		},
		ValidArgsFunction: c.comp.VolumeListCompletion,
	}
	volumeEncryptionSecretManifestCmd := &cobra.Command{
		Use:   "encryption-secret-manifest",
		Short: "print a secret manifest for volume encryption",
		Long:  "This command helps you with the creation of a secret to encrypt volumes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.volumeEncryptionSecretManifest()
		},
	}
	volumeClusterInfoCmd := &cobra.Command{
		Use:   "clusterinfo",
		Short: "show storage cluster infos",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.volumeClusterInfo()
		},
	}

	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "manage snapshots",
		Long:  "list/find/delete snapshot",
	}
	snapshotListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list snapshot",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.snapshotFind()
		},
	}
	snapshotDescribeCmd := &cobra.Command{
		Use:   "describe <snapshot>",
		Short: "describes a snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.snapshotDescribe(args)
		},
	}
	snapshotDeleteCmd := &cobra.Command{
		Use:     "delete <snapshot>",
		Aliases: []string{"destroy", "rm", "remove"},
		Short:   "delete a snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.snapshotDelete(args)
		},
	}

	qosCmd := &cobra.Command{
		Use:   "qos",
		Short: "manage qos policies",
		Long:  "list qos policies",
	}
	qosListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list qos policies",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.listQoSPolicies()
		},
	}

	snapshotListCmd.Flags().StringP("snapshotid", "", "", "snapshotid to filter [optional]")
	snapshotListCmd.Flags().StringP("project", "", "", "project to filter")
	snapshotListCmd.Flags().StringP("name", "", "", "name to filter")
	snapshotListCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")

	snapshotDescribeCmd.Flags().StringP("project", "", "", "project to filter")
	snapshotDeleteCmd.Flags().StringP("project", "", "", "project to filter")

	genericcli.Must(snapshotListCmd.MarkFlagRequired("project"))
	genericcli.Must(snapshotDescribeCmd.MarkFlagRequired("project"))
	genericcli.Must(snapshotDeleteCmd.MarkFlagRequired("project"))

	genericcli.Must(snapshotListCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	genericcli.Must(snapshotListCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))

	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotDescribeCmd)
	snapshotCmd.AddCommand(snapshotDeleteCmd)
	volumeCmd.AddCommand(snapshotCmd)

	qosCmd.AddCommand(qosListCmd)
	volumeCmd.AddCommand(qosCmd)

	volumeCmd.AddCommand(volumeListCmd)
	volumeCmd.AddCommand(volumeDeleteCmd)
	volumeCmd.AddCommand(volumeDescribeCmd)
	volumeCmd.AddCommand(volumeManifestCmd)
	volumeCmd.AddCommand(volumeEncryptionSecretManifestCmd)
	volumeCmd.AddCommand(volumeClusterInfoCmd)
	volumeCmd.AddCommand(volumeSetQoSCmd)

	volumeListCmd.Flags().StringP("volumeid", "", "", "volumeid to filter [optional]")
	volumeListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	volumeListCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")
	volumeListCmd.Flags().StringP("tenant", "", "", "tenant to filter [optional]")
	volumeListCmd.Flags().Bool("only-unbound", false, "show only unbound volumes that are not connected to any hosts, pv may be still present. [optional]")

	genericcli.Must(volumeListCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	genericcli.Must(volumeListCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	genericcli.Must(volumeListCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))

	volumeManifestCmd.Flags().StringP("name", "", "restored-pv", "name of the PersistentVolume")
	volumeManifestCmd.Flags().StringP("namespace", "", "default", "namespace for the PersistentVolume")
	volumeManifestCmd.Flags().StringP("storage-class-name", "", "", "the storage class for the PersistentVolume")

	volumeEncryptionSecretManifestCmd.Flags().StringP("namespace", "", "default", "namespace for the PersistentVolume encryption secret")
	volumeEncryptionSecretManifestCmd.Flags().StringP("passphrase", "", "please-change-me", "passphrase for the PersistentVolume encryption")

	volumeClusterInfoCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")
	genericcli.Must(volumeClusterInfoCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))

	volumeSetQoSCmd.Flags().StringP("qos-id", "", "", "the id of the new qos policy of the volume")
	volumeSetQoSCmd.Flags().StringP("qos-name", "", "", "the name of the new qos policy of the volume")

	genericcli.Must(volumeSetQoSCmd.RegisterFlagCompletionFunc("qos-id", c.comp.PolicyIDListCompletion))
	genericcli.Must(volumeSetQoSCmd.RegisterFlagCompletionFunc("qos-name", c.comp.PolicyNameListCompletion))

	return volumeCmd
}

func (c *config) volumeFind() error {
	if helper.AtLeastOneViperStringFlagGiven("volumeid", "project", "partition", "tenant") {
		params := volume.NewFindVolumesParams()
		ifr := &models.V1VolumeFindRequest{
			VolumeID:    helper.ViperString("volumeid"),
			ProjectID:   helper.ViperString("project"),
			PartitionID: helper.ViperString("partition"),
			TenantID:    helper.ViperString("tenant"),
		}
		params.SetBody(ifr)
		resp, err := c.cloud.Volume.FindVolumes(params, nil)
		if err != nil {
			return err
		}
		volumes := resp.Payload
		if viper.GetBool("only-unbound") {
			volumes = onlyUnboundVolumes(volumes)
		}
		return c.listPrinter.Print(volumes)
	}
	resp, err := c.cloud.Volume.ListVolumes(nil, nil)
	if err != nil {
		return err
	}
	volumes := resp.Payload
	if viper.GetBool("only-unbound") {
		volumes = onlyUnboundVolumes(volumes)
	}
	return c.listPrinter.Print(volumes)
}

func onlyUnboundVolumes(volumes []*models.V1VolumeResponse) (result []*models.V1VolumeResponse) {
	for _, v := range volumes {
		if len(v.ConnectedHosts) > 0 {
			continue
		}
		v := v
		result = append(result, v)
	}
	return result
}

func (c *config) volumeDescribe(args []string) error {
	vol, err := c.getVolumeFromArgs(args)
	if err != nil {
		return err
	}
	return c.listPrinter.Print(vol)
}

func (c *config) volumeDelete(args []string) error {
	vol, err := c.getVolumeFromArgs(args)
	if err != nil {
		return err
	}

	if len(vol.ConnectedHosts) > 0 {
		return fmt.Errorf("volume is still connected to this node:%s", vol.ConnectedHosts)
	}

	if !viper.GetBool("yes-i-really-mean-it") {
		fmt.Printf(`
delete volume: %q, all data will be lost forever.
If used in cronjob for example, volume might not be connected now, but required at a later point in time.
`, *vol.VolumeID)
		err = helper.Prompt("Are you sure? (y/n)", "y")
		if err != nil {
			return err
		}
	}

	resp, err := c.cloud.Volume.DeleteVolume(volume.NewDeleteVolumeParams().WithID(*vol.VolumeID), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *config) volumeSetQoS(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}
	policyId := helper.ViperString("qos-id")
	policyName := helper.ViperString("qos-name")

	if pointer.SafeDeref(policyId) == "" && pointer.SafeDeref(policyName) == "" {
		return fmt.Errorf("either qos-id or qos-name must be specified")
	}
	if pointer.SafeDeref(policyId) != "" && pointer.SafeDeref(policyName) != "" {
		return fmt.Errorf("either qos-id or qos-name must be specified, not both")
	}

	params := volume.NewSetVolumeQoSPolicyParams().
		WithID(id).
		WithBody(&models.V1VolumeSetQoSPolicyRequest{
			QoSPolicyID:   policyId,
			QoSPolicyName: policyName,
		})

	resp, err := c.cloud.Volume.SetVolumeQoSPolicy(params, nil)
	if err != nil {
		return err
	}
	return c.listPrinter.Print(resp.Payload)
}

func (c *config) volumeClusterInfo() error {
	params := volume.NewClusterInfoParams().WithPartitionid(helper.ViperString("partition"))
	resp, err := c.cloud.Volume.ClusterInfo(params, nil)
	if err != nil {
		return err
	}
	return c.listPrinter.Print(resp.Payload)
}

func (c *config) volumeManifest(args []string) error {
	volume, err := c.getVolumeFromArgs(args)
	if err != nil {
		return err
	}

	var (
		name      = viper.GetString("name")
		namespace = viper.GetString("namespace")
		sc        = viper.GetString("storage-class-name")
	)

	if sc == "" {
		return fmt.Errorf("a storage class name must be provided, this cannot be derived automatically")
	}

	return output.VolumeManifest(*volume, name, namespace, sc)
}

func (c *config) volumeEncryptionSecretManifest() error {
	namespace := viper.GetString("namespace")
	passphrase := viper.GetString("passphrase")
	return output.VolumeEncryptionSecretManifest(namespace, passphrase)
}

func (c *config) getVolumeFromArgs(args []string) (*models.V1VolumeResponse, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no volume given")
	}

	id := args[0]
	resp, err := c.cloud.Volume.GetVolume(volume.NewGetVolumeParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

// Snapshots

func (c *config) getSnapshotFromArgs(args []string) (*models.V1SnapshotResponse, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no snapshot given")
	}
	id := args[0]
	projectid := helper.ViperString("project")
	resp, err := c.cloud.Volume.GetSnapshot(volume.NewGetSnapshotParams().WithID(id).WithProjectID(projectid), nil)
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *config) snapshotFind() error {
	params := volume.NewFindSnapshotsParams()
	ifr := &models.V1SnapshotFindRequest{
		SnapshotID:  helper.ViperString("snapshotid"),
		ProjectID:   helper.ViperString("project"),
		Name:        helper.ViperString("name"),
		PartitionID: helper.ViperString("partition"),
	}
	params.SetBody(ifr)
	resp, err := c.cloud.Volume.FindSnapshots(params, nil)
	if err != nil {
		return err
	}
	return c.listPrinter.Print(resp.Payload)
}

func (c *config) snapshotDescribe(args []string) error {
	snap, err := c.getSnapshotFromArgs(args)
	if err != nil {
		return err
	}
	return c.listPrinter.Print(snap)
}

func (c *config) snapshotDelete(args []string) error {
	snap, err := c.getSnapshotFromArgs(args)
	if err != nil {
		return err
	}

	if !viper.GetBool("yes-i-really-mean-it") {
		fmt.Printf(`
delete snapshot: %q, all data will be lost forever.
`, *snap.SnapshotID)
		err = helper.Prompt("Are you sure? (y/n)", "y")
		if err != nil {
			return err
		}
	}

	projectid := helper.ViperString("project")
	resp, err := c.cloud.Volume.DeleteSnapshot(volume.NewDeleteSnapshotParams().WithID(*snap.SnapshotID).WithProjectID(projectid), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *config) listQoSPolicies() error {
	resp, err := c.cloud.Volume.ListPolicies(volume.NewListPoliciesParams(), nil)
	if err != nil {
		return err
	}
	return c.listPrinter.Print(resp.Payload)
}
