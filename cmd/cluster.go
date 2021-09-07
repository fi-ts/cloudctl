package cmd

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/fi-ts/cloud-go/api/client/cluster"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/cmd/output"

	"github.com/Masterminds/semver/v3"
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
		Use:     "delete <clusterid>",
		Short:   "delete a cluster",
		Aliases: []string{"rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterDelete(args)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}
	clusterDescribeCmd = &cobra.Command{
		Use:   "describe <clusterid>",
		Short: "describe a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterDescribe(args)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}
	clusterKubeconfigCmd = &cobra.Command{
		Use:   "kubeconfig <clusterid>",
		Short: "get cluster kubeconfig",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterKubeconfig(args)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}

	clusterReconcileCmd = &cobra.Command{
		Use:   "reconcile <clusterid>",
		Short: "trigger cluster reconciliation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return reconcileCluster(args)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}
	clusterUpdateCmd = &cobra.Command{
		Use:   "update <clusterid>",
		Short: "update a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateCluster(args)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
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
		Use:     "ls <clusterid>",
		Aliases: []string{"list"},
		Short:   "list machines of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterMachines(args)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}
	clusterIssuesCmd = &cobra.Command{
		Use:     "issues [<clusterid>]",
		Aliases: []string{"problems", "warnings"},
		Short:   "lists cluster issues, shows required actions explicitly when id argument is given",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterIssues(args)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}
	clusterMachineSSHCmd = &cobra.Command{
		Use:   "ssh <clusterid>",
		Short: "ssh access a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterMachineSSH(args, false)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}
	clusterMachineConsoleCmd = &cobra.Command{
		Use:   "console <clusterid>",
		Short: "console access a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterMachineSSH(args, true)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}
	clusterMachineResetCmd = &cobra.Command{
		Use:   "reset <clusterid>",
		Short: "hard power reset of a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterMachineReset(args)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}
	clusterMachineCycleCmd = &cobra.Command{
		Use:   "cycle <clusterid>",
		Short: "soft power cycle of a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterMachineCycle(args)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}
	clusterMachineReinstallCmd = &cobra.Command{
		Use:   "reinstall <clusterid>",
		Short: "reinstall OS image onto a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterMachineReinstall(args)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}
	clusterLogsCmd = &cobra.Command{
		Use:   "logs <clusterid>",
		Short: "get logs for the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return clusterLogs(args)
		},
		ValidArgsFunction: clusterListCompletionFunc,
		PreRun:            bindPFlags,
	}
)

func init() {
	clusterCreateCmd.Flags().String("name", "", "name of the cluster, max 10 characters. [required]")
	clusterCreateCmd.Flags().String("description", "", "description of the cluster. [optional]")
	clusterCreateCmd.Flags().String("project", "", "project where this cluster should belong to. [required]")
	clusterCreateCmd.Flags().String("partition", "", "partition of the cluster. [required]")
	clusterCreateCmd.Flags().String("purpose", "evaluation", "purpose of the cluster, can be one of production|development|evaluation. SLA is only given on production clusters. [optional]")
	clusterCreateCmd.Flags().String("version", "", "kubernetes version of the cluster. defaults to latest available, check cluster inputs for possible values. [optional]")
	clusterCreateCmd.Flags().String("machinetype", "", "machine type to use for the nodes. [optional]")
	clusterCreateCmd.Flags().String("machineimage", "", "machine image to use for the nodes, must be in the form of <name>-<version> [optional]")
	clusterCreateCmd.Flags().String("firewalltype", "", "machine type to use for the firewall. [optional]")
	clusterCreateCmd.Flags().String("firewallimage", "", "machine image to use for the firewall. [optional]")
	clusterCreateCmd.Flags().String("firewallcontroller", "", "version of the firewall-controller to use. [optional]")
	clusterCreateCmd.Flags().String("cri", "", "container runtime to use, only docker|containerd supported as alternative actually. [optional]")
	clusterCreateCmd.Flags().Int32("minsize", 1, "minimal workers of the cluster.")
	clusterCreateCmd.Flags().Int32("maxsize", 1, "maximal workers of the cluster.")
	clusterCreateCmd.Flags().String("maxsurge", "1", "max number (e.g. 1) or percentage (e.g. 10%) of workers created during a update of the cluster.")
	clusterCreateCmd.Flags().String("maxunavailable", "1", "max number (e.g. 1) or percentage (e.g. 10%) of workers that can be unavailable during a update of the cluster.")
	clusterCreateCmd.Flags().StringSlice("labels", []string{}, "labels of the cluster")
	clusterCreateCmd.Flags().StringSlice("external-networks", []string{}, "external networks of the cluster")
	clusterCreateCmd.Flags().StringSlice("egress", []string{}, "static egress ips per network, must be in the form <network>:<ip>; e.g.: --egress internet:1.2.3.4,extnet:123.1.1.1 --egress internet:1.2.3.5 [optional]")
	clusterCreateCmd.Flags().BoolP("allowprivileged", "", false, "allow privileged containers the cluster.")
	clusterCreateCmd.Flags().String("audit", "on", "audit logging of cluster API access; can be off, on (default) or splunk (Logging to a predefined or custom splunk endpoint). [optional]")
	clusterCreateCmd.Flags().Duration("healthtimeout", 0, "period (e.g. \"24h\") after which an unhealthy node is declared failed and will be replaced. [optional]")
	clusterCreateCmd.Flags().Duration("draintimeout", 0, "period (e.g. \"3h\") after which a draining node will be forcefully deleted. [optional]")

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
	err = clusterCreateCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.RegisterFlagCompletionFunc("external-networks", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return networkListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.RegisterFlagCompletionFunc("version", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return versionListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.RegisterFlagCompletionFunc("machinetype", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineTypeListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.RegisterFlagCompletionFunc("machineimage", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineImageListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.RegisterFlagCompletionFunc("firewalltype", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firewallTypeListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.RegisterFlagCompletionFunc("firewallimage", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firewallImageListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.RegisterFlagCompletionFunc("firewallcontroller", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firewallControllerVersionListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.RegisterFlagCompletionFunc("purpose", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"production", "development", "evaluation"}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.RegisterFlagCompletionFunc("cri", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"docker", "containerd"}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterCreateCmd.RegisterFlagCompletionFunc("audit", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"off", "on", "splunk"}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	// Cluster list --------------------------------------------------------------------
	clusterListCmd.Flags().String("id", "", "show clusters of given id")
	clusterListCmd.Flags().String("name", "", "show clusters of given name")
	clusterListCmd.Flags().String("project", "", "show clusters of given project")
	clusterListCmd.Flags().String("partition", "", "show clusters in partition")
	clusterListCmd.Flags().String("tenant", "", "show clusters of given tenant")
	err = clusterListCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterListCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	// Cluster update --------------------------------------------------------------------
	clusterUpdateCmd.Flags().String("workergroup", "", "the name of the worker group to apply updates to, only required when there are multiple worker groups.")
	clusterUpdateCmd.Flags().Int32("minsize", 0, "minimal workers of the cluster.")
	clusterUpdateCmd.Flags().Int32("maxsize", 0, "maximal workers of the cluster.")
	clusterUpdateCmd.Flags().String("version", "", "kubernetes version of the cluster.")
	clusterUpdateCmd.Flags().String("firewalltype", "", "machine type to use for the firewall.")
	clusterUpdateCmd.Flags().String("firewallimage", "", "machine image to use for the firewall.")
	clusterUpdateCmd.Flags().String("firewallcontroller", "", "version of the firewall-controller to use.")
	clusterUpdateCmd.Flags().String("machinetype", "", "machine type to use for the nodes.")
	clusterUpdateCmd.Flags().String("machineimage", "", "machine image to use for the nodes, must be in the form of <name>-<version> ")
	clusterUpdateCmd.Flags().StringSlice("addlabels", []string{}, "labels to add to the cluster")
	clusterUpdateCmd.Flags().StringSlice("removelabels", []string{}, "labels to remove from the cluster")
	clusterUpdateCmd.Flags().BoolP("allowprivileged", "", false, "allow privileged containers the cluster, please add --yes-i-really-mean-it")
	clusterUpdateCmd.Flags().String("audit", "on", "audit logging of cluster API access; can be off, on or splunk (Logging to a predefined or custom splunk endpoint).")
	clusterUpdateCmd.Flags().String("purpose", "", "purpose of the cluster, can be one of production|development|evaluation. SLA is only given on production clusters.")
	clusterUpdateCmd.Flags().StringSlice("egress", []string{}, "static egress ips per network, must be in the form <networkid>:<semicolon-separated ips>; e.g.: --egress internet:1.2.3.4;1.2.3.5 --egress extnet:123.1.1.1 [optional]. Use --egress none to remove all ingress rules.")
	clusterUpdateCmd.Flags().StringSlice("external-networks", []string{}, "external networks of the cluster")
	clusterUpdateCmd.Flags().Duration("healthtimeout", 0, "period (e.g. \"24h\") after which an unhealthy node is declared failed and will be replaced.")
	clusterUpdateCmd.Flags().Duration("draintimeout", 0, "period (e.g. \"3h\") after which a draining node will be forcefully deleted.")
	clusterUpdateCmd.Flags().String("maxsurge", "", "max number (e.g. 1) or percentage (e.g. 10%) of workers created during a update of the cluster.")
	clusterUpdateCmd.Flags().String("maxunavailable", "", "max number (e.g. 1) or percentage (e.g. 10%) of workers that can be unavailable during a update of the cluster.")
	clusterUpdateCmd.Flags().BoolP("autoupdate-kubernetes", "", false, "enables automatic updates of the kubernetes patch version of the cluster")
	clusterUpdateCmd.Flags().BoolP("autoupdate-machineimages", "", false, "enables automatic updates of the worker node images of the cluster, be aware that this deletes worker nodes!")

	err = clusterUpdateCmd.RegisterFlagCompletionFunc("version", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return versionListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterUpdateCmd.RegisterFlagCompletionFunc("firewalltype", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firewallTypeListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterUpdateCmd.RegisterFlagCompletionFunc("firewallimage", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firewallImageListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterUpdateCmd.RegisterFlagCompletionFunc("firewallcontroller", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firewallControllerVersionListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterUpdateCmd.RegisterFlagCompletionFunc("machinetype", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineTypeListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterUpdateCmd.RegisterFlagCompletionFunc("machineimage", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineImageListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterUpdateCmd.RegisterFlagCompletionFunc("purpose", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"production", "development", "evaluation"}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterUpdateCmd.RegisterFlagCompletionFunc("audit", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"off", "on", "splunk"}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	clusterMachineSSHCmd.Flags().String("machineid", "", "machine to connect to.")
	err = clusterMachineSSHCmd.MarkFlagRequired("machineid")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterMachineSSHCmd.RegisterFlagCompletionFunc("machineid", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return clusterMachineListCompletion(args, false)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	clusterMachineConsoleCmd.Flags().String("machineid", "", "machine to connect to.")
	err = clusterMachineConsoleCmd.MarkFlagRequired("machineid")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterMachineConsoleCmd.RegisterFlagCompletionFunc("machineid", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return clusterMachineListCompletion(args, true)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	clusterMachineResetCmd.Flags().String("machineid", "", "machine to reset.")
	err = clusterMachineResetCmd.MarkFlagRequired("machineid")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterMachineResetCmd.RegisterFlagCompletionFunc("machineid", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return clusterMachineListCompletion(args, true)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	clusterMachineCycleCmd.Flags().String("machineid", "", "machine to reset.")
	err = clusterMachineCycleCmd.MarkFlagRequired("machineid")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterMachineCycleCmd.RegisterFlagCompletionFunc("machineid", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return clusterMachineListCompletion(args, true)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	clusterMachineReinstallCmd.Flags().String("machineid", "", "machine to reinstall.")
	clusterMachineReinstallCmd.Flags().String("machineimage", "", "image to reinstall (optional).")
	err = clusterMachineReinstallCmd.MarkFlagRequired("machineid")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = clusterMachineReinstallCmd.RegisterFlagCompletionFunc("machineid", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return clusterMachineListCompletion(args, true)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	clusterMachineCmd.AddCommand(clusterMachineListCmd)
	clusterMachineCmd.AddCommand(clusterMachineSSHCmd)
	clusterMachineCmd.AddCommand(clusterMachineConsoleCmd)
	clusterMachineCmd.AddCommand(clusterMachineResetCmd)
	clusterMachineCmd.AddCommand(clusterMachineCycleCmd)
	clusterMachineCmd.AddCommand(clusterMachineReinstallCmd)

	clusterReconcileCmd.Flags().Bool("retry", false, "Executes a cluster \"retry\" operation instead of regular \"reconcile\".")
	clusterReconcileCmd.Flags().Bool("maintain", false, "Executes a cluster \"maintain\" operation instead of regular \"reconcile\".")

	clusterIssuesCmd.Flags().String("id", "", "show clusters of given id")
	clusterIssuesCmd.Flags().String("name", "", "show clusters of given name")
	clusterIssuesCmd.Flags().String("project", "", "show clusters of given project")
	clusterIssuesCmd.Flags().String("partition", "", "show clusters in partition")
	clusterIssuesCmd.Flags().String("tenant", "", "show clusters of given tenant")

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
	firewallController := viper.GetString("firewallcontroller")

	cri := viper.GetString("cri")

	minsize := viper.GetInt32("minsize")
	maxsize := viper.GetInt32("maxsize")
	maxsurge := viper.GetString("maxsurge")
	maxunavailable := viper.GetString("maxunavailable")

	healthtimeout := viper.GetDuration("healthtimeout")
	draintimeout := viper.GetDuration("draintimeout")

	allowprivileged := viper.GetBool("allowprivileged")
	audit := viper.GetString("audit")

	labels := viper.GetStringSlice("labels")

	// FIXME helper and validation
	networks := viper.GetStringSlice("external-networks")
	egress := viper.GetStringSlice("egress")
	maintenanceBegin := "220000+0100"
	maintenanceEnd := "233000+0100"

	version := viper.GetString("version")
	if version == "" {
		request := cluster.NewListConstraintsParams()
		constraints, err := cloud.Cluster.ListConstraints(request, nil)
		if err != nil {
			return err
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

	labelMap, err := helper.LabelsToMap(labels)
	if err != nil {
		log.Fatal(err)
	}

	switch cri {
	case "containerd":
	case "docker":
	case "":
	default:
		log.Fatalf("provided cri:%s is not supported, only docker or containerd at the moment", cri)
	}

	var (
		clusterAudit  bool
		auditToSplunk bool
	)
	switch audit {
	case "off":
		clusterAudit = false
		auditToSplunk = false
	case "on":
		clusterAudit = true
		auditToSplunk = false
	case "splunk":
		clusterAudit = true
		auditToSplunk = true
	case "":
	default:
		log.Fatalf("Audit value %s is not supported; choose \"off\", \"on\" or \"splunk\".", audit)
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
		FirewallSize:              &firewallType,
		FirewallImage:             &firewallImage,
		FirewallControllerVersion: &firewallController,
		Kubernetes: &models.V1Kubernetes{
			AllowPrivilegedContainers: &allowprivileged,
			Version:                   &version,
		},
		Audit: &models.V1Audit{
			ClusterAudit:  &clusterAudit,
			AuditToSplunk: &auditToSplunk,
		},
		Maintenance: &models.V1Maintenance{
			TimeWindow: &models.V1MaintenanceTimeWindow{
				Begin: &maintenanceBegin,
				End:   &maintenanceEnd,
			},
		},
		AdditionalNetworks: networks,
		PartitionID:        &partition,
	}

	egressRules := makeEgressRules(egress)
	if len(egressRules) > 0 {
		scr.EgressRules = egressRules
	}

	if healthtimeout != 0 {
		scr.Workers[0].HealthTimeout = int64(healthtimeout)
	}

	if draintimeout != 0 {
		scr.Workers[0].DrainTimeout = int64(draintimeout)
	}

	request := cluster.NewCreateClusterParams()
	request.SetBody(scr)
	shoot, err := cloud.Cluster.CreateCluster(request, nil)
	if err != nil {
		return err
	}
	return printer.Print(shoot.Payload)
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
		response, err := cloud.Cluster.FindClusters(fcp, nil)
		if err != nil {
			return err
		}
		return printer.Print(response.Payload)
	}

	request := cluster.NewListClustersParams()
	shoots, err := cloud.Cluster.ListClusters(request, nil)
	if err != nil {
		return err
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
	credentials, err := cloud.Cluster.GetClusterKubeconfigTpl(request, nil)
	if err != nil {
		return err
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

	mergedKubeconfig, err := helper.EnrichKubeconfigTpl(kubeconfigContent, authContext)
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
	credentials, err := cloud.Cluster.GetSSHKeyPair(request, nil)
	if err != nil {
		return nil, err
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

	shoot, err := cloud.Cluster.ReconcileCluster(request, nil)
	if err != nil {
		return err
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
	firewallController := viper.GetString("firewallcontroller")
	firewallNetworks := viper.GetStringSlice("external-networks")
	machineType := viper.GetString("machinetype")
	machineImageAndVersion := viper.GetString("machineimage")
	purpose := viper.GetString("purpose")
	addLabels := viper.GetStringSlice("addlabels")
	removeLabels := viper.GetStringSlice("removelabels")
	egress := viper.GetStringSlice("egress")
	maxsurge := viper.GetString("maxsurge")
	maxunavailable := viper.GetString("maxunavailable")

	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	resp, err := cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}
	current := resp.Payload

	healthtimeout := viper.GetDuration("healthtimeout")
	draintimeout := viper.GetDuration("draintimeout")

	request := cluster.NewUpdateClusterParams()
	cur := &models.V1ClusterUpdateRequest{
		ID: &ci,
		Maintenance: &models.V1Maintenance{
			AutoUpdate: &models.V1MaintenanceAutoUpdate{
				KubernetesVersion: current.Maintenance.AutoUpdate.KubernetesVersion,
				MachineImage:      current.Maintenance.AutoUpdate.MachineImage,
			},
		},
	}

	if minsize != 0 || maxsize != 0 || machineImageAndVersion != "" || machineType != "" || healthtimeout != 0 || draintimeout != 0 || maxsurge != "" || maxunavailable != "" {
		workers := current.Workers

		var worker *models.V1Worker
		if workergroupname != "" {
			for _, w := range workers {
				if w.Name != nil && *w.Name == workergroupname {
					worker = w
					break
				}
			}
			if worker == nil {
				return fmt.Errorf("no worker group found by name: %s", workergroupname)
			}
		} else if len(workers) == 1 {
			worker = workers[0]
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

		mcmMigrated := false
		for _, feature := range current.ControlPlaneFeatureGates {
			if feature == "machineControllerManagerOOT" {
				mcmMigrated = true
				break
			}
		}

		if healthtimeout != 0 {
			if !mcmMigrated {
				log.Fatal("custom healthtimeout requires feature: machineControllerManagerOOT")
			}
			worker.HealthTimeout = int64(healthtimeout)
		}

		if draintimeout != 0 {
			if !mcmMigrated {
				log.Fatal("custom draintimeout requires feature: machineControllerManagerOOT")
			}
			worker.DrainTimeout = int64(draintimeout)
		}

		if maxsurge != "" {
			worker.MaxSurge = &maxsurge
		}

		if maxunavailable != "" {
			worker.MaxUnavailable = &maxunavailable
		}

		cur.Workers = append(cur.Workers, workers...)
	}

	if viper.IsSet("autoupdate-kubernetes") {
		auto := viper.GetBool("autoupdate-kubernetes")
		cur.Maintenance.AutoUpdate.KubernetesVersion = &auto
	}
	if viper.IsSet("autoupdate-machineimages") {
		auto := viper.GetBool("autoupdate-machineimages")
		cur.Maintenance.AutoUpdate.MachineImage = &auto
	}

	updateCausesDowntime := false
	if firewallImage != "" {
		if current.FirewallImage != nil && *current.FirewallImage != firewallImage {
			updateCausesDowntime = true
		}
		cur.FirewallImage = &firewallImage
	}
	if firewallType != "" {
		if current.FirewallSize != nil && *current.FirewallSize != firewallType {
			updateCausesDowntime = true
		}
		cur.FirewallSize = &firewallType
	}
	if firewallController != "" {
		cur.FirewallControllerVersion = &firewallController
	}
	if len(firewallNetworks) > 0 {
		if !sets.NewString(firewallNetworks...).Equal(sets.NewString(current.AdditionalNetworks...)) {
			updateCausesDowntime = true
		}
		cur.AdditionalNetworks = firewallNetworks
	}

	if purpose != "" {
		cur.Purpose = &purpose
	}

	if len(addLabels) > 0 || len(removeLabels) > 0 {
		labelMap := current.Labels

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

	auditFlags := &models.V1Audit{}
	audit := viper.GetString("audit")
	switch audit {
	case "off":
		ca := false
		as := false
		auditFlags.ClusterAudit = &ca
		auditFlags.AuditToSplunk = &as
	case "on":
		ca := true
		as := false
		auditFlags.ClusterAudit = &ca
		auditFlags.AuditToSplunk = &as
	case "splunk":
		ca := true
		as := true
		auditFlags.ClusterAudit = &ca
		auditFlags.AuditToSplunk = &as
	case "":
	default:
		log.Fatalf("Audit value %s is not supported; choose \"off\", \"on\" or \"splunk\".", audit)
	}
	cur.Audit = auditFlags

	cur.EgressRules = makeEgressRules(egress)

	if updateCausesDowntime && !viper.GetBool("yes-i-really-mean-it") {
		fmt.Println("This cluster update will cause downtime.")
		err = helper.Prompt("Are you sure? (y/n)", "y")
		if err != nil {
			return err
		}
	}

	request.SetBody(cur)
	shoot, err := cloud.Cluster.UpdateCluster(request, nil)
	if err != nil {
		return err
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
	resp, err := cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
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
	c, err := cloud.Cluster.DeleteCluster(request, nil)
	if err != nil {
		return err
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
	shoot, err := cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
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
			response, err := cloud.Cluster.FindClusters(fcp, nil)
			if err != nil {
				return err
			}
			return printer.Print(output.ShootIssuesResponses(response.Payload))
		}

		request := cluster.NewListClustersParams().WithReturnMachines(&boolTrue)
		shoots, err := cloud.Cluster.ListClusters(request, nil)
		if err != nil {
			return err
		}
		return printer.Print(output.ShootIssuesResponses(shoots.Payload))
	}

	ci, err := clusterID("issues", args)
	if err != nil {
		return err
	}
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	shoot, err := cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
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
	shoot, err := cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}

	if printer.Type() != "table" {
		return printer.Print(shoot.Payload)
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
	shoot, err := cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}
	var conditions []*models.V1beta1Condition
	var lastOperation *models.V1beta1LastOperation
	var lastErrors []*models.V1beta1LastError
	if shoot.Payload != nil && shoot.Payload.Status != nil {
		conditions = shoot.Payload.Status.Conditions
		lastOperation = shoot.Payload.Status.LastOperation
		lastErrors = shoot.Payload.Status.LastErrors
	}

	if printer.Type() != "table" {
		type s struct {
			Conditions    []*models.V1beta1Condition
			LastOperation *models.V1beta1LastOperation
			LastErrors    []*models.V1beta1LastError
		}
		return printer.Print(s{
			Conditions:    conditions,
			LastOperation: lastOperation,
			LastErrors:    lastErrors,
		})
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
	sc, err := cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return err
	}

	return output.YAMLPrinter{}.Print(sc)
}

func clusterMachineReset(args []string) error {
	cid, err := clusterID("reset", args)
	if err != nil {
		return err
	}
	mid := viper.GetString("machineid")

	request := cluster.NewResetMachineParams()
	request.SetID(cid)
	request.Body = &models.V1ClusterMachineResetRequest{Machineid: &mid}

	shoot, err := cloud.Cluster.ResetMachine(request, nil)
	if err != nil {
		return err
	}

	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)

	return printer.Print(ms)
}

func clusterMachineCycle(args []string) error {
	cid, err := clusterID("reset", args)
	if err != nil {
		return err
	}
	mid := viper.GetString("machineid")

	request := cluster.NewCycleMachineParams()
	request.SetID(cid)
	request.Body = &models.V1ClusterMachineCycleRequest{Machineid: &mid}

	shoot, err := cloud.Cluster.CycleMachine(request, nil)
	if err != nil {
		return err
	}

	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)

	return printer.Print(ms)
}

func clusterMachineReinstall(args []string) error {
	cid, err := clusterID("reinstall", args)
	if err != nil {
		return err
	}
	mid := viper.GetString("machineid")
	img := viper.GetString("machineimage")

	request := cluster.NewReinstallMachineParams()
	request.SetID(cid)
	request.Body = &models.V1ClusterMachineReinstallRequest{Machineid: &mid}
	if img != "" {
		request.Body.Imageid = img
	}

	shoot, err := cloud.Cluster.ReinstallMachine(request, nil)
	if err != nil {
		return err
	}

	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)

	return printer.Print(ms)
}

func clusterMachineSSH(args []string, console bool) error {
	cid, err := clusterID("ssh", args)
	if err != nil {
		return err
	}
	mid := viper.GetString("machineid")

	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(cid)
	shoot, err := cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}

	keypair, err := sshKeyPair(cid)
	if err != nil {
		return err
	}
	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)
	for _, m := range ms {
		if *m.ID == mid {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("unable determine home directory:%w", err)
			}
			privateKeyFile := path.Join(home, "."+programName, "."+cid+".id_rsa")
			err = os.WriteFile(privateKeyFile, keypair.privatekey, 0600)
			if err != nil {
				return fmt.Errorf("unable to write private key:%s error:%w", privateKeyFile, err)
			}
			defer os.Remove(privateKeyFile)
			if console {
				fmt.Printf("access console via ssh\n")
				authContext, err := getAuthContext(viper.GetString("kubeConfig"))
				if err != nil {
					return err
				}
				err = os.Setenv("LC_METAL_STACK_OIDC_TOKEN", authContext.IDToken)
				if err != nil {
					return err
				}
				bmcConsolePort := "5222"
				err = ssh("-i", privateKeyFile, mid+"@"+consoleHost, "-p", bmcConsolePort)
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

func makeEgressRules(egressFlagValue []string) []*models.V1EgressRule {
	if len(egressFlagValue) == 0 {
		return nil
	}

	if len(egressFlagValue) == 1 && egressFlagValue[0] == "none" {
		return []*models.V1EgressRule{}
	}

	m := map[string]models.V1EgressRule{}
	for _, e := range egressFlagValue {
		parts := strings.Split(e, ":")
		if len(parts) != 2 {
			log.Fatalf("egress config needs format <network>:<ip> but got %q", e)
		}
		n, ip := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if net.ParseIP(ip) == nil {
			log.Fatalf("egress config contains an invalid IP %s for network %s", ip, n)
		}

		if _, ok := m[n]; !ok {
			m[n] = models.V1EgressRule{
				NetworkID: &n,
			}
		}

		element := m[n]
		element.IPs = append(element.IPs, ip)
		m[n] = element
	}

	egressRules := []*models.V1EgressRule{}
	for _, e := range m {
		r := e
		egressRules = append(egressRules, &r)
	}
	return egressRules
}
