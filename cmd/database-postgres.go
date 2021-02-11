package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/fi-ts/cloud-go/api/client/database"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/go-openapi/strfmt"
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
	postgresDescribeCmd = &cobra.Command{
		Use:   "describe <postgres>",
		Short: "describe a postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresDescribe(args)
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
	postgresCmd.AddCommand(postgresDescribeCmd)

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
	postgresCreateCmd.Flags().StringP("maintenance-weekday", "", "Sun", "weekday of the automatic maintenance [optional]")
	postgresCreateCmd.Flags().StringP("maintenance-start", "", "22:30:00 +0000", "start time of the automatic maintenance [optional]")
	postgresCreateCmd.Flags().StringP("maintenance-end", "", "23:30:00 +0000", "end time of the automatic maintenance [optional]")

	postgresCreateCmd.Flags().StringP("s3-url", "", "", "s3-url to backup to [optional]")
	postgresCreateCmd.Flags().StringP("s3-accesskey", "", "", "s3-accesskey to backup to [optional]")
	postgresCreateCmd.Flags().StringP("s3-secretkey", "", "", "s3-secretkey to backup to [optional]")
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
	postgresListCmd.Flags().StringP("description", "", "", "description to filter [optional]")
	postgresListCmd.Flags().StringP("tenant", "", "", "tenant to filter [optional]")
	postgresListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	postgresListCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")

	// postgresUpdateCmd.Flags().StringP("name", "", "restored-pv", "name of the PersistentPostgres")
	// postgresUpdateCmd.Flags().StringP("namespace", "", "default", "namespace for the PersistentPostgres")

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
	mweekday := viper.GetString("maintenance-weekday")
	ms := viper.GetString("maintenance-start")
	me := viper.GetString("maintenance-end")
	s3URL := viper.GetString("s3-url")
	s3Accesskey := viper.GetString("s3-accesskey")
	s3Secretkey := viper.GetString("s3-secretkey")
	var backup models.V1Backup
	if s3URL != "" && s3Accesskey != "" && s3Secretkey != "" {
		backup = models.V1Backup{
			S3BucketURL: s3URL,
			Accesskey:   s3Accesskey,
			Secretkey:   s3Secretkey,
		}
	}

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
		Maintenance: &models.V1MaintenanceWindow{
			Weekday: parseWeekday(mweekday),
			TimeWindow: &models.V1TimeWindow{
				Start: strfmt.DateTime(parseTime(ms)),
				End:   strfmt.DateTime(parseTime(me)),
			},
		},
		Backup: &backup,
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
			Description: helper.ViperString("description"),
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
	params := database.NewDeletePostgresParams().WithID(*pg.ID)
	resp, err := cloud.Database.DeletePostgres(params, nil)
	if err != nil {
		return err
	}

	return printer.Print(resp.Payload)
}

func postgresDescribe(args []string) error {
	postgres, err := getPostgresFromArgs(args)
	if err != nil {
		return err
	}

	return printer.Print(postgres)
}

func getPostgresFromArgs(args []string) (*models.V1PostgresResponse, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no postgres id given")
	}

	id := args[0]
	params := database.NewGetPostgresParams().WithID(id)
	resp, err := cloud.Database.GetPostgres(params, nil)
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func parseWeekday(weekday string) int32 {
	switch weekday {
	case "Sun", "SUN", "sun":
		return 0
	case "Mon", "MON", "mon":
		return 1
	case "Tue", "TUE", "tue":
		return 2
	case "Wed", "WED", "wed":
		return 3
	case "Thu", "THU", "thu":
		return 4
	case "Fri", "FRI", "fri":
		return 5
	case "Sat", "SAT", "sat":
		return 6
	case "All", "ALL", "all":
		return 7
	default:
		fmt.Printf("error parsing weekday:%s", weekday)
		return 0
	}
}

func parseTime(t string) time.Time {
	result, err := time.Parse("15:04:05 -0700", t)
	if err != nil {
		fmt.Printf("error parsing time:%v", err)
		return time.Date(0, 0, 0, 23, 0, 0, 0, time.UTC)
	}
	return result
}
