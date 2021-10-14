package cmd

import (
	"fmt"

	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/spf13/cobra"
)

func newHealthCmd(c *config) *cobra.Command {
	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "show health information",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := c.cloud.Health.Health(nil, nil)
			if err != nil {
				return err
			}

			must(output.New().Print(resp.Payload))

			fmt.Println()

			return output.New().Print(resp.Payload.Services)
		},
		PreRun: bindPFlags,
	}
	return healthCmd
}
