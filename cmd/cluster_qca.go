package cmd

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type qcaCmd struct {
	c *config
}

func newClusterQCACmd(c *config) *cobra.Command {
	w := qcaCmd{
		c: c,
	}

	clusterQCACmd := &cobra.Command{
		Use:   "qca",
		Short: "configure a cluster's qualys cloud agent configuration",
		Long:  `qualys cloud agent scans and monitors the nodes of a cluster`,
	}

	configureCmd := &cobra.Command{
		Use:   "configure --cluster-id=<clusterid>",
		Short: "configure the qca settings for this cluster (only allowed for provider tenant)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.configure()
		},
	}

	showCmd := &cobra.Command{
		Use:   "show --cluster-id=<clusterid>",
		Short: "show the current qca configuration for this cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.show()
		},
	}

	clusterQCACmd.PersistentFlags().String("cluster-id", "", "the id of the cluster to apply the xdr configuration to")
	genericcli.Must(clusterQCACmd.MarkPersistentFlagRequired("cluster-id"))
	genericcli.Must(clusterQCACmd.RegisterFlagCompletionFunc("cluster-id", c.comp.ClusterListCompletion))

	configureCmd.Flags().Bool("disabled", false, "disables the entire xdr functionality")
	configureCmd.Flags().String("proxy", "", "the proxy for the qca configuration, either in the form ip:port or as a parseable url")

	clusterQCACmd.AddCommand(configureCmd, showCmd)

	return clusterQCACmd
}

func (c *qcaCmd) configure() error {

	qcaConfiguration := &models.V1QualysCloudAgent{}

	if viper.IsSet("disabled") {
		qcaConfiguration.Disabled = new(viper.GetBool("disabled"))
	}

	if viper.IsSet("proxy") {
		proxy, err := parseQCAProxy(viper.GetString("proxy"))
		if err != nil {
			return err
		}
		qcaConfiguration.Proxy = proxy
	}

	_, err := c.c.cloud.Cluster.UpdateCluster(cluster.NewUpdateClusterParams().WithBody(&models.V1ClusterUpdateRequest{
		ID:        new(viper.GetString("cluster-id")),
		QCAConfig: qcaConfiguration,
	}), nil)
	if err != nil {
		return err
	}

	return nil
}

func parseQCAProxy(proxy string) (string, error) {
	if proxy == "" {
		return "", fmt.Errorf("proxy must not be empty")
	}

	if strings.Contains(proxy, "://") {
		u, err := url.Parse(proxy)
		if err != nil {
			return "", fmt.Errorf("unable to parse proxy url %q: %w", proxy, err)
		}
		if u.Host == "" {
			return "", fmt.Errorf("invalid proxy url %q: missing host", proxy)
		}
		return proxy, nil
	}

	if _, _, err := net.SplitHostPort(proxy); err != nil {
		return "", fmt.Errorf("invalid proxy %q, must be in the form ip:port or a parseable url", proxy)
	}

	return proxy, nil
}

func (c *qcaCmd) show() error {
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(viper.GetString("cluster-id"))
	shoot, err := c.c.cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}

	return c.c.describePrinter.Print(shoot.Payload.QCAConfig)
}
