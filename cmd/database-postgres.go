package cmd

import (
	"fmt"
	"log"

	"github.com/fi-ts/cloud-go/api/client/database"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/spf13/cobra"
)

var (
	postgresCmd = &cobra.Command{
		Use:   "postgres",
		Short: "manage postgres",
		Long:  "list/find/delete postgress",
	}
	postgresCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresCreate()
		},
		PreRun: bindPFlags,
	}
	postgresApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "apply postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresApply(args)
		},
		PreRun: bindPFlags,
	}

	postgresListCmd = &cobra.Command{
		Use:     "list",
		Short:   "list postgres",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresFind()
		},
		PreRun: bindPFlags,
	}
	postgresDeleteCmd = &cobra.Command{
		Use:     "delete <postgres>",
		Aliases: []string{"rm", "destroy", "remove", "delete"},
		Short:   "delete a postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresDelete(args)
		},
		PreRun: bindPFlags,
	}
	postgresUpdateCmd = &cobra.Command{
		Use:   "update <postgres>",
		Short: "update a postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresUpdate(args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	rootCmd.AddCommand(postgresCmd)

	postgresCmd.AddCommand(postgresCreateCmd)
	postgresCmd.AddCommand(postgresApplyCmd)
	postgresCmd.AddCommand(postgresListCmd)
	postgresCmd.AddCommand(postgresDeleteCmd)
	postgresCmd.AddCommand(postgresUpdateCmd)

	postgresListCmd.Flags().StringP("id", "", "", "postgres id to filter [optional]")
	postgresListCmd.Flags().StringP("name", "", "", "name to filter [optional]")
	postgresListCmd.Flags().StringP("tenant", "", "", "tenant to filter [optional]")
	postgresListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	postgresListCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")

	postgresUpdateCmd.Flags().StringP("name", "", "restored-pv", "name of the PersistentPostgres")
	postgresUpdateCmd.Flags().StringP("namespace", "", "default", "namespace for the PersistentPostgres")

	err := postgresListCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	err = postgresListCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
func postgresCreate() error {
	return nil
}
func postgresApply(args []string) error {
	return nil
}
func postgresFind() error {
	if helper.AtLeastOneViperStringFlagGiven("postgresid", "project", "partition") {
		params := database.NewFindPostgresParams()
		ifr := &models.V1PostgresFindRequest{
			ID:          helper.ViperString("id"),
			Name:        helper.ViperString("name"),
			Tenant:      helper.ViperString("tenant"),
			ProjectID:   helper.ViperString("project"),
			PartitionID: helper.ViperString("partition"),
		}
		params.SetBody(ifr)
		resp, err := cloud.Database.FindPostgres(params, nil)
		if err != nil {
			return err
		}
		return printer.Print(resp.Payload)
	}
	resp, err := cloud.Database.ListPostgres(nil, nil)
	if err != nil {
		return err
	}
	return printer.Print(resp.Payload)
}

func postgresDelete(args []string) error {
	pg, err := getPostgresFromArgs(args)
	if err != nil {
		return err
	}
	params := &database.DeletePostgresParams{}
	params.SetID(*pg.ID)
	resp, err := cloud.Database.DeletePostgres(params, nil)
	if err != nil {
		return err
	}

	return printer.Print(resp.Payload)
}

func postgresUpdate(args []string) error {
	// postgres, err := getPostgresFromArgs(args)
	// if err != nil {
	// 	return err
	// }
	// name := viper.GetString("name")
	// namespace := viper.GetString("namespace")

	return nil
}
func getPostgresFromArgs(args []string) (*models.V1PostgresResponse, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no postgres id given")
	}

	id := args[0]
	params := database.NewFindPostgresParams()
	ifr := &models.V1PostgresFindRequest{
		ID: &id,
	}
	params.SetBody(ifr)
	resp, err := cloud.Database.FindPostgres(params, nil)
	if err != nil {
		return nil, err
	}
	if len(resp.Payload) < 1 {
		return nil, fmt.Errorf("no postgres for id:%s found", id)
	}
	if len(resp.Payload) > 1 {
		return nil, fmt.Errorf("more than one postgres for id:%s found", id)
	}
	return resp.Payload[0], nil
}
