package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/fi-ts/cloudctl/pkg/client/mssql"
	"github.com/spf13/cobra"
)

type mssqlStore interface {
	List() ([]mssql.Database, error)
	Create(name, version, project string, storageGB int) (*mssql.Database, error)
	Delete(id, project string) error
}

var newMSSQLStore = func() mssqlStore { return mssql.NewStore() }

func newDBCmd(c *config) *cobra.Command {
	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "[Preview] manage database services",
		Long:  "[Preview] manage database services (mssql, ...)",
	}

	dbCmd.AddCommand(newDBMSSQLCmd(c))

	return dbCmd
}

func newDBMSSQLCmd(c *config) *cobra.Command {
	mssqlCmd := &cobra.Command{
		Use:   "mssql",
		Short: "[Preview] manage mssql databases",
		Long:  "[Preview] manage Microsoft SQL Server databases",
	}

	var (
		wide    bool
		order   string
		project string
	)

	mssqlListCmd := &cobra.Command{
		Use:     "list",
		Short:   "[Preview] list mssql databases",
		Long:    "[Preview] list all mssql databases",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.dbList(wide, order, project)
		},
	}
	mssqlListCmd.Flags().BoolVarP(&wide, "wide", "w", false, "display with additional columns")
	mssqlListCmd.Flags().StringVar(&order, "order", "", "order by column(s) (comma separated, prefix with - for descending). Supported: id, name, project, version, storage, host, port, status, created")
	mssqlListCmd.Flags().StringVar(&project, "project", "", "filter databases by project")

	var (
		createName    string
		createVersion string
		createProject string
		createStorage int
	)

	mssqlCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "[Preview] create a new mssql database",
		Long:  "[Preview] create a new mssql database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.dbCreate(createName, createVersion, createProject, createStorage)
		},
	}
	mssqlCreateCmd.Flags().StringVar(&createName, "name", "", "name of the database")
	mssqlCreateCmd.Flags().StringVar(&createVersion, "version", "", "SQL Server version (2019, 2022)")
	mssqlCreateCmd.Flags().StringVar(&createProject, "project", "", "project of the database")
	mssqlCreateCmd.Flags().IntVar(&createStorage, "storage-gb", mssql.DefaultStorage, "storage size in GB")
	if err := mssqlCreateCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	if err := mssqlCreateCmd.MarkFlagRequired("version"); err != nil {
		panic(err)
	}
	if err := mssqlCreateCmd.MarkFlagRequired("project"); err != nil {
		panic(err)
	}

	var deleteProject string

	mssqlDeleteCmd := &cobra.Command{
		Use:     "delete <database-id>",
		Short:   "[Preview] delete a mssql database",
		Long:    "[Preview] delete a mssql database by id",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.dbDelete(args[0], deleteProject)
		},
	}
	mssqlDeleteCmd.Flags().StringVar(&deleteProject, "project", "", "project of the database")
	if err := mssqlDeleteCmd.MarkFlagRequired("project"); err != nil {
		panic(err)
	}

	mssqlCmd.AddCommand(mssqlListCmd)
	mssqlCmd.AddCommand(mssqlCreateCmd)
	mssqlCmd.AddCommand(mssqlDeleteCmd)

	return mssqlCmd
}

func (c *config) dbList(wide bool, order, project string) error {
	store := newMSSQLStore()

	dbs, err := store.List()
	if err != nil {
		return fmt.Errorf("could not list databases: %w", err)
	}

	filtered := filterDBsByProject(dbs, project)
	sorted := sortDatabases(filtered, order)
	renderDatabasesTable(c.out, sorted, wide)

	return nil
}

func (c *config) dbCreate(name, version, project string, storageGB int) error {
	store := newMSSQLStore()

	db, err := store.Create(name, version, project, storageGB)
	if err != nil {
		if isDBValidationError(err) {
			return err
		}
		return fmt.Errorf("could not complete creation: %w", err)
	}

	renderCreatedDatabase(c.out, *db)

	return nil
}

func (c *config) dbDelete(id, project string) error {
	if project == "" {
		return fmt.Errorf("required flag --project not set")
	}

	store := newMSSQLStore()

	if err := store.Delete(id, project); err != nil {
		return err
	}

	fmt.Fprintf(c.out, "deleted database %s\n", id)

	return nil
}

func isDBValidationError(err error) bool {
	msg := err.Error()
	return strings.HasPrefix(msg, "invalid database name") ||
		strings.HasPrefix(msg, "invalid version") ||
		strings.HasPrefix(msg, "invalid storage_gb")
}

func filterDBsByProject(databases []mssql.Database, project string) []mssql.Database {
	if project == "" {
		return databases
	}

	filtered := make([]mssql.Database, 0, len(databases))
	for _, db := range databases {
		if db.Project == project {
			filtered = append(filtered, db)
		}
	}
	return filtered
}

func sortDatabases(databases []mssql.Database, order string) []mssql.Database {
	sorted := make([]mssql.Database, len(databases))
	copy(sorted, databases)

	if strings.TrimSpace(order) == "" {
		return sorted
	}

	cols := strings.Split(order, ",")

	sort.SliceStable(sorted, func(i, j int) bool {
		for _, col := range cols {
			col = strings.TrimSpace(col)
			if col == "" {
				continue
			}

			descending := false
			if strings.HasPrefix(col, "-") {
				descending = true
				col = strings.TrimPrefix(col, "-")
			}

			cmp := compareDatabaseField(sorted[i], sorted[j], col)
			if cmp == 0 {
				continue
			}
			if descending {
				return cmp > 0
			}
			return cmp < 0
		}
		return false
	})

	return sorted
}

func compareDatabaseField(a, b mssql.Database, col string) int {
	switch col {
	case "id":
		return strings.Compare(a.ID, b.ID)
	case "name":
		return strings.Compare(a.Name, b.Name)
	case "project":
		return strings.Compare(a.Project, b.Project)
	case "version":
		return strings.Compare(a.Version, b.Version)
	case "storage":
		return intCompare(a.StorageGB, b.StorageGB)
	case "host":
		return strings.Compare(a.Host, b.Host)
	case "port":
		return intCompare(a.Port, b.Port)
	case "status":
		return strings.Compare(a.Status, b.Status)
	case "created":
		return strings.Compare(a.CreatedTimestamp, b.CreatedTimestamp)
	default:
		return 0
	}
}

func intCompare(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

const (
	colDBID        = "ID"
	colDBName      = "Name"
	colDBProject   = "Project"
	colDBVersion   = "Version"
	colDBStorageGB = "Storage_GB"
	colDBHost      = "Host"
	colDBPort      = "Port"
	colDBStatus    = "Status"
	colDBCreated   = "Created"
)

func renderDatabasesTable(w io.Writer, databases []mssql.Database, wide bool) {
	headers := []string{colDBID, colDBName, colDBProject, colDBVersion, colDBStatus}
	if wide {
		headers = append(headers, colDBStorageGB, colDBHost, colDBPort, colDBCreated)
	}

	rows := make([][]string, 0, len(databases))
	for _, db := range databases {
		row := []string{db.ID, db.Name, db.Project, db.Version, db.Status}
		if wide {
			row = append(row, fmt.Sprintf("%d", db.StorageGB), db.Host, fmt.Sprintf("%d", db.Port), db.CreatedTimestamp)
		}
		rows = append(rows, row)
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	writeTableRow(w, headers, widths)
	for _, row := range rows {
		writeTableRow(w, row, widths)
	}
}

func renderCreatedDatabase(w io.Writer, db mssql.Database) {
	headers := []string{colDBID, colDBName, colDBProject, colDBVersion, colDBStorageGB, colDBStatus, colDBCreated}
	values := []string{db.ID, db.Name, db.Project, db.Version, fmt.Sprintf("%d", db.StorageGB), db.Status, db.CreatedTimestamp}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for i, v := range values {
		if len(v) > widths[i] {
			widths[i] = len(v)
		}
	}

	writeTableRow(w, headers, widths)
	writeTableRow(w, values, widths)
}
