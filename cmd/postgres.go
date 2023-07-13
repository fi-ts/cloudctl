package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fi-ts/cloud-go/api/client/database"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func newPostgresCmd(c *config) *cobra.Command {
	postgresCmd := &cobra.Command{
		Use:   "postgres",
		Short: "manage postgres",
		Long: `
Create an manage postgres databases.

To create a postgres database you first need to specify/create a backup-config where the database should store the backups.
This can be done either by auto-create a S3 endpoint or provide all S3 details.

Example Workflow:

1. Display all possible locations where postgres databases can be created:

# cloudctl postgres partition
NAME            ALLOWEDTENANTS
dc1             any

2. Create a backup-config with retention count and schedule

# cloudctl postgres backup-config auto-create --name daily-for-one-week --project <your-project-id> --partition dc1 --retention 7 --schedule "45 3 * * 0"
ID                                      NAME                    PROJECT                                 SCHEDULE        RETENTION       S3                                                              CREATEDBY
3094421c-ee11-4155-b4d9-7fdac116c0ff    daily-for-one-week      b621eb99-4888-4911-93fc-95854fc030e8    45 3 * * *       7               https://s3.dev.example/backup-3094421c      <Achim Muster>[achim.muster@example.com]

3. Create a postgres database

# cloudctl postgres create --description accounting-db-test --project <your-project-id> --partition dc1 --backup-config <backup-config-id-from-above>
ID                                      DESCRIPTION             PARTITION       TENANT  PROJECT                                 CPU     BUFFER  STORAGE BACKUP-CONFIG                           REPLICAS VERSION AGE     STATUS
890b1601-6cc3-46cd-86a6-d4479bc1528d    accounting-db-test      dc1             fits    b621eb99-4888-4911-93fc-95854fc030e8    500m    500m    10Gi    3094421c-ee11-4155-b4d9-7fdac116c0ff    1        12      0s

4. Check if it is running with

# cloudctl postgres ls --description accounting-db-test
ID                                      DESCRIPTION             PARTITION       TENANT  PROJECT                                 CPU     BUFFER  STORAGE BACKUP-CONFIG                           REPLICAS VERSION AGE     STATUS
890b1601-6cc3-46cd-86a6-d4479bc1528d    accounting-db-test      dc1             fits    b621eb99-4888-4911-93fc-95854fc030e8    500m    500m    10Gi    3094421c-ee11-4155-b4d9-7fdac116c0ff    1        12      1m 21s  Running

5. Connect to the database

# cloudctl postgres connectionstring 890b1601-6cc3-46cd-86a6-d4479bc1528d
PGPASSWORD=J34JnhbtPQ2s1znPmp1pWNRv9EPvbsUvQ3OLh2ycyVAcMmK6upazlzM4JAELpaC0 psql --host=1.2.3.4 --port=32004 --username=postgres
PGPASSWORD=imQw7KYOdha2wLODlQFFEodN8eyfoFYKOUmXnxFwQJHIeOQtjZJmMGDskpa3SARr psql --host=1.2.3.4 --port=32004 --username=standby

# PGPASSWORD=J34JnhbtPQ2s1znPmp1pWNRv9EPvbsUvQ3OLh2ycyVAcMmK6upazlzM4JAELpaC0 psql --host=1.2.3.4 --port=32004 --username=postgres
psql (12.6 (Ubuntu 12.6-0ubuntu0.20.10.1))
SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, bits: 256, compression: off)
Type "help" for help.

postgres=#

6. You can create more databases, all using the same backup-config
`,
	}
	postgresCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresCreate()
		},
		PreRun: bindPFlags,
	}
	postgresCreateStandbyCmd := &cobra.Command{
		Use:   "create-standby",
		Short: "create postgres standby",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresCreateStandby()
		},
		PreRun: bindPFlags,
	}
	postgresPromoteToPrimaryCmd := &cobra.Command{
		Use:   "promote-to-primary",
		Short: "promote a standby instance to become a replication primary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresPromoteToPrimary(args)
		},
		PreRun: bindPFlags,
	}
	postgresDemoteToStandbyCmd := &cobra.Command{
		Use:   "demote-to-standby",
		Short: "demote a the replication primary to become a standby instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresDemoteToStandby(args)
		},
		PreRun: bindPFlags,
	}
	postgresRestoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "restore postgres from existing one",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresRestore()
		},
		PreRun: bindPFlags,
	}
	postgresApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "apply postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresApply()
		},
		PreRun: bindPFlags,
	}
	postgresEditCmd := &cobra.Command{
		Use:   "edit",
		Short: "edit postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresEdit(args)
		},
		PreRun: bindPFlags,
	}
	postgresAcceptRestoreCmd := &cobra.Command{
		Use:   "restore-accepted",
		Short: "confirm the restore of a database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresAcceptRestore(args)
		},
		PreRun: bindPFlags,
	}
	postgresListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list postgres",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresFind()
		},
		PreRun: bindPFlags,
	}
	postgresListBackupsCmd := &cobra.Command{
		Use:   "list-backups",
		Short: "list postgres backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresListBackups(args)
		},
		PreRun: bindPFlags,
	}
	postgresDeleteCmd := &cobra.Command{
		Use:     "delete <postgres>",
		Aliases: []string{"destroy", "rm", "remove"},
		Short:   "delete a postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresDelete(args)
		},
		PreRun: bindPFlags,
	}
	postgresDescribeCmd := &cobra.Command{
		Use:   "describe <postgres>",
		Short: "describe a postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresDescribe(args)
		},
		PreRun: bindPFlags,
	}
	postgresConnectionStringCmd := &cobra.Command{
		Use:   "connectionstring <postgres>",
		Short: "return the connectionstring for a postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresConnectionString(args)
		},
		PreRun: bindPFlags,
	}
	postgresVersionsCmd := &cobra.Command{
		Use:   "version",
		Short: "describe all postgres versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresVersions()
		},
		PreRun: bindPFlags,
	}
	postgresPartitionsCmd := &cobra.Command{
		Use:   "partition",
		Short: "describe all partitions where postgres might be deployed",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresPartitions()
		},
		PreRun: bindPFlags,
	}
	postgresBackupCmd := &cobra.Command{
		Use:   "backup-config",
		Short: "manage postgres backup configuration",
		Long:  "list/find/delete postgres backup configuration",
	}
	postgresBackupCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresBackupCreate(false)
		},
		PreRun: bindPFlags,
	}
	postgresBackupAutoCreateCmd := &cobra.Command{
		Use:   "auto-create",
		Short: "auto create backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresBackupCreate(true)
		},
		PreRun: bindPFlags,
	}
	postgresBackupUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "update backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresBackupUpdate()
		},
		PreRun: bindPFlags,
	}
	postgresBackupListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresBackupList()
		},
		PreRun: bindPFlags,
	}
	postgresBackupDescribeCmd := &cobra.Command{
		Use:   "describe",
		Short: "describe backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresBackupDescribe(args)
		},
		PreRun: bindPFlags,
	}
	postgresBackupDeleteCmd := &cobra.Command{
		Use:     "delete <backup-config>",
		Aliases: []string{"rm", "destroy", "remove", "delete"},
		Short:   "delete a backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.postgresBackupDelete(args)
		},
		PreRun: bindPFlags,
	}

	postgresCmd.AddCommand(postgresBackupCmd)

	postgresCmd.AddCommand(postgresCreateCmd)
	postgresCmd.AddCommand(postgresCreateStandbyCmd)
	postgresCmd.AddCommand(postgresPromoteToPrimaryCmd)
	postgresCmd.AddCommand(postgresDemoteToStandbyCmd)
	postgresCmd.AddCommand(postgresRestoreCmd)
	postgresCmd.AddCommand(postgresApplyCmd)
	postgresCmd.AddCommand(postgresEditCmd)
	postgresCmd.AddCommand(postgresAcceptRestoreCmd)
	postgresCmd.AddCommand(postgresListCmd)
	postgresCmd.AddCommand(postgresListBackupsCmd)
	postgresCmd.AddCommand(postgresDeleteCmd)
	postgresCmd.AddCommand(postgresDescribeCmd)
	postgresCmd.AddCommand(postgresVersionsCmd)
	postgresCmd.AddCommand(postgresPartitionsCmd)
	postgresCmd.AddCommand(postgresConnectionStringCmd)

	postgresBackupCmd.AddCommand(postgresBackupCreateCmd)
	postgresBackupCmd.AddCommand(postgresBackupAutoCreateCmd)
	postgresBackupCmd.AddCommand(postgresBackupUpdateCmd)
	postgresBackupCmd.AddCommand(postgresBackupListCmd)
	postgresBackupCmd.AddCommand(postgresBackupDescribeCmd)
	postgresBackupCmd.AddCommand(postgresBackupDeleteCmd)

	// Create
	postgresCreateCmd.Flags().StringP("description", "", "", "description of the database")
	postgresCreateCmd.Flags().StringP("project", "", "", "project of the database")
	postgresCreateCmd.Flags().StringP("partition", "", "", "partition where the database should be created")
	postgresCreateCmd.Flags().IntP("replicas", "", 1, "replicas of the database")
	postgresCreateCmd.Flags().StringP("version", "", "12", "version of the database") // FIXME add possible values
	postgresCreateCmd.Flags().StringSliceP("sources", "", []string{"0.0.0.0/0"}, "networks which should be allowed to connect in CIDR notation")
	postgresCreateCmd.Flags().StringSliceP("labels", "", []string{}, "labels to add to that postgres database")
	postgresCreateCmd.Flags().StringP("cpu", "", "500m", "cpus for the database")
	postgresCreateCmd.Flags().StringP("buffer", "", "64Mi", "shared buffer for the database")
	postgresCreateCmd.Flags().StringP("storage", "", "10Gi", "storage for the database")
	postgresCreateCmd.Flags().StringP("backup-config", "", "", "backup to use")
	postgresCreateCmd.Flags().StringSliceP("maintenance", "", []string{"Sun:22:00-23:00"}, "time specification of the automatic maintenance in the form Weekday:HH:MM-HH-MM [optional]")
	postgresCreateCmd.Flags().BoolP("audit-logs", "", true, "enable audit logs for the database")
	must(postgresCreateCmd.MarkFlagRequired("description"))
	must(postgresCreateCmd.MarkFlagRequired("project"))
	must(postgresCreateCmd.MarkFlagRequired("partition"))
	must(postgresCreateCmd.MarkFlagRequired("backup-config"))
	must(postgresCreateCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(postgresCreateCmd.RegisterFlagCompletionFunc("partition", c.comp.PostgresListPartitionsCompletion))
	must(postgresCreateCmd.RegisterFlagCompletionFunc("version", c.comp.PostgresListVersionsCompletion))

	// CreateStandby
	postgresCreateStandbyCmd.Flags().StringP("primary-postgres-id", "", "", "id of the primary database")
	postgresCreateStandbyCmd.Flags().StringP("description", "", "", "description of the database")
	postgresCreateStandbyCmd.Flags().StringP("partition", "", "", "partition where the database should be created")
	postgresCreateStandbyCmd.Flags().IntP("replicas", "", 1, "replicas of the database")
	postgresCreateStandbyCmd.Flags().StringSliceP("labels", "", []string{}, "labels to add to that postgres database")
	postgresCreateStandbyCmd.Flags().StringP("backup-config", "", "", "backup to use")
	postgresCreateStandbyCmd.Flags().StringSliceP("maintenance", "", []string{"Sun:22:00-23:00"}, "time specification of the automatic maintenance in the form Weekday:HH:MM-HH-MM [optional]")
	must(postgresCreateStandbyCmd.MarkFlagRequired("primary-postgres-id"))
	must(postgresCreateStandbyCmd.MarkFlagRequired("description"))
	must(postgresCreateStandbyCmd.MarkFlagRequired("partition"))
	must(postgresCreateStandbyCmd.MarkFlagRequired("backup-config"))
	must(postgresCreateStandbyCmd.RegisterFlagCompletionFunc("primary-postgres-id", c.comp.PostgresListCompletion))
	must(postgresCreateStandbyCmd.RegisterFlagCompletionFunc("partition", c.comp.PostgresListPartitionsCompletion))

	// PromoteToPrimary
	postgresPromoteToPrimaryCmd.Flags().BoolP("synchronous", "", false, "make the replication synchronous")

	// Restore
	postgresRestoreCmd.Flags().StringP("source-postgres-id", "", "", "if of the primary database")
	postgresRestoreCmd.Flags().StringP("timestamp", "", time.Now().Format(time.RFC3339), "point-in-time to restore to")
	postgresRestoreCmd.Flags().StringP("version", "", "", "postgres version of the database")
	postgresRestoreCmd.Flags().StringP("description", "", "", "description of the database")
	postgresRestoreCmd.Flags().StringP("partition", "", "", "partition where the database should be created. Changing the partition compared to the source database requires administrative privileges")
	postgresRestoreCmd.Flags().StringSliceP("labels", "", []string{}, "labels to add to that postgres database")
	postgresRestoreCmd.Flags().StringSliceP("maintenance", "", []string{"Sun:22:00-23:00"}, "time specification of the automatic maintenance in the form Weekday:HH:MM-HH-MM [optional]")
	must(postgresRestoreCmd.MarkFlagRequired("source-postgres-id"))
	must(postgresRestoreCmd.RegisterFlagCompletionFunc("source-postgres-id", c.comp.PostgresListCompletion))
	must(postgresRestoreCmd.RegisterFlagCompletionFunc("partition", c.comp.PostgresListPartitionsCompletion))

	// List
	postgresListCmd.Flags().StringP("id", "", "", "postgres id to filter [optional]")
	postgresListCmd.Flags().StringP("description", "", "", "description to filter [optional]")
	postgresListCmd.Flags().StringP("tenant", "", "", "tenant to filter [optional]")
	postgresListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	postgresListCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")

	must(postgresListCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(postgresListCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))

	postgresApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
	Example postgres update:

	# cloudctl postgres describe postgres1 -o yaml > postgres1.yaml
	# vi postgres1.yaml
	## either via stdin
	# cat postgres1.yaml | cloudctl postgres apply -f -
	## or via file
	# cloudctl postgres apply -f postgres1.yaml
	`)

	postgresConnectionStringCmd.Flags().StringP("type", "", "psql", "the type of the connectionstring to create, can be one of psql|jdbc")
	must(postgresConnectionStringCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"jdbc", "psql"}, cobra.ShellCompDirectiveNoFileComp
	}))

	postgresBackupCreateCmd.Flags().StringP("name", "", "", "name of the backup config")
	postgresBackupCreateCmd.Flags().StringP("project", "", "", "project of the backup config")
	postgresBackupCreateCmd.Flags().StringP("schedule", "", "30 00 * * *", "backup schedule in cron syntax")
	postgresBackupCreateCmd.Flags().Int32P("retention", "", int32(10), "number of backups per database to retain")
	postgresBackupCreateCmd.Flags().StringP("s3-endpoint", "", "", "s3 endpoint to backup to")
	postgresBackupCreateCmd.Flags().StringP("s3-region", "", "", "s3 region to backup to [optional]")
	postgresBackupCreateCmd.Flags().StringP("s3-bucketname", "", "", "s3 bucketname to backup to")
	postgresBackupCreateCmd.Flags().StringP("s3-accesskey", "", "", "s3-accesskey")
	postgresBackupCreateCmd.Flags().StringP("s3-secretkey", "", "", "s3-secretkey")
	postgresBackupCreateCmd.Flags().StringP("s3-encryptionkey", "", "", "s3 encryption key, enables sse (server side encryption) if given [optional]")
	must(postgresBackupCreateCmd.MarkFlagRequired("name"))
	must(postgresBackupCreateCmd.MarkFlagRequired("project"))
	must(postgresBackupCreateCmd.MarkFlagRequired("s3-endpoint"))
	must(postgresBackupCreateCmd.MarkFlagRequired("s3-accesskey"))
	must(postgresBackupCreateCmd.MarkFlagRequired("s3-secretkey"))

	postgresBackupAutoCreateCmd.Flags().StringP("name", "", "", "name of the backup config")
	postgresBackupAutoCreateCmd.Flags().StringP("project", "", "", "project of the backup config")
	postgresBackupAutoCreateCmd.Flags().StringP("schedule", "", "30 00 * * *", "backup schedule in cron syntax")
	postgresBackupAutoCreateCmd.Flags().Int32P("retention", "", int32(10), "number of backups per database to retain")
	postgresBackupAutoCreateCmd.Flags().StringP("partition", "", "", "the postgres partition this backup configuration is mainly used in. This e.g. automatically selects the recommended S3 partition for the (auto-created) S3 bucket.")
	must(postgresBackupAutoCreateCmd.MarkFlagRequired("name"))
	must(postgresBackupAutoCreateCmd.MarkFlagRequired("project"))
	must(postgresBackupAutoCreateCmd.MarkFlagRequired("partition"))

	postgresBackupUpdateCmd.Flags().StringP("id", "", "", "id of the database backup")
	postgresBackupUpdateCmd.Flags().StringP("schedule", "", "", "backup schedule in cron syntax [optional]")
	postgresBackupUpdateCmd.Flags().Int32P("retention", "", int32(0), "number of backups per database to retain [optional]")
	must(postgresBackupUpdateCmd.MarkFlagRequired("id"))

	return postgresCmd
}

func (c *config) postgresCreate() error {
	desc := viper.GetString("description")
	project := viper.GetString("project")
	partition := viper.GetString("partition")
	replicas := viper.GetInt32("replicas")
	version := viper.GetString("version")
	sources := viper.GetStringSlice("sources")
	labels := viper.GetStringSlice("labels")
	cpu := viper.GetString("cpu")
	buffer := viper.GetString("buffer")
	backupConfig := viper.GetString("backup-config")
	storage := viper.GetString("storage")
	maintenance := viper.GetStringSlice("maintenance")
	auditLogs := viper.GetBool("audit-logs")

	labelMap, err := helper.LabelsToMap(labels)
	if err != nil {
		return err
	}
	pcr := &models.V1PostgresCreateRequest{
		Description:       desc,
		ProjectID:         project,
		PartitionID:       partition,
		NumberOfInstances: replicas,
		Version:           version,
		Backup:            backupConfig,
		Size: &models.V1PostgresSize{
			CPU:          cpu,
			SharedBuffer: buffer,
			StorageSize:  storage,
		},
		AccessList: &models.V1AccessList{
			SourceRanges: sources,
		},
		Maintenance: maintenance,
		Labels:      labelMap,
		AuditLogs:   auditLogs,
	}
	request := database.NewCreatePostgresParams()
	request.SetBody(pcr)

	response, err := c.cloud.Database.CreatePostgres(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}

func (c *config) postgresCreateStandby() error {
	primaryPostgresID := viper.GetString("primary-postgres-id")
	desc := viper.GetString("description")
	partition := viper.GetString("partition")
	labels := viper.GetStringSlice("labels")
	backupConfig := viper.GetString("backup-config")
	maintenance := viper.GetStringSlice("maintenance")

	labelMap, err := helper.LabelsToMap(labels)
	if err != nil {
		return err
	}
	pcsr := &models.V1PostgresCreateStandbyRequest{
		PrimaryID:   &primaryPostgresID,
		Description: desc,
		PartitionID: partition,
		Backup:      backupConfig,
		Maintenance: maintenance,
		Labels:      labelMap,
	}
	request := database.NewCreatePostgresStandbyParams()
	request.SetBody(pcsr)

	response, err := c.cloud.Database.CreatePostgresStandby(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}

func (c *config) postgresPromoteToPrimary(args []string) error {
	id, err := c.postgresID("promote-to-primary", args)
	if err != nil {
		return err
	}

	params := database.NewGetPostgresParams().WithID(id)
	resp, err := c.cloud.Database.GetPostgres(params, nil)
	if err != nil {
		return err
	}
	current := resp.Payload

	// copy the (minimum) current config
	body := &models.V1PostgresUpdateRequest{
		ProjectID:      current.ProjectID,
		ID:             current.ID,
		Connection:     current.Connection,
		AuditLogs:      current.AuditLogs,
		PostgresParams: current.PostgresParams,
	}

	// abort if there is no configured connection
	if body.Connection == nil {
		return fmt.Errorf("standalone postgres cluster detected, cannot be promoted to primary")
	}

	// promote to primary
	body.Connection.LocalSideIsPrimary = true
	if viper.IsSet("synchronous") {
		// also set the sync flag if given
		body.Connection.Synchronous = viper.GetBool("synchronous")
	}

	// send the update request
	req := database.NewUpdatePostgresParams()
	req.Body = body
	uresp, err := c.cloud.Database.UpdatePostgres(req, nil)
	if err != nil {
		return err
	}
	return output.New().Print(uresp.Payload)
}

func (c *config) postgresDemoteToStandby(args []string) error {
	id, err := c.postgresID("demote-to-standby", args)
	if err != nil {
		return err
	}

	params := database.NewGetPostgresParams().WithID(id)
	resp, err := c.cloud.Database.GetPostgres(params, nil)
	if err != nil {
		return err
	}
	current := resp.Payload

	// copy the (minimum) current config
	body := &models.V1PostgresUpdateRequest{
		ProjectID:      current.ProjectID,
		ID:             current.ID,
		Connection:     current.Connection,
		AuditLogs:      current.AuditLogs,
		PostgresParams: current.PostgresParams,
	}

	// abort if there is no configured connection
	if body.Connection == nil {
		return fmt.Errorf("standalone postgres cluster detected, cannot be demoted to standby")
	}

	// demote to standby
	body.Connection.LocalSideIsPrimary = false

	// send the update request
	req := database.NewUpdatePostgresParams()
	req.Body = body
	uresp, err := c.cloud.Database.UpdatePostgres(req, nil)
	if err != nil {
		return err
	}
	return output.New().Print(uresp.Payload)
}

func (c *config) postgresRestore() error {
	srcID := viper.GetString("source-postgres-id")
	desc := viper.GetString("description")
	partition := viper.GetString("partition")
	labels := viper.GetStringSlice("labels")
	version := viper.GetString("version")
	maintenance := viper.GetStringSlice("maintenance")
	timestamp := viper.GetString("timestamp")

	labelMap, err := helper.LabelsToMap(labels)
	if err != nil {
		return err
	}
	pcsr := &models.V1PostgresRestoreRequest{
		SourceID:    &srcID,
		Description: desc,
		PartitionID: partition,
		Version:     version,
		Maintenance: maintenance,
		Labels:      labelMap,
		Timestamp:   timestamp,
	}
	request := database.NewRestorePostgresParams()
	request.SetBody(pcsr)

	response, err := c.cloud.Database.RestorePostgres(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}

func (c *config) postgresApply() error {
	var pcrs []models.V1PostgresCreateRequest
	var pcr models.V1PostgresCreateRequest

	var purs []models.V1PostgresUpdateRequest
	var pur models.V1PostgresUpdateRequest

	err := helper.ReadFrom(viper.GetString("file"), &pur, func(data interface{}) {
		udoc, ok := data.(*models.V1PostgresUpdateRequest)
		if ok {
			purs = append(purs, *udoc)
			// the request needs to be renewed as otherwise the pointers in the request struct will
			// always point to same last value in the multi-document loop
			pur = models.V1PostgresUpdateRequest{}
		}
	})
	if err != nil {
		return err
	}

	err = helper.ReadFrom(viper.GetString("file"), &pcr, func(data interface{}) {
		cdoc, ok := data.(*models.V1PostgresCreateRequest)
		if ok {
			pcrs = append(pcrs, *cdoc)
			// the request needs to be renewed as otherwise the pointers in the request struct will
			// always point to same last value in the multi-document loop
			pcr = models.V1PostgresCreateRequest{}
		}
	})

	if err != nil {
		return err
	}
	response := []*models.V1PostgresResponse{}
	for i, par := range purs {
		if par.ID == nil {
			continue
		}
		params := database.NewGetPostgresParams().WithID(*par.ID)
		resp, err := c.cloud.Database.GetPostgres(params, nil)
		if err != nil {
			return err
		}
		if resp.Payload.ID != nil {
			// existing postgres, update
			request := database.NewUpdatePostgresParams()
			request.SetBody(&purs[i])

			updatedPG, err := c.cloud.Database.UpdatePostgres(request, nil)
			if err != nil {
				return err
			}
			response = append(response, updatedPG.Payload)
			continue
		}
	}
	for i := range pcrs {

		// no postgres found, create it
		request := database.NewCreatePostgresParams()
		request.SetBody(&pcrs[i])

		createdPG, err := c.cloud.Database.CreatePostgres(request, nil)
		if err != nil {
			return err
		}
		response = append(response, createdPG.Payload)
		continue
	}
	return output.New().Print(response)
}

func (c *config) postgresEdit(args []string) error {
	id, err := c.postgresID("edit", args)
	if err != nil {
		return err
	}

	getFunc := func(id string) ([]byte, error) {
		params := database.NewGetPostgresParams().WithID(id)
		resp, err := c.cloud.Database.GetPostgres(params, nil)
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
		uresp, err := c.cloud.Database.UpdatePostgres(pup, nil)
		if err != nil {
			return err
		}
		return output.New().Print(uresp.Payload)
	}
	return helper.Edit(id, getFunc, updateFunc)
}

func (c *config) postgresAcceptRestore(args []string) error {
	pg, err := c.getPostgresFromArgs(args)
	if err != nil {
		return err
	}

	must(output.New().Print(pg))

	fmt.Println("Has the restore finished successfully?")
	err = helper.Prompt("(type yes to proceed):", "yes")
	if err != nil {
		return err
	}

	params := database.NewAcceptPostgresRestoreParams().WithID(*pg.ID)
	resp, err := c.cloud.Database.AcceptPostgresRestore(params, nil)
	if err != nil {
		return err
	}

	return output.New().Print(resp.Payload)
}

func readPostgresUpdateRequests(filename string) ([]models.V1PostgresUpdateRequest, error) {
	var purs []models.V1PostgresUpdateRequest
	var pur models.V1PostgresUpdateRequest
	err := helper.ReadFrom(filename, &pur, func(data interface{}) {
		doc := data.(*models.V1PostgresUpdateRequest)
		purs = append(purs, *doc)
	})
	if err != nil {
		return purs, err
	}
	if len(purs) != 1 {
		return purs, fmt.Errorf("postgres update error more or less than one postgres given:%d", len(purs))
	}
	return purs, nil
}

func (c *config) postgresFind() error {
	if helper.AtLeastOneViperStringFlagGiven("id", "description", "tenant", "project", "partition") {
		params := database.NewFindPostgresParams()
		ifr := &models.V1PostgresFindRequest{}
		id := helper.ViperString("id")
		if id != nil {
			ifr.ID = *id
		}
		description := helper.ViperString("description")
		if description != nil {
			ifr.Description = *description
		}
		tenant := helper.ViperString("tenant")
		if tenant != nil {
			ifr.Tenant = *tenant
		}
		projectID := helper.ViperString("project")
		if projectID != nil {
			ifr.ProjectID = *projectID
		}
		partitionID := helper.ViperString("partition")
		if partitionID != nil {
			ifr.PartitionID = *partitionID
		}

		params.SetBody(ifr)
		resp, err := c.cloud.Database.FindPostgres(params, nil)
		if err != nil {
			return err
		}
		return output.New().Print(resp.Payload)
	}
	resp, err := c.cloud.Database.ListPostgres(nil, nil)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Payload)
}

func (c *config) postgresDelete(args []string) error {
	pg, err := c.getPostgresFromArgs(args)
	if err != nil {
		return err
	}

	if !viper.GetBool("yes-i-really-mean-it") {
		must(output.New().Print(pg))

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
	}

	params := database.NewDeletePostgresParams().WithID(*pg.ID)
	resp, err := c.cloud.Database.DeletePostgres(params, nil)
	if err != nil {
		return err
	}

	return output.New().Print(resp.Payload)
}

func (c *config) postgresDescribe(args []string) error {
	postgres, err := c.getPostgresFromArgs(args)
	if err != nil {
		return err
	}

	return output.New().Print(postgres)
}

func (c *config) postgresListBackups(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no postgres id given")
	}

	id := args[0]
	params := database.NewGetPostgresBackupsParams().WithID(id)
	resp, err := c.cloud.Database.GetPostgresBackups(params, nil)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Payload)
}

func (c *config) postgresConnectionString(args []string) error {
	t := viper.GetString("type")

	postgres, err := c.getPostgresFromArgs(args)
	if err != nil {
		return err
	}

	params := database.NewGetPostgresSecretsParams().WithID(*postgres.ID)
	resp, err := c.cloud.Database.GetPostgresSecrets(params, nil)
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
			fmt.Printf("jdbc:postgresql://%s:%d/postgres?user=%s&password=%s&ssl=true&tcpKeepAlive=true\n", ip, port, user, password)
		case "psql":
			fmt.Printf("PGPASSWORD=%s psql --host=%s --port=%d --username=%s\n", password, ip, port, user)
		default:
			return fmt.Errorf("unknown connectionstring type:%s", t)
		}
	}
	return nil
}

func (c *config) postgresBackupCreate(autocreate bool) error {
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

	bcr := &models.V1PostgresBackupConfigCreateRequest{
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
		bcr.Secret = &models.V1PostgresBackupSecret{
			Accesskey: s3Accesskey,
			Secretkey: s3Secretkey,
		}
		if s3Encryptionkey != "" {
			bcr.Secret.S3encryptionkey = s3Encryptionkey
		}
	}
	request := database.NewCreatePostgresBackupConfigParams()
	request.SetBody(bcr)

	response, err := c.cloud.Database.CreatePostgresBackupConfig(request, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}
func (c *config) postgresBackupUpdate() error {
	id := viper.GetString("id")

	request := database.NewGetBackupConfigParams().WithID(id)
	resp, err := c.cloud.Database.GetBackupConfig(request, nil)
	if err != nil {
		return err
	}
	if resp == nil || resp.Payload == nil {
		return fmt.Errorf("given backup %s does not exist", id)
	}

	schedule := viper.GetString("schedule")
	retention := viper.GetInt32("retention")

	bur := &models.V1PostgresBackupConfigUpdateRequest{
		ID: id,
	}
	if schedule != "" {
		bur.Schedule = schedule
	}
	if retention != 0 {
		bur.Retention = retention
	}

	req := database.NewUpdatePostgresBackupConfigParams()
	req.SetBody(bur)

	response, err := c.cloud.Database.UpdatePostgresBackupConfig(req, nil)
	if err != nil {
		return err
	}

	return output.New().Print(response.Payload)
}

func (c *config) postgresBackupList() error {

	request := database.NewListPostgresBackupConfigsParams()
	resp, err := c.cloud.Database.ListPostgresBackupConfigs(request, nil)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Payload)
}
func (c *config) postgresBackupDescribe(args []string) error {

	if len(args) < 1 {
		return fmt.Errorf("missing backup id")
	}
	if len(args) > 1 {
		return fmt.Errorf("only a single backup id is supported")
	}
	id := args[0]

	gbcp := database.NewGetBackupConfigParams().WithID(id)
	resp, err := c.cloud.Database.GetBackupConfig(gbcp, nil)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Payload)
}
func (c *config) postgresBackupDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing backup id")
	}
	if len(args) > 1 {
		return fmt.Errorf("only a single backup id is supported")
	}
	id := args[0]

	// try to fetch that backup-config first
	gbcp := database.NewGetBackupConfigParams().WithID(args[0])
	_, err := c.cloud.Database.GetBackupConfig(gbcp, nil)
	if err != nil {
		return err
	}

	idParts := strings.Split(id, "-")
	firstPartOfID := idParts[0]
	lastPartOfID := idParts[len(idParts)-1]
	fmt.Println("Please answer some security questions to delete this postgres database backup")
	err = helper.Prompt("first part of ID:", firstPartOfID)
	if err != nil {
		return err
	}
	err = helper.Prompt("last part of ID:", lastPartOfID)
	if err != nil {
		return err
	}

	request := database.NewDeletePostgresBackupConfigParams().WithID(id)
	resp, err := c.cloud.Database.DeletePostgresBackupConfig(request, nil)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Payload)

}

func (c *config) postgresVersions() error {
	params := database.NewGetPostgresVersionsParams()
	resp, err := c.cloud.Database.GetPostgresVersions(params, nil)
	if err != nil {
		return err
	}

	return output.New().Print(resp.Payload)
}
func (c *config) postgresPartitions() error {
	params := database.NewGetPostgresPartitionsParams()
	resp, err := c.cloud.Database.GetPostgresPartitions(params, nil)
	if err != nil {
		return err
	}

	return output.New().Print(resp.Payload)
}
func (c *config) getPostgresFromArgs(args []string) (*models.V1PostgresResponse, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no postgres id given")
	}

	id := args[0]
	params := database.NewGetPostgresParams().WithID(id)
	resp, err := c.cloud.Database.GetPostgres(params, nil)
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *config) postgresID(verb string, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("postgres %s requires postgresID as argument", verb)
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", fmt.Errorf("postgres %s requires exactly one postgresID as argument", verb)
}
