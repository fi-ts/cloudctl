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
	}
	policyCmd := &cobra.Command{
		Use:     "policy --cluster-id=<clusterid>",
		Aliases: []string{"pol"},
		Short:   "manage the audit policy for this cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.auditPolicy()
		},
	}
	splunkCmd := &cobra.Command{
		Use:   "splunk --cluster-id=<clusterid>",
		Short: "configure splunk as an audit backend, if enabled without any specific configuration, the provider's default configuration will be used",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.splunk()
		},
	}
	s3Cmd := &cobra.Command{
		Use:   "s3 --cluster-id=<clusterid>",
		Short: "configure s3 as an audit backend",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.s3()
		},
	}
	clusterForwardingCmd := &cobra.Command{
		Use:   "cluster-forwarding --cluster-id=<clusterid>",
		Short: "configure forwarding the audit logs to an audittailer pod in the cluster (not recommended for production, see long help text)",
		Long:  "the approach has several downsides such as dependency on the stability of the VPN, possible corruption of the audit logs through malicious users in the cluster, etc. therefore this backend is not a recommended for production use-cases.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.clusterForwarding()
		},
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

	s3Cmd.Flags().Bool("enabled", false, "enables s3 audit backend for this cluster.")
	s3Cmd.Flags().String("access-key", "", "the s3 access key to configure.")
	s3Cmd.Flags().String("secret-key", "", "the s3 secret key to configure.")
	s3Cmd.Flags().String("bucket", "", "the s3 bucket to configure.")
	s3Cmd.Flags().String("endpoint", "", "the s3 endpoint to configure.")
	s3Cmd.Flags().String("prefix", "", "the s3 prefix to configure.")
	s3Cmd.Flags().String("region", "", "the s3 region to configure.")
	s3Cmd.Flags().String("key-format", "", "the s3 key format to configure.")
	s3Cmd.Flags().Bool("tls", true, "enables tls.")
	s3Cmd.Flags().String("total-file-size", "", "the s3 total file size to configure.")
	s3Cmd.Flags().String("upload-timeout", "", "the s3 upload timeout to configure.")
	s3Cmd.Flags().Bool("use-compression", false, "enables compression.")

	clusterAuditCmd.AddCommand(modeCmd, policyCmd, splunkCmd, s3Cmd, clusterForwardingCmd)

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
	auditConfiguration := &models.V1Audit{}

	if auditConfiguration.Backends == nil {
		auditConfiguration.Backends = &models.V1AuditBackends{}
	}
	if auditConfiguration.Backends.Splunk == nil {
		auditConfiguration.Backends.Splunk = &models.V1AuditBackendSplunk{}
	}

	if viper.IsSet("enabled") {
		auditConfiguration.Backends.Splunk.Enabled = pointer.Pointer(viper.GetBool("enabled"))
	}
	if viper.IsSet("host") {
		auditConfiguration.Backends.Splunk.Host = pointer.Pointer(viper.GetString("host"))
	}
	if viper.IsSet("index") {
		auditConfiguration.Backends.Splunk.Index = pointer.Pointer(viper.GetString("index"))
	}
	if viper.IsSet("port") {
		auditConfiguration.Backends.Splunk.Port = pointer.Pointer(viper.GetString("port"))
	}
	if viper.IsSet("token") {
		auditConfiguration.Backends.Splunk.Token = pointer.Pointer(viper.GetString("token"))
	}
	if viper.IsSet("ca") {
		ca, err := os.ReadFile(viper.GetString("ca"))
		if err != nil {
			return err
		}

		auditConfiguration.Backends.Splunk.TLS = pointer.Pointer(true)
		auditConfiguration.Backends.Splunk.Ca = pointer.Pointer(string(ca))
	}

	_, err := c.c.cloud.Cluster.UpdateCluster(cluster.NewUpdateClusterParams().WithBody(&models.V1ClusterUpdateRequest{
		ID:    pointer.Pointer(viper.GetString("cluster-id")),
		Audit: auditConfiguration,
	}), nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *auditCmd) s3() error {
	auditConfiguration := &models.V1Audit{}

	if auditConfiguration.Backends == nil {
		auditConfiguration.Backends = &models.V1AuditBackends{}
	}
	if auditConfiguration.Backends.S3 == nil {
		auditConfiguration.Backends.S3 = &models.V1AuditBackendS3{}
	}

	if viper.IsSet("enabled") {
		auditConfiguration.Backends.S3.Enabled = pointer.Pointer(viper.GetBool("enabled"))
	}
	if viper.IsSet("access-key") {
		auditConfiguration.Backends.S3.AccessKey = pointer.Pointer(viper.GetString("access-key"))
	}
	if viper.IsSet("secret-key") {
		auditConfiguration.Backends.S3.SecretKey = pointer.Pointer(viper.GetString("secret-key"))
	}
	if viper.IsSet("bucket") {
		auditConfiguration.Backends.S3.Bucket = pointer.Pointer(viper.GetString("bucket"))
	}
	if viper.IsSet("endpoint") {
		auditConfiguration.Backends.S3.Endpoint = pointer.Pointer(viper.GetString("endpoint"))
	}
	if viper.IsSet("prefix") {
		auditConfiguration.Backends.S3.Prefix = pointer.Pointer(viper.GetString("prefix"))
	}
	if viper.IsSet("region") {
		auditConfiguration.Backends.S3.Region = pointer.Pointer(viper.GetString("region"))
	}
	if viper.IsSet("key-format") {
		auditConfiguration.Backends.S3.S3KeyFormat = pointer.Pointer(viper.GetString("key-format"))
	}
	if viper.IsSet("tls") {
		auditConfiguration.Backends.S3.TLSEnabled = pointer.Pointer(viper.GetBool("tls"))
	}
	if viper.IsSet("total-file-size") {
		auditConfiguration.Backends.S3.TotalFileSize = pointer.Pointer(viper.GetString("total-file-size"))
	}
	if viper.IsSet("upload-timeout") {
		auditConfiguration.Backends.S3.UploadTimeout = pointer.Pointer(viper.GetString("upload-timeout"))
	}
	if viper.IsSet("use-compression") {
		auditConfiguration.Backends.S3.UseCompression = pointer.Pointer(viper.GetBool("use-compression"))
	}

	_, err := c.c.cloud.Cluster.UpdateCluster(cluster.NewUpdateClusterParams().WithBody(&models.V1ClusterUpdateRequest{
		ID:    pointer.Pointer(viper.GetString("cluster-id")),
		Audit: auditConfiguration,
	}), nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *auditCmd) clusterForwarding() error {
	auditConfiguration := &models.V1Audit{}

	if auditConfiguration.Backends == nil {
		auditConfiguration.Backends = &models.V1AuditBackends{}
	}
	if auditConfiguration.Backends.ClusterForwarding == nil {
		auditConfiguration.Backends.ClusterForwarding = &models.V1AuditBackendClusterForwarding{}
	}

	if viper.IsSet("enabled") {
		auditConfiguration.Backends.ClusterForwarding.Enabled = pointer.Pointer(viper.GetBool("enabled"))
	}

	_, err := c.c.cloud.Cluster.UpdateCluster(cluster.NewUpdateClusterParams().WithBody(&models.V1ClusterUpdateRequest{
		ID:    pointer.Pointer(viper.GetString("cluster-id")),
		Audit: auditConfiguration,
	}), nil)
	if err != nil {
		return err
	}

	return nil
}
