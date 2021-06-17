package cmd

import (
	"fmt"

	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/metal-stack/v"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the client and server version information",
	Long:  "print the client and server version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		v := api.Version{
			Client: v.V.String(),
		}

		resp, err := cloud.Version.Info(nil, nil)
		if err == nil {
			v.Server = resp.Payload
		}

		if err2 := printer.Print(v); err2 != nil {
			return err2
		}
		if err != nil {
			return fmt.Errorf("failed to get server info: %w", err)
		}
		return nil
	},
	PreRun: bindPFlags,
}
