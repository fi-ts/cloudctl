package cmd

import (
	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type xdrCmd struct {
	c *config
}

func newClusterXdrCmd(c *config) *cobra.Command {
	w := xdrCmd{
		c: c,
	}

	clusterXdrCmd := &cobra.Command{
		Use:   "xdr",
		Short: "configure a cluster's cortex xdr configuration",
		Long:  `cortex xdr is a detection and response app to stop attacks`,
	}

	configureCmd := &cobra.Command{
		Use:   "configure --cluster-id=<clusterid>",
		Short: "configure the xdr settings for this cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.configure()
		},
	}

	showCmd := &cobra.Command{
		Use:   "show --cluster-id=<clusterid>",
		Short: "show the current xdr configuration for this cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.show()
		},
	}

	clusterXdrCmd.PersistentFlags().String("cluster-id", "", "the id of the cluster to apply the xdr configuration to")
	genericcli.Must(clusterXdrCmd.MarkPersistentFlagRequired("cluster-id"))
	genericcli.Must(clusterXdrCmd.RegisterFlagCompletionFunc("cluster-id", c.comp.ClusterListCompletion))

	configureCmd.Flags().Bool("disabled", false, "disables the entire xdr functionality")
	configureCmd.Flags().String("distributionid", "", "the distribution id for the xdr configuration")
	configureCmd.Flags().StringSlice("proxies", []string{}, "proxy list for the xdr configuration")
	configureCmd.Flags().String("customtag", "", "custom tag for the xdr configuration")
	configureCmd.Flags().Bool("noproxy", false, "disables proxy usage for the xdr configuration")

	clusterXdrCmd.AddCommand(configureCmd, showCmd)

	return clusterXdrCmd
}

func (c *xdrCmd) configure() error {
	xdrConfiguration := &models.V1XDR{}

	if viper.IsSet("disabled") {
		xdrConfiguration.Disabled = pointer.Pointer(viper.GetBool("disabled"))
	}

	if viper.IsSet("distributionid") {
		xdrConfiguration.DistributionID = pointer.Pointer(viper.GetString("distributionid"))
	}

	if viper.IsSet("proxies") {
		xdrConfiguration.ProxyList = viper.GetStringSlice("proxies")
	}

	if viper.IsSet("customtag") {
		xdrConfiguration.CustomTag = pointer.Pointer(viper.GetString("customtag"))
	}

	if viper.IsSet("noproxy") {
		xdrConfiguration.NoProxy = pointer.Pointer(viper.GetBool("noproxy"))
	}

	_, err := c.c.cloud.Cluster.UpdateCluster(cluster.NewUpdateClusterParams().WithBody(&models.V1ClusterUpdateRequest{
		ID:        pointer.Pointer(viper.GetString("cluster-id")),
		XDRConfig: xdrConfiguration,
	}), nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *xdrCmd) show() error {
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(viper.GetString("cluster-id"))
	shoot, err := c.c.cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return err
	}

	return c.c.describePrinter.Print(shoot.Payload.XDRConfig)
}
