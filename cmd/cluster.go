package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"slices"
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
	"github.com/metal-stack/metal-lib/pkg/genericcli"
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
	utiltaints "github.com/gardener/machine-controller-manager/pkg/util/taints"
)

func newClusterCmd(c *config) *cobra.Command {
	clusterCmd := &cobra.Command{
		Use:   "cluster",
		Short: "manage clusters",
	}
	clusterCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterCreate()
		},
	}

	clusterListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list clusters",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterList()
		},
	}
	clusterDeleteCmd := &cobra.Command{
		Use:     "delete <clusterid>",
		Short:   "delete a cluster",
		Aliases: []string{"destroy", "rm", "remove"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterDelete(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterDescribeCmd := &cobra.Command{
		Use:   "describe <clusterid>",
		Short: "describe a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterDescribe(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterKubeconfigCmd := &cobra.Command{
		Use:   "kubeconfig <clusterid>",
		Short: "get cluster kubeconfig",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterKubeconfig(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}

	clusterReconcileCmd := &cobra.Command{
		Use:   "reconcile <clusterid>",
		Short: "trigger cluster reconciliation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.reconcileCluster(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterUpdateCmd := &cobra.Command{
		Use:   "update <clusterid>",
		Short: "update a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.updateCluster(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterInputsCmd := &cobra.Command{
		Use:   "inputs",
		Short: "get possible cluster inputs like k8s versions, etc.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterInputs()
		},
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
	}
	clusterIssuesCmd := &cobra.Command{
		Use:     "issues [<clusterid>]",
		Aliases: []string{"problems", "warnings"},
		Short:   "lists cluster issues, shows required actions explicitly when id argument is given",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterIssues(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterMonitoringSecretCmd := &cobra.Command{
		Use:   "monitoring-secret <clusterid>",
		Short: "returns the endpoint and access credentials to the monitoring of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMonitoringSecret(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterMachineSSHCmd := &cobra.Command{
		Use:   "ssh <clusterid>",
		Short: "ssh access a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachineSSH(args, false)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterMachineConsoleCmd := &cobra.Command{
		Use:   "console <clusterid>",
		Short: "console access a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachineSSH(args, true)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterMachineResetCmd := &cobra.Command{
		Use:   "reset <clusterid>",
		Short: "hard power reset of a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachineReset(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterMachineCycleCmd := &cobra.Command{
		Use:   "cycle <clusterid>",
		Short: "soft power cycle of a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachineCycle(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterMachineReinstallCmd := &cobra.Command{
		Use:   "reinstall <clusterid>",
		Short: "reinstall OS image onto a machine/firewall of the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachineReinstall(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterMachinePackagesCmd := &cobra.Command{
		Use:   "packages <clusterid>",
		Short: "show packages of the os image which is installed on this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterMachinePackages(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterLogsCmd := &cobra.Command{
		Use:   "logs <clusterid>",
		Short: "get logs for the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterLogs(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
	}
	clusterDNSManifestCmd := &cobra.Command{
		Use:   "dns-manifest <clusterid>",
		Short: "create a manifest for an ingress or service type loadbalancer, creating a DNS entry and valid certificate within your cluster domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.clusterDNSManifest(args)
		},
		ValidArgsFunction: c.comp.ClusterListCompletion,
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
	clusterCreateCmd.Flags().Int32("minsize", 1, "minimal workers of the cluster.")
	clusterCreateCmd.Flags().Int32("maxsize", 1, "maximal workers of the cluster.")
	clusterCreateCmd.Flags().String("maxsurge", "1", "max number (e.g. 1) or percentage (e.g. 10%) of workers created during a update of the cluster.")
	clusterCreateCmd.Flags().String("maxunavailable", "0", "max number (e.g. 0) or percentage (e.g. 10%) of workers that can be unavailable during a update of the cluster.")
	clusterCreateCmd.Flags().StringSlice("labels", []string{}, "labels of the cluster")
	clusterCreateCmd.Flags().StringSlice("external-networks", []string{}, "external networks of the cluster")
	clusterCreateCmd.Flags().StringSlice("egress", []string{}, "static egress ips per network, must be in the form <network>:<ip>; e.g.: --egress internet:1.2.3.4,extnet:123.1.1.1 --egress internet:1.2.3.5 [optional]")
	clusterCreateCmd.Flags().String("default-pod-security-standard", "", "sets default pod security standard for clusters >= v1.23.x, defaults to restricted on clusters >= v1.25 (valid values: empty string, privileged, baseline, restricted)")
	clusterCreateCmd.Flags().Duration("healthtimeout", 0, "period (e.g. \"24h\") after which an unhealthy node is declared failed and will be replaced. [optional]")
	clusterCreateCmd.Flags().Duration("draintimeout", 0, "period (e.g. \"3h\") after which a draining node will be forcefully deleted. [optional]")
	clusterCreateCmd.Flags().Bool("encrypted-storage-classes", false, "enables the deployment of encrypted duros storage classes into the cluster. please refer to the user manual to properly use volume encryption. [optional]")
	clusterCreateCmd.Flags().BoolP("autoupdate-kubernetes", "", false, "enables automatic updates of the kubernetes patch version of the cluster [optional]")
	clusterCreateCmd.Flags().BoolP("autoupdate-machineimages", "", false, "enables automatic updates of the worker node images of the cluster, be aware that this rolls worker nodes! [optional]")
	clusterCreateCmd.Flags().Bool("autoupdate-firewallimage", false, "enables automatic updates of the firewall image, be aware that this rolls firewalls! [optional]")
	clusterCreateCmd.Flags().String("maintenance-begin", "220000+0100", "defines the beginning of the nightly maintenance time window (e.g. for autoupdates) in the format HHMMSS+ZONE, e.g. \"220000+0100\". [optional]")
	clusterCreateCmd.Flags().String("maintenance-end", "233000+0100", "defines the end of the nightly maintenance time window (e.g. for autoupdates) in the format HHMMSS+ZONE, e.g. \"233000+0100\". [optional]")
	clusterCreateCmd.Flags().String("default-storage-class", "", "set default storage class to given name, must be one of the managed storage classes")
	clusterCreateCmd.Flags().String("max-pods-per-node", "", "set number of maximum pods per node (default: 510). Lower numbers allow for more node per cluster. [optional]")
	clusterCreateCmd.Flags().String("cni", "", "the network plugin used in this cluster, defaults to calico. please note that cilium support is still Alpha and we are happy to receive feedback. [optional]")
	clusterCreateCmd.Flags().Bool("enable-calico-ebpf", false, "enables calico cni to use eBPF data plane and DSR configuration, for increased performance and preserving source IP addresses. [optional]")
	clusterCreateCmd.Flags().BoolP("enable-node-local-dns", "", false, "enables node local dns cache on the cluster nodes. [optional].")
	clusterCreateCmd.Flags().BoolP("disable-forwarding-to-upstream-dns", "", false, "disables direct forwarding of queries to external dns servers when node-local-dns is enabled. All dns queries will go through coredns. [optional].")
	clusterCreateCmd.Flags().StringSlice("kube-apiserver-acl-allowed-cidrs", []string{}, "comma-separated list of external CIDRs allowed to connect to the kube-apiserver (e.g. \"212.34.68.0/24,212.34.89.0/27\")")
	clusterCreateCmd.Flags().Bool("enable-kube-apiserver-acl", false, "restricts access from outside to the kube-apiserver to the source ip addresses set by --kube-apiserver-acl-allowed-cidrs [optional].")
	clusterCreateCmd.Flags().String("network-isolation", "", "defines restrictions to external network communication for the cluster, can be one of baseline|restricted|isolated. baseline sets no special restrictions to external networks, restricted by default only allows external traffic to explicitly allowed destinations, forbidden disallows communication with external networks except for a limited set of networks. Please consult the documentation for detailed descriptions of the individual modes as these cannot be altered anymore after creation. [optional]")
	clusterCreateCmd.Flags().Bool("high-availability-control-plane", false, "enables a high availability control plane for the cluster, cannot be disabled again")
	clusterCreateCmd.Flags().Int64("kubelet-pod-pid-limit", 0, "controls the maximum number of process IDs per pod allowed by the kubelet")

	genericcli.Must(clusterCreateCmd.MarkFlagRequired("name"))
	genericcli.Must(clusterCreateCmd.MarkFlagRequired("project"))
	genericcli.Must(clusterCreateCmd.MarkFlagRequired("partition"))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("seed", c.comp.SeedListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("external-networks", c.comp.NetworkListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("version", c.comp.VersionListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("machinetype", c.comp.MachineTypeListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("machineimage", c.comp.MachineImageListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("firewalltype", c.comp.FirewallTypeListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("firewallimage", c.comp.FirewallImageListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("firewallcontroller", c.comp.FirewallControllerVersionListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("purpose", c.comp.ClusterPurposeListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("default-pod-security-standard", c.comp.PodSecurityListCompletion))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("cni", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"calico\tcalico networking plugin. this is the cluster default.",
			"cilium\tcilium networking plugin. please note that cilium support is still Alpha and we are happy to receive feedback.",
		}, cobra.ShellCompDirectiveNoFileComp
	}))
	genericcli.Must(clusterCreateCmd.RegisterFlagCompletionFunc("network-isolation", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			models.V1ClusterCreateRequestNetworkAccessTypeBaseline + "\tno special restrictions for external network traffic, service type loadbalancer possible in all networks",
			models.V1ClusterCreateRequestNetworkAccessTypeRestricted + "\texternal network traffic needs to be allowed explicitly, own cluster wide network policies possible, service type loadbalancer possible in all networks",
			models.V1ClusterCreateRequestNetworkAccessTypeForbidden + "\texternal network traffic is not possible except for allowed networks , own cluster wide network policies not possible, service type loadbalancer possible only in allowed networks, for the allowed networks please see cluster inputs",
		}, cobra.ShellCompDirectiveNoFileComp
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
	genericcli.Must(clusterListCmd.RegisterFlagCompletionFunc("id", c.comp.ClusterListCompletion))
	genericcli.Must(clusterListCmd.RegisterFlagCompletionFunc("name", c.comp.ClusterNameCompletion))
	genericcli.Must(clusterListCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	genericcli.Must(clusterListCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	genericcli.Must(clusterListCmd.RegisterFlagCompletionFunc("seed", c.comp.SeedListCompletion))
	genericcli.Must(clusterListCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(clusterListCmd.RegisterFlagCompletionFunc("purpose", c.comp.ClusterPurposeListCompletion))

	// Cluster update --------------------------------------------------------------------
	clusterUpdateCmd.Flags().String("workergroup", "", "the name of the worker group to apply updates to, only required when there are multiple worker groups.")
	clusterUpdateCmd.Flags().Bool("remove-workergroup", false, "if set, removes the targeted worker group")
	clusterUpdateCmd.Flags().StringSlice("workerlabels", []string{}, "labels of the worker group (syncs to kubernetes node resource after some time, too)")
	clusterUpdateCmd.Flags().StringSlice("workerannotations", []string{}, "annotations of the worker group (syncs to kubernetes node resource after some time, too)")
	clusterUpdateCmd.Flags().StringSlice("workertaints", []string{}, "list of taints to set for nodes of the worker group. (use empty string to remove previous set taints)")
	clusterUpdateCmd.Flags().String("workerversion", "", "set custom kubernetes version of the worker group independent of the api server. note that the worker version may only be two minor version older than the api server as stated in the official kubernetes version skew policy. (set to \"\" to remove custom kubernetes version)")
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
	clusterUpdateCmd.Flags().String("default-pod-security-standard", "", "set default pod security standard for cluster >=v 1.23.x, send empty string explicitly to disable pod security standards (valid values: empty string, privileged, baseline, restricted)")
	clusterUpdateCmd.Flags().String("purpose", "", fmt.Sprintf("purpose of the cluster, can be one of %s. SLA is only given on production clusters.", strings.Join(completion.ClusterPurposes, "|")))
	clusterUpdateCmd.Flags().StringSlice("egress", []string{}, "static egress ips per network, must be in the form <networkid>:<semicolon-separated ips>; e.g.: --egress internet:1.2.3.4;1.2.3.5 --egress extnet:123.1.1.1 [optional]. Use \"--egress none\" to remove all egress rules.")
	clusterUpdateCmd.Flags().StringSlice("external-networks", []string{}, "external networks of the cluster")
	clusterUpdateCmd.Flags().Duration("healthtimeout", 0, "period (e.g. \"24h\") after which an unhealthy node is declared failed and will be replaced. (0 = provider-default)")
	clusterUpdateCmd.Flags().Duration("draintimeout", 0, "period (e.g. \"3h\") after which a draining node will be forcefully deleted. (0 = provider-default)")
	clusterUpdateCmd.Flags().String("maxsurge", "", "max number (e.g. 1) or percentage (e.g. 10%) of workers created during a update of the cluster.")
	clusterUpdateCmd.Flags().String("maxunavailable", "", "max number (e.g. 0) or percentage (e.g. 10%) of workers that can be unavailable during a update of the cluster.")
	clusterUpdateCmd.Flags().BoolP("autoupdate-kubernetes", "", false, "enables automatic updates of the kubernetes patch version of the cluster")
	clusterUpdateCmd.Flags().BoolP("autoupdate-machineimages", "", false, "enables automatic updates of the worker node images of the cluster, be aware that this deletes worker nodes!")
	clusterUpdateCmd.Flags().Bool("autoupdate-firewallimage", false, "enables automatic updates of the firewall image, be aware that this rolls firewalls! [optional]")
	clusterUpdateCmd.Flags().String("maintenance-begin", "", "defines the beginning of the nightly maintenance time window (e.g. for autoupdates) in the format HHMMSS+ZONE, e.g. \"220000+0100\". [optional]")
	clusterUpdateCmd.Flags().String("maintenance-end", "", "defines the end of the nightly maintenance time window (e.g. for autoupdates) in the format HHMMSS+ZONE, e.g. \"233000+0100\". [optional]")
	clusterUpdateCmd.Flags().Bool("encrypted-storage-classes", false, "enables the deployment of encrypted duros storage classes into the cluster. please refer to the user manual to properly use volume encryption.")
	clusterUpdateCmd.Flags().String("default-storage-class", "", "set default storage class to given name, must be one of the managed storage classes")
	clusterUpdateCmd.Flags().BoolP("disable-custom-default-storage-class", "", false, "if set to true, no default class is deployed, you have to set one of your storageclasses manually to default")
	clusterUpdateCmd.Flags().BoolP("enable-node-local-dns", "", false, "enables node local dns cache on the cluster nodes. [optional]. WARNING: changing this value will lead to rolling of the worker nodes [optional]")
	clusterUpdateCmd.Flags().BoolP("disable-forwarding-to-upstream-dns", "", false, "disables direct forwarding of queries to external dns servers when node-local-dns is enabled. All dns queries will go through coredns [optional].")
	clusterUpdateCmd.Flags().StringSlice("kube-apiserver-acl-set-allowed-cidrs", []string{}, "comma-separated list of external CIDRs allowed to connect to the kube-apiserver (e.g. \"212.34.68.0/24,212.34.89.0/27\")")
	clusterUpdateCmd.Flags().StringSlice("kube-apiserver-acl-add-to-allowed-cidrs", []string{}, "comma-separated list of external CIDRs to add to the allowed CIDRs to connect to the kube-apiserver (e.g. \"212.34.68.0/24,212.34.89.0/27\")")
	clusterUpdateCmd.Flags().StringSlice("kube-apiserver-acl-remove-from-allowed-cidrs", []string{}, "comma-separated list of external CIDRs to be removed from the allowed CIDRs to connect to the kube-apiserver (e.g. \"212.34.68.0/24,212.34.89.0/27\")")
	clusterUpdateCmd.Flags().Bool("enable-kube-apiserver-acl", false, "restricts access from outside to the kube-apiserver to the source ip addresses set by --kube-apiserver-acl-* [optional].")
	clusterUpdateCmd.Flags().Bool("high-availability-control-plane", false, "enables a high availability control plane for the cluster, cannot be disabled again")
	clusterUpdateCmd.Flags().Int64("kubelet-pod-pid-limit", 0, "controls the maximum number of process IDs per pod allowed by the kubelet")
	clusterUpdateCmd.Flags().Bool("enable-calico-ebpf", false, "enables calico cni to use eBPF data plane and DSR configuration, for increased performance and preserving source IP addresses. [optional]")

	genericcli.Must(clusterUpdateCmd.RegisterFlagCompletionFunc("version", c.comp.VersionListCompletion))
	genericcli.Must(clusterUpdateCmd.RegisterFlagCompletionFunc("workerversion", c.comp.VersionListCompletion))
	genericcli.Must(clusterUpdateCmd.RegisterFlagCompletionFunc("firewalltype", c.comp.FirewallTypeListCompletion))
	genericcli.Must(clusterUpdateCmd.RegisterFlagCompletionFunc("firewallimage", c.comp.FirewallImageListCompletion))
	genericcli.Must(clusterUpdateCmd.RegisterFlagCompletionFunc("seed", c.comp.SeedListCompletion))
	genericcli.Must(clusterUpdateCmd.RegisterFlagCompletionFunc("firewallcontroller", c.comp.FirewallControllerVersionListCompletion))
	genericcli.Must(clusterUpdateCmd.RegisterFlagCompletionFunc("machinetype", c.comp.MachineTypeListCompletion))
	genericcli.Must(clusterUpdateCmd.RegisterFlagCompletionFunc("machineimage", c.comp.MachineImageListCompletion))
	genericcli.Must(clusterUpdateCmd.RegisterFlagCompletionFunc("purpose", c.comp.ClusterPurposeListCompletion))
	genericcli.Must(clusterUpdateCmd.RegisterFlagCompletionFunc("default-pod-security-standard", c.comp.PodSecurityListCompletion))

	clusterInputsCmd.Flags().String("partition", "", "partition of the constraints.")
	genericcli.Must(clusterInputsCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))

	// Cluster dns manifest --------------------------------------------------------------------
	clusterDNSManifestCmd.Flags().String("type", "ingress", "either of type ingress or service")
	clusterDNSManifestCmd.Flags().String("name", "<name>", "the resource name")
	clusterDNSManifestCmd.Flags().String("namespace", "default", "the resource's namespace")
	clusterDNSManifestCmd.Flags().Int("ttl", 180, "the ttl set to the created dns entry")
	clusterDNSManifestCmd.Flags().Bool("with-certificate", true, "whether to request a let's encrypt certificate for the requested dns entry or not")
	clusterDNSManifestCmd.Flags().String("backend-name", "my-backend", "the name of the backend")
	clusterDNSManifestCmd.Flags().Int32("backend-port", 443, "the port of the backend")
	clusterDNSManifestCmd.Flags().String("ingress-class", "nginx", "the ingress class name")
	genericcli.Must(clusterDNSManifestCmd.RegisterFlagCompletionFunc("type", cobra.FixedCompletions([]string{"ingress", "service"}, cobra.ShellCompDirectiveNoFileComp)))

	// Cluster machine ... --------------------------------------------------------------------
	clusterMachineSSHCmd.Flags().String("machineid", "", "machine to connect to.")
	genericcli.Must(clusterMachineSSHCmd.MarkFlagRequired("machineid"))
	genericcli.Must(clusterMachineSSHCmd.RegisterFlagCompletionFunc("machineid", c.comp.ClusterFirewallListCompletion))

	clusterMachineConsoleCmd.Flags().String("machineid", "", "machine to connect to.")
	genericcli.Must(clusterMachineConsoleCmd.MarkFlagRequired("machineid"))
	genericcli.Must(clusterMachineConsoleCmd.RegisterFlagCompletionFunc("machineid", c.comp.ClusterMachineListCompletion))

	clusterMachineResetCmd.Flags().String("machineid", "", "machine to reset.")
	genericcli.Must(clusterMachineResetCmd.MarkFlagRequired("machineid"))
	genericcli.Must(clusterMachineResetCmd.RegisterFlagCompletionFunc("machineid", c.comp.ClusterMachineListCompletion))

	clusterMachineCycleCmd.Flags().String("machineid", "", "machine to reset.")
	genericcli.Must(clusterMachineCycleCmd.MarkFlagRequired("machineid"))
	genericcli.Must(clusterMachineCycleCmd.RegisterFlagCompletionFunc("machineid", c.comp.ClusterMachineListCompletion))

	clusterMachineReinstallCmd.Flags().String("machineid", "", "machine to reinstall.")
	clusterMachineReinstallCmd.Flags().String("machineimage", "", "image to reinstall (optional).")
	genericcli.Must(clusterMachineReinstallCmd.MarkFlagRequired("machineid"))
	genericcli.Must(clusterMachineReinstallCmd.RegisterFlagCompletionFunc("machineid", c.comp.ClusterMachineListCompletion))

	clusterMachinePackagesCmd.Flags().String("machineid", "", "machine to connect to.")
	genericcli.Must(clusterMachinePackagesCmd.MarkFlagRequired("machineid"))
	genericcli.Must(clusterMachinePackagesCmd.RegisterFlagCompletionFunc("machineid", c.comp.ClusterMachineListCompletion))

	clusterMachineCmd.AddCommand(clusterMachineListCmd)
	clusterMachineCmd.AddCommand(clusterMachineSSHCmd)
	clusterMachineCmd.AddCommand(clusterMachineConsoleCmd)
	clusterMachineCmd.AddCommand(clusterMachineResetCmd)
	clusterMachineCmd.AddCommand(clusterMachineCycleCmd)
	clusterMachineCmd.AddCommand(clusterMachineReinstallCmd)
	clusterMachineCmd.AddCommand(clusterMachinePackagesCmd)

	clusterReconcileCmd.Flags().String("operation", models.V1ClusterReconcileRequestOperationReconcile, fmt.Sprintf("Executes a cluster \"reconcile\" operation, can be one of %s.", strings.Join(completion.ClusterReconcileOperations, "|")))
	genericcli.Must(clusterReconcileCmd.RegisterFlagCompletionFunc("operation", c.comp.ClusterReconcileOperationCompletion))

	clusterIssuesCmd.Flags().String("id", "", "show clusters of given id")
	clusterIssuesCmd.Flags().String("name", "", "show clusters of given name")
	clusterIssuesCmd.Flags().String("project", "", "show clusters of given project")
	clusterIssuesCmd.Flags().String("partition", "", "show clusters in partition")
	clusterIssuesCmd.Flags().String("tenant", "", "show clusters of given tenant")

	genericcli.Must(clusterIssuesCmd.RegisterFlagCompletionFunc("name", c.comp.ClusterNameCompletion))
	genericcli.Must(clusterIssuesCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	genericcli.Must(clusterIssuesCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	genericcli.Must(clusterIssuesCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))

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
	clusterCmd.AddCommand(clusterDNSManifestCmd)
	clusterCmd.AddCommand(clusterMonitoringSecretCmd)
	clusterCmd.AddCommand(newClusterAuditCmd(c))

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
	encryptedStorageClasses := strconv.FormatBool(viper.GetBool("encrypted-storage-classes"))
	enableNodeLocalDNS := viper.GetBool("enable-node-local-dns")
	disableForwardToUpstreamDNS := viper.GetBool("disable-forwarding-to-upstream-dns")
	highAvailability := strconv.FormatBool(viper.GetBool("high-availability-control-plane"))
	podpidLimit := viper.GetInt64("kubelet-pod-pid-limit")
	calicoEbpf := strconv.FormatBool(viper.GetBool("enable-calico-ebpf"))

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

	var defaultPodSecurityStandard *string
	if viper.IsSet("default-pod-security-standard") {
		defaultPodSecurityStandard = pointer.Pointer(viper.GetString("default-pod-security-standard"))
	}

	var networkAccessType *string
	if viper.IsSet("network-isolation") {
		networkAccessType = pointer.Pointer(viper.GetString("network-isolation"))
		switch *networkAccessType {
		case models.V1ClusterCreateRequestNetworkAccessTypeForbidden:
			fmt.Printf(`
WARNING: You are going to create a cluster which has no internet access with the following consequences:
- pulling images is only possible from private registries you provide, these registries must be resolvable from the public dns, their IP must be located in one of the allowed networks (see cluster inputs), and must be secured with a trusted TLS certificate
- service type loadbalancer can only be created in networks which are specified in the allowed networks (see cluster inputs)
- cluster wide network policies can only be created in certain network ranges which are specified in the allowed networks (see cluster inputs)
- It is not possible to change this cluster back to %q after creation
`, models.V1ClusterCreateRequestNetworkAccessTypeBaseline)
			err := helper.Prompt("Are you sure? (y/n)", "y")
			if err != nil {
				return err
			}
		case models.V1ClusterCreateRequestNetworkAccessTypeRestricted:
			fmt.Printf(`
WARNING: You are going to create a cluster that has no default internet access with the following consequences:
- pulling images is only possible from private registries you provide, these registries must be resolvable from the public dns and must be secured with a trusted TLS certificate
- you can create cluster wide network policies to the outside world without restrictions
- pulling container images from registries requires to create a corresponding CWNP to these registries
- It is not possible to change this cluster back to %q after creation
`, models.V1ClusterCreateRequestNetworkAccessTypeBaseline)
			err := helper.Prompt("Are you sure? (y/n)", "y")
			if err != nil {
				return err
			}
		case models.V1ClusterCreateRequestNetworkAccessTypeBaseline:
			// Noop
		}
	}

	labels := viper.GetStringSlice("labels")

	// FIXME helper and validation
	networks := viper.GetStringSlice("external-networks")
	egress := viper.GetStringSlice("egress")
	maintenanceBegin := viper.GetString("maintenance-begin")
	maintenanceEnd := viper.GetString("maintenance-end")

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
			},
		},
		FirewallSize:              &firewallType,
		FirewallImage:             &firewallImage,
		FirewallControllerVersion: &firewallController,
		Kubernetes: &models.V1Kubernetes{
			Version:                    &version,
			DefaultPodSecurityStandard: defaultPodSecurityStandard,
		},
		Maintenance: &models.V1Maintenance{
			TimeWindow: &models.V1MaintenanceTimeWindow{
				Begin: &maintenanceBegin,
				End:   &maintenanceEnd,
			},
		},
		AdditionalNetworks: networks,
		PartitionID:        &partition,
		ClusterFeatures: &models.V1ClusterFeatures{
			LogAcceptedConnections: &logAcceptedConnections,
			DurosStorageEncryption: &encryptedStorageClasses,
		},
		CustomDefaultStorageClass: customDefaultStorageClass,
		Cni:                       cni,
		NetworkAccessType:         networkAccessType,
	}

	if viper.IsSet("autoupdate-kubernetes") ||
		viper.IsSet("autoupdate-machineimages") ||
		viper.IsSet("autoupdate-firewallimage") ||
		purpose == string(v1beta1.ShootPurposeEvaluation) {

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
		if viper.IsSet("autoupdate-firewallimage") {
			auto := viper.GetBool("autoupdate-firewallimage")
			scr.Maintenance.AutoUpdate.FirewallImage = &auto
		}
	}

	if viper.IsSet("max-pods-per-node") {
		scr.Kubernetes.MaxPodsPerNode = viper.GetInt32("max-pods-per-node")
	}
	if seed != "" {
		scr.SeedName = seed
	}

	if viper.IsSet("enable-node-local-dns") {
		if scr.SystemComponents == nil {
			scr.SystemComponents = &models.V1SystemComponents{}
		}
		if scr.SystemComponents.NodeLocalDNS == nil {
			scr.SystemComponents.NodeLocalDNS = &models.V1NodeLocalDNS{}
		}

		scr.SystemComponents.NodeLocalDNS.Enabled = &enableNodeLocalDNS
	}
	if viper.IsSet("disable-forwarding-to-upstream-dns") {
		if scr.SystemComponents == nil {
			scr.SystemComponents = &models.V1SystemComponents{}
		}
		if scr.SystemComponents.NodeLocalDNS == nil {
			scr.SystemComponents.NodeLocalDNS = &models.V1NodeLocalDNS{}
		}
		scr.SystemComponents.NodeLocalDNS.DisableForwardToUpstreamDNS = &disableForwardToUpstreamDNS
	}

	if viper.IsSet("kube-apiserver-acl-allowed-cidrs") || viper.IsSet("enable-kube-apiserver-acl") {
		if !viper.GetBool("yes-i-really-mean-it") && viper.IsSet("enable-kube-apiserver-acl") {
			return fmt.Errorf("--enable-kube-apiserver-acl is set but you forgot to add --yes-i-really-mean-it")
		}

		if viper.GetBool("enable-kube-apiserver-acl") {
			fmt.Println("WARNING: Restricting access to the kube-apiserver prevents FI-TS operators from helping you in case of any issues in your cluster.")
			err = helper.Prompt("Are you sure? (y/n)", "y")
			if err != nil {
				return err
			}
		}

		scr.KubeAPIServerACL = &models.V1KubeAPIServerACL{
			CIDRs:    viper.GetStringSlice("kube-apiserver-acl-allowed-cidrs"),
			Disabled: pointer.Pointer(!viper.GetBool("enable-kube-apiserver-acl")),
		}
	}

	if viper.IsSet("enable-calico-ebpf") {
		if activate, _ := strconv.ParseBool(calicoEbpf); activate {
			if err := genericcli.PromptCustom(&genericcli.PromptConfig{
				Message:     "Enabling the Calico eBPF feature gate is still a beta feature. Be aware that this may impact the network policies in your cluster as source IP addresses are preserved with this configuration.",
				ShowAnswers: true,
				Out:         c.out,
			}); err != nil {
				return err
			}
		}

		scr.ClusterFeatures.CalicoEbpfDataplane = &calicoEbpf
	}

	if viper.IsSet("high-availability-control-plane") {
		if ha, _ := strconv.ParseBool(highAvailability); ha {
			if err := genericcli.PromptCustom(&genericcli.PromptConfig{
				Message:     "Enabling the HA control plane feature gate is still a beta feature. You cannot use it in combination with the cluster forwarding backend of the audit extension. Please be aware that you cannot revert this feature gate after it was enabled.",
				ShowAnswers: true,
				Out:         c.out,
			}); err != nil {
				return err
			}
		}

		scr.ClusterFeatures.HighAvailability = &highAvailability
	}

	if viper.IsSet("kubelet-pod-pid-limit") {
		if !viper.GetBool("yes-i-really-mean-it") {
			return fmt.Errorf("--kubelet-pod-pid-limit can only be changed in combination with --yes-i-really-mean-it because this change can lead to pods not starting anymore in the cluster")
		}
		scr.Kubernetes.PodPIDsLimit = &podpidLimit
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
	return c.describePrinter.Print(shoot.Payload)
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
		return c.listPrinter.Print(response.Payload)
	}

	request := cluster.NewListClustersParams()
	shoots, err := c.cloud.Cluster.ListClusters(request, nil)
	if err != nil {
		return err
	}
	return c.listPrinter.Print(shoots.Payload)
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
	vpn        *models.V1VPN
}

func (c *config) sshKeyPair(clusterID string) (*sshkeypair, error) {
	request := cluster.NewGetSSHKeyPairParams()
	request.SetID(clusterID)
	credentials, err := c.cloud.Cluster.GetSSHKeyPair(request, nil)
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
		vpn:        credentials.Payload.VPN,
	}, nil
}

func (c *config) reconcileCluster(args []string) error {
	ci, err := c.clusterID("reconcile", args)
	if err != nil {
		return err
	}

	request := cluster.NewReconcileClusterParams()
	request.SetID(ci)

	operation := viper.GetString("operation")
	request.Body = &models.V1ClusterReconcileRequest{Operation: &operation}

	shoot, err := c.cloud.Cluster.ReconcileCluster(request, nil)
	if err != nil {
		return err
	}
	return c.describePrinter.Print(shoot.Payload)
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
	workertaintsslice := viper.GetStringSlice("workertaints")
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

	enableNodeLocalDNS := viper.GetBool("enable-node-local-dns")
	disableForwardToUpstreamDNS := viper.GetBool("disable-forwarding-to-upstream-dns")

	defaultStorageClass := viper.GetString("default-storage-class")
	disableDefaultStorageClass := viper.GetBool("disable-custom-default-storage-class")

	encryptedStorageClasses := strconv.FormatBool(viper.GetBool("encrypted-storage-classes"))
	highAvailability := strconv.FormatBool(viper.GetBool("high-availability-control-plane"))
	calicoEbpf := strconv.FormatBool(viper.GetBool("enable-calico-ebpf"))

	podpidLimit := viper.GetInt64("kubelet-pod-pid-limit")

	workerlabels, err := helper.LabelsToMap(workerlabelslice)
	if err != nil {
		return err
	}
	workerannotations, err := helper.LabelsToMap(workerannotationsslice)
	if err != nil {
		return err
	}
	coreworkertaints, _, err := utiltaints.ParseTaints(workertaintsslice)
	if err != nil {
		return fmt.Errorf("specified taints are invalid: %w", err)
	}

	var workertaints []*models.V1Taint
	for _, t := range coreworkertaints {
		t := t
		workertaints = append(workertaints, &models.V1Taint{
			Key:    &t.Key,
			Value:  t.Value,
			Effect: (*string)(&t.Effect),
		})
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
	if viper.IsSet("encrypted-storage-classes") {
		clusterFeatures.DurosStorageEncryption = &encryptedStorageClasses
	}
	if viper.IsSet("logacceptedconns") {
		clusterFeatures.LogAcceptedConnections = &logAcceptedConnections
	}
	if viper.IsSet("enable-calico-ebpf") {
		if activate, _ := strconv.ParseBool(calicoEbpf); activate {
			if err := genericcli.PromptCustom(&genericcli.PromptConfig{
				Message:     "Enabling the Calico eBPF feature gate is still a beta feature. Be aware that this may impact the network policies in your cluster as source IP addresses are preserved with this configuration.",
				ShowAnswers: true,
				Out:         c.out,
			}); err != nil {
				return err
			}
		}

		clusterFeatures.CalicoEbpfDataplane = &calicoEbpf
	}
	if viper.IsSet("high-availability-control-plane") {
		if v, _ := strconv.ParseBool(highAvailability); v {
			if err := genericcli.PromptCustom(&genericcli.PromptConfig{
				Message:     "Enabling the HA control plane feature gate is still a beta feature. You cannot use it in combination with the cluster forwarding backend of the audit extension. Please be aware that you cannot revert this feature gate after it was enabled.",
				ShowAnswers: true,
				Out:         c.out,
			}); err != nil {
				return err
			}
		}

		clusterFeatures.HighAvailability = &highAvailability
	}

	workergroupKubernetesVersion := viper.GetString("workerversion")

	request := cluster.NewUpdateClusterParams()
	cur := &models.V1ClusterUpdateRequest{
		ID: &ci,
		Maintenance: &models.V1Maintenance{
			AutoUpdate: &models.V1MaintenanceAutoUpdate{
				KubernetesVersion: current.Maintenance.AutoUpdate.KubernetesVersion,
				MachineImage:      current.Maintenance.AutoUpdate.MachineImage,
				FirewallImage:     current.Maintenance.AutoUpdate.FirewallImage,
			},
		},
		ClusterFeatures:           &clusterFeatures,
		CustomDefaultStorageClass: customDefaultStorageClass,
	}

	if workergroupname != "" ||
		minsize != 0 || maxsize != 0 || maxsurge != "" || maxunavailable != "" ||
		machineImageAndVersion != "" || machineType != "" ||
		viper.IsSet("healthtimeout") || viper.IsSet("draintimeout") ||
		viper.IsSet("workerlabels") || viper.IsSet("workerannotations") || viper.IsSet("workertaints") || viper.IsSet("workerversion") {

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
				fmt.Printf("Adding a new worker group to cluster:%q. Please note that running multiple worker groups leads to higher basic costs of the cluster!\n", *current.Name)
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
					Taints:         workertaints,
				}

				workers = append(workers, worker)
			}
		} else if len(workers) == 1 {
			worker = workers[0]
		} else {
			return fmt.Errorf("there are multiple worker groups, please specify the worker group you want to update with --workergroup")
		}

		if removeworkergroup {
			if worker == nil {
				return fmt.Errorf("worker group %s not found", workergroupname)
			}

			fmt.Printf("WARNING. Removing a worker group from cluster:%q cannot be undone and causes the loss of local data on the deleted nodes.\n", *current.Name)
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
					fmt.Printf("WARNING. New maxsize of cluster:%q is lower than currently active machines. A random worker node which is still in use will be removed.\n", *current.Name)
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

			if viper.IsSet("workertaints") {
				worker.Taints = workertaints
			}

			if viper.IsSet("workerversion") {
				if pointer.SafeDeref(worker.KubernetesVersion) != "" && workergroupKubernetesVersion == "" {
					fmt.Printf("WARNING. Removing the worker version override of cluster:%q may update your worker nodes to the version of the api server.\n", *current.Name)
					err = helper.Prompt("Are you sure? (y/n)", "y")
					if err != nil {
						return err
					}
				}
				worker.KubernetesVersion = pointer.Pointer(workergroupKubernetesVersion)
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
	if viper.IsSet("autoupdate-firewallimage") {
		auto := viper.GetBool("autoupdate-firewallimage")
		cur.Maintenance.AutoUpdate.FirewallImage = &auto
	}
	if viper.IsSet("maintenance-begin") {
		begin := viper.GetString("maintenance-begin")
		if cur.Maintenance.TimeWindow == nil {
			cur.Maintenance.TimeWindow = &models.V1MaintenanceTimeWindow{}
		}
		cur.Maintenance.TimeWindow.Begin = &begin
	}
	if viper.IsSet("maintenance-end") {
		if cur.Maintenance.TimeWindow == nil {
			cur.Maintenance.TimeWindow = &models.V1MaintenanceTimeWindow{}
		}
		end := viper.GetString("maintenance-end")
		cur.Maintenance.TimeWindow.End = &end
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

	if viper.IsSet("kube-apiserver-acl-set-allowed-cidrs") || viper.IsSet("enable-kube-apiserver-acl") ||
		viper.IsSet("kube-apiserver-acl-add-to-allowed-cidrs") || viper.IsSet("kube-apiserver-acl-remove-from-allowed-cidrs") {

		newACL := current.KubeAPIServerACL
		if newACL == nil {
			newACL = &models.V1KubeAPIServerACL{
				CIDRs:    []string{},
				Disabled: pointer.Pointer(true),
			}
		}

		if viper.IsSet("enable-kube-apiserver-acl") {
			if !viper.GetBool("yes-i-really-mean-it") {
				return fmt.Errorf("--enable-kube-apiserver-acl is set but you forgot to add --yes-i-really-mean-it")
			}
			newACL.Disabled = pointer.Pointer(!viper.GetBool("enable-kube-apiserver-acl"))
		}

		if viper.IsSet("enable-kube-apiserver-acl") && viper.GetBool("enable-kube-apiserver-acl") {
			fmt.Printf("WARNING: Restricting access of cluster:%q to the kube-apiserver prevents FI-TS operators from helping you in case of any issues in your cluster.\n", *current.Name)
			err = helper.Prompt("Are you sure? (y/n)", "y")
			if err != nil {
				return err
			}
		}

		for _, r := range viper.GetStringSlice("kube-apiserver-acl-remove-from-allowed-cidrs") {
			newACL.CIDRs = slices.DeleteFunc(newACL.CIDRs, func(s string) bool {
				return s == r
			})
		}
		newACL.CIDRs = append(newACL.CIDRs, viper.GetStringSlice("kube-apiserver-acl-add-to-allowed-cidrs")...)

		if viper.IsSet("kube-apiserver-acl-set-allowed-cidrs") {
			newACL.CIDRs = viper.GetStringSlice("kube-apiserver-acl-set-allowed-cidrs")
		}

		slices.Sort(newACL.CIDRs)
		newACL.CIDRs = slices.Compact(newACL.CIDRs)
		cur.KubeAPIServerACL = newACL
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
	if viper.IsSet("default-pod-security-standard") {
		if !viper.GetBool("yes-i-really-mean-it") {
			return fmt.Errorf("--default-pod-security-standard is set but you forgot to add --yes-i-really-mean-it")
		}
		k8s.DefaultPodSecurityStandard = pointer.Pointer(viper.GetString("default-pod-security-standard"))
	}

	if viper.IsSet("kubelet-pod-pid-limit") {
		if !viper.GetBool("yes-i-really-mean-it") {
			return fmt.Errorf("--kubelet-pod-pid-limit can only be changed in combination with --yes-i-really-mean-it because this change can lead to pods not starting anymore in the cluster")
		}
		k8s.PodPIDsLimit = &podpidLimit
	}

	cur.Kubernetes = k8s
	cur.EgressRules = makeEgressRules(egress)

	if viper.IsSet("enable-node-local-dns") {
		if !viper.GetBool("yes-i-really-mean-it") {
			return fmt.Errorf("setting --enable-node-local-dns will lead to rolling of worker nodes. Please add --yes-i-really-mean-it")
		}

		if cur.SystemComponents == nil {
			cur.SystemComponents = &models.V1SystemComponents{}
		}
		if cur.SystemComponents.NodeLocalDNS == nil {
			cur.SystemComponents.NodeLocalDNS = &models.V1NodeLocalDNS{}
		}
		cur.SystemComponents.NodeLocalDNS.Enabled = &enableNodeLocalDNS

	}
	if viper.IsSet("disable-forwarding-to-upstream-dns") {
		if cur.SystemComponents == nil {
			cur.SystemComponents = &models.V1SystemComponents{}
		}
		if cur.SystemComponents.NodeLocalDNS == nil {
			cur.SystemComponents.NodeLocalDNS = &models.V1NodeLocalDNS{}
		}
		cur.SystemComponents.NodeLocalDNS.DisableForwardToUpstreamDNS = &disableForwardToUpstreamDNS
	}

	if updateCausesDowntime && !viper.GetBool("yes-i-really-mean-it") {
		fmt.Printf("This update of cluster:%q will cause downtime.\n", *current.Name)
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
	return c.describePrinter.Print(shoot.Payload)
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

	genericcli.Must(c.listPrinter.Print(resp.Payload))

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
	return c.describePrinter.Print(cl.Payload)
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
	return c.describePrinter.Print(shoot.Payload)
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
			return c.listPrinter.Print(output.ShootIssuesResponses(response.Payload))
		}

		request := cluster.NewListClustersParams().WithReturnMachines(&boolTrue)
		shoots, err := c.cloud.Cluster.ListClusters(request, nil)
		if err != nil {
			return err
		}
		return c.listPrinter.Print(output.ShootIssuesResponses(shoots.Payload))
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
	return c.listPrinter.Print(output.ShootIssuesResponse(shoot.Payload))
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

	if viper.GetString("output-format") != "table" {
		return c.describePrinter.Print(shoot.Payload)
	}

	fmt.Println("Cluster:")
	genericcli.Must(c.listPrinter.Print(shoot.Payload))

	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)
	fmt.Println("\nMachines:")

	// TODO: when migrating to new table printer from metal-lib, use existing listprinter!
	// return c.listPrinter.Print(ms)
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

	if viper.GetString("output-format") != "table" {
		type s struct {
			Conditions    []*models.V1beta1Condition
			LastOperation *models.V1beta1LastOperation
			LastErrors    []*models.V1beta1LastError
		}
		return c.describePrinter.Print(s{
			Conditions:    conditions,
			LastOperation: lastOperation,
			LastErrors:    lastErrors,
		})
	}

	fmt.Println("Conditions:")
	err = c.listPrinter.Print(conditions)
	if err != nil {
		return err
	}

	fmt.Println("\nLast Errors:")
	err = c.listPrinter.Print(lastErrors)
	if err != nil {
		return err
	}

	fmt.Println("\nLast Operation:")
	return c.listPrinter.Print(lastOperation)
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

	return c.describePrinter.Print(sc.Payload)
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

	return c.listPrinter.Print(ms)
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

	return c.listPrinter.Print(ms)
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

	return c.listPrinter.Print(ms)
}

func (c *config) clusterMachinePackages(args []string) error {
	cid, err := c.clusterID("packages", args)
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

	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)
	for _, m := range ms {
		if *m.ID == mid {
			if m.Allocation == nil || m.Allocation.Image == nil {
				continue
			}
			id := *m.ID
			url := m.Allocation.Image.URL
			packageURL := strings.Replace(url, "img.tar.lz4", "packages.txt", 1)
			//nolint:gosec,noctx
			res, err := http.Head(packageURL)
			if err != nil {
				return fmt.Errorf("image:%s does not have a package list", id)
			}
			defer res.Body.Close()
			if res.StatusCode >= 400 {
				return fmt.Errorf("image:%s does not have a package list", id)
			}
			//nolint:gosec,noctx
			getResp, err := http.Get(packageURL)
			if err != nil {
				return err
			}
			defer getResp.Body.Close()
			content, err := io.ReadAll(getResp.Body)
			if err != nil {
				return err
			}
			fmt.Printf("%s", string(content))
			return nil
		}
	}
	return fmt.Errorf("machine:%s not found in cluster:%s", mid, cid)
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

	return c.describePrinter.Print(secret.Payload)
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

	keypair, err := c.sshKeyPair(cid)
	if err != nil {
		return err
	}
	ms := shoot.Payload.Machines
	ms = append(ms, shoot.Payload.Firewalls...)
	for _, m := range ms {
		if *m.ID != mid {
			continue
		}
		if console {
			fmt.Printf("access console via ssh\n")
			authContext, err := api.GetAuthContext(viper.GetString("kubeconfig"))
			if err != nil {
				return err
			}
			bmcConsolePort := 5222
			err = c.sshClient(mid, c.consoleHost, keypair.privatekey, bmcConsolePort, &authContext.IDToken)
			return err
		}
		networks := m.Allocation.Networks
		switch *m.Allocation.Role {
		case "firewall":
			if keypair.vpn != nil {
				return c.firewallSSHViaVPN(*m.ID, keypair.privatekey, keypair.vpn)
			}

			for _, nw := range networks {
				if *nw.Underlay || *nw.Private {
					continue
				}
				for _, ip := range nw.Ips {
					if portOpen(ip, "22", time.Second) {
						err := c.sshClient("metal", ip, keypair.privatekey, 22, nil)
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

	return fmt.Errorf("machine:%s not found in cluster:%s", mid, cid)
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
