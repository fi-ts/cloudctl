package cmd

import (
	"errors"
	"fmt"

	"github.com/fi-ts/cloud-go/api/client/health"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/spf13/cobra"
)

func newHealthCmd(c *config) *cobra.Command {
	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "show health information",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := c.cloud.Health.Health(nil, nil)
			if err != nil {
				var r *health.HealthInternalServerError
				if errors.As(err, &r) {
					resp = health.NewHealthOK()
					resp.Payload = r.Payload
				} else {
					return err
				}
			}

			genericcli.Must(c.listPrinter.Print(resp.Payload))

			fmt.Println()

			return c.listPrinter.Print(resp.Payload.Services)
		},
	}
	return healthCmd
}
