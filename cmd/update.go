package cmd

import (
	"github.com/metal-stack/updater"
	"github.com/spf13/cobra"
)

func newUpdateCmd(c *config, name string) *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "update the program",
	}
	updateCheckCmd := &cobra.Command{
		Use:   "check",
		Short: "check for update of the program",
		RunE: func(cmd *cobra.Command, args []string) error {
			desired, err := getDesiredVersion(c)
			if err != nil {
				return err
			}
			u, err := updater.New("fi-ts", name, name, desired)
			if err != nil {
				return err
			}
			return u.Check()
		},
	}
	updateDoCmd := &cobra.Command{
		Use:   "do",
		Short: "do the update of the program",
		RunE: func(cmd *cobra.Command, args []string) error {
			desired, err := getDesiredVersion(c)
			if err != nil {
				return err
			}
			u, err := updater.New("fi-ts", name, name, desired)
			if err != nil {
				return err
			}
			return u.Do()
		},
	}

	updateCmd.AddCommand(updateCheckCmd)
	updateCmd.AddCommand(updateDoCmd)

	return updateCmd
}

func getDesiredVersion(c *config) (*string, error) {
	resp, err := c.cloud.Version.Info(nil, nil)
	if err != nil {
		return nil, err
	}
	if resp.Payload != nil && resp.Payload.MinClientVersion != nil {
		return resp.Payload.MinClientVersion, nil
	}
	return nil, nil
}
