package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"gopkg.in/yaml.v3"

	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/fi-ts/cloudctl/cmd/helper"
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
		Description:     "manage projects, a project organizes cloud resources regarding tenancy, quotas, billing and authentication",
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
			genericcli.Must(cmd.MarkFlagRequired("name"))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("id", "", "show projects of given id")
			cmd.Flags().String("name", "", "show projects of given name")
			cmd.Flags().String("tenant", "", "show projects of given tenant")
			genericcli.Must(cmd.RegisterFlagCompletionFunc("id", c.comp.ProjectListCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
		},
		ApplyCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
	Example project update:

	# cloudctl project describe project1 -o yaml > project1.yaml
	# vi project1.yaml
	## either via stdin
	# cat project1.yaml | cloudctl project apply -f -
	## or via file
	# cloudctl project apply -f project1.yaml
	`)
		},
	}

	return genericcli.NewCmds(cmdsConfig, newMachineReservationsCmd(c))
}

func (c projectCmd) Get(id string) (*models.V1ProjectResponse, error) {
	resp, err := c.cloud.Project.FindProject(project.NewFindProjectParams().WithID(id), nil)
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

func (c projectCmd) Describe(id string) (*models.V1ProjectResponse, error) {
	request := project.NewFindProjectParams()
	request.SetID(id)
	resp, err := c.cloud.Project.FindProject(request, nil)
	return resp.Payload, err
}

func (c projectCmd) Delete(id string) (*models.V1ProjectResponse, error) {
	request := project.NewDeleteProjectParams().WithID(id)
	response, err := c.cloud.Project.DeleteProject(request, nil)
	return response.Payload, err
}

func (c projectCmd) List() ([]*models.V1ProjectResponse, error) {
	id := viper.GetString("id")
	name := viper.GetString("name")
	tenant := viper.GetString("tenant")
	if id != "" || name != "" || tenant != "" {
		pfr := project.NewFindProjectsParams().WithBody(&models.V1ProjectFindRequest{
			ID:       id,
			Name:     name,
			TenantID: tenant,
		})

		response, err := c.cloud.Project.FindProjects(pfr, nil)
		return response.Payload.Projects, err
	}

	request := project.NewListProjectsParams()
	response, err := c.cloud.Project.ListProjects(request, nil)
	return response.Payload.Projects, err
}

func (c *config) projectID(verb string, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("project %s requires projectID as argument", verb)
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", fmt.Errorf("project %s requires exactly one projectID as argument", verb)
}

func (c projectCmd) ApplyFromFile(from string) (genericcli.BulkResults[*models.V1ProjectResponse], error) {
	var pars []models.V1ProjectCreateRequest
	var par models.V1ProjectCreateRequest
	err := helper.ReadFrom(viper.GetString("file"), &par, func(data interface{}) {
		doc := data.(*models.V1ProjectCreateRequest)
		pars = append(pars, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		par = models.V1ProjectCreateRequest{}
	})
	if err != nil {
		return nil, err
	}
	var response genericcli.BulkResults[*models.V1ProjectResponse]
	for i, par := range pars {
		request := project.NewFindProjectParams()
		request.SetID(par.Meta.ID)
		p, err := c.cloud.Project.FindProject(request, nil)
		if err != nil {
			var r *project.FindProjectDefault
			if !errors.As(err, &r) {
				return response, err
			}
			if r.Code() != http.StatusNotFound {
				return response, err
			}
		}
		if p == nil || p.Payload == nil {
			params := project.NewCreateProjectParams()
			params.SetBody(&pars[i])
			resp, err := c.cloud.Project.CreateProject(params, nil)
			if err != nil {
				response = append(response, genericcli.BulkResult[*models.V1ProjectResponse]{
					Result: resp.Payload,
					Action: genericcli.BulkErrorOnCreate,
					Error:  err,
				})
				return response, err
			}
			response = append(response, genericcli.BulkResult[*models.V1ProjectResponse]{
				Result: resp.Payload,
				Action: genericcli.BulkCreated,
				Error:  nil,
			})
			continue
		}
		if p.Payload.Meta != nil {
			params := project.NewUpdateProjectParams()
			pur := &models.V1ProjectUpdateRequest{}
			if par.Description != "" {
				pur.Description = par.Description
			}
			if par.Name != "" {
				pur.Name = par.Name
			}
			if par.Quotas != nil {
				pur.Quotas = par.Quotas
			}
			if par.Meta != nil {
				pur.Meta = par.Meta
			}
			if par.TenantID != "" {
				pur.TenantID = par.TenantID
			}
			params.SetBody(pur)
			resp, err := c.cloud.Project.UpdateProject(params, nil)
			if err != nil {
				response = append(response, genericcli.BulkResult[*models.V1ProjectResponse]{
					Result: resp.Payload,
					Action: genericcli.BulkErrorOnUpdate,
					Error:  err,
				})
				return response, err
			}
			response = append(response, genericcli.BulkResult[*models.V1ProjectResponse]{
				Result: resp.Payload,
				Action: genericcli.BulkUpdated,
				Error:  nil,
			})
			continue
		}
	}
	return response, err
}

func (c projectCmd) Edit(args []string) (*models.V1ProjectResponse, error) {
	id, err := c.projectID("edit", args)
	if err != nil {
		return nil, err
	}

	getFunc := func(id string) ([]byte, error) {
		request := project.NewFindProjectParams()
		request.SetID(id)
		resp, err := c.cloud.Project.FindProject(request, nil)
		if err != nil {
			return nil, fmt.Errorf("project describe error:%w", err)
		}
		content, err := yaml.Marshal(resp.Payload)
		if err != nil {
			return nil, err
		}
		return content, nil
	}
	updateFunc := func(filename string) error {
		purs, err := readProjectUpdateRequests(filename)
		if err != nil {
			return err
		}
		if len(purs) != 1 {
			return fmt.Errorf("project update error more or less than one project given:%d", len(purs))
		}
		pup := project.NewUpdateProjectParams()
		pup.Body = &purs[0]
		uresp, err := c.cloud.Project.UpdateProject(pup, nil)
		if err != nil {
			return err
		}
		return c.listPrinter.Print(uresp.Payload)
	}

	return nil, helper.Edit(id, getFunc, updateFunc)
}

func readProjectUpdateRequests(filename string) ([]models.V1ProjectUpdateRequest, error) {
	var purs []models.V1ProjectUpdateRequest
	var pur models.V1ProjectUpdateRequest
	err := helper.ReadFrom(filename, &pur, func(data interface{}) {
		doc := data.(*models.V1ProjectUpdateRequest)
		purs = append(purs, *doc)
	})
	if err != nil {
		return purs, err
	}
	if len(purs) != 1 {
		return purs, fmt.Errorf("project update error more or less than one project given:%d", len(purs))
	}
	return purs, nil
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
