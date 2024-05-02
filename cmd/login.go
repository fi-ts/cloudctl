package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/metal-stack/metal-lib/auth"
	"github.com/metal-stack/v"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newLoginCmd(c *config) *cobra.Command {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "login user and receive token",
		Long:  "login and receive token that will be used to authenticate commands.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				console io.Writer
				handler auth.TokenHandlerFunc
			)

			if viper.GetBool("print-only") {
				// do not store, only print to console
				handler = printTokenHandler
			} else {
				cs, err := api.GetContexts()
				if err != nil {
					return err
				}
				console = os.Stdout
				handler = auth.NewUpdateKubeConfigHandler(viper.GetString("kubeconfig"), console, auth.WithContextName(api.FormatContextName(api.CloudContext, cs.CurrentContext)))
			}

			scopes := auth.DexScopes
			ctx := api.MustDefaultContext()
			if ctx.IssuerType == "generic" {
				scopes = auth.GenericScopes
			} else if ctx.CustomScopes != "" {
				cs := strings.Split(ctx.CustomScopes, ",")
				for i := range cs {
					cs[i] = strings.TrimSpace(cs[i])
				}
				scopes = cs
			}

			config := auth.Config{
				ClientID:     ctx.ClientID,
				ClientSecret: ctx.ClientSecret,
				IssuerURL:    ctx.IssuerURL,
				Scopes:       scopes,
				TokenHandler: handler,
				Console:      console,
				Debug:        viper.GetBool("debug"),
				Log:          c.log,
			}

			if ctx.IssuerType == "generic" {
				config.SuccessMessage = fmt.Sprintf(`Please close this page and return to your terminal. Manage your session on: <a href=%q>%s</a>`, ctx.IssuerURL+"/account", ctx.IssuerURL+"/account")
			}

			err := auth.OIDCFlow(config)
			if err != nil {
				return err
			}

			resp, err := c.cloud.Version.Info(nil, helper.ClientNoAuth())
			if err != nil {
				return err
			}
			if resp.Payload != nil && resp.Payload.MinClientVersion != nil {
				minVersion := *resp.Payload.MinClientVersion
				parsedMinVersion, err := semver.NewVersion(minVersion)
				if err != nil {
					return fmt.Errorf("required cloudctl minimum version:%q is not semver parsable:%w", minVersion, err)
				}

				// This is a developer build
				if !strings.HasPrefix(v.Version, "v") {
					return nil
				}

				thisVersion, err := semver.NewVersion(v.Version)
				if err != nil {
					return fmt.Errorf("cloudctl version:%q is not semver parsable:%w", v.Version, err)
				}

				if thisVersion.LessThan(parsedMinVersion) {
					return fmt.Errorf("your cloudctl version:%s is smaller than the required minimum version:%s, please run `cloudctl update do` to update to the supported version", thisVersion, minVersion)
				}

				if !thisVersion.Equal(parsedMinVersion) {
					fmt.Println()
					fmt.Printf("WARNING: Your cloudctl version %q might not compatible with the cloud-api (supported version is %q). Please run `cloudctl update do` to update to the supported version.", thisVersion, minVersion)
					fmt.Println()
				}
			}

			return nil
		},
	}
	loginCmd.Flags().Bool("print-only", false, "If true, the token is printed to stdout")
	return loginCmd
}

func printTokenHandler(tokenInfo auth.TokenInfo) error {
	fmt.Println(tokenInfo.IDToken)
	return nil
}
