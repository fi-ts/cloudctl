package cmd

import (
	"fmt"

	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/metal-stack/metal-lib/auth"
	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "logout user from OIDC SSO session",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := api.MustDefaultContext()

			err := auth.Logout(auth.Config{
				IssuerURL: ctx.IssuerURL,
			})
			if err != nil {
				return err
			}

			fmt.Println("OIDC SSO session successfully logged out. Token is not revoked and are valid until expiration.")

			return nil
		},
		PreRun: bindPFlags,
	}
	return logoutCmd
}
