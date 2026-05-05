package cmd

import (
	"fmt"

	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const statusURL = "https://status.fits.cloud"

func newStatusCmd(c *config) *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "open the status page in the browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			authContext, err := api.GetAuthContext(viper.GetString("kubeconfig"))
			if err != nil {
				return fmt.Errorf("no valid session found, please run `cloudctl login` first: %w", err)
			}

			url := fmt.Sprintf("%s/auth/login?token=%s", statusURL, authContext.IDToken)

			if err := helper.OpenBrowser(url); err != nil {
				fmt.Fprintln(c.out, "Could not open browser. Please open this URL manually:")
				fmt.Fprintln(c.out, url)
				return nil
			}

			fmt.Fprintln(c.out, "Opening status page in browser...")

			return nil
		},
	}
	return statusCmd
}
