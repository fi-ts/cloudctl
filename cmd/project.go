package cmd

import (
	"errors"
	"fmt"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"

	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/fi-ts/cloudctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type projectCmd struct {
	*config
}

func newProjectCmd(c *config) *cobra.Command {
	w := projectCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, *models.V1ProjectResponse]{
		BinaryName:      binaryName,
		GenericCLI:      genericcli.NewGenericCLI(w).WithFS(c.fs),
		Singular:        "project",
		Plural:          "projects",
		Description:     "a project organizes cloud resources regarding tenancy, quotas, billing and authentication",
		Sorter:          sorters.ProjectSorter(),
		ValidArgsFn:     c.comp.ProjectListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("name", "", "name of the project, max 10 characters. [required]")
			cmd.Flags().String("description", "", "description of the project. [required]")
			cmd.Flags().String("tenant", "", "create project for given tenant")
			cmd.Flags().StringSlice("label", nil, "add initial label, can be given multiple times to add multiple labels, e.g. --label=foo --label=bar")
			cmd.Flags().StringSlice("annotation", nil, "add initial annotation, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar")
			cmd.Flags().Int32("cluster-quota", 0, "cluster quota")
			cmd.Flags().Int32("machine-quota", 0, "machine quota")
			cmd.Flags().Int32("ip-quota", 0, "ip quota")

			cmd.MarkFlagsMutuallyExclusive("file", "name")
			cmd.MarkFlagsRequiredTogether("name", "description")
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("name", "", "", "show projects of given name")
			cmd.Flags().StringP("id", "", "", "show projects of given id")
			cmd.Flags().StringP("tenant", "", "", "show projects of given id")
		},
	}

	return genericcli.NewCmds(cmdsConfig)
}

func (c projectCmd) Get(id string) (*models.V1ProjectResponse, error) {
	resp, err := c.cloud.Project.FindProject(project.NewFindProjectParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c projectCmd) List() ([]*models.V1ProjectResponse, error) {
	resp, err := c.cloud.Project.FindProjects(project.NewFindProjectsParams().WithBody(&models.V1ProjectFindRequest{
		ID:       viper.GetString("id"),
		Name:     viper.GetString("name"),
		TenantID: viper.GetString("tenant"),
	}), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload.Projects, nil
}

func (c projectCmd) Delete(id string) (*models.V1ProjectResponse, error) {
	resp, err := c.cloud.Project.DeleteProject(project.NewDeleteProjectParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c projectCmd) Create(rq *models.V1ProjectCreateRequest) (*models.V1ProjectResponse, error) {
	resp, err := c.cloud.Project.CreateProject(project.NewCreateProjectParams().WithBody(rq), nil)
	if err != nil {
		var r *project.CreateProjectConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c projectCmd) Update(rq *models.V1ProjectUpdateRequest) (*models.V1ProjectResponse, error) {
	resp, err := c.cloud.Project.FindProject(project.NewFindProjectParams().WithID(rq.Meta.ID), nil)
	if err != nil {
		return nil, err
	}

	rq.Meta.Version = resp.Payload.Meta.Version

	updateResp, err := c.cloud.Project.UpdateProject(project.NewUpdateProjectParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}

func (c projectCmd) Convert(r *models.V1ProjectResponse) (string, *models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, error) {
	if r.Meta == nil {
		return "", nil, nil, fmt.Errorf("meta is nil")
	}
	return r.Meta.ID, projectResponseToCreate(r), projectResponseToUpdate(r), nil
}

func projectResponseToCreate(r *models.V1ProjectResponse) *models.V1ProjectCreateRequest {
	return &models.V1ProjectCreateRequest{
		Meta: &models.V1Meta{
			Apiversion:  r.Meta.Apiversion,
			Kind:        r.Meta.Kind,
			ID:          r.Meta.ID,
			Annotations: r.Meta.Annotations,
			Labels:      r.Meta.Labels,
			Version:     r.Meta.Version,
		},
		Description: r.Description,
		Name:        r.Name,
		Quotas:      r.Quotas,
		TenantID:    r.TenantID,
	}
}

func projectResponseToUpdate(r *models.V1ProjectResponse) *models.V1ProjectUpdateRequest {
	return &models.V1ProjectUpdateRequest{
		Meta: &models.V1Meta{
			Apiversion:  r.Meta.Apiversion,
			Kind:        r.Meta.Kind,
			ID:          r.Meta.ID,
			Annotations: r.Meta.Annotations,
			Labels:      r.Meta.Labels,
			Version:     r.Meta.Version,
		},
		Description: r.Description,
		Name:        r.Name,
		Quotas:      r.Quotas,
		TenantID:    r.TenantID,
	}
}
