package cmd

import (
	"fmt"

	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/sorters"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type machineReservationsCmd struct {
	*config
}

func newMachineReservationsCmd(c *config) *cobra.Command {
	w := machineReservationsCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1MachineReservationCreateRequest, *models.V1MachineReservationUpdateRequest, *models.V1MachineReservationResponse]{
		BinaryName:      binaryName,
		GenericCLI:      genericcli.NewGenericCLI(w).WithFS(c.fs),
		Args:            []string{"project", "size"},
		Singular:        "machine-reservation",
		Plural:          "machine-reservations",
		Description:     "manage machine reservations, ids must be provided in the form <project>@<size>",
		Sorter:          sorters.MachineReservationsSorter(),
		ValidArgsFn:     c.comp.MachineReservationListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("project", "", "show reservations of given project")
			cmd.Flags().String("size", "", "show reservations of given size")
			cmd.Flags().String("tenant", "", "show reservations of given tenant")
			genericcli.Must(cmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
		},
		UpdateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().Int32("amount", 0, "the amount of machines to reserve")
			cmd.Flags().String("description", "", "the description of the reservation")
			cmd.Flags().StringSlice("partitions", nil, "the partitions in which this reservation is being made")
			cmd.Flags().Bool("force", false, "allows overbooking of a partition")
			genericcli.Must(cmd.RegisterFlagCompletionFunc("partitions", c.comp.PartitionListCompletion))
		},
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("project", "", "the project of the reservation")
			cmd.Flags().String("size", "", "the size of the reservation")
			cmd.Flags().Int32("amount", 0, "the amount of machines to reserve")
			cmd.Flags().String("description", "", "the description of the reservation")
			cmd.Flags().StringSlice("partitions", nil, "the partitions in which this reservation is being made")
			cmd.Flags().Bool("force", false, "allows overbooking of a partition")
			genericcli.Must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("partitions", c.comp.PartitionListCompletion))
		},
		CreateRequestFromCLI: func() (*models.V1MachineReservationCreateRequest, error) {
			return &models.V1MachineReservationCreateRequest{
				Amount:       pointer.PointerOrNil(viper.GetInt32("amount")),
				Description:  pointer.PointerOrNil(viper.GetString("description")),
				Partitionids: viper.GetStringSlice("partitions"),
				Projectid:    pointer.PointerOrNil(viper.GetString("project")),
				Sizeid:       pointer.PointerOrNil(viper.GetString("size")),
			}, nil
		},
		UpdateRequestFromCLI: func(args []string) (*models.V1MachineReservationUpdateRequest, error) {
			ids, err := genericcli.GetExactlyNArgs(2, args)
			if err != nil {
				return nil, err
			}

			return &models.V1MachineReservationUpdateRequest{
				Amount:       pointer.PointerOrNil(viper.GetInt32("amount")),
				Description:  pointer.PointerOrNil(viper.GetString("description")),
				Partitionids: viper.GetStringSlice("partitions"),
				Projectid:    &ids[w.projectIndex()],
				Sizeid:       &ids[w.sizeIndex()],
			}, nil
		},
	}

	usageCmd := &cobra.Command{
		Use:               "usage",
		Short:             "shows the current usage of machine reservations",
		ValidArgsFunction: c.comp.MachineReservationListCompletion,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return w.machineReservationsUsage()
		},
	}

	usageCmd.Flags().String("project", "", "show reservations of given project")
	usageCmd.Flags().String("size", "", "show reservations of given size")
	usageCmd.Flags().String("tenant", "", "show reservations of given tenant")
	genericcli.Must(usageCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
	genericcli.Must(usageCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	genericcli.Must(usageCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))

	return genericcli.NewCmds(cmdsConfig, usageCmd)
}

func (m machineReservationsCmd) Convert(r *models.V1MachineReservationResponse) ([]string, *models.V1MachineReservationCreateRequest, *models.V1MachineReservationUpdateRequest, error) {
	return []string{*r.Projectid, *r.Sizeid}, toMachineReservationCreateRequest(r), toMachineReservationUpdateRequest(r), nil
}

func toMachineReservationCreateRequest(r *models.V1MachineReservationResponse) *models.V1MachineReservationCreateRequest {
	return &models.V1MachineReservationCreateRequest{
		Amount:       r.Amount,
		Description:  &r.Description,
		Partitionids: r.Partitionids,
		Projectid:    r.Projectid,
		Sizeid:       r.Sizeid,
	}
}

func toMachineReservationUpdateRequest(r *models.V1MachineReservationResponse) *models.V1MachineReservationUpdateRequest {
	return &models.V1MachineReservationUpdateRequest{
		Amount:       r.Amount,
		Description:  &r.Description,
		Partitionids: r.Partitionids,
		Projectid:    r.Projectid,
		Sizeid:       r.Sizeid,
	}
}

func (m machineReservationsCmd) Create(rq *models.V1MachineReservationCreateRequest) (*models.V1MachineReservationResponse, error) {
	resp, err := m.cloud.Project.CreateMachineReservation(project.NewCreateMachineReservationParams().
		WithBody(rq).
		WithForce(pointer.Pointer(viper.GetBool("force"))), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (m machineReservationsCmd) Delete(id ...string) (*models.V1MachineReservationResponse, error) {
	resp, err := m.cloud.Project.DeleteMachineReservation(project.NewDeleteMachineReservationParams().
		WithProject(pointer.Pointer(id[m.projectIndex()])).
		WithSize(pointer.Pointer(id[m.sizeIndex()])), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (m machineReservationsCmd) Get(id ...string) (*models.V1MachineReservationResponse, error) {
	all, err := m.List()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch machine reservations: %w", err)
	}

	for _, rv := range all {
		if id[m.projectIndex()] != pointer.SafeDeref(rv.Projectid) {
			continue
		}
		if id[m.sizeIndex()] != pointer.SafeDeref(rv.Sizeid) {
			continue
		}

		return rv, nil
	}

	return nil, fmt.Errorf("no reservation found with size %q for project %q", id[1], id[0])
}

func (m machineReservationsCmd) List() ([]*models.V1MachineReservationResponse, error) {
	resp, err := m.cloud.Project.ListMachineReservations(project.NewListMachineReservationsParams().
		WithBody(&models.V1MachineReservationFindRequest{
			Projectid: pointer.PointerOrNil(viper.GetString("project")),
			Sizeid:    pointer.PointerOrNil(viper.GetString("size")),
			Tenant:    pointer.PointerOrNil(viper.GetString("tenant")),
		}), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (m machineReservationsCmd) Update(rq *models.V1MachineReservationUpdateRequest) (*models.V1MachineReservationResponse, error) {
	resp, err := m.cloud.Project.UpdateMachineReservation(project.NewUpdateMachineReservationParams().WithBody(rq).
		WithForce(pointer.Pointer(viper.GetBool("force"))), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (m machineReservationsCmd) machineReservationsUsage() error {
	resp, err := m.cloud.Project.MachineReservationsUsage(project.NewMachineReservationsUsageParams().
		WithBody(&models.V1MachineReservationFindRequest{
			Projectid: pointer.PointerOrNil(viper.GetString("project")),
			Sizeid:    pointer.PointerOrNil(viper.GetString("size")),
			Tenant:    pointer.PointerOrNil(viper.GetString("tenant")),
		}), nil)
	if err != nil {
		return err
	}

	return m.listPrinter.Print(resp.Payload)
}

func (m machineReservationsCmd) projectIndex() int {
	return 0
}
func (m machineReservationsCmd) sizeIndex() int {
	return 1
}
