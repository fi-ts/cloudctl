package cmd

import (
	"fmt"
	"os"

	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type auditCmd struct {
	c *config
}

func newClusterAuditCmd(c *config) *cobra.Command {
	w := auditCmd{
		c: c,
	}

	clusterAuditCmd := &cobra.Command{
		Use:   "audit --cluster-id=<clusterid>",
		Short: "configure a cluster's kube-apiserver audit configuration",
		Long: `audit logs are captured through a webhook and buffered for up to 1GB next to the cluster's kube-apiserver. multiple backends are supported and can run simultaneously:
- Splunk
- Cluster Forwarding (not recommended for production use)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if viper.IsSet("disabled") {
				return w.disable()
			}
			return fmt.Errorf("no command specified")
		},
		PreRun: bindPFlags,
	}
	modeCmd := &cobra.Command{
		Use:   "mode --cluster-id=<clusterid>",
		Short: "set the audit webhook mode for this cluster",
		Long: `the webhook mode of the cluster, one of:
- batch: Buffer events and asynchronously process them in batches.
- blocking: Block API server responses on processing each individual event.
- blocking-strict: Same as blocking, but when there is a failure during audit logging at the RequestReceived stage, the whole request to the kube-apiserver fails. This is the default.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.mode(args)
		},
		ValidArgs: []string{
			"batch\tBuffer events and asynchronously process them in batches.",
			"blocking\tBlock API server responses on processing each individual event.",
			"blocking-strict\tSame as blocking, but when there is a failure during audit logging at the RequestReceived stage, the whole request to the kube-apiserver fails. This is the default.",
		},
		PreRun: bindPFlags,
	}
	policyCmd := &cobra.Command{
		Use:     "policy --cluster-id=<clusterid>",
		Aliases: []string{"pol"},
		Short:   "manage the audit policy for this cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.auditPolicy()
		},
		PreRun: bindPFlags,
	}
	splunkCmd := &cobra.Command{
		Use:   "splunk --cluster-id=<clusterid>",
		Short: "configure splunk as an audit backend, if enabled without any specific configuration, the provider's default configuration will be used",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.splunk()
		},
		PreRun: bindPFlags,
	}
	clusterForwardingCmd := &cobra.Command{
		Use:   "cluster-forwarding --cluster-id=<clusterid>",
		Short: "configure forwarding the audit logs to an audittailer pod in the cluster (not recommended for production, see long help text)",
		Long:  "the approach has several downsides such as dependency on the stability of the VPN, possible corruption of the audit logs through malicious users in the cluster, etc. therefore this backend is not a recommended for production use-cases.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.clusterForwarding()
		},
		PreRun: bindPFlags,
	}

	clusterAuditCmd.Flags().Bool("disabled", false, "disables the entire audit functionality, enable again with --disabled=false, requires --yes-i-really-mean-it flag")

	clusterAuditCmd.PersistentFlags().String("cluster-id", "", "the id of the cluster to apply the audit configuration to")
	genericcli.Must(clusterAuditCmd.MarkPersistentFlagRequired("cluster-id"))
	genericcli.Must(clusterAuditCmd.RegisterFlagCompletionFunc("cluster-id", c.comp.ClusterListCompletion))

	policyCmd.Flags().String("from-file", "", "reads and applies the audit policy from the given file path")
	policyCmd.Flags().Bool("remove", false, "removes the custom audit policy")
	policyCmd.Flags().Bool("show", false, "shows the current audit policy")
	policyCmd.MarkFlagsMutuallyExclusive("from-file", "remove", "show")
	policyCmd.MarkFlagsOneRequired("from-file", "remove", "show")

	clusterForwardingCmd.Flags().Bool("enabled", false, "enables cluster-forwarding audit backend for this cluster.")

	splunkCmd.Flags().Bool("enabled", false, "enables splunk audit backend for this cluster, if enabled without any specific settings, the provider-default splunk backend will be used.")
	splunkCmd.Flags().String("host", "", "the splunk host to configure.")
	splunkCmd.Flags().String("index", "", "the splunk index to configure.")
	splunkCmd.Flags().String("port", "", "the splunk port to configure.")
	splunkCmd.Flags().String("token", "", "the splunk token used to authenticate against the splunk endpoint.")
	splunkCmd.Flags().String("ca", "", "the path to a ca used for tls connection to splunk endpoint.")

	clusterAuditCmd.AddCommand(modeCmd, policyCmd, splunkCmd, clusterForwardingCmd)

	return clusterAuditCmd
}

func (c *auditCmd) mode(args []string) error {
	mode, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	_, err = c.c.cloud.Cluster.UpdateCluster(cluster.NewUpdateClusterParams().WithBody(&models.V1ClusterUpdateRequest{
		ID: pointer.Pointer(viper.GetString("cluster-id")),
		Audit: &models.V1Audit{
			WebhookMode: pointer.Pointer(mode),
		},
	}), nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *auditCmd) auditPolicy() error {
	if viper.GetBool("show") {
		resp, err := c.c.cloud.Cluster.GetAuditPolicy(cluster.NewGetAuditPolicyParams().WithID(viper.GetString("cluster-id")), nil)
		if err != nil {
			return err
		}

		fmt.Println(pointer.SafeDeref(resp.Payload.Raw))

		return nil
	}

	if viper.GetBool("remove") {
		_, err := c.c.cloud.Cluster.UpdateCluster(cluster.NewUpdateClusterParams().WithBody(&models.V1ClusterUpdateRequest{
			ID: pointer.Pointer(viper.GetString("cluster-id")),
			Audit: &models.V1Audit{
				AuditPolicy: pointer.Pointer(""),
			},
		}), nil)
		if err != nil {
			return err
		}

		return nil
	}

	if viper.IsSet("from-file") {
		policy, err := os.ReadFile(viper.GetString("from-file"))
		if err != nil {
			return err
		}

		_, err = c.c.cloud.Cluster.UpdateCluster(cluster.NewUpdateClusterParams().WithBody(&models.V1ClusterUpdateRequest{
			ID: pointer.Pointer(viper.GetString("cluster-id")),
			Audit: &models.V1Audit{
				AuditPolicy: pointer.Pointer(string(policy)),
			},
		}), nil)
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("either --show, --remove or --from-file needs to be used")
}

func (c *auditCmd) disable() error {
	disabled := viper.GetBool("disabled")
	if disabled && !viper.GetBool("yes-i-really-mean-it") {
		return fmt.Errorf("disabling cluster auditing requires --yes-i-really-mean-it")
	}

	_, err := c.c.cloud.Cluster.UpdateCluster(cluster.NewUpdateClusterParams().WithBody(&models.V1ClusterUpdateRequest{
		ID: pointer.Pointer(viper.GetString("cluster-id")),
		Audit: &models.V1Audit{
			Disabled: pointer.Pointer(disabled),
		},
	}), nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *auditCmd) splunk() error {
	auditConfigration := &models.V1Audit{}

	if auditConfigration.Backends == nil {
		auditConfigration.Backends = &models.V1AuditBackends{}
	}
	if auditConfigration.Backends.Splunk == nil {
		auditConfigration.Backends.Splunk = &models.V1AuditBackendSplunk{}
	}

	if viper.IsSet("enabled") {
		auditConfigration.Backends.Splunk.Enabled = pointer.Pointer(viper.GetBool("enabled"))
	}
	if viper.IsSet("host") {
		auditConfigration.Backends.Splunk.Host = pointer.Pointer(viper.GetString("host"))
	}
	if viper.IsSet("index") {
		auditConfigration.Backends.Splunk.Index = pointer.Pointer(viper.GetString("index"))
	}
	if viper.IsSet("port") {
		auditConfigration.Backends.Splunk.Port = pointer.Pointer(viper.GetString("port"))
	}
	if viper.IsSet("token") {
		auditConfigration.Backends.Splunk.Token = pointer.Pointer(viper.GetString("token"))
	}
	if viper.IsSet("ca") {
		ca, err := os.ReadFile(viper.GetString("ca"))
		if err != nil {
			return err
		}

		auditConfigration.Backends.Splunk.TLS = pointer.Pointer(true)
		auditConfigration.Backends.Splunk.Ca = pointer.Pointer(string(ca))
	}

	_, err := c.c.cloud.Cluster.UpdateCluster(cluster.NewUpdateClusterParams().WithBody(&models.V1ClusterUpdateRequest{
		ID:    pointer.Pointer(viper.GetString("cluster-id")),
		Audit: auditConfigration,
	}), nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *auditCmd) clusterForwarding() error {
	auditConfigration := &models.V1Audit{}

	if auditConfigration.Backends == nil {
		auditConfigration.Backends = &models.V1AuditBackends{}
	}
	if auditConfigration.Backends.ClusterForwarding == nil {
		auditConfigration.Backends.ClusterForwarding = &models.V1AuditBackendClusterForwarding{}
	}

	if viper.IsSet("enabled") {
		auditConfigration.Backends.ClusterForwarding.Enabled = pointer.Pointer(viper.GetBool("enabled"))
	}

	_, err := c.c.cloud.Cluster.UpdateCluster(cluster.NewUpdateClusterParams().WithBody(&models.V1ClusterUpdateRequest{
		ID:    pointer.Pointer(viper.GetString("cluster-id")),
		Audit: auditConfigration,
	}), nil)
	if err != nil {
		return err
	}

	return nil
}
