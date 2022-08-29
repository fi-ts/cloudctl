package cmd

import (
	"fmt"

	"github.com/fi-ts/cloud-go/api/client/volume"
	"github.com/fi-ts/cloudctl/cmd/sorters"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type volumeCmd struct {
	*config
}

func newVolumeCmd(c *config) *cobra.Command {
	w := volumeCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[any, any, *models.V1VolumeResponse]{
		BinaryName: binaryName,
		GenericCLI: genericcli.NewGenericCLI[any, any, *models.V1VolumeResponse](w).WithFS(c.fs),
		OnlyCmds: genericcli.OnlyCmds(
			genericcli.ListCmd,
			genericcli.DescribeCmd,
			genericcli.DeleteCmd,
		),
		Singular:        "volume",
		Plural:          "volumes",
		Description:     "manage persistent cloud storage volumes",
		Sorter:          sorters.VolumeSorter(),
		ValidArgsFn:     c.comp.VolumeListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("id", "", "", "volumeid to filter [optional]")
			cmd.Flags().StringP("project", "", "", "project to filter [optional]")
			cmd.Flags().StringP("partition", "", "", "partition to filter [optional]")
			cmd.Flags().StringP("tenant", "", "", "tenant to filter [optional]")
			cmd.Flags().Bool("only-unbound", false, "show only unbound volumes that are not connected to any hosts, pv may be still present. [optional]")

			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			must(cmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
			must(cmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
		},
	}

	manifestCmd := &cobra.Command{
		Use:   "manifest <volume>",
		Short: "print a manifest for a volume",
		Long:  "this is only useful for volumes which are not used in any k8s cluster. With the PersistenVolumeClaim given you can reuse it in a new cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.volumeManifest(args)
		},
		ValidArgsFunction: c.comp.VolumeListCompletion,
	}

	encryptionSecretManifestCmd := &cobra.Command{
		Use:   "encryption-secret-manifest",
		Short: "print a secret manifest for volume encryption",
		Long:  "This command helps you with the creation of a secret to encrypt volumes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.volumeEncryptionSecretManifest()
		},
	}

	clusterInfoCmd := &cobra.Command{
		Use:   "clusterinfo",
		Short: "show storage cluster infos",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.volumeClusterInfo()
		},
	}

	manifestCmd.Flags().StringP("name", "", "restored-pv", "name of the PersistentVolume")
	manifestCmd.Flags().StringP("namespace", "", "default", "namespace for the PersistentVolume")

	encryptionSecretManifestCmd.Flags().StringP("namespace", "", "default", "namespace for the PersistentVolume encryption secret")
	encryptionSecretManifestCmd.Flags().StringP("passphrase", "", "please-change-me", "passphrase for the PersistentVolume encryption")

	clusterInfoCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")
	must(clusterInfoCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))

	return genericcli.NewCmds(cmdsConfig, manifestCmd, encryptionSecretManifestCmd, clusterInfoCmd, newSnapshotCmd(c))
}

func (c volumeCmd) Get(id string) (*models.V1VolumeResponse, error) {
	resp, err := c.client.Volume.GetVolume(volume.NewGetVolumeParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c volumeCmd) List() ([]*models.V1VolumeResponse, error) {
	resp, err := c.client.Volume.FindVolumes(volume.NewFindVolumesParams().WithBody(&models.V1VolumeFindRequest{
		VolumeID:    pointer.PointerOrNil(viper.GetString("id")),
		ProjectID:   pointer.PointerOrNil(viper.GetString("project")),
		PartitionID: pointer.PointerOrNil(viper.GetString("partition")),
		TenantID:    pointer.PointerOrNil(viper.GetString("tenant")),
	}), nil)
	if err != nil {
		return nil, err
	}

	volumes := resp.Payload
	if viper.GetBool("only-unbound") {
		volumes = onlyUnboundVolumes(volumes)
	}

	return volumes, nil
}

func (c volumeCmd) Delete(id string) (*models.V1VolumeResponse, error) {
	vol, err := c.Get(id)
	if err != nil {
		return nil, err
	}

	if len(vol.ConnectedHosts) > 0 {
		return nil, fmt.Errorf("volume is still connected to this node:%s", vol.ConnectedHosts)
	}

	if !viper.GetBool("yes-i-really-mean-it") {
		fmt.Fprintf(c.out, `
delete volume: %q, all data will be lost forever.
If used in cronjob for example, volume might not be connected now, but required at a later point in time.
`, *vol.VolumeID)
		err = helper.Prompt("Are you sure? (y/n)", "y")
		if err != nil {
			return nil, err
		}
	}

	resp, err := c.client.Volume.DeleteVolume(volume.NewDeleteVolumeParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c volumeCmd) Create(rq any) (*models.V1VolumeResponse, error) {
	return nil, fmt.Errorf("not implemented for volumes, managed through Kubernetes")
}

func (c volumeCmd) Update(rq any) (*models.V1VolumeResponse, error) {
	return nil, fmt.Errorf("not implemented for volumes, managed through Kubernetes")
}

func (c volumeCmd) Convert(r *models.V1VolumeResponse) (string, any, any, error) {
	if r.VolumeID == nil {
		return "", nil, nil, fmt.Errorf("id is nil")
	}
	return *r.VolumeID, nil, nil, nil
}

func (c volumeCmd) ToUpdate(r *models.V1VolumeResponse) (any, error) {
	return nil, fmt.Errorf("not implemented for volumes, managed through Kubernetes")
}

func onlyUnboundVolumes(volumes []*models.V1VolumeResponse) (result []*models.V1VolumeResponse) {
	// TODO: this filter should be moved to the server
	for _, v := range volumes {
		if len(v.ConnectedHosts) > 0 {
			continue
		}
		v := v
		result = append(result, v)
	}
	return result
}

func (c volumeCmd) volumeClusterInfo() error {
	resp, err := c.client.Volume.ClusterInfo(volume.NewClusterInfoParams().WithPartitionid(helper.ViperString("partition")), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c volumeCmd) volumeManifest(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	volume, err := c.Get(id)
	if err != nil {
		return err
	}

	name := viper.GetString("name")
	namespace := viper.GetString("namespace")

	return output.VolumeManifest(*volume, name, namespace)
}

func (c volumeCmd) volumeEncryptionSecretManifest() error {
	namespace := viper.GetString("namespace")
	passphrase := viper.GetString("passphrase")
	return output.VolumeEncryptionSecretManifest(namespace, passphrase)
}

// Snapshots

type snapshotCmd struct {
	*config
}

func newSnapshotCmd(c *config) *cobra.Command {
	w := snapshotCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[any, any, *models.V1SnapshotResponse]{
		BinaryName: binaryName,
		GenericCLI: genericcli.NewGenericCLI[any, any, *models.V1SnapshotResponse](w).WithFS(c.fs),
		OnlyCmds: genericcli.OnlyCmds(
			genericcli.ListCmd,
			genericcli.DescribeCmd,
			genericcli.DeleteCmd,
		),
		Singular:        "snapshot",
		Plural:          "snapshots",
		Description:     "manage persistent cloud storage volume snapshots",
		Sorter:          sorters.SnapshotSorter(),
		ValidArgsFn:     c.comp.SnapshotListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("id", "", "", "snapshotid to filter [optional]")
			cmd.Flags().StringP("project", "", "", "project to filter")
			cmd.Flags().StringP("name", "", "", "name to filter")
			cmd.Flags().StringP("partition", "", "", "partition to filter [optional]")

			must(cmd.MarkFlagRequired("project"))

			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			must(cmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
		},
		DeleteCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("project", "", "", "project to filter")
			must(cmd.MarkFlagRequired("project"))
		},
		DescribeCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("project", "", "", "project to filter")
			must(cmd.MarkFlagRequired("project"))
		},
	}

	return genericcli.NewCmds(cmdsConfig)
}

func (c snapshotCmd) Get(id string) (*models.V1SnapshotResponse, error) {
	resp, err := c.client.Volume.GetSnapshot(volume.NewGetSnapshotParams().WithID(id).WithProjectID(pointer.PointerOrNil(viper.GetString("project"))), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c snapshotCmd) List() ([]*models.V1SnapshotResponse, error) {
	resp, err := c.client.Volume.FindSnapshots(volume.NewFindSnapshotsParams().WithBody(&models.V1SnapshotFindRequest{
		SnapshotID:  pointer.PointerOrNil(viper.GetString("id")),
		ProjectID:   pointer.PointerOrNil(viper.GetString("project")),
		Name:        pointer.PointerOrNil(viper.GetString("name")),
		PartitionID: pointer.PointerOrNil(viper.GetString("partition")),
	}), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c snapshotCmd) Delete(id string) (*models.V1SnapshotResponse, error) {
	if !viper.GetBool("yes-i-really-mean-it") {
		fmt.Fprintf(c.out, `
delete snapshot: %q, all data will be lost forever.
`, id)
		err := helper.Prompt("Are you sure? (y/n)", "y")
		if err != nil {
			return nil, err
		}
	}

	resp, err := c.client.Volume.DeleteSnapshot(volume.NewDeleteSnapshotParams().WithID(id).WithProjectID(pointer.PointerOrNil(viper.GetString("project"))), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c snapshotCmd) Create(rq any) (*models.V1SnapshotResponse, error) {
	return nil, fmt.Errorf("not implemented for volumes, managed through Kubernetes")
}

func (c snapshotCmd) Update(rq any) (*models.V1SnapshotResponse, error) {
	return nil, fmt.Errorf("not implemented for volumes, managed through Kubernetes")
}

func (c snapshotCmd) Convert(r *models.V1SnapshotResponse) (string, any, any, error) {
	if r.SnapshotID == nil {
		return "", nil, nil, fmt.Errorf("id is nil")
	}
	return *r.SnapshotID, nil, nil, nil
}
