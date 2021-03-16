package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fi-ts/cloud-go/api/client/database"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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
			return postgresApply()
		},
		PreRun: bindPFlags,
	}
	postgresEditCmd = &cobra.Command{
		Use:   "edit",
		Short: "edit postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresEdit(args)
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
	postgresConnectionStringtCmd = &cobra.Command{
		Use:   "connectionstring <postgres>",
		Short: "return the connectionstring for a postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresConnectionString(args)
		},
		PreRun: bindPFlags,
	}
	postgresVersionsCmd = &cobra.Command{
		Use:   "version",
		Short: "describe all postgres versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresVersions()
		},
		PreRun: bindPFlags,
	}
	postgresPartitionsCmd = &cobra.Command{
		Use:   "partition",
		Short: "describe all partitions where postgres might be deployed",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresPartitions()
		},
		PreRun: bindPFlags,
	}
	postgresBackupCmd = &cobra.Command{
		Use:   "backup-config",
		Short: "manage postgres backup configuration",
		Long:  "list/find/delete postgres backup configuration",
	}
	postgresBackupCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresBackupCreate(false)
		},
		PreRun: bindPFlags,
	}
	postgresBackupAutoCreateCmd = &cobra.Command{
		Use:   "auto-create",
		Short: "auto create backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresBackupCreate(true)
		},
		PreRun: bindPFlags,
	}
	postgresBackupUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "update backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresBackupUpdate()
		},
		PreRun: bindPFlags,
	}
	postgresBackupListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresBackupGet(args)
		},
		PreRun: bindPFlags,
	}
	postgresBackupDeleteCmd = &cobra.Command{
		Use:     "delete <backup-config>",
		Aliases: []string{"rm", "destroy", "remove", "delete"},
		Short:   "delete a backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return postgresBackupDelete(args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	rootCmd.AddCommand(postgresCmd)
	postgresCmd.AddCommand(postgresBackupCmd)

	postgresCmd.AddCommand(postgresCreateCmd)
	postgresCmd.AddCommand(postgresApplyCmd)
	postgresCmd.AddCommand(postgresEditCmd)
	postgresCmd.AddCommand(postgresListCmd)
	postgresCmd.AddCommand(postgresDeleteCmd)
	postgresCmd.AddCommand(postgresDescribeCmd)
	postgresCmd.AddCommand(postgresVersionsCmd)
	postgresCmd.AddCommand(postgresPartitionsCmd)
	postgresCmd.AddCommand(postgresConnectionStringtCmd)

	postgresBackupCmd.AddCommand(postgresBackupCreateCmd)
	postgresBackupCmd.AddCommand(postgresBackupAutoCreateCmd)
	postgresBackupCmd.AddCommand(postgresBackupUpdateCmd)
	postgresBackupCmd.AddCommand(postgresBackupListCmd)
	postgresBackupCmd.AddCommand(postgresBackupDeleteCmd)

	// Create
	postgresCreateCmd.Flags().StringP("description", "", "", "description of the database")
	postgresCreateCmd.Flags().StringP("tenant", "", "", "tenant of the database, requires on-behalf rights [optional]")
	postgresCreateCmd.Flags().StringP("project", "", "", "project of the database")
	postgresCreateCmd.Flags().StringP("partition", "", "", "partition where the database should be created")
	postgresCreateCmd.Flags().IntP("instances", "", 1, "instances of the database")
	postgresCreateCmd.Flags().StringP("version", "", "12", "version of the database") // FIXME add possible values
	postgresCreateCmd.Flags().StringSliceP("sources", "", []string{"0.0.0.0/0"}, "networks which should be allowed to connect in CIDR notation")
	postgresCreateCmd.Flags().StringSliceP("labels", "", []string{}, "labels to add to that postgres database")
	postgresCreateCmd.Flags().StringP("cpu", "", "500m", "cpus for the database")
	postgresCreateCmd.Flags().StringP("buffer", "", "500m", "shared buffer for the database")
	postgresCreateCmd.Flags().StringP("storage", "", "10Gi", "storage for the database")
	postgresCreateCmd.Flags().StringP("backup", "", "", "backup to use")
	postgresCreateCmd.Flags().StringSliceP("maintenance", "", []string{"Sun:22:00-23:00"}, "time specification of the automatic maintenance in the form Weekday:HH:MM-HH-MM [optional]")
	err := postgresCreateCmd.MarkFlagRequired("description")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresCreateCmd.MarkFlagRequired("project")
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
		return postgresListPartitionsCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresCreateCmd.RegisterFlagCompletionFunc("version", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return postgresListVersionsCompletion()
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

	postgresApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
	Example postgres update:

	# cloudctl postgres describe postgres1 -o yaml > postgres1.yaml
	# vi postgres1.yaml
	## either via stdin
	# cat postgres1.yaml | cloudctl postgres apply -f -
	## or via file
	# cloudctl postgres apply -f postgres1.yaml
	`)

	postgresConnectionStringtCmd.Flags().StringP("type", "", "psql", "the type of the connectionstring to create, can be one of psql|jdbc")
	err = postgresConnectionStringtCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"jdbc", "psql"}, cobra.ShellCompDirectiveDefault
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	postgresBackupCreateCmd.Flags().StringP("name", "", "", "name of the database backup")
	postgresBackupCreateCmd.Flags().StringP("project", "", "", "project of the database backup")
	postgresBackupCreateCmd.Flags().StringP("schedule", "", "30 00 * * *", "backup schedule in cron syntax")
	postgresBackupCreateCmd.Flags().Int32P("retention", "", int32(10), "backup retention days")
	postgresBackupCreateCmd.Flags().BoolP("autocreate", "", false, "automatically create s3 backup bucket")
	postgresBackupCreateCmd.Flags().StringP("partition", "", "", "if autocreate is set to true, use this partition to create the backup bucket")
	postgresBackupCreateCmd.Flags().StringP("s3-endpoint", "", "", "s3 endpooint to backup to")
	postgresBackupCreateCmd.Flags().StringP("s3-region", "", "", "s3 region to backup to [optional]")
	postgresBackupCreateCmd.Flags().StringP("s3-bucketname", "", "", "s3 bucketname to backup to")
	postgresBackupCreateCmd.Flags().StringP("s3-accesskey", "", "", "s3-accesskey")
	postgresBackupCreateCmd.Flags().StringP("s3-secretkey", "", "", "s3-secretkey")
	postgresBackupCreateCmd.Flags().StringP("s3-encryptionkey", "", "", "s3 encryption key, enables sse (server side encryption) if given [optional]")
	err = postgresBackupCreateCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresBackupCreateCmd.MarkFlagRequired("project")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresBackupCreateCmd.MarkFlagRequired("schedule")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresBackupCreateCmd.MarkFlagRequired("s3-endpoint")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresBackupCreateCmd.MarkFlagRequired("s3-accesskey")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresBackupCreateCmd.MarkFlagRequired("s3-secretkey")
	if err != nil {
		log.Fatal(err.Error())
	}

	postgresBackupAutoCreateCmd.Flags().StringP("name", "", "", "name of the database backup")
	postgresBackupAutoCreateCmd.Flags().StringP("project", "", "", "project of the database backup")
	postgresBackupAutoCreateCmd.Flags().StringP("schedule", "", "30 00 * * *", "backup schedule in cron syntax")
	postgresBackupAutoCreateCmd.Flags().Int32P("retention", "", int32(10), "backup retention days")
	postgresBackupAutoCreateCmd.Flags().StringP("partition", "", "", "use this partition to create the backup bucket")
	err = postgresBackupAutoCreateCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresBackupAutoCreateCmd.MarkFlagRequired("project")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresBackupAutoCreateCmd.MarkFlagRequired("schedule")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresBackupAutoCreateCmd.MarkFlagRequired("retention")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = postgresBackupAutoCreateCmd.MarkFlagRequired("partition")
	if err != nil {
		log.Fatal(err.Error())
	}

	postgresBackupUpdateCmd.Flags().StringP("id", "", "", "id of the database backup")
	postgresBackupUpdateCmd.Flags().StringP("schedule", "", "", "backup schedule in cron syntax [optional]")
	postgresBackupUpdateCmd.Flags().Int32P("retention", "", int32(10), "backup retention days [optional]")
	err = postgresBackupUpdateCmd.MarkFlagRequired("id")
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
	labels := viper.GetStringSlice("labels")
	cpu := viper.GetString("cpu")
	buffer := viper.GetString("buffer")
	backup := viper.GetString("backup")
	storage := viper.GetString("storage")
	maintenance := viper.GetStringSlice("maintenance")

	labelMap, err := helper.LabelsToMap(labels)
	if err != nil {
		return err
	}
	pcr := &models.V1PostgresCreateRequest{
		Description:       desc,
		Tenant:            tenant,
		ProjectID:         project,
		PartitionID:       partition,
		NumberOfInstances: instances,
		Version:           version,
		Backup:            backup,
		Size: &models.V1Size{
			CPU:          cpu,
			SharedBuffer: buffer,
			StorageSize:  storage,
		},
		AccessList: &models.V1AccessList{
			SourceRanges: sources,
		},
		Maintenance: maintenance,
		Labels:      labelMap,
	}
	request := database.NewCreatePostgresParams()
	request.SetBody(pcr)

	response, err := cloud.Database.CreatePostgres(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}

func postgresApply() error {
	var pars []models.V1PostgresCreateRequest
	var par models.V1PostgresCreateRequest
	err := helper.ReadFrom(viper.GetString("file"), &par, func(data interface{}) {
		doc := data.(*models.V1PostgresCreateRequest)
		pars = append(pars, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		par = models.V1PostgresCreateRequest{}
	})
	if err != nil {
		return err
	}
	response := []*models.V1PostgresResponse{}
	for i, par := range pars {
		params := database.NewGetPostgresParams().WithID(*par.ID)
		resp, err := cloud.Database.GetPostgres(params, nil)
		if err != nil {
			return err
		}
		if resp.Payload == nil {
			// no postgres found, create it
			request := database.NewCreatePostgresParams()
			request.SetBody(&par)

			createdPG, err := cloud.Database.CreatePostgres(request, nil)
			if err != nil {
				return err
			}
			response = append(response, createdPG.Payload)
			continue
		}
		if resp.Payload.ID != nil {
			// existing postgres, update
			request := database.NewUpdatePostgresParams()
			request.SetBody(&pars[i])

			updatedPG, err := cloud.Database.UpdatePostgres(request, nil)
			if err != nil {
				return err
			}
			response = append(response, updatedPG.Payload)
			continue
		}
	}
	return printer.Print(response)
}

func postgresEdit(args []string) error {
	id, err := postgresID("edit", args)
	if err != nil {
		return err
	}

	getFunc := func(id string) ([]byte, error) {
		params := database.NewGetPostgresParams().WithID(id)
		resp, err := cloud.Database.GetPostgres(params, nil)
		if err != nil {
			return nil, err
		}
		content, err := yaml.Marshal(resp.Payload)
		if err != nil {
			return nil, err
		}
		return content, nil
	}
	updateFunc := func(filename string) error {
		purs, err := readPostgresUpdateRequests(filename)
		if err != nil {
			return err
		}
		if len(purs) != 1 {
			return fmt.Errorf("postgres update error more or less than one postgres given:%d", len(purs))
		}
		pup := database.NewUpdatePostgresParams()
		pup.Body = &purs[0]
		uresp, err := cloud.Database.UpdatePostgres(pup, nil)
		if err != nil {
			return err
		}
		return printer.Print(uresp.Payload)
	}
	return helper.Edit(id, getFunc, updateFunc)
}

func readPostgresUpdateRequests(filename string) ([]models.V1PostgresCreateRequest, error) {
	var pcrs []models.V1PostgresCreateRequest
	var pcr models.V1PostgresCreateRequest
	err := helper.ReadFrom(filename, &pcr, func(data interface{}) {
		doc := data.(*models.V1PostgresCreateRequest)
		pcrs = append(pcrs, *doc)
	})
	if err != nil {
		return pcrs, err
	}
	if len(pcrs) != 1 {
		return pcrs, fmt.Errorf("postgres update error more or less than one postgres given:%d", len(pcrs))
	}
	return pcrs, nil
}

func postgresFind() error {
	if helper.AtLeastOneViperStringFlagGiven("id", "description", "tenant", "project", "partition") {
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

	printer.Print(pg)

	idParts := strings.Split(*pg.ID, "-")
	firstPartOfPostgresID := idParts[0]
	lastPartOfPostgresID := idParts[len(idParts)-1]
	fmt.Println("Please answer some security questions to delete this postgres database")
	err = helper.Prompt("first part of ID:", firstPartOfPostgresID)
	if err != nil {
		return err
	}
	err = helper.Prompt("last part of ID:", lastPartOfPostgresID)
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

	printer.Print(postgres)

	// FIXME this is a ugly hack to reset the printer and have a new header.
	initPrinter()
	fmt.Println("\nBackups:")
	return printer.Print(postgres.Backups)
}
func postgresConnectionString(args []string) error {
	t := viper.GetString("type")

	postgres, err := getPostgresFromArgs(args)
	if err != nil {
		return err
	}

	params := database.NewGetPostgresSecretsParams().WithID(*postgres.ID)
	resp, err := cloud.Database.GetPostgresSecrets(params, nil)
	if err != nil {
		return err
	}
	ip := "localhost"
	port := int32(5432)
	if postgres.Status.Socket != nil {
		ip = postgres.Status.Socket.IP
		port = postgres.Status.Socket.Port
	}

	userpassword := make(map[string]string)
	if resp.Payload.UserSecret != nil && len(resp.Payload.UserSecret) > 0 {
		for _, user := range resp.Payload.UserSecret {
			userpassword[user.Username] = user.Password
		}
	}
	if len(userpassword) == 0 {
		userpassword["unknown"] = "unknown"
	}
	for user, password := range userpassword {
		switch t {
		case "jdbc":
			fmt.Printf("jdbc:postgresql://%s:%d/postgres?user=%s&password=%s&ssl=true\n", ip, port, user, password)
		case "psql":
			fmt.Printf("PGPASSWORD=%s psql --host=%s --port=%d --username=%s\n", password, ip, port, user)
		default:
			return fmt.Errorf("unknown connectionstring type:%s", t)
		}
	}
	return nil
}

func postgresBackupCreate(autocreate bool) error {
	name := viper.GetString("name")
	project := viper.GetString("project")
	schedule := viper.GetString("schedule")
	retention := viper.GetInt32("retention")
	partition := viper.GetString("partition")
	s3Endpoint := viper.GetString("s3-endpoint")
	s3Region := viper.GetString("s3-region")
	s3BucketName := viper.GetString("s3-bucketname")
	s3Accesskey := viper.GetString("s3-accesskey")
	s3Secretkey := viper.GetString("s3-secretkey")
	s3Encryptionkey := viper.GetString("s3-encryptionkey")

	bcr := &models.V1BackupCreateRequest{
		Name:      name,
		ProjectID: project,
		Schedule:  schedule,
		Retention: retention,
	}
	if autocreate {
		bcr.Autocreate = true
		bcr.Partition = partition
	} else {
		bcr.S3Endpoint = s3Endpoint
		bcr.S3BucketName = s3BucketName
		if s3Region != "" {
			bcr.S3Region = s3Region
		}
		bcr.Secret = &models.V1BackupSecret{
			Accesskey: s3Accesskey,
			Secretkey: s3Secretkey,
		}
		if s3Encryptionkey != "" {
			bcr.Secret.S3encryptionkey = s3Encryptionkey
		}
	}
	request := database.NewCreatePostgresBackupParams()
	request.SetBody(bcr)

	response, err := cloud.Database.CreatePostgresBackup(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}
func postgresBackupUpdate() error {
	id := viper.GetString("id")

	request := database.NewGetPostgresBackupsParams().WithID(id)
	resp, err := cloud.Database.GetPostgresBackups(request, nil)
	if err != nil {
		return err
	}
	if resp == nil || resp.Payload == nil {
		return fmt.Errorf("given backup %s does not exist", id)
	}

	schedule := viper.GetString("schedule")
	retention := viper.GetInt32("retention")

	bur := &models.V1BackupUpdateRequest{
		ID: id,
	}
	if schedule != "" {
		bur.Schedule = schedule
	}
	if retention != 0 {
		bur.Retention = retention
	}

	req := database.NewUpdatePostgresBackupParams()
	req.SetBody(bur)

	response, err := cloud.Database.UpdatePostgresBackup(req, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}

func postgresBackupGet(args []string) error {
	if len(args) == 0 {
		request := database.NewListPostgresBackupsParams()
		resp, err := cloud.Database.ListPostgresBackups(request, nil)
		if err != nil {
			return err
		}
		return printer.Print(resp.Payload)
	}

	if len(args) > 0 {
		request := database.NewGetPostgresBackupsParams().WithID(args[0])
		resp, err := cloud.Database.GetPostgresBackups(request, nil)
		if err != nil {
			return err
		}
		return printer.Print(resp.Payload)
	}
	return fmt.Errorf("only backup id allowed")
}
func postgresBackupDelete(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("missing backup id")
	}
	id := args[0]

	idParts := strings.Split(id, "-")
	firstPartOfID := idParts[0]
	lastPartOfID := idParts[len(idParts)-1]
	fmt.Println("Please answer some security questions to delete this postgres database backup")
	err := helper.Prompt("first part of ID:", firstPartOfID)
	if err != nil {
		return err
	}
	err = helper.Prompt("last part of ID:", lastPartOfID)
	if err != nil {
		return err
	}

	request := database.NewDeletePostgresBackupParams().WithID(id)
	resp, err := cloud.Database.DeletePostgresBackup(request, nil)
	if err != nil {
		return err
	}
	return printer.Print(resp.Payload)

}

func postgresVersions() error {
	params := database.NewGetPostgresVersionsParams()
	resp, err := cloud.Database.GetPostgresVersions(params, nil)
	if err != nil {
		return err
	}

	return printer.Print(resp.Payload)
}
func postgresPartitions() error {
	params := database.NewGetPostgresPartitionsParams()
	resp, err := cloud.Database.GetPostgresPartitions(params, nil)
	if err != nil {
		return err
	}

	return printer.Print(resp.Payload)
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

func postgresID(verb string, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("postgres %s requires postgresID as argument", verb)
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", fmt.Errorf("postgres %s requires exactly one postgresID as argument", verb)
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
