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
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/fatih/color"
	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/gosimple/slug"
	"github.com/metal-stack/metal-lib/auth"
	"github.com/metal-stack/metal-lib/pkg/pointer"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/completion"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/fi-ts/cloudctl/pkg/api"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
)

type auditConfigOptionsMap map[string]struct {
	Config      *models.V1Audit
	Description string
}

func (a auditConfigOptionsMap) Names(withDescription bool) []string {
	var names []string
	for name, opt := range a {
		if withDescription {
			names = append(names, fmt.Sprintf("%s\t%s", name, opt.Description))
		} else {
			names = append(names, name)
		}
	}
	return names
}

var (
	// options
	auditConfigOptions = auditConfigOptionsMap{
		"off": {
			Description: "turn off the kube-apiserver auditlog",
			Config: &models.V1Audit{
				ClusterAudit:  pointer.Pointer(false),
				AuditToSplunk: pointer.Pointer(false),
			},
		},
		"on": {
			Description: "turn on the kube-apiserver auditlog, and expose it as container log of the audittailer deployment in the audit namespace",
			Config: &models.V1Audit{
				ClusterAudit:  pointer.Pointer(true),
				AuditToSplunk: pointer.Pointer(false),
			},
		},
		"splunk": {
			Description: "also forward the auditlog to a splunk HEC endpoint. create a custom splunk config manifest with \"cloudctl cluster splunk-config-manifest\"",
			Config: &models.V1Audit{
				ClusterAudit:  pointer.Pointer(true),
				AuditToSplunk: pointer.Pointer(true),
			},
		},
	}
)

func newClusterCmd(c *config) *cobra.Command {
	clusterCmd := &cobra.Command{
		Use:   "cluster",
		Short: "manage clusters",
		Long:  "TODO",
	}
	clusterCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterCreate()
		},
		PreRun: bindPFlags,
	}

	clusterListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list clusters",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterList()
		},
		PreRun: bindPFlags,
	}
	clusterDeleteCmd := &cobra.Command{
		Use:     "delete <clusterid>",
		Short:   "delete a cluster",
		Aliases: []string{"destroy", "rm", "remove"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterDelete(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterDescribeCmd := &cobra.Command{
		Use:   "describe <clusterid>",
		Short: "describe a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterDescribe(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterKubeconfigCmd := &cobra.Command{
		Use:   "kubeconfig <clusterid>",
		Short: "get cluster kubeconfig",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterKubeconfig(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}

	clusterReconcileCmd := &cobra.Command{
		Use:   "reconcile <clusterid>",
		Short: "trigger cluster reconciliation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.reconcileCluster(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterUpdateCmd := &cobra.Command{
		Use:   "update <clusterid>",
		Short: "update a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.updateCluster(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterInputsCmd := &cobra.Command{
		Use:   "inputs",
		Short: "get possible cluster inputs like k8s versions, etc.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterInputs()
		},
		PreRun: bindPFlags,
	}
	clusterMachineCmd := &cobra.Command{
		Use:     "machine",
		Aliases: []string{"machines"},
		Short:   "list and access machines in the cluster",
	}
	clusterMachineListCmd := &cobra.Command{
		Use:     "ls <clusterid>",
		Aliases: []string{"list"},
		Short:   "list machines of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachines(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterIssuesCmd := &cobra.Command{
		Use:     "issues [<clusterid>]",
		Aliases: []string{"problems", "warnings"},
		Short:   "lists cluster issues, shows required actions explicitly when id argument is given",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterIssues(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterMonitoringSecretCmd := &cobra.Command{
		Use:   "monitoring-secret <clusterid>",
		Short: "returns the endpoint and access credentials to the monitoring of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMonitoringSecret(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterMachineSSHCmd := &cobra.Command{
		Use:   "ssh <clusterid>",
		Short: "ssh access a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachineSSH(args, false)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterMachineConsoleCmd := &cobra.Command{
		Use:   "console <clusterid>",
		Short: "console access a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachineSSH(args, true)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterMachineResetCmd := &cobra.Command{
		Use:   "reset <clusterid>",
		Short: "hard power reset of a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachineReset(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterMachineCycleCmd := &cobra.Command{
		Use:   "cycle <clusterid>",
		Short: "soft power cycle of a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachineCycle(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterMachineReinstallCmd := &cobra.Command{
		Use:   "reinstall <clusterid>",
		Short: "reinstall OS image onto a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachineReinstall(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterLogsCmd := &cobra.Command{
		Use:   "logs <clusterid>",
		Short: "get logs for the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterLogs(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}
	clusterSplunkConfigManifestCmd := &cobra.Command{
		Use:   "splunk-config-manifest",
		Short: "create a manifest for a custom splunk configuration, overriding the default settings for splunk auditing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterSplunkConfigManifest()
		},
		PreRun: bindPFlags,
	}
	clusterDNSManifestCmd := &cobra.Command{
		Use:   "dns-manifest <clusterid>",
		Short: "create a manifest for an ingress or service type loadbalancer, creating a DNS entry and valid certificate within your cluster domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterDNSManifest(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
		PreRun:            bindPFlags,
	}

	clusterCreateCmd.Flags().String("name", "", "name of the cluster, max 10 characters. [required]")
	clusterCreateCmd.Flags().String("description", "", "description of the cluster. [optional]")
	clusterCreateCmd.Flags().String("project", "", "project where this cluster should belong to. [required]")
	clusterCreateCmd.Flags().String("partition", "", "partition of the cluster. [required]")
	clusterCreateCmd.Flags().String("seed", "", "name of seed where this cluster should be scheduled. [optional]")
	clusterCreateCmd.Flags().String("purpose", "evaluation", fmt.Sprintf("purpose of the cluster, can be one of %s. SLA is only given on production clusters. [optional]", strings.Join(completion.ClusterPurposes, "|")))
	clusterCreateCmd.Flags().String("version", "", "kubernetes version of the cluster. defaults to latest available, check cluster inputs for possible values. [optional]")
	clusterCreateCmd.Flags().String("machinetype", "", "machine type to use for the nodes. [optional]")
	clusterCreateCmd.Flags().String("machineimage", "", "machine image to use for the nodes, must be in the form of <name>-<version> [optional]")
	clusterCreateCmd.Flags().String("firewalltype", "", "machine type to use for the firewall. [optional]")
	clusterCreateCmd.Flags().String("firewallimage", "", "machine image to use for the firewall. [optional]")
	clusterCreateCmd.Flags().String("firewallcontroller", "", "version of the firewall-controller to use. [optional]")
	clusterCreateCmd.Flags().BoolP("logacceptedconns", "", false, "also log accepted connections on the cluster firewall [optional]")
	clusterCreateCmd.Flags().String("cri", "", "container runtime to use, only docker|containerd supported as alternative actually. [optional]")
	clusterCreateCmd.Flags().Int32("minsize", 1, "minimal workers of the cluster.")
	clusterCreateCmd.Flags().Int32("maxsize", 1, "maximal workers of the cluster.")
	clusterCreateCmd.Flags().String("maxsurge", "1", "max number (e.g. 1) or percentage (e.g. 10%) of workers created during a update of the cluster.")
	clusterCreateCmd.Flags().String("maxunavailable", "0", "max number (e.g. 0) or percentage (e.g. 10%) of workers that can be unavailable during a update of the cluster.")
	clusterCreateCmd.Flags().StringSlice("labels", []string{}, "labels of the cluster")
	clusterCreateCmd.Flags().StringSlice("external-networks", []string{}, "external networks of the cluster")
	clusterCreateCmd.Flags().StringSlice("egress", []string{}, "static egress ips per network, must be in the form <network>:<ip>; e.g.: --egress internet:1.2.3.4,extnet:123.1.1.1 --egress internet:1.2.3.5 [optional]")
	clusterCreateCmd.Flags().BoolP("allowprivileged", "", false, "allow privileged containers the cluster.")
	clusterCreateCmd.Flags().String("audit", "on", "audit logging of cluster API access; can be off, on (default) or splunk (logging to a predefined or custom splunk endpoint). [optional]")
	clusterCreateCmd.Flags().Duration("healthtimeout", 0, "period (e.g. \"24h\") after which an unhealthy node is declared failed and will be replaced. [optional]")
	clusterCreateCmd.Flags().Duration("draintimeout", 0, "period (e.g. \"3h\") after which a draining node will be forcefully deleted. [optional]")
	clusterCreateCmd.Flags().BoolP("reversed-vpn", "", false, "enables usage of reversed-vpn instead of konnectivity tunnel for worker connectivity. [optional]")
	clusterCreateCmd.Flags().BoolP("autoupdate-kubernetes", "", false, "enables automatic updates of the kubernetes patch version of the cluster [optional]")
	clusterCreateCmd.Flags().BoolP("autoupdate-machineimages", "", false, "enables automatic updates of the worker node images of the cluster, be aware that this deletes worker nodes! [optional]")
	clusterCreateCmd.Flags().String("default-storage-class", "", "set default storage class to given name, must be one of the managed storage classes")
	clusterCreateCmd.Flags().String("max-pods-per-node", "", "set number of maximum pods per node (default: 510). Lower numbers allow for more node per cluster. [optional]")
	clusterCreateCmd.Flags().String("cni", "", "the network plugin used in this cluster, defaults to calico. please note that cilium support is still Alpha and we are happy to receive feedback. [optional]")

	must(clusterCreateCmd.MarkFlagRequired("name"))
	must(clusterCreateCmd.MarkFlagRequired("project"))
	must(clusterCreateCmd.MarkFlagRequired("partition"))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("seed", c.comp.SeedListCompletion))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("external-networks", c.comp.NetworkListCompletion))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("version", c.comp.VersionListCompletion))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("machinetype", c.comp.MachineTypeListCompletion))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("machineimage", c.comp.MachineImageListCompletion))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("firewalltype", c.comp.FirewallTypeListCompletion))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("firewallimage", c.comp.FirewallImageListCompletion))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("firewallcontroller", c.comp.FirewallControllerVersionListCompletion))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("purpose", c.comp.ClusterPurposeListCompletion))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("cri", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"docker", "containerd"}, cobra.ShellCompDirectiveNoFileComp
	}))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("cni", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"calico\tcalico networking plugin. this is the cluster default.",
			"cilium\tcilium networking plugin. please note that cilium support is still Alpha and we are happy to receive feedback.",
		}, cobra.ShellCompDirectiveNoFileComp
	}))
	must(clusterCreateCmd.RegisterFlagCompletionFunc("audit", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return auditConfigOptions.Names(true),
			cobra.ShellCompDirectiveNoFileComp
	}))

	clusterDescribeCmd.Flags().Bool("no-machines", false, "does not return in the output")

	// Cluster list --------------------------------------------------------------------
	clusterListCmd.Flags().String("id", "", "show clusters of given id")
	clusterListCmd.Flags().String("name", "", "show clusters of given name")
	clusterListCmd.Flags().String("project", "", "show clusters of given project")
	clusterListCmd.Flags().String("partition", "", "show clusters in partition")
	clusterListCmd.Flags().String("seed", "", "show clusters in seed")
	clusterListCmd.Flags().String("tenant", "", "show clusters of given tenant")
	clusterListCmd.Flags().StringSlice("labels", nil, "show clusters of given labels")
	clusterListCmd.Flags().String("purpose", "", "show clusters of given purpose")
	must(clusterListCmd.RegisterFlagCompletionFunc("name", c.comp.ClusterNameCompletion))
	must(clusterListCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(clusterListCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	must(clusterListCmd.RegisterFlagCompletionFunc("seed", c.comp.SeedListCompletion))
	must(clusterListCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	must(clusterListCmd.RegisterFlagCompletionFunc("purpose", c.comp.ClusterPurposeListCompletion))

	// Cluster update --------------------------------------------------------------------
	clusterUpdateCmd.Flags().String("workergroup", "", "the name of the worker group to apply updates to, only required when there are multiple worker groups.")
	clusterUpdateCmd.Flags().Bool("remove-workergroup", false, "if set, removes the targeted worker group")
	clusterUpdateCmd.Flags().StringSlice("workerlabels", []string{}, "labels of the worker group (syncs to kubernetes node resource after some time, too)")
	clusterUpdateCmd.Flags().StringSlice("workerannotations", []string{}, "annotations of the worker group (syncs to kubernetes node resource after some time, too)")
	clusterUpdateCmd.Flags().Int32("minsize", 0, "minimal workers of the cluster.")
	clusterUpdateCmd.Flags().Int32("maxsize", 0, "maximal workers of the cluster.")
	clusterUpdateCmd.Flags().String("version", "", "kubernetes version of the cluster.")
	clusterUpdateCmd.Flags().String("seed", "", "name of seed where this cluster should be scheduled.")
	clusterUpdateCmd.Flags().String("firewalltype", "", "machine type to use for the firewall.")
	clusterUpdateCmd.Flags().String("firewallimage", "", "machine image to use for the firewall.")
	clusterUpdateCmd.Flags().String("firewallcontroller", "", "version of the firewall-controller to use.")
	clusterUpdateCmd.Flags().BoolP("logacceptedconns", "", false, "enables logging of accepted connections on the cluster firewall")
	clusterUpdateCmd.Flags().String("machinetype", "", "machine type to use for the nodes.")
	clusterUpdateCmd.Flags().String("machineimage", "", "machine image to use for the nodes, must be in the form of <name>-<version> ")
	clusterUpdateCmd.Flags().StringSlice("addlabels", []string{}, "labels to add to the cluster")
	clusterUpdateCmd.Flags().StringSlice("removelabels", []string{}, "labels to remove from the cluster")
	clusterUpdateCmd.Flags().BoolP("allowprivileged", "", false, "allow privileged containers the cluster, please add --yes-i-really-mean-it")
	clusterUpdateCmd.Flags().String("audit", "on", "audit logging of cluster API access; can be off, on or splunk (logging to a predefined or custom splunk endpoint).")
	clusterUpdateCmd.Flags().String("purpose", "", fmt.Sprintf("purpose of the cluster, can be one of %s. SLA is only given on production clusters.", strings.Join(completion.ClusterPurposes, "|")))
	clusterUpdateCmd.Flags().StringSlice("egress", []string{}, "static egress ips per network, must be in the form <networkid>:<semicolon-separated ips>; e.g.: --egress internet:1.2.3.4;1.2.3.5 --egress extnet:123.1.1.1 [optional]. Use --egress none to remove all egress rules.")
	clusterUpdateCmd.Flags().StringSlice("external-networks", []string{}, "external networks of the cluster")
	clusterUpdateCmd.Flags().Duration("healthtimeout", 0, "period (e.g. \"24h\") after which an unhealthy node is declared failed and will be replaced. (0 = provider-default)")
	clusterUpdateCmd.Flags().Duration("draintimeout", 0, "period (e.g. \"3h\") after which a draining node will be forcefully deleted. (0 = provider-default)")
	clusterUpdateCmd.Flags().String("maxsurge", "", "max number (e.g. 1) or percentage (e.g. 10%) of workers created during a update of the cluster.")
	clusterUpdateCmd.Flags().String("maxunavailable", "", "max number (e.g. 0) or percentage (e.g. 10%) of workers that can be unavailable during a update of the cluster.")
	clusterUpdateCmd.Flags().BoolP("autoupdate-kubernetes", "", false, "enables automatic updates of the kubernetes patch version of the cluster")
	clusterUpdateCmd.Flags().BoolP("autoupdate-machineimages", "", false, "enables automatic updates of the worker node images of the cluster, be aware that this deletes worker nodes!")
	clusterUpdateCmd.Flags().BoolP("reversed-vpn", "", false, "enables usage of reversed-vpn instead of konnectivity tunnel for worker connectivity.")
	clusterUpdateCmd.Flags().String("default-storage-class", "", "set default storage class to given name, must be one of the managed storage classes")
	clusterUpdateCmd.Flags().BoolP("disable-custom-default-storage-class", "", false, "if set to true, no default class is deployed, you have to set one of your storageclasses manually to default")

	must(clusterUpdateCmd.RegisterFlagCompletionFunc("version", c.comp.VersionListCompletion))
	must(clusterUpdateCmd.RegisterFlagCompletionFunc("firewalltype", c.comp.FirewallTypeListCompletion))
	must(clusterUpdateCmd.RegisterFlagCompletionFunc("firewallimage", c.comp.FirewallImageListCompletion))
	must(clusterUpdateCmd.RegisterFlagCompletionFunc("seed", c.comp.SeedListCompletion))
	must(clusterUpdateCmd.RegisterFlagCompletionFunc("firewallcontroller", c.comp.FirewallControllerVersionListCompletion))
	must(clusterUpdateCmd.RegisterFlagCompletionFunc("machinetype", c.comp.MachineTypeListCompletion))
	must(clusterUpdateCmd.RegisterFlagCompletionFunc("machineimage", c.comp.MachineImageListCompletion))
	must(clusterUpdateCmd.RegisterFlagCompletionFunc("purpose", c.comp.ClusterPurposeListCompletion))
	must(clusterUpdateCmd.RegisterFlagCompletionFunc("audit", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return auditConfigOptions.Names(true),
			cobra.ShellCompDirectiveNoFileComp
	}))

	clusterInputsCmd.Flags().String("partition", "", "partition of the constraints.")
	must(clusterInputsCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))

	// Cluster splunk config manifest --------------------------------------------------------------------
	clusterSplunkConfigManifestCmd.Flags().String("token", "", "the hec token to use for this cluster's audit logs")
	clusterSplunkConfigManifestCmd.Flags().String("index", "", "the splunk index to use for this cluster's audit logs")
	clusterSplunkConfigManifestCmd.Flags().String("hechost", "", "the hostname or IP of the splunk HEC endpoint")
	clusterSplunkConfigManifestCmd.Flags().Int("hecport", 0, "port on which the splunk HEC endpoint is listening")
	clusterSplunkConfigManifestCmd.Flags().Bool("tls", false, "whether to use TLS encryption. You do need to specify a CA file.")
	clusterSplunkConfigManifestCmd.Flags().String("cafile", "", "the path to the file containing the ca certificate (chain) for the splunk HEC endpoint")
	clusterSplunkConfigManifestCmd.Flags().String("cabase64", "", "the base64-encoded ca certificate (chain) for the splunk HEC endpoint")

	// Cluster dns manifest --------------------------------------------------------------------
	clusterDNSManifestCmd.Flags().String("type", "ingress", "either of type ingress or service")
	clusterDNSManifestCmd.Flags().String("name", "<name>", "the resource name")
	clusterDNSManifestCmd.Flags().String("namespace", "default", "the resource's namespace")
	clusterDNSManifestCmd.Flags().Int("ttl", 180, "the ttl set to the created dns entry")
	clusterDNSManifestCmd.Flags().Bool("with-certificate", true, "whether to request a let's encrypt certificate for the requested dns entry or not")
	clusterDNSManifestCmd.Flags().String("backend-name", "my-backend", "the name of the backend")
	clusterDNSManifestCmd.Flags().Int32("backend-port", 443, "the port of the backend")
	clusterDNSManifestCmd.Flags().String("ingress-class", "nginx", "the ingress class name")
	must(clusterDNSManifestCmd.RegisterFlagCompletionFunc("type", cobra.FixedCompletions([]string{"ingress", "service"}, cobra.ShellCompDirectiveNoFileComp)))

	// Cluster machine ... --------------------------------------------------------------------
	clusterMachineSSHCmd.Flags().String("machineid", "", "machine to connect to.")
	must(clusterMachineSSHCmd.MarkFlagRequired("machineid"))
	must(clusterMachineSSHCmd.RegisterFlagCompletionFunc("machineid", c.comp.ClusterFirewallListCompletion))

	clusterMachineConsoleCmd.Flags().String("machineid", "", "machine to connect to.")
	must(clusterMachineConsoleCmd.MarkFlagRequired("machineid"))
	must(clusterMachineConsoleCmd.RegisterFlagCompletionFunc("machineid", c.comp.ClusterMachineListCompletion))

	clusterMachineResetCmd.Flags().String("machineid", "", "machine to reset.")
	must(clusterMachineResetCmd.MarkFlagRequired("machineid"))
	must(clusterMachineResetCmd.RegisterFlagCompletionFunc("machineid", c.comp.ClusterMachineListCompletion))

	clusterMachineCycleCmd.Flags().String("machineid", "", "machine to reset.")
	must(clusterMachineCycleCmd.MarkFlagRequired("machineid"))
	must(clusterMachineCycleCmd.RegisterFlagCompletionFunc("machineid", c.comp.ClusterMachineListCompletion))

	clusterMachineReinstallCmd.Flags().String("machineid", "", "machine to reinstall.")
	clusterMachineReinstallCmd.Flags().String("machineimage", "", "image to reinstall (optional).")
	must(clusterMachineReinstallCmd.MarkFlagRequired("machineid"))
	must(clusterMachineReinstallCmd.RegisterFlagCompletionFunc("machineid", c.comp.ClusterMachineListCompletion))

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

	must(clusterIssuesCmd.RegisterFlagCompletionFunc("name", c.comp.ClusterNameCompletion))
	must(clusterIssuesCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(clusterIssuesCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	must(clusterIssuesCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))

	clusterKubeconfigCmd.Flags().Bool("merge", false, "merges the cluster's kubeconfig into the current active kubeconfig, otherwise an individual kubeconfig is printed to console only")
	clusterKubeconfigCmd.Flags().Bool("set-context", false, "when setting the merge parameter to true, immediately activates the cluster's context")

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
	clusterCmd.AddCommand(clusterSplunkConfigManifestCmd)
	clusterCmd.AddCommand(clusterDNSManifestCmd)
	clusterCmd.AddCommand(clusterMonitoringSecretCmd)

	return clusterCmd
}

func (c *config) clusterCreate() error {
	name := viper.GetString("name")
	desc := viper.GetString("description")
	partition := viper.GetString("partition")
	seed := viper.GetString("seed")
	project := viper.GetString("project")
	purpose := viper.GetString("purpose")
	machineType := viper.GetString("machinetype")
	machineImageAndVersion := viper.GetString("machineimage")
	firewallType := viper.GetString("firewalltype")
	firewallImage := viper.GetString("firewallimage")
	firewallController := viper.GetString("firewallcontroller")
	logAcceptedConnections := strconv.FormatBool(viper.GetBool("logacceptedconns"))

	cri := viper.GetString("cri")
	var cni string
	if viper.IsSet("cni") {
		cni = viper.GetString("cni")
	}

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

	reversedVPN := strconv.FormatBool(viper.GetBool("reversed-vpn"))

	version := viper.GetString("version")
	if version == "" {
		request := cluster.NewListConstraintsParams()
		constraints, err := c.cloud.Cluster.ListConstraints(request, nil)
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

	auditConfig, ok := auditConfigOptions[audit]
	if !ok {
		return fmt.Errorf("audit value %s is not supported; choose one of %v", audit, auditConfigOptions.Names(false))
	}

	var customDefaultStorageClass *models.V1CustomDefaultStorageClass
	if viper.IsSet("default-storage-class") {
		class := viper.GetString("default-storage-class")
		customDefaultStorageClass = &models.V1CustomDefaultStorageClass{
			ClassName: &class,
		}
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
		Audit: auditConfig.Config,
		Maintenance: &models.V1Maintenance{
			TimeWindow: &models.V1MaintenanceTimeWindow{
				Begin: &maintenanceBegin,
				End:   &maintenanceEnd,
			},
		},
		AdditionalNetworks: networks,
		PartitionID:        &partition,
		ClusterFeatures: &models.V1ClusterFeatures{
			ReversedVPN:            &reversedVPN,
			LogAcceptedConnections: &logAcceptedConnections,
		},
		CustomDefaultStorageClass: customDefaultStorageClass,
		Cni:                       cni,
	}

	if viper.IsSet("autoupdate-kubernetes") || viper.IsSet("autoupdate-machineimages") || purpose == string(v1beta1.ShootPurposeEvaluation) {
		scr.Maintenance.AutoUpdate = &models.V1MaintenanceAutoUpdate{}

		// default to true for evaluation clusters
		if purpose == string(v1beta1.ShootPurposeEvaluation) {
			scr.Maintenance.AutoUpdate.KubernetesVersion = pointer.Pointer(true)
		}
		if viper.IsSet("autoupdate-kubernetes") {
			auto := viper.GetBool("autoupdate-kubernetes")
			scr.Maintenance.AutoUpdate.KubernetesVersion = &auto
		}
		if viper.IsSet("autoupdate-machineimages") {
			auto := viper.GetBool("autoupdate-machineimages")
			scr.Maintenance.AutoUpdate.MachineImage = &auto
		}
	}

	if viper.IsSet("max-pods-per-node") {
		scr.Kubernetes.MaxPodsPerNode = viper.GetInt32("max-pods-per-node")
	}
	if seed != "" {
		scr.SeedName = seed
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
	shoot, err := c.cloud.Cluster.CreateCluster(request, nil)
	if err != nil {
		return err
	}
	return output.New().Print(shoot.Payload)
}

func (c *config) clusterList() error {
	id := viper.GetString("id")
	name := viper.GetString("name")
	tenant := viper.GetString("tenant")
	partition := viper.GetString("partition")
	seed := viper.GetString("seed")
	project := viper.GetString("project")
	purpose := viper.GetString("purpose")
	labels := viper.GetStringSlice("labels")
	var cfr *models.V1ClusterFindRequest
	if id != "" || name != "" || tenant != "" || partition != "" || seed != "" || project != "" || purpose != "" || len(labels) > 0 {
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
		if seed != "" {
			cfr.SeedName = &seed
		}
		if purpose != "" {
			cfr.Purpose = &purpose
		}
		if len(labels) > 0 {
			labelMap := map[string]string{}
			for _, l := range labels {
				parts := strings.SplitN(l, "=", 2)
				if len(parts) != 2 {
					log.Fatalf("provided labels must be in the form <key>=<value>, found: %s", l)
				}
				labelMap[parts[0]] = parts[1]
			}
			cfr.Labels = labelMap
		}
	}
	if cfr != nil {
		fcp := cluster.NewFindClustersParams()
		fcp.SetBody(cfr)
		response, err := c.cloud.Cluster.FindClusters(fcp, nil)
		if err != nil {
			return err
		}
		return output.New().Print(response.Payload)
	}

	request := cluster.NewListClustersParams()
	shoots, err := c.cloud.Cluster.ListClusters(request, nil)
	if err != nil {
		return err
	}
	return output.New().Print(shoots.Payload)
}

func (c *config) clusterKubeconfig(args []string) error {
	id, err := c.clusterID("credentials", args)
	if err != nil {
		return err
	}

	request := cluster.NewGetClusterKubeconfigTplParams()
	request.SetID(id)
	credentials, err := c.cloud.Cluster.GetClusterKubeconfigTpl(request, nil)
	if err != nil {
		return err
	}

	kubeconfigTpl := *credentials.Payload.Kubeconfig // is a kubeconfig with only a single cluster entry

	kubeconfigFile := viper.GetString("kubeconfig")
	authContext, err := api.GetAuthContext(kubeconfigFile)
	if err != nil {
		return err
	}
	if !authContext.AuthProviderOidc {
		return fmt.Errorf("active user %s has no oidc authProvider, check config", authContext.User)
	}

	if !viper.GetBool("merge") {
		mergedKubeconfig, err := helper.EnrichKubeconfigTpl(kubeconfigTpl, authContext)
		if err != nil {
			return err
		}

		fmt.Println(string(mergedKubeconfig))
		return nil
	}

	currentCfg, filename, _, err := auth.LoadKubeConfig(kubeconfigFile)
	if err != nil {
		return err
	}

	clusterResp, err := c.cloud.Cluster.FindCluster(cluster.NewFindClusterParams().WithID(id), nil)
	if err != nil {
		return err
	}

	contextName := slug.Make(*clusterResp.Payload.Name)

	if viper.GetBool("set-context") {
		auth.SetCurrentContext(currentCfg, contextName)
	}

	mergedKubeconfig, err := helper.MergeKubeconfigTpl(currentCfg, kubeconfigTpl, contextName, *clusterResp.Payload.Name, authContext)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, mergedKubeconfig, 0600)
	if err != nil {
		return err
	}

	fmt.Printf("%s merged context %q into %s\n", color.GreenString("✔"), contextName, filename)

	return nil
}

type sshkeypair struct {
	privatekey []byte
	publickey  []byte
}

func (c *config) sshKeyPair(clusterID string) (*sshkeypair, *models.V1VPN, error) {
	request := cluster.NewGetSSHKeyPairParams()
	request.SetID(clusterID)
	credentials, err := c.cloud.Cluster.GetSSHKeyPair(request, nil)
	if err != nil {
		return nil, nil, err
	}
	privateKey, err := base64.StdEncoding.DecodeString(*credentials.Payload.SSHKeyPair.PrivateKey)
	if err != nil {
		return nil, nil, err
	}
	publicKey, err := base64.StdEncoding.DecodeString(*credentials.Payload.SSHKeyPair.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return &sshkeypair{
		privatekey: privateKey,
		publickey:  publicKey,
	}, credentials.Payload.VPN, nil
}

func (c *config) reconcileCluster(args []string) error {
	ci, err := c.clusterID("reconcile", args)
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

	shoot, err := c.cloud.Cluster.ReconcileCluster(request, nil)
	if err != nil {
		return err
	}
	return output.New().Print(shoot.Payload)
}

func (c *config) updateCluster(args []string) error {
	ci, err := c.clusterID("update", args)
	if err != nil {
		return err
	}
	workergroupname := viper.GetString("workergroup")
	removeworkergroup := viper.GetBool("remove-workergroup")
	workerlabelslice := viper.GetStringSlice("workerlabels")
	workerannotationsslice := viper.GetStringSlice("workerannotations")
	minsize := viper.GetInt32("minsize")
	maxsize := viper.GetInt32("maxsize")
	version := viper.GetString("version")
	seed := viper.GetString("seed")
	firewallType := viper.GetString("firewalltype")
	firewallImage := viper.GetString("firewallimage")
	firewallController := viper.GetString("firewallcontroller")
	firewallNetworks := viper.GetStringSlice("external-networks")
	logAcceptedConnections := strconv.FormatBool(viper.GetBool("logacceptedconns"))
	machineType := viper.GetString("machinetype")
	machineImageAndVersion := viper.GetString("machineimage")
	purpose := viper.GetString("purpose")
	addLabels := viper.GetStringSlice("addlabels")
	removeLabels := viper.GetStringSlice("removelabels")
	egress := viper.GetStringSlice("egress")
	maxsurge := viper.GetString("maxsurge")
	maxunavailable := viper.GetString("maxunavailable")

	defaultStorageClass := viper.GetString("default-storage-class")
	disableDefaultStorageClass := viper.GetBool("disable-custom-default-storage-class")

	reversedVPN := strconv.FormatBool(viper.GetBool("reversed-vpn"))

	workerlabels, err := helper.LabelsToMap(workerlabelslice)
	if err != nil {
		return err
	}
	workerannotations, err := helper.LabelsToMap(workerannotationsslice)
	if err != nil {
		return err
	}

	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	resp, err := c.cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}
	current := resp.Payload

	healthtimeout := viper.GetDuration("healthtimeout")
	draintimeout := viper.GetDuration("draintimeout")

	customDefaultStorageClass := current.CustomDefaultStorageClass
	if viper.IsSet("default-storage-class") && disableDefaultStorageClass {
		return fmt.Errorf("either default-storage-class or disable-custom-default-storage-class may be specified, not both")
	}

	if disableDefaultStorageClass {
		customDefaultStorageClass = nil
	}

	if viper.IsSet("default-storage-class") {
		customDefaultStorageClass = &models.V1CustomDefaultStorageClass{
			ClassName: &defaultStorageClass,
		}
	}

	var clusterFeatures models.V1ClusterFeatures
	if viper.IsSet("reversed-vpn") {
		clusterFeatures.ReversedVPN = &reversedVPN
	}
	if viper.IsSet("logacceptedconns") {
		clusterFeatures.LogAcceptedConnections = &logAcceptedConnections
	}

	request := cluster.NewUpdateClusterParams()
	cur := &models.V1ClusterUpdateRequest{
		ID: &ci,
		Maintenance: &models.V1Maintenance{
			AutoUpdate: &models.V1MaintenanceAutoUpdate{
				KubernetesVersion: current.Maintenance.AutoUpdate.KubernetesVersion,
				MachineImage:      current.Maintenance.AutoUpdate.MachineImage,
			},
		},
		ClusterFeatures:           &clusterFeatures,
		CustomDefaultStorageClass: customDefaultStorageClass,
	}

	if workergroupname != "" ||
		minsize != 0 || maxsize != 0 || maxsurge != "" || maxunavailable != "" ||
		machineImageAndVersion != "" || machineType != "" ||
		viper.IsSet("healthtimeout") || viper.IsSet("draintimeout") ||
		viper.IsSet("workerlabels") || viper.IsSet("workerannotations") {

		workers := current.Workers

		var worker *models.V1Worker
		if workergroupname != "" {
			for _, w := range workers {
				if w.Name != nil && *w.Name == workergroupname {
					worker = w
					break
				}
			}
			if worker == nil && !removeworkergroup {
				fmt.Println("Adding a new worker group to the cluster.")
				err = helper.Prompt("Are you sure? (y/n)", "y")
				if err != nil {
					return err
				}

				worker = &models.V1Worker{
					Name:           &workergroupname,
					Minimum:        pointer.Pointer(int32(1)),
					Maximum:        pointer.Pointer(int32(1)),
					MaxSurge:       pointer.Pointer("1"),
					MaxUnavailable: pointer.Pointer("0"),
					Labels:         workerlabels,
					Annotations:    workerannotations,
				}
				workers = append(workers, worker)
			}
		} else if len(workers) == 1 {
			worker = workers[0]
		} else {
			return fmt.Errorf("there are multiple worker groups, please specify the worker group you want to update with --workergroup")
		}

		if removeworkergroup {
			fmt.Println("WARNING. Removing a worker group cannot be undone and causes the loss of local data on the deleted nodes.")
			err = helper.Prompt("Are you sure? (y/n)", "y")
			if err != nil {
				return err
			}

			var newWorkers []*models.V1Worker
			for _, w := range workers {
				w := w
				if w.Name != nil && *w.Name == *worker.Name {
					continue
				}
				newWorkers = append(newWorkers, w)
			}

			cur.Workers = newWorkers
		} else {
			if minsize != 0 {
				worker.Minimum = &minsize
			}
			if maxsize != 0 {
				c := 0
				for _, m := range current.Machines {
					for _, t := range m.Tags {
						if t == fmt.Sprintf("%s=%s", string(constants.LabelWorkerPool), *worker.Name) {
							c++
						}
					}
				}
				if int(maxsize) < c {
					fmt.Println("WARNING. New maxsize is lower than currently active machines. A random worker node which is still in use will be removed.")
					err = helper.Prompt("Are you sure? (y/n)", "y")
					if err != nil {
						return err
					}
				}
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

			if viper.IsSet("healthtimeout") {
				worker.HealthTimeout = int64(healthtimeout)
			}

			if viper.IsSet("draintimeout") {
				worker.DrainTimeout = int64(draintimeout)
			}

			if viper.IsSet("workerlabels") {
				worker.Labels = workerlabels
			}

			if viper.IsSet("workerannotations") {
				worker.Annotations = workerannotations
			}

			if maxsurge != "" {
				worker.MaxSurge = &maxsurge
			}

			if maxunavailable != "" {
				worker.MaxUnavailable = &maxunavailable
			}

			cur.Workers = append(cur.Workers, workers...)
		}
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
		if *cur.Maintenance.AutoUpdate.KubernetesVersion && *current.Purpose == string(v1beta1.ShootPurposeEvaluation) && purpose != string(v1beta1.ShootPurposeEvaluation) {
			fmt.Print("\nHint: Kubernetes auto updates will still be enabled after this update.\n\n")
		}
		cur.Purpose = &purpose
	}

	if seed != "" && current.Status.SeedName != seed {
		updateCausesDowntime = true
		cur.SeedName = &seed
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

	if viper.IsSet("audit") {
		audit := viper.GetString("audit")
		auditConfig, ok := auditConfigOptions[audit]
		if !ok {
			return fmt.Errorf("audit value %s is not supported; choose one of %v", audit, auditConfigOptions.Names(false))
		}
		cur.Audit = auditConfig.Config
	}

	cur.EgressRules = makeEgressRules(egress)

	if updateCausesDowntime && !viper.GetBool("yes-i-really-mean-it") {
		fmt.Println("This cluster update will cause downtime.")
		err = helper.Prompt("Are you sure? (y/n)", "y")
		if err != nil {
			return err
		}
	}

	request.SetBody(cur)
	shoot, err := c.cloud.Cluster.UpdateCluster(request, nil)
	if err != nil {
		return err
	}
	return output.New().Print(shoot.Payload)
}

func (c *config) clusterDelete(args []string) error {
	ci, err := c.clusterID("delete", args)
	if err != nil {
		return err
	}

	// we discussed that users are not able to skip the cluster deletion prompt
	// with the --yes-i-really-mean-it flag because deleting our clusters with
	// local storage only could lead to very big problems for users.
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	resp, err := c.cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}

	must(output.New().Print(resp.Payload))

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
	cl, err := c.cloud.Cluster.DeleteCluster(request, nil)
	if err != nil {
		return err
	}
	return output.New().Print(cl.Payload)
}

func (c *config) clusterDescribe(args []string) error {
	ci, err := c.clusterID("describe", args)
	if err != nil {
		return err
	}
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	if viper.GetBool("no-machines") {
		findRequest.WithReturnMachines(pointer.Pointer(false))
	}
	shoot, err := c.cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}
	return output.New().Print(shoot.Payload)
}

func (c *config) clusterIssues(args []string) error {
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
			response, err := c.cloud.Cluster.FindClusters(fcp, nil)
			if err != nil {
				return err
			}
			return output.New().Print(output.ShootIssuesResponses(response.Payload))
		}

		request := cluster.NewListClustersParams().WithReturnMachines(&boolTrue)
		shoots, err := c.cloud.Cluster.ListClusters(request, nil)
		if err != nil {
			return err
		}
		return output.New().Print(output.ShootIssuesResponses(shoots.Payload))
	}

	ci, err := c.clusterID("issues", args)
	if err != nil {
		return err
	}
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	shoot, err := c.cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}
	return output.New().Print(output.ShootIssuesResponse(shoot.Payload))
}

func (c *config) clusterMachines(args []string) error {
	ci, err := c.clusterID("machines", args)
	if err != nil {
		return err
	}
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	shoot, err := c.cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}

	if output.New().Type() != "table" {
		return output.New().Print(shoot.Payload)
	}

	fmt.Println("Cluster:")
	must(output.New().Print(shoot.Payload))

	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)
	fmt.Println("\nMachines:")
	return output.New().Print(ms)
}

func (c *config) clusterLogs(args []string) error {
	ci, err := c.clusterID("logs", args)
	if err != nil {
		return err
	}
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(ci)
	shoot, err := c.cloud.Cluster.FindCluster(findRequest, nil)
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

	if output.New().Type() != "table" {
		type s struct {
			Conditions    []*models.V1beta1Condition
			LastOperation *models.V1beta1LastOperation
			LastErrors    []*models.V1beta1LastError
		}
		return output.New().Print(s{
			Conditions:    conditions,
			LastOperation: lastOperation,
			LastErrors:    lastErrors,
		})
	}

	fmt.Println("Conditions:")
	err = output.New().Print(conditions)
	if err != nil {
		return err
	}

	fmt.Println("\nLast Errors:")
	err = output.New().Print(lastErrors)
	if err != nil {
		return err
	}

	fmt.Println("\nLast Operation:")
	return output.New().Print(lastOperation)
}

func (c *config) clusterInputs() error {
	request := cluster.NewListConstraintsParams()
	partition := viper.GetString("partition")
	if partition != "" {
		request.WithPartition(&partition)
	}
	sc, err := c.cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(sc)
}

func (c *config) clusterSplunkConfigManifest() error {
	secret := corev1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "splunk-config", Namespace: "kube-system"},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{},
		Data:       map[string][]byte{},
	}
	if viper.IsSet("token") {
		secret.StringData["hecToken"] = viper.GetString("token")
	}
	if viper.IsSet("index") {
		secret.StringData["index"] = viper.GetString("index")
	}
	if viper.IsSet("hechost") {
		secret.StringData["hecHost"] = viper.GetString("hechost")
	}
	if viper.IsSet("hecport") {
		secret.StringData["hecPort"] = strconv.Itoa(viper.GetInt("hecport"))
	}
	if viper.IsSet("tls") {
		if !viper.IsSet("cafile") && !viper.IsSet("cabase64") {
			return fmt.Errorf("you need to supply a ca certificate when using TLS")
		}
		secret.StringData["tlsEnabled"] = strconv.FormatBool(viper.GetBool("tls"))
	}
	if viper.IsSet("cafile") {
		if viper.IsSet("cabase64") {
			return fmt.Errorf("specify ca certificate either through cafile or through cabase64, do not use both flags")
		}
		hecCAFile, err := os.ReadFile(viper.GetString("cafile"))
		if err != nil {
			return err
		}
		secret.StringData["hecCAFile"] = string(hecCAFile)
	}
	if viper.IsSet("cabase64") {
		hecCAFileString := viper.GetString("cabase64")
		_, err := base64.StdEncoding.DecodeString(hecCAFileString)
		if err != nil {
			return fmt.Errorf("unable to decode ca file string:%w", err)
		}
		secret.Data["hecCAFile"] = []byte(hecCAFileString)
	}

	helper.MustPrintKubernetesResource(secret)

	return nil
}

func (c *config) clusterDNSManifest(args []string) error {
	ci, err := c.clusterID("dns-manifest", args)
	if err != nil {
		return err
	}

	cluster, err := c.cloud.Cluster.FindCluster(cluster.NewFindClusterParams().WithID(ci).WithReturnMachines(pointer.Pointer(false)), nil)
	if err != nil {
		return err
	}

	domain := fmt.Sprintf("%s.%s", viper.GetString("name"), pointer.SafeDeref(cluster.Payload.DNSEndpoint))

	var resource runtime.Object

	annotations := map[string]string{
		"dns.gardener.cloud/class":    "garden",
		"dns.gardener.cloud/ttl":      strconv.Itoa(viper.GetInt("ttl")),
		"dns.gardener.cloud/dnsnames": domain,
	}

	if viper.GetBool("with-certificate") {
		annotations["cert.gardener.cloud/purpose"] = "managed"
	}

	switch t := viper.GetString("type"); t {
	case "ingress":
		ingress := &networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Ingress",
				APIVersion: networkingv1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      viper.GetString("name"),
				Namespace: viper.GetString("namespace"),
				Labels: map[string]string{
					"app": viper.GetString("name"),
				},
				Annotations: annotations,
			},
			Spec: networkingv1.IngressSpec{
				IngressClassName: pointer.Pointer(viper.GetString("ingress-class")),
				Rules: []networkingv1.IngressRule{
					{
						Host: domain,
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										PathType: pointer.Pointer(networkingv1.PathTypePrefix),
										Path:     "/",
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: viper.GetString("backend-name"),
												Port: networkingv1.ServiceBackendPort{
													Number: viper.GetInt32("backend-port"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		if viper.GetBool("with-certificate") {
			ingress.Spec.TLS = []networkingv1.IngressTLS{
				{
					Hosts:      []string{domain},
					SecretName: fmt.Sprintf("%s-tls-secret", viper.GetString("name")),
				},
			}
		}

		resource = ingress
	case "service":
		service := &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      viper.GetString("name"),
				Namespace: viper.GetString("namespace"),
				Labels: map[string]string{
					"app": viper.GetString("name"),
				},
				Annotations: annotations,
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Selector: map[string]string{
					"app": viper.GetString("name"),
				},
				Ports: []corev1.ServicePort{
					{
						Name:       viper.GetString("backend-name"),
						Port:       viper.GetInt32("backend-port"),
						TargetPort: intstr.FromInt(int(viper.GetInt32("backend-port"))),
						Protocol:   corev1.ProtocolTCP,
					},
				},
			},
		}

		if viper.GetBool("with-certificate") {
			service.Annotations["cert.gardener.cloud/secretname"] = fmt.Sprintf("%s-tls-secret", viper.GetString("name"))
		}

		resource = service
	default:
		return fmt.Errorf("type must be one of %s, found: %s", []string{"loadbalancer", "service"}, t)
	}

	helper.MustPrintKubernetesResource(resource)

	return nil
}

func (c *config) clusterMachineReset(args []string) error {
	cid, err := c.clusterID("reset", args)
	if err != nil {
		return err
	}
	mid := viper.GetString("machineid")

	request := cluster.NewResetMachineParams()
	request.SetID(cid)
	request.Body = &models.V1ClusterMachineResetRequest{Machineid: &mid}

	shoot, err := c.cloud.Cluster.ResetMachine(request, nil)
	if err != nil {
		return err
	}

	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)

	return output.New().Print(ms)
}

func (c *config) clusterMachineCycle(args []string) error {
	cid, err := c.clusterID("reset", args)
	if err != nil {
		return err
	}
	mid := viper.GetString("machineid")

	request := cluster.NewCycleMachineParams()
	request.SetID(cid)
	request.Body = &models.V1ClusterMachineCycleRequest{Machineid: &mid}

	shoot, err := c.cloud.Cluster.CycleMachine(request, nil)
	if err != nil {
		return err
	}

	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)

	return output.New().Print(ms)
}

func (c *config) clusterMachineReinstall(args []string) error {
	cid, err := c.clusterID("reinstall", args)
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

	shoot, err := c.cloud.Cluster.ReinstallMachine(request, nil)
	if err != nil {
		return err
	}

	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)

	return output.New().Print(ms)
}

func (c *config) clusterMonitoringSecret(args []string) error {
	cid, err := c.clusterID("monitoring-secret", args)
	if err != nil {
		return err
	}

	secret, err := c.cloud.Cluster.GetMonitoringSecret(cluster.NewGetMonitoringSecretParams().WithID(cid), nil)
	if err != nil {
		return err
	}

	return output.New().Print(secret.Payload)
}

func (c *config) clusterMachineSSH(args []string, console bool) error {
	cid, err := c.clusterID("ssh", args)
	if err != nil {
		return err
	}
	mid := viper.GetString("machineid")

	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(cid)
	shoot, err := c.cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}

	keypair, vpn, err := c.sshKeyPair(cid)
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
			privateKeyFile := path.Join(home, "."+c.name, "."+cid+".id_rsa")
			err = os.WriteFile(privateKeyFile, keypair.privatekey, 0600)
			if err != nil {
				return fmt.Errorf("unable to write private key:%s error:%w", privateKeyFile, err)
			}
			defer os.Remove(privateKeyFile)
			if console {
				fmt.Printf("access console via ssh\n")
				authContext, err := api.GetAuthContext(viper.GetString("kubeconfig"))
				if err != nil {
					return err
				}
				err = os.Setenv("LC_METAL_STACK_OIDC_TOKEN", authContext.IDToken)
				if err != nil {
					return err
				}
				bmcConsolePort := "5222"
				err = runSSH("-i", privateKeyFile, mid+"@"+c.consoleHost, "-p", bmcConsolePort)
				return err
			}
			networks := m.Allocation.Networks
			switch *m.Allocation.Role {
			case "firewall":
				if vpn != nil {
					return c.firewallSSHViaVPN(*m.ID, keypair.privatekey, vpn)
				}

				for _, nw := range networks {
					if *nw.Underlay || *nw.Private {
						continue
					}
					for _, ip := range nw.Ips {
						if portOpen(ip, "22", time.Second) {
							err := runSSH("-i", privateKeyFile, "metal"+"@"+ip)
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
				return fmt.Errorf("unknown machine role:%s", *m.Allocation.Role)
			}
		}
	}

	return fmt.Errorf("machine:%s not found in cluster:%s", mid, cid)
}

func runSSH(args ...string) error {
	path, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("unable to locate ssh in path")
	}
	args = append(args, "-o", "StrictHostKeyChecking=No")
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

func (c *config) clusterID(verb string, args []string) (string, error) {
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
