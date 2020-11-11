package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/metal-stack/metal-lib/auth"
	"gopkg.in/yaml.v3"

	"github.com/fi-ts/cloud-go/api/client/cluster"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/cmd/output"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mitchellh/mapstructure"
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
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return clusterListCompletion()
		},
		PreRun: bindPFlags,
	}
	clusterDescribeCmd = &cobra.Command{
		Use:   "describe <uid>",
		Short: "describe a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterDescribe(args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return clusterListCompletion()
		},
		PreRun: bindPFlags,
	}
	clusterApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "create/update a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterApply(args)
		},
		PreRun: bindPFlags,
	}
	clusterKubeconfigCmd = &cobra.Command{
		Use:   "kubeconfig <uid>",
		Short: "get cluster kubeconfig",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterKubeconfig(args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return clusterListCompletion()
		},
		PreRun: bindPFlags,
	}

	clusterReconcileCmd = &cobra.Command{
		Use:   "reconcile <uid>",
		Short: "trigger cluster reconciliation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return reconcileCluster(args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return clusterListCompletion()
		},
		PreRun: bindPFlags,
	}
	clusterUpdateCmd = &cobra.Command{
		Use:   "update <uid>",
		Short: "update a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateCluster(args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return clusterListCompletion()
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
	clusterMachineCmd = &cobra.Command{
		Use:     "machine",
		Aliases: []string{"machines"},
		Short:   "list and access machines in the cluster",
	}
	clusterMachineListCmd = &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "list machines of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterMachines(args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return clusterListCompletion()
		},
		PreRun: bindPFlags,
	}
	clusterIssuesCmd = &cobra.Command{
		Use:     "issues [<uid>]",
		Aliases: []string{"problems", "warnings"},
		Short:   "lists cluster issues, shows required actions explicitly when id argument is given",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterIssues(args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return clusterListCompletion()
		},
		PreRun: bindPFlags,
	}
	clusterMachineSSHCmd = &cobra.Command{
		Use:   "ssh <clusterid>",
		Short: "ssh access a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterMachineSSH(args, false)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return clusterListCompletion()
		},
		PreRun: bindPFlags,
	}
	clusterMachineConsoleCmd = &cobra.Command{
		Use:   "console <clusterid>",
		Short: "console access a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterMachineSSH(args, true)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return clusterListCompletion()
		},
		PreRun: bindPFlags,
	}
	clusterLogsCmd = &cobra.Command{
		Use:   "logs",
		Short: "get logs for the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterLogs(args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return clusterListCompletion()
		},
		PreRun: bindPFlags,
	}
)

func init() {
	clusterCreateCmd.Flags().String("name", "", "name of the cluster, max 10 characters. [required]")
	clusterCreateCmd.Flags().String("description", "", "description of the cluster. [optional]")
	clusterCreateCmd.Flags().String("project", "", "project where this cluster should belong to. [required]")
	clusterCreateCmd.Flags().String("partition", "", "partition of the cluster. [required]")
	clusterCreateCmd.Flags().String("purpose", "evaluation", "purpose of the cluster, can be one of production|testing|development|evaluation. SLA is only given on production clusters. [optional]")
	clusterCreateCmd.Flags().String("version", "", "kubernetes version of the cluster. defaults to latest available, check cluster inputs for possible values. [optional]")
	clusterCreateCmd.Flags().String("machinetype", "", "machine type to use for the nodes. [optional]")
	clusterCreateCmd.Flags().String("machineimage", "", "machine image to use for the nodes, must be in the form of <name>-<version> [optional]")
	clusterCreateCmd.Flags().String("firewalltype", "", "machine type to use for the firewall. [optional]")
	clusterCreateCmd.Flags().String("firewallimage", "", "machine image to use for the firewall. [optional]")
	clusterCreateCmd.Flags().String("cri", "docker", "container runtime to use, only docker|containerd supported as alternative actually. [optional]")
	clusterCreateCmd.Flags().Int32("minsize", 1, "minimal workers of the cluster.")
	clusterCreateCmd.Flags().Int32("maxsize", 1, "maximal workers of the cluster.")
	clusterCreateCmd.Flags().String("maxsurge", "1", "max number (e.g. 1) or percentage (e.g. 10%) of workers created during a update of the cluster.")
	clusterCreateCmd.Flags().String("maxunavailable", "1", "max number (e.g. 1) or percentage (e.g. 10%) of workers that can be unavailable during a update of the cluster.")
	clusterCreateCmd.Flags().StringSlice("labels", []string{}, "labels of the cluster")
	clusterCreateCmd.Flags().StringSlice("external-networks", []string{}, "external networks of the cluster")
	clusterCreateCmd.Flags().BoolP("allowprivileged", "", false, "allow privileged containers the cluster.")

	err := clusterCreateCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.MarkFlagRequired("project")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.MarkFlagRequired("partition")
	if err != nil {
		log.Fatal(err.Error())
	}
	clusterCreateCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion()
	})
	clusterCreateCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion()
	})
	clusterCreateCmd.RegisterFlagCompletionFunc("external-networks", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return networkListCompletion()
	})
	clusterCreateCmd.RegisterFlagCompletionFunc("version", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return versionListCompletion()
	})
	clusterCreateCmd.RegisterFlagCompletionFunc("machinetype", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineTypeListCompletion()
	})
	clusterCreateCmd.RegisterFlagCompletionFunc("machineimage", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineImageListCompletion()
	})
	clusterCreateCmd.RegisterFlagCompletionFunc("firewalltype", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firewallTypeListCompletion()
	})
	clusterCreateCmd.RegisterFlagCompletionFunc("firewallimage", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firewallImageListCompletion()
	})
	clusterCreateCmd.RegisterFlagCompletionFunc("purpose", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"production", "testing", "development", "evaluation"}, cobra.ShellCompDirectiveDefault
	})
	clusterCreateCmd.RegisterFlagCompletionFunc("cri", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"docker", "containerd"}, cobra.ShellCompDirectiveDefault
	})

	// Cluster list --------------------------------------------------------------------
	clusterListCmd.Flags().String("id", "", "show clusters of given id")
	clusterListCmd.Flags().String("name", "", "show clusters of given name")
	clusterListCmd.Flags().String("project", "", "show clusters of given project")
	clusterListCmd.Flags().String("partition", "", "show clusters in partition")
	clusterListCmd.Flags().String("tenant", "", "show clusters of given tenant")
	clusterListCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion()
	})
	clusterListCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion()
	})

	// Cluster update --------------------------------------------------------------------
	clusterUpdateCmd.Flags().String("workergroup", "", "the name of the worker group to apply updates to, only required when there are multiple worker groups.")
	clusterUpdateCmd.Flags().Int32("minsize", 0, "minimal workers of the cluster.")
	clusterUpdateCmd.Flags().Int32("maxsize", 0, "maximal workers of the cluster.")
	clusterUpdateCmd.Flags().String("version", "", "kubernetes version of the cluster.")
	clusterUpdateCmd.Flags().String("firewalltype", "", "machine type to use for the firewall.")
	clusterUpdateCmd.Flags().String("firewallimage", "", "machine image to use for the firewall.")
	clusterUpdateCmd.Flags().String("machinetype", "", "machine type to use for the nodes.")
	clusterUpdateCmd.Flags().String("machineimage", "", "machine image to use for the nodes, must be in the form of <name>-<version> ")
	clusterUpdateCmd.Flags().StringSlice("addlabels", []string{}, "labels to add to the cluster")
	clusterUpdateCmd.Flags().StringSlice("removelabels", []string{}, "labels to remove from the cluster")
	clusterUpdateCmd.Flags().BoolP("allowprivileged", "", false, "allow privileged containers the cluster, please add --yes-i-really-mean-it")
	clusterUpdateCmd.Flags().String("purpose", "", "purpose of the cluster, can be one of production|testing|development|evaluation. SLA is only given on production clusters.")
	clusterUpdateCmd.RegisterFlagCompletionFunc("version", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return versionListCompletion()
	})
	clusterUpdateCmd.RegisterFlagCompletionFunc("firewalltype", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firewallTypeListCompletion()
	})
	clusterUpdateCmd.RegisterFlagCompletionFunc("firewallimage", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firewallImageListCompletion()
	})
	clusterUpdateCmd.RegisterFlagCompletionFunc("machinetype", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineTypeListCompletion()
	})
	clusterUpdateCmd.RegisterFlagCompletionFunc("machineimage", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineImageListCompletion()
	})
	clusterUpdateCmd.RegisterFlagCompletionFunc("purpose", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"production", "testing", "development", "evaluation"}, cobra.ShellCompDirectiveDefault
	})

	clusterMachineSSHCmd.Flags().String("machineid", "", "machine to connect to.")
	clusterMachineSSHCmd.MarkFlagRequired("machineid")
	clusterMachineSSHCmd.RegisterFlagCompletionFunc("machineid", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// FIXME howto implement flag based completion for a already given clusterid
		fmt.Printf("args:%v\n", args)
		return clusterMachineListCompletion("123")
	})
	clusterMachineConsoleCmd.Flags().String("machineid", "", "machine to connect to.")
	clusterMachineConsoleCmd.MarkFlagRequired("machineid")
	clusterMachineConsoleCmd.RegisterFlagCompletionFunc("machineid", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// FIXME howto implement flag based completion for a already given clusterid
		fmt.Printf("args:%v\n", args)
		return clusterMachineListCompletion("123")
	})
	clusterMachineCmd.AddCommand(clusterMachineListCmd)
	clusterMachineCmd.AddCommand(clusterMachineSSHCmd)
	clusterMachineCmd.AddCommand(clusterMachineConsoleCmd)

	clusterReconcileCmd.Flags().Bool("retry", false, "Executes a cluster \"retry\" operation instead of regular \"reconcile\".")
	clusterReconcileCmd.Flags().Bool("maintain", false, "Executes a cluster \"maintain\" operation instead of regular \"reconcile\".")

	clusterIssuesCmd.Flags().String("id", "", "show clusters of given id")
	clusterIssuesCmd.Flags().String("name", "", "show clusters of given name")
	clusterIssuesCmd.Flags().String("project", "", "show clusters of given project")
	clusterIssuesCmd.Flags().String("partition", "", "show clusters in partition")
	clusterIssuesCmd.Flags().String("tenant", "", "show clusters of given tenant")

	clusterApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
	Example cluster update:

	# cloudctl cluster describe cluster1 -o yaml > cluster1.yaml
	# vi cluster1.yaml
	## either via stdin
	# cat cluster1.yaml | cloudctl cluster apply -f -
	## or via file
	# cloudctl cluster apply -f cluster1.yaml
	`)

	clusterCmd.AddCommand(clusterApplyCmd)
	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCmd.AddCommand(clusterListCmd)
	clusterCmd.AddCommand(clusterKubeconfigCmd)
	clusterCmd.AddCommand(clusterDeleteCmd)
	clusterCmd.AddCommand(clusterDescribeCmd)
	clusterCmd.AddCommand(clusterInputsCmd)
	clusterCmd.AddCommand(clusterReconcileCmd)
	clusterCmd.AddCommand(clusterUpdateCmd)
	clusterCmd.AddCommand(clusterMachineCmd)
	clusterCmd.AddCommand(clusterLogsCmd)
	clusterCmd.AddCommand(clusterIssuesCmd)
}

func clusterCreate() error {
	name := viper.GetString("name")
	desc := viper.GetString("description")
	partition := viper.GetString("partition")
	project := viper.GetString("project")
	purpose := viper.GetString("purpose")
	machineType := viper.GetString("machinetype")
	machineImageAndVersion := viper.GetString("machineimage")
	firewallType := viper.GetString("firewalltype")
	firewallImage := viper.GetString("firewallimage")

	cri := viper.GetString("cri")

	minsize := viper.GetInt32("minsize")
	maxsize := viper.GetInt32("maxsize")
	maxsurge := viper.GetString("maxsurge")
	maxunavailable := viper.GetString("maxunavailable")

	allowprivileged := viper.GetBool("allowprivileged")

	labels := viper.GetStringSlice("labels")

	// FIXME helper and validation
	networks := viper.GetStringSlice("external-networks")
	autoUpdateKubernetes := false
	autoUpdateMachineImage := false
	maintenanceBegin := "220000+0100"
	maintenanceEnd := "233000+0100"

	version := viper.GetString("version")
	if version == "" {
		request := cluster.NewListConstraintsParams()
		constraints, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
		if err != nil {
			switch e := err.(type) {
			case *cluster.ListConstraintsDefault:
				return output.HTTPError(e.Payload)
			default:
				return output.UnconventionalError(err)
			}
		}

		availableVersions := constraints.Payload.KubernetesVersions
		if len(availableVersions) == 0 {
			log.Fatalf("no kubernetes versions available to deploy")
		}

		sortedVersions := make([]*semver.Version, len(availableVersions))
		for i, r := range availableVersions {
			v, err := semver.NewVersion(r)
			if err != nil {
				log.Fatalf("Error parsing version: %s", err)
			}

			sortedVersions[i] = v
		}

		sort.Sort(semver.Collection(sortedVersions))

		version = sortedVersions[len(sortedVersions)-1].String()
	}

	machineImage := models.V1MachineImage{}
	if machineImageAndVersion != "" {
		machineImageParts := strings.Split(machineImageAndVersion, "-")
		if len(machineImageParts) == 2 {
			machineImage = models.V1MachineImage{
				Name:    &machineImageParts[0],
				Version: &machineImageParts[1],
			}
		} else {
			log.Fatalf("given machineimage:%s is invalid must be in the form <name>-<version>", machineImageAndVersion)
		}
	}

	labelMap := make(map[string]string)
	for _, l := range labels {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) != 2 {
			log.Fatalf("provided labels must be in the form <key>=<value>, found: %s", l)
		}
		labelMap[parts[0]] = parts[1]
	}

	switch cri {
	case "containerd":
	case "docker":
	default:
		log.Fatalf("provided cri:%s is not supported, only docker or containerd at the moment", cri)
	}

	scr := &models.V1ClusterCreateRequest{
		ProjectID:   &project,
		Name:        &name,
		Labels:      labelMap,
		Description: &desc,
		Purpose:     &purpose,
		Workers: []*models.V1Worker{
			{
				Minimum:        &minsize,
				Maximum:        &maxsize,
				MaxSurge:       &maxsurge,
				MaxUnavailable: &maxunavailable,
				MachineType:    &machineType,
				MachineImage:   &machineImage,
				CRI:            &cri,
			},
		},
		FirewallSize:  &firewallType,
		FirewallImage: &firewallImage,
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
		PartitionID:        &partition,
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

func clusterApply(args []string) error {
	var inputFiles []map[string]interface{}
	var genericInput map[string]interface{}
	err := helper.ReadFrom(viper.GetString("file"), &genericInput, func(data interface{}) {
		doc := data.(*map[string]interface{})
		inputFiles = append(inputFiles, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		genericInput = make(map[string]interface{})
	})
	if err != nil {
		return err
	}
	var response []*models.V1ClusterResponse
	for _, input := range inputFiles {
		request := cluster.NewFindClusterParams()
		rawID, ok := input["id"]
		if !ok {
			return fmt.Errorf("id needs to be specified")
		}
		id, ok := rawID.(string)
		if !ok {
			return fmt.Errorf("id needs to be a string")
		}
		request.SetID(id)
		c, err := cloud.Cluster.FindCluster(request, cloud.Auth)
		if err != nil {
			switch e := err.(type) {
			case *cluster.FindClusterDefault:
				return output.HTTPError(e.Payload)
			default:
				return output.UnconventionalError(err)
			}
		}
		if c.Payload == nil {
			var ccr *models.V1ClusterCreateRequest
			err = mapstructure.Decode(input, ccr)
			if err != nil {
				return fmt.Errorf("could not decode into cluster create request: %v", err)
			}
			params := cluster.NewCreateClusterParams()
			params.SetBody(ccr)
			resp, err := cloud.Cluster.CreateCluster(params, cloud.Auth)
			if err != nil {
				switch e := err.(type) {
				case *cluster.CreateClusterDefault:
					return output.HTTPError(e.Payload)
				default:
					return output.UnconventionalError(err)
				}
			}
			response = append(response, resp.Payload)
			continue
		}
		var cur *models.V1ClusterUpdateRequest
		err = mapstructure.Decode(input, cur)
		if err != nil {
			return fmt.Errorf("could not decode into cluster update request: %v", err)
		}
		params := cluster.NewUpdateClusterParams()
		params.SetBody(cur)
		resp, err := cloud.Cluster.UpdateCluster(params, cloud.Auth)
		if err != nil {
			return err
		}
		response = append(response, resp.Payload)
	}
	return printer.Print(response)
}

func clusterList() error {
	id := viper.GetString("id")
	name := viper.GetString("name")
	tenant := viper.GetString("tenant")
	partition := viper.GetString("partition")
	project := viper.GetString("project")
	var cfr *models.V1ClusterFindRequest
	if id != "" || name != "" || tenant != "" || partition != "" || project != "" {
		cfr = &models.V1ClusterFindRequest{}

		if id != "" {
			cfr.ID = &id
		}
		if name != "" {
			cfr.Name = &name
		}
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

func clusterKubeconfig(args []string) error {
	ci, err := clusterID("credentials", args)
	if err != nil {
		return err
	}
	request := cluster.NewGetClusterKubeconfigTplParams()
	request.SetID(ci)
	credentials, err := cloud.Cluster.GetClusterKubeconfigTpl(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.GetClusterKubeconfigTplDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	// kubeconfig with cluster
	kubeconfigContent := *credentials.Payload.Kubeconfig

	kubeconfigFile := viper.GetString("kubeConfig")
	authContext, err := getAuthContext(kubeconfigFile)
	if err != nil {
		return err
	}
	if !authContext.AuthProviderOidc {
		return fmt.Errorf("active user %s has no oidc authProvider, check config", authContext.User)
	}

	cfg := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(kubeconfigContent), cfg)
	if err != nil {
		return err
	}
	// identify clustername
	clusterNames, err := auth.GetClusterNames(cfg)
	if err != nil {
		return err
	}
	if len(clusterNames) != 1 {
		return fmt.Errorf("expected one cluster in config, got %d", len(clusterNames))
	}

	userName := authContext.User
	clusterName := clusterNames[0]
	contextName := fmt.Sprintf("%s@%s", userName, clusterName)

	// merge with current user credentials
	err = auth.AddUser(cfg, *authContext)
	if err != nil {
		return err
	}
	err = auth.AddContext(cfg, contextName, clusterName, userName)
	if err != nil {
		return err
	}
	auth.SetCurrentContext(cfg, contextName)

	mergedKubeconfig, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// print kubeconfig
	fmt.Println(string(mergedKubeconfig))
	return nil
}

type sshkeypair struct {
	privatekey []byte
	publickey  []byte
}

func sshKeyPair(clusterID string) (*sshkeypair, error) {
	request := cluster.NewGetSSHKeyPairParams()
	request.SetID(clusterID)
	credentials, err := cloud.Cluster.GetSSHKeyPair(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.GetSSHKeyPairDefault:
			return nil, output.HTTPError(e.Payload)
		default:
			return nil, output.UnconventionalError(err)
		}
	}
	privateKey, err := base64.StdEncoding.DecodeString(*credentials.Payload.SSHKeyPair.PrivateKey)
	if err != nil {
		return nil, err
	}
	publicKey, err := base64.StdEncoding.DecodeString(*credentials.Payload.SSHKeyPair.PublicKey)
	if err != nil {
		return nil, err
	}
	return &sshkeypair{
		privatekey: privateKey,
		publickey:  publicKey,
	}, nil
}

func reconcileCluster(args []string) error {
	ci, err := clusterID("reconcile", args)
	if err != nil {
		return err
	}

	request := cluster.NewReconcileClusterParams()
	request.SetID(ci)

	if helper.ViperBool("retry") != nil && helper.ViperBool("maintain") != nil {
		return fmt.Errorf("--retry and --maintain are mutually exclusive")
	}

	var operation *string
	if viper.GetBool("retry") {
		o := "retry"
		operation = &o
	}
	if viper.GetBool("maintain") {
		o := "maintain"
		operation = &o
	}
	request.Body = &models.V1ClusterReconcileRequest{Operation: operation}

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
	workergroupname := viper.GetString("workergroup")
	minsize := viper.GetInt32("minsize")
	maxsize := viper.GetInt32("maxsize")
	version := viper.GetString("version")
	firewallType := viper.GetString("firewalltype")
	firewallImage := viper.GetString("firewallimage")
	machineType := viper.GetString("machinetype")
	machineImageAndVersion := viper.GetString("machineimage")
	purpose := viper.GetString("purpose")
	addLabels := viper.GetStringSlice("addlabels")
	removeLabels := viper.GetStringSlice("removelabels")

	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	current, err := cloud.Cluster.FindCluster(findRequest, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.FindClusterDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	request := cluster.NewUpdateClusterParams()
	cur := &models.V1ClusterUpdateRequest{
		ID: &ci,
	}

	if minsize != 0 || maxsize != 0 || machineImageAndVersion != "" || machineType != "" {
		var worker *models.V1Worker
		if workergroupname != "" {
			for _, w := range current.Payload.Workers {
				if w.Name != nil && *w.Name == workergroupname {
					worker = w
					break
				}
			}
			if worker == nil {
				return fmt.Errorf("no worker group found by name: %s", workergroupname)
			}
		} else if len(current.Payload.Workers) == 1 {
			worker = current.Payload.Workers[0]
		} else {
			return fmt.Errorf("there are multiple worker groups, please specify the worker group you want to update with --workergroup")
		}

		if minsize != 0 {
			worker.Minimum = &minsize
		}
		if maxsize != 0 {
			worker.Maximum = &maxsize
		}

		if machineImageAndVersion != "" {
			machineImage := models.V1MachineImage{}
			machineImageParts := strings.Split(machineImageAndVersion, "-")
			if len(machineImageParts) == 2 {
				machineImage = models.V1MachineImage{
					Name:    &machineImageParts[0],
					Version: &machineImageParts[1],
				}
			} else {
				log.Fatalf("given machineimage:%s is invalid must be in the form <name>-<version>", machineImageAndVersion)
			}
			worker.MachineImage = &machineImage
		}

		if machineType != "" {
			worker.MachineType = &machineType
		}

		cur.Workers = append(cur.Workers, worker)
	}

	if firewallImage != "" {
		cur.FirewallImage = &firewallImage
	}
	if firewallType != "" {
		cur.FirewallSize = &firewallType
	}
	if purpose != "" {
		cur.Purpose = &purpose
	}

	if len(addLabels) > 0 || len(removeLabels) > 0 {
		labelMap := current.Payload.Labels

		for _, l := range removeLabels {
			parts := strings.SplitN(l, "=", 2)
			delete(labelMap, parts[0])
		}
		for _, l := range addLabels {
			parts := strings.SplitN(l, "=", 2)
			if len(parts) != 2 {
				log.Fatalf("provided labels must be in the form <key>=<value>, found: %s", l)
			}
			labelMap[parts[0]] = parts[1]
		}

		cur.Labels = labelMap
	}

	k8s := &models.V1Kubernetes{}
	if version != "" {
		k8s.Version = &version
	}
	if viper.IsSet("allowprivileged") {
		if !viper.GetBool("yes-i-really-mean-it") {
			return fmt.Errorf("allowprivileged is set but you forgot to add --yes-i-really-mean-it")
		}
		allowPrivileged := viper.GetBool("allowprivileged")
		k8s.AllowPrivilegedContainers = &allowPrivileged
	}
	cur.Kubernetes = k8s

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

	// we discussed that users are not able to skip the cluster deletion prompt
	// with the --yes-i-really-mean-it flag because deleting our clusters with
	// local storage only could lead to very big problems for users.
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	resp, err := cloud.Cluster.FindCluster(findRequest, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.FindClusterDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	printer.Print(resp.Payload)
	firstPartOfClusterID := strings.Split(*resp.Payload.ID, "-")[0]
	fmt.Println("Please answer some security questions to delete this cluster")
	err = helper.Prompt("first part of clusterID:", firstPartOfClusterID)
	if err != nil {
		return err
	}
	err = helper.Prompt("Clustername:", *resp.Payload.Name)
	if err != nil {
		return err
	}

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
	return printer.Print(shoot.Payload)
}

func clusterIssues(args []string) error {
	if len(args) == 0 {
		id := viper.GetString("id")
		name := viper.GetString("name")
		tenant := viper.GetString("tenant")
		partition := viper.GetString("partition")
		project := viper.GetString("project")
		boolTrue := true
		var cfr *models.V1ClusterFindRequest
		if id != "" || name != "" || tenant != "" || partition != "" || project != "" {
			cfr = &models.V1ClusterFindRequest{}

			if id != "" {
				cfr.ID = &id
			}
			if name != "" {
				cfr.Name = &name
			}
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
			fcp := cluster.NewFindClustersParams().WithReturnMachines(&boolTrue)
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
			return printer.Print(output.ShootIssuesResponses(response.Payload))
		}

		request := cluster.NewListClustersParams().WithReturnMachines(&boolTrue)
		shoots, err := cloud.Cluster.ListClusters(request, cloud.Auth)
		if err != nil {
			switch e := err.(type) {
			case *cluster.ListClustersDefault:
				return output.HTTPError(e.Payload)
			default:
				return output.UnconventionalError(err)
			}
		}
		return printer.Print(output.ShootIssuesResponses(shoots.Payload))
	}

	ci, err := clusterID("issues", args)
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
	return printer.Print(output.ShootIssuesResponse(shoot.Payload))
}

func clusterMachines(args []string) error {
	ci, err := clusterID("machines", args)
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
	fmt.Println("Cluster:")
	printer.Print(shoot.Payload)

	// FIXME this is a ugly hack to reset the printer and have a new header.
	initPrinter()

	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)
	fmt.Println("\nMachines:")
	return printer.Print(ms)
}

func clusterLogs(args []string) error {
	ci, err := clusterID("logs", args)
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
	var conditions []*models.V1beta1Condition
	var lastOperation *models.V1beta1LastOperation
	var lastErrors []*models.V1beta1LastError
	if shoot.Payload != nil && shoot.Payload.Status != nil {
		conditions = shoot.Payload.Status.Conditions
		lastOperation = shoot.Payload.Status.LastOperation
		lastErrors = shoot.Payload.Status.LastErrors
	}

	fmt.Println("Conditions:")
	err = printer.Print(conditions)
	if err != nil {
		return err
	}

	// FIXME this is a ugly hack to reset the printer and have a new header.
	initPrinter()

	fmt.Println("\nLast Errors:")
	err = printer.Print(lastErrors)
	if err != nil {
		return err
	}

	// FIXME this is a ugly hack to reset the printer and have a new header.
	initPrinter()

	fmt.Println("\nLast Operation:")
	return printer.Print(lastOperation)
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

func clusterMachineSSH(args []string, console bool) error {
	cid, err := clusterID("ssh", args)
	if err != nil {
		return err
	}
	mid := viper.GetString("machineid")

	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(cid)
	shoot, err := cloud.Cluster.FindCluster(findRequest, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *cluster.FindClusterDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	keypair, err := sshKeyPair(cid)
	if err != nil {
		return err
	}
	for _, m := range shoot.Payload.Machines {
		if *m.ID == mid {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("unable determine home directory:%v", err)
			}
			privateKeyFile := path.Join(home, "."+programName, "."+cid+".id_rsa")
			err = ioutil.WriteFile(privateKeyFile, keypair.privatekey, 0600)
			if err != nil {
				return fmt.Errorf("unable to write private key:%s error:%v", privateKeyFile, err)
			}
			defer os.Remove(privateKeyFile)
			if console {
				fmt.Printf("access console via ssh\n")
				bmcConsolePort := "5222"
				err := ssh("-i", privateKeyFile, mid+"@"+cloud.ConsoleHost, "-p", bmcConsolePort)
				return err
			}
			networks := m.Allocation.Networks
			feature := m.Allocation.Image.Features[0]
			switch feature {
			case "firewall":
				for _, nw := range networks {
					if *nw.Underlay || *nw.Private {
						continue
					}
					for _, ip := range nw.Ips {
						if portOpen(ip, "22", time.Second) {
							err := ssh("-i", privateKeyFile, "metal"+"@"+ip)
							return err
						}
					}
				}
				return fmt.Errorf("no ip with a open ssh port found")
			case "machine":
				// FIXME metal user is not allowed to execute
				// ip vrf exec <tenantvrf> ssh <machineip>
				return fmt.Errorf("machine access via ssh not implemented")
			default:
				return fmt.Errorf("unknown machine type:%s", feature)
			}
		}
	}

	return fmt.Errorf("machine:%s not found in cluster:%s", mid, cid)
}

func ssh(args ...string) error {
	path, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("unable to locate ssh in path")
	}
	fmt.Printf("%s %s\n", path, strings.Join(args, " "))
	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd.Run()
}

func portOpen(ip string, port string, timeout time.Duration) bool {
	address := net.JoinHostPort(ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	if conn != nil {
		_ = conn.Close()
		return true
	}
	return false
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
