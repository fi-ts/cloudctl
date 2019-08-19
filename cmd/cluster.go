package cmd

import (
	"fmt"

	"git.f-i-ts.de/cloud-native/cloudctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clusterCmd = &cobra.Command{
		Use:   "cluster",
		Short: "manage clusters",
		Long:  "TODO",
	}
	clusterCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterCreate()
		},
		PreRun: bindPFlags,
	}

	clusterListCmd = &cobra.Command{
		Use:   "list",
		Short: "list clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterList()
		},
		PreRun: bindPFlags,
	}
)

func init() {
	clusterCreateCmd.Flags().StringP("name", "", "", "name of the cluster. [required]")
	clusterCreateCmd.Flags().StringP("description", "", "", "description of the cluster. [required]")
	clusterCreateCmd.Flags().StringP("purpose", "", "production", "purpose of the cluster, can be one of production|dev|eval. [required]")
	clusterCreateCmd.Flags().StringP("owner", "", "", "owner of the cluster. [required]")
	clusterCreateCmd.Flags().StringP("partition", "", "nbg-w8101", "partition of the cluster. [required]")
	clusterCreateCmd.Flags().StringP("version", "", "1.14.3", "kubernetes version of the cluster. [required]")
	clusterCreateCmd.Flags().IntP("minsize", "", 1, "minimal workers of the cluster.")
	clusterCreateCmd.Flags().IntP("maxsize", "", 1, "maximal workers of the cluster.")
	clusterCreateCmd.Flags().IntP("maxsurge", "", 1, "max number of workers created during a update of the cluster.")
	clusterCreateCmd.Flags().IntP("maxunavailable", "", 1, "max number of workers that can be unavailable during a update of the cluster.")
	clusterCreateCmd.Flags().StringSlice("labels", []string{}, "labels of the cluster")
	clusterCreateCmd.Flags().StringSlice("external-networks", []string{}, "external networks of the cluster")
	clusterCreateCmd.Flags().BoolP("allowprivileged", "", false, "allow privileged containers the cluster.")

	clusterCreateCmd.MarkFlagRequired("owner")

	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCmd.AddCommand(clusterListCmd)

}

func clusterCreate() error {
	owner := viper.GetString("owner")
	desc := viper.GetString("description")
	purpose := viper.GetString("purpose")
	partition := viper.GetString("partition")
	// FIXME helper and validation
	networks := viper.GetStringSlice("external-networks")
	scr := &api.ShootCreateRequest{
		CreatedBy:            owner,
		Owner:                owner,
		Name:                 viper.GetString("name"),
		Description:          &desc,
		Purpose:              &purpose,
		LoadBalancerProvider: api.DefaultLoadBalancerProvider,
		MachineImage:         api.DefaultMachineImage,
		FirewallImage:        api.DefaultFirewallImage,
		FirewallSize:         api.DefaultFirewallSize,
		Workers: []api.Worker{
			{
				Name:           "default-worker",
				MachineType:    api.DefaultMachineType,
				AutoScalerMin:  viper.GetInt("minsize"),
				AutoScalerMax:  viper.GetInt("maxsize"),
				MaxSurge:       viper.GetInt("maxsurge"),
				MaxUnavailable: viper.GetInt("maxunavailable"),
				VolumeType:     api.DefaultVolumeType,
				VolumeSize:     api.DefaultVolumeSize,
			},
		},
		Kubernetes: api.Kubernetes{
			AllowPrivilegedContainers: viper.GetBool("allowprivileged"),
			Version:                   viper.GetString("version"),
		},
		Maintenance: api.Maintenance{
			AutoUpdate: &api.MaintenanceAutoUpdate{
				KubernetesVersion: false,
				MachineImage:      false,
			},
			TimeWindow: &api.MaintenanceTimeWindow{
				Begin: "220000+0100",
				End:   "233000+0100",
			},
		},
		Networks: networks,
		Zones:    []string{partition},
	}

	shoot, err := gardener.CreateShoot(scr)
	if err != nil {
		return err
	}
	fmt.Printf("Shoot:%s created\n", shoot.GetName())

	return printer.Print(shoot)
}

func clusterList() error {
	// projects, err := client.GardenV1beta1().Projects().List(metav1.ListOptions{})

	// if err != nil {
	// 	return err
	// }
	// for _, project := range projects.Items {
	// 	fmt.Println(project.Name)
	// }
	return nil
}
