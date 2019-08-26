package cmd

import (
	"fmt"
	"git.f-i-ts.de/cloud-native/metallib/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login user and receive token",
	Long:  "login and receive token that will be used to authenticate commands.",
	RunE: func(cmd *cobra.Command, args []string) error {

		var console io.Writer
		var handler auth.TokenHandlerFunc
		if viper.GetBool("printOnly") {
			// do not print to console
			handler = printTokenHandler
		} else {
			console = os.Stdout
			handler = auth.NewUpdateKubeConfigHandler(viper.GetString("kubeConfig"), console)
		}

		config := auth.Config{
			ClientID:     viper.GetString("clientId"),
			ClientSecret: viper.GetString("clientSecret"),
			IssuerURL:    viper.GetString("issuerUrl"),
			TokenHandler: handler,
			Console:      console,
			Debug:        viper.GetBool("debug"),
		}

		fmt.Println()

		return auth.OIDCFlow(config)
	},
	PreRun: bindPFlags,
}

func printTokenHandler(tokenInfo auth.TokenInfo) error {

	fmt.Println(tokenInfo.IDToken)
	return nil
}

func init() {

	loginCmd.Flags().String("clientId", "auth-go-cli", "The clientId for the registered app at the OIDC-Provider")
	loginCmd.Flags().String("clientSecret", "AuGx99dsxS1hcHAtc9VfcmV1", "The clientSecret for the registered app at the OIDC-Provider")
	loginCmd.Flags().String("issuerUrl", "https://dex.test.fi-ts.io/dex", "URL of the issuer for the token")
	loginCmd.Flags().Bool("printOnly", false, "If true, the token is printed to stdout")
}
