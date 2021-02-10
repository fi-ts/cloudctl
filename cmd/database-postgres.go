package cmd

import (
	"fmt"
	"log"

	"github.com/fi-ts/cloud-go/api/client/database"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	// Create
	postgresCreateCmd.Flags().StringP("description", "", "", "description of the database")
	postgresCreateCmd.Flags().StringP("tenant", "", "", "tenant of the database, requires on-behalf rights [optional]")
	postgresCreateCmd.Flags().StringP("project", "", "", "project of the database")
	postgresCreateCmd.Flags().StringP("partition", "", "", "partition where the database should be created")
	postgresCreateCmd.Flags().IntP("instances", "", 1, "instances of the database")
	postgresCreateCmd.Flags().StringP("version", "", "12", "version of the database") // FIXME add possible values
	postgresCreateCmd.Flags().StringSliceP("sources", "", []string{"0.0.0.0/0"}, "networks which should be allowed to connect")
	postgresCreateCmd.Flags().StringP("cpu", "", "500m", "cpus for the database")
	postgresCreateCmd.Flags().StringP("buffer", "", "500m", "shared buffer for the database")
	postgresCreateCmd.Flags().StringP("storage", "", "10Gi", "storage for the database")
	// TODO Maintenance
	err := postgresCreateCmd.MarkFlagRequired("project")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresCreateCmd.MarkFlagRequired("partition")
	if err != nil {
		log.Fatal(err.Error())
	}

	err = postgresCreateCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	err = postgresCreateCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresCreateCmd.RegisterFlagCompletionFunc("version", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		versions := []string{
			"11",
			"11.10",
			"12",
			"12.5",
			"13",
			"13.1",
		}
		return versions, cobra.ShellCompDirectiveDefault
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	// List
	postgresListCmd.Flags().StringP("id", "", "", "postgres id to filter [optional]")
	postgresListCmd.Flags().StringP("name", "", "", "name to filter [optional]")
	postgresListCmd.Flags().StringP("tenant", "", "", "tenant to filter [optional]")
	postgresListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	postgresListCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")

	postgresUpdateCmd.Flags().StringP("name", "", "restored-pv", "name of the PersistentPostgres")
	postgresUpdateCmd.Flags().StringP("namespace", "", "default", "namespace for the PersistentPostgres")

	err = postgresListCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
	desc := viper.GetString("description")
	tenant := viper.GetString("tenant")
	project := viper.GetString("project")
	partition := viper.GetString("partition")
	instances := viper.GetInt32("instances")
	version := viper.GetString("version")
	sources := viper.GetStringSlice("sources")
	cpu := viper.GetString("cpu")
	buffer := viper.GetString("buffer")
	storage := viper.GetString("storage")
	pcr := &models.V1PostgresCreateRequest{
		Description:       desc,
		Tenant:            tenant,
		ProjectID:         project,
		PartitionID:       partition,
		NumberOfInstances: instances,
		Version:           version,
		Size: &models.V1Size{
			CPU:          cpu,
			SharedBuffer: buffer,
			StorageSize:  storage,
		},
		AccessList: &models.V1AccessList{
			SourceRanges: sources,
		},
	}
	request := database.NewCreatePostgresParams()
	request.SetBody(pcr)

	response, err := cloud.Database.CreatePostgres(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
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
