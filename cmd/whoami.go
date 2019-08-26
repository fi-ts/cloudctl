package cmd

import (
	"fmt"
	"git.f-i-ts.de/cloud-native/metallib/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "shows current user",
	Long:  "shows the current user, that will be used to authenticate commands.",
	RunE: func(cmd *cobra.Command, args []string) error {

		kubeconfig := viper.GetString("kubeConfig")
		authContext, err := auth.CurrentAuthContext(kubeconfig)
		if err != nil {
			return err
		}

		if !authContext.AuthProviderOidc {
			return fmt.Errorf("active user %s has no oidc authProvider, check config", authContext.User)
		}

		fmt.Println(authContext.User)
		return nil
	},
	PreRun: bindPFlags,
}
