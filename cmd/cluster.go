package cmd

import (
	"fmt"

	"git.f-i-ts.de/cloud-native/cloudctl/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	clusterCreateCmd.Flags().StringP("owner", "", "", "owner of the cluster. [required]")
	clusterCreateCmd.Flags().StringP("partition", "", "", "partition of the cluster. [required]")
	clusterCreateCmd.Flags().StringSlice("labels", []string{}, "labels of the cluster")
	clusterCreateCmd.Flags().StringSlice("external-networks", []string{}, "external networks of the cluster")

	clusterCreateCmd.MarkFlagRequired("owner")

	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCmd.AddCommand(clusterListCmd)

}

func clusterCreate() error {
	// spec:
	// namespace: garden-<cluster-id>
	// createdBy:
	//     apiGroup: rbac.authorization.k8s.io
	//     kind: User
	//     name: heinz.schenk@f-i-ts.de
	// members:
	// - apiGroup: rbac.authorization.k8s.io
	//     kind: User
	//     name: heinz.schenk@f-i-ts.de
	// owner:
	//     apiGroup: rbac.authorization.k8s.io
	//     kind: User
	//     name: heinz.schenk@f-i-ts.de

	owner := viper.GetString("owner")

	project, err := pkg.CreateProject(client, owner)
	if err != nil {
		return err
	}
	fmt.Printf("Project:%s Namespace:%s UID: %s created\n", project.GetName(), project.GetNamespace(), project.GetUID())

	sb, err := pkg.CreateSecretBinding(client, project)
	if err != nil {
		return err
	}
	fmt.Printf("SecretBinding:%s created\n", sb.GetName())

	shoot, err := pkg.CreateShoot(client, project, sb, "production", "1.14.3")
	if err != nil {
		return err
	}
	fmt.Printf("Shoot:%s created\n", shoot.GetName())

	return nil
}

func clusterList() error {
	projects, err := client.GardenV1beta1().Projects().List(metav1.ListOptions{})

	if err != nil {
		return err
	}
	for _, project := range projects.Items {
		fmt.Println(project.Name)
	}
	return nil
}
