package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/fi-ts/cloudctl/pkg/client/llm"
	"github.com/spf13/cobra"
)

type llmStore interface {
	List() ([]llm.LLMEndpoint, error)
	Create(name, model, project string) (*llm.LLMEndpoint, error)
	Delete(endpointID, project string) error
}

var newLLMStore = func() llmStore { return llm.NewStore() }

func newLLMCmd(c *config) *cobra.Command {
	llmCmd := &cobra.Command{
		Use:   "llm",
		Short: "[Preview] manage llm endpoints",
		Long:  "[Preview] manage large language model endpoints",
	}

	var (
		wide    bool
		order   string
		project string
	)

	llmListCmd := &cobra.Command{
		Use:     "list",
		Short:   "[Preview] list llm endpoints",
		Long:    "[Preview] list all llm endpoints",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.llmList(wide, order, project)
		},
	}
	llmListCmd.Flags().BoolVarP(&wide, "wide", "w", false, "display with additional columns")
	llmListCmd.Flags().StringVar(&order, "order", "", "order by column(s) (comma separated, prefix with - for descending). Supported: id, name, model, project, status, created, address")
	llmListCmd.Flags().StringVar(&project, "project", "", "filter endpoints by project")

	var (
		createName    string
		createModel   string
		createProject string
	)

	llmCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "[Preview] create a new llm endpoint",
		Long:  "[Preview] create a new llm endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.llmCreate(createName, createModel, createProject)
		},
	}
	llmCreateCmd.Flags().StringVar(&createName, "name", "", "name of the endpoint")
	llmCreateCmd.Flags().StringVar(&createModel, "model", "", "model of the endpoint")
	llmCreateCmd.Flags().StringVar(&createProject, "project", "", "project of the endpoint")
	if err := llmCreateCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	if err := llmCreateCmd.MarkFlagRequired("model"); err != nil {
		panic(err)
	}
	if err := llmCreateCmd.MarkFlagRequired("project"); err != nil {
		panic(err)
	}

	var deleteProject string

	llmDeleteCmd := &cobra.Command{
		Use:     "delete <endpoint-id>",
		Short:   "[Preview] delete an llm endpoint",
		Long:    "[Preview] delete an llm endpoint by id",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.llmDelete(args[0], deleteProject)
		},
	}
	llmDeleteCmd.Flags().StringVar(&deleteProject, "project", "", "project of the endpoint")
	if err := llmDeleteCmd.MarkFlagRequired("project"); err != nil {
		panic(err)
	}

	llmCmd.AddCommand(llmListCmd)
	llmCmd.AddCommand(llmCreateCmd)
	llmCmd.AddCommand(llmDeleteCmd)

	return llmCmd
}

func (c *config) llmList(wide bool, order, project string) error {
	store := newLLMStore()

	eps, err := store.List()
	if err != nil {
		return fmt.Errorf("could not list endpoints: %w", err)
	}

	filtered := filterByProject(eps, project)
	sorted := sortEndpoints(filtered, order)
	renderEndpointsTable(c.out, sorted, wide)

	return nil
}

func (c *config) llmCreate(name, model, project string) error {
	store := newLLMStore()

	ep, err := store.Create(name, model, project)
	if err != nil {
		if isValidationError(err) {
			return err
		}
		return fmt.Errorf("could not complete creation: %w", err)
	}

	renderCreatedEndpoint(c.out, *ep)

	return nil
}

func (c *config) llmDelete(endpointID, project string) error {
	if project == "" {
		return fmt.Errorf("required flag --project not set")
	}

	store := newLLMStore()

	if err := store.Delete(endpointID, project); err != nil {
		return err
	}

	fmt.Fprintf(c.out, "deleted endpoint %s\n", endpointID)

	return nil
}

func isValidationError(err error) bool {
	msg := err.Error()
	return strings.HasPrefix(msg, "invalid endpoint name") || strings.HasPrefix(msg, "invalid model")
}

func filterByProject(endpoints []llm.LLMEndpoint, project string) []llm.LLMEndpoint {
	if project == "" {
		return endpoints
	}

	filtered := make([]llm.LLMEndpoint, 0, len(endpoints))
	for _, e := range endpoints {
		if e.Project == project {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func sortEndpoints(endpoints []llm.LLMEndpoint, order string) []llm.LLMEndpoint {
	sorted := make([]llm.LLMEndpoint, len(endpoints))
	copy(sorted, endpoints)

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

			cmp := compareEndpointField(sorted[i], sorted[j], col)
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

func compareEndpointField(a, b llm.LLMEndpoint, col string) int {
	switch col {
	case "id":
		return strings.Compare(a.EndpointID, b.EndpointID)
	case "name":
		return strings.Compare(a.EndpointName, b.EndpointName)
	case "model":
		return strings.Compare(a.Model, b.Model)
	case "project":
		return strings.Compare(a.Project, b.Project)
	case "status":
		return strings.Compare(a.Status, b.Status)
	case "created":
		return strings.Compare(a.CreatedTimestamp, b.CreatedTimestamp)
	case "address":
		return strings.Compare(a.EndpointAddress, b.EndpointAddress)
	default:
		return 0
	}
}

const (
	colEndpointID       = "Endpoint_Id"
	colEndpointName     = "Endpoint_Name"
	colModel            = "Model"
	colProject          = "Project"
	colStatus           = "Status"
	colCreatedTimestamp = "Created_Timestamp"
	colEndpointAddress  = "Endpoint_Address"
)

func renderEndpointsTable(w io.Writer, endpoints []llm.LLMEndpoint, wide bool) {
	headers := []string{colEndpointID, colEndpointName, colModel, colProject, colStatus}
	if wide {
		headers = append(headers, colCreatedTimestamp, colEndpointAddress)
	}

	rows := make([][]string, 0, len(endpoints))
	for _, e := range endpoints {
		row := []string{e.EndpointID, e.EndpointName, e.Model, e.Project, e.Status}
		if wide {
			row = append(row, e.CreatedTimestamp, e.EndpointAddress)
		}
		rows = append(rows, row)
	}

	// First pass: compute each column's width as the widest value.
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

func writeTableRow(w io.Writer, cells []string, widths []int) {
	fields := make([]string, len(cells))
	for i, cell := range cells {
		if i == len(cells)-1 {
			fields[i] = cell
			continue
		}
		fields[i] = fmt.Sprintf("%-*s", widths[i], cell)
	}
	fmt.Fprintln(w, strings.Join(fields, "  "))
}

func renderCreatedEndpoint(w io.Writer, e llm.LLMEndpoint) {
	headers := []string{colEndpointID, colEndpointName, colModel, colProject, colStatus, colCreatedTimestamp}
	values := []string{e.EndpointID, e.EndpointName, e.Model, e.Project, e.Status, e.CreatedTimestamp}

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
