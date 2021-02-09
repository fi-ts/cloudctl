package cmd

import (
	"github.com/spf13/cobra"
)

var (
	databaseCmd = &cobra.Command{
		Use:     "database",
		Aliases: []string{"db"},
		Short:   "manage databases",
		Long:    "TODO",
	}
)

func init() {
	rootCmd.AddCommand(databaseCmd)

	databaseCmd.AddCommand(postgresCmd)
}
