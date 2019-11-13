package cmd

import (
	"encoding/base64"
	"fmt"
	"log"

	"git.f-i-ts.de/cloud-native/cloudctl/api/client/cluster"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
	"git.f-i-ts.de/cloud-native/cloudctl/cmd/helper"
	output "git.f-i-ts.de/cloud-native/cloudctl/cmd/output"

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
		Use:     "list",
		Short:   "list clusters",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterList()
		},
		PreRun: bindPFlags,
	}
	clusterDeleteCmd = &cobra.Command{
		Use:     "delete <uid>",
		Short:   "delete a cluster",
		Aliases: []string{"rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterDelete(args)
		},
		PreRun: bindPFlags,
	}
	clusterDescribeCmd = &cobra.Command{
		Use:   "describe <uid>",
		Short: "describe a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterDescribe(args)
		},
		PreRun: bindPFlags,
	}
	clusterCredentialsCmd = &cobra.Command{
		Use:   "credentials <uid>",
		Short: "get cluster credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterCredentials(args)
		},
		PreRun: bindPFlags,
	}
	clusterSSHKeyPairCmd = &cobra.Command{
		Use:   "sshkeypair <uid>",
		Short: "get cluster sshkeypair",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterSSHKeyPair(args)
		},
		PreRun: bindPFlags,
	}
	clusterReconcileCmd = &cobra.Command{
		Use:   "reconcile <uid>",
		Short: "trigger cluster reconciliation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return reconcileCluster(args)
		},
		PreRun: bindPFlags,
	}
	clusterUpdateCmd = &cobra.Command{
		Use:   "update <uid>",
		Short: "update a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateCluster(args)
		},
		PreRun: bindPFlags,
	}
	clusterInputsCmd = &cobra.Command{
		Use:   "inputs",
		Short: "get possible cluster inputs like k8s versions, etc.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterInputs()
		},
		PreRun: bindPFlags,
	}
)

func init() {
	clusterCreateCmd.Flags().String("name", "", "name of the cluster, max 10 characters. [required]")
	clusterCreateCmd.Flags().String("description", "", "description of the cluster. [required]")
	clusterCreateCmd.Flags().String("project", "", "project where this cluster should belong to. [required]")
	clusterCreateCmd.Flags().String("partition", "nbg-w8101", "partition of the cluster. [required]")
	clusterCreateCmd.Flags().String("version", "1.14.3", "kubernetes version of the cluster. [required]")
	clusterCreateCmd.Flags().Int32("minsize", 1, "minimal workers of the cluster.")
	clusterCreateCmd.Flags().Int32("maxsize", 1, "maximal workers of the cluster.")
	clusterCreateCmd.Flags().String("maxsurge", "1", "max number (e.g. 1) or percentage (e.g. 10%) of workers created during a update of the cluster.")
	clusterCreateCmd.Flags().String("maxunavailable", "1", "max number (e.g. 1) or percentage (e.g. 10%) of workers that can be unavailable during a update of the cluster.")
	clusterCreateCmd.Flags().StringSlice("labels", []string{}, "labels of the cluster")
	clusterCreateCmd.Flags().StringSlice("external-networks", []string{"internet"}, "external networks of the cluster, can be internet,mpls")
	clusterCreateCmd.Flags().BoolP("allowprivileged", "", false, "allow privileged containers the cluster.")
	clusterCreateCmd.Flags().BoolP("defaultingress", "", true, "deploy a default ingress controller")

	err := clusterCreateCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.MarkFlagRequired("project")
	if err != nil {
		log.Fatal(err.Error())
	}

	clusterListCmd.Flags().String("project", "", "show clusters of given project")
	clusterListCmd.Flags().String("partition", "", "show clusters in partition")
	clusterListCmd.Flags().String("tenant", "", "show clusters of given tenant")

	clusterUpdateCmd.Flags().Int32("minsize", 0, "minimal workers of the cluster.")
	clusterUpdateCmd.Flags().Int32("maxsize", 0, "maximal workers of the cluster.")
	clusterUpdateCmd.Flags().String("version", "", "kubernetes version of the cluster.")

	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCmd.AddCommand(clusterListCmd)
	clusterCmd.AddCommand(clusterCredentialsCmd)
	clusterCmd.AddCommand(clusterDeleteCmd)
	clusterCmd.AddCommand(clusterDescribeCmd)
	clusterCmd.AddCommand(clusterInputsCmd)
	clusterCmd.AddCommand(clusterReconcileCmd)
	clusterCmd.AddCommand(clusterSSHKeyPairCmd)
	clusterCmd.AddCommand(clusterUpdateCmd)
}

func clusterCreate() error {
	name := viper.GetString("name")
	desc := viper.GetString("description")
	partition := viper.GetString("partition")
	project := viper.GetString("project")

	minsize := viper.GetInt32("minsize")
	maxsize := viper.GetInt32("maxsize")
	maxsurge := viper.GetString("maxsurge")
	maxunavailable := viper.GetString("maxunavailable")

	allowprivileged := viper.GetBool("allowprivileged")
	version := viper.GetString("version")
	defaultingress := viper.GetBool("defaultingress")

	// FIXME helper and validation
	networks := viper.GetStringSlice("external-networks")
	autoUpdateKubernetes := false
	autoUpdateMachineImage := false
	maintenanceBegin := "220000+0100"
	maintenanceEnd := "233000+0100"

	kubernetesEnabled := false

	scr := &models.V1ClusterCreateRequest{
		ProjectID:   &project,
		Name:        &name,
		Description: &desc,
		Workers: []*models.V1Worker{
			{
				AutoScalerMin:  &minsize,
				AutoScalerMax:  &maxsize,
				MaxSurge:       &maxsurge,
				MaxUnavailable: &maxunavailable,
			},
		},
		Kubernetes: &models.V1Kubernetes{
			AllowPrivilegedContainers: &allowprivileged,
			Version:                   &version,
		},
		Maintenance: &models.V1Maintenance{
			AutoUpdate: &models.V1MaintenanceAutoUpdate{
				KubernetesVersion: &autoUpdateKubernetes,
				MachineImage:      &autoUpdateMachineImage,
			},
			TimeWindow: &models.V1MaintenanceTimeWindow{
				Begin: &maintenanceBegin,
				End:   &maintenanceEnd,
			},
		},
		AdditionalNetworks: networks,
		Zones:              []string{partition},
		Addons: &models.V1Addons{
			KubernetesDashboard: &kubernetesEnabled,
			NginxIngress:        &defaultingress,
		},
	}
	request := cluster.NewCreateClusterParams()
	request.SetBody(scr)
	shoot, err := cloud.Cluster.CreateCluster(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.CreateClusterConflict:
			return output.HTTPError(e.Payload)
		case *cluster.CreateClusterDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	return printer.Print(shoot.Payload)
}

func clusterList() error {
	tenant := viper.GetString("tenant")
	partition := viper.GetString("partition")
	project := viper.GetString("project")
	var cfr *models.V1ClusterFindRequest
	if tenant != "" || partition != "" || project != "" {
		cfr = &models.V1ClusterFindRequest{}
		if tenant != "" {
			cfr.Tenant = &tenant
		}
		if project != "" {
			cfr.ProjectID = &project
		}
		if partition != "" {
			cfr.PartitionID = &partition
		}
	}
	if cfr != nil {
		fcp := cluster.NewFindClustersParams()
		fcp.SetBody(cfr)
		response, err := cloud.Cluster.FindClusters(fcp, cloud.Auth)
		if err != nil {
			switch e := err.(type) {
			case *cluster.FindClustersDefault:
				return output.HTTPError(e.Payload)
			default:
				return output.UnconventionalError(err)
			}
		}
		return printer.Print(response.Payload)
	}

	request := cluster.NewListClustersParams()
	shoots, err := cloud.Cluster.ListClusters(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.ListClustersDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	return printer.Print(shoots.Payload)
}
func clusterCredentials(args []string) error {
	ci, err := clusterID("credentials", args)
	if err != nil {
		return err
	}
	request := cluster.NewGetClusterCredentialsParams()
	request.SetID(ci)
	credentials, err := cloud.Cluster.GetClusterCredentials(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.GetClusterCredentialsDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	fmt.Println(*credentials.Payload.Kubeconfig)
	return nil
}

func clusterSSHKeyPair(args []string) error {
	ci, err := clusterID("credentials", args)
	if err != nil {
		return err
	}
	request := cluster.NewGetSSHKeyPairParams()
	request.SetID(ci)
	credentials, err := cloud.Cluster.GetSSHKeyPair(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.GetSSHKeyPairDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	privateKey, err := base64.StdEncoding.DecodeString(*credentials.Payload.SSHKeyPair.PrivateKey)
	if err != nil {
		return err
	}
	publicKey, err := base64.StdEncoding.DecodeString(*credentials.Payload.SSHKeyPair.PublicKey)
	if err != nil {
		return err
	}
	fmt.Printf("private key:\n%s\n", privateKey)
	fmt.Printf("public  key:\n%s\n", publicKey)
	return nil
}

func reconcileCluster(args []string) error {
	ci, err := clusterID("reconcile", args)
	if err != nil {
		return err
	}
	request := cluster.NewReconcileClusterParams()
	request.SetID(ci)
	shoot, err := cloud.Cluster.ReconcileCluster(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.ReconcileClusterDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	return printer.Print(shoot.Payload)
}

func updateCluster(args []string) error {
	ci, err := clusterID("update", args)
	if err != nil {
		return err
	}
	minsize := viper.GetInt32("minsize")
	maxsize := viper.GetInt32("maxsize")
	version := viper.GetString("version")

	request := cluster.NewUpdateClusterParams()
	cur := &models.V1ClusterUpdateRequest{
		ID: &ci,
	}
	worker := &models.V1Worker{}
	cur.Workers = append(cur.Workers, worker)
	if minsize != 0 || maxsize != 0 {
		if minsize != 0 {
			cur.Workers[0].AutoScalerMin = &minsize
		}
		if maxsize != 0 {
			cur.Workers[0].AutoScalerMax = &maxsize
		}
	}
	if version != "" {
		cur.Kubernetes = &models.V1Kubernetes{
			Version: &version,
		}
	}
	request.SetBody(cur)
	shoot, err := cloud.Cluster.UpdateCluster(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.UpdateClusterDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	return printer.Print(shoot.Payload)
}

func clusterDelete(args []string) error {
	ci, err := clusterID("delete", args)
	if err != nil {
		return err
	}
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	shoot, err := cloud.Cluster.FindCluster(findRequest, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.FindClusterDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	printer.Print(shoot)
	helper.Prompt("Press Enter to delete above cluster.")
	request := cluster.NewDeleteClusterParams()
	request.SetID(ci)
	c, err := cloud.Cluster.DeleteCluster(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.DeleteClusterDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	return printer.Print(c.Payload)
}
func clusterDescribe(args []string) error {
	ci, err := clusterID("describe", args)
	if err != nil {
		return err
	}
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	shoot, err := cloud.Cluster.FindCluster(findRequest, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.FindClusterDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	return output.YAMLPrinter{}.Print(shoot.Payload)
}

func clusterInputs() error {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.ListConstraintsDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	return output.YAMLPrinter{}.Print(sc)
}

func clusterID(verb string, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("cluster %s requires clusterID as argument", verb)
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", fmt.Errorf("cluster %s requires exactly one clusterID as argument", verb)
}
