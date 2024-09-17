package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"gopkg.in/yaml.v3"

	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newProjectCmd(c *config) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "manage projects",
		Long:  "a project organizes cloud resources regarding tenancy, quotas, billing and authentication",
	}
	projectCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.projectCreate()
		},
	}
	projectDescribeCmd := &cobra.Command{
		Use:   "describe <projectID>",
		Short: "describe a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.projectDescribe(args)
		},
		ValidArgsFunction: c.comp.ProjectListCompletion,
	}
	projectDeleteCmd := &cobra.Command{
		Use:     "delete <projectID>",
		Aliases: []string{"destroy", "rm", "remove"},
		Short:   "delete a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.projectDelete(args)
		},
		ValidArgsFunction: c.comp.ProjectListCompletion,
	}
	projectApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.projectApply()
		},
	}
	projectEditCmd := &cobra.Command{
		Use:   "edit <projectID>",
		Short: "edit a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.projectEdit(args)
		},
		ValidArgsFunction: c.comp.ProjectListCompletion,
	}
	projectListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list projects",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.projectList()
		},
	}

	projectCreateCmd.Flags().String("name", "", "name of the project, max 10 characters. [required]")
	projectCreateCmd.Flags().String("description", "", "description of the project. [required]")
	projectCreateCmd.Flags().String("tenant", "", "create project for given tenant")
	projectCreateCmd.Flags().StringSlice("label", nil, "add initial label, can be given multiple times to add multiple labels, e.g. --label=foo --label=bar")
	projectCreateCmd.Flags().StringSlice("annotation", nil, "add initial annotation, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar")
	projectCreateCmd.Flags().Int32("cluster-quota", 0, "cluster quota")
	projectCreateCmd.Flags().Int32("machine-quota", 0, "machine quota")
	projectCreateCmd.Flags().Int32("ip-quota", 0, "ip quota")
	genericcli.Must(projectCreateCmd.MarkFlagRequired("name"))
	genericcli.Must(projectCreateCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))

	projectListCmd.Flags().String("id", "", "show projects of given id")
	projectListCmd.Flags().String("name", "", "show projects of given name")
	projectListCmd.Flags().String("tenant", "", "show projects of given tenant")
	genericcli.Must(projectListCmd.RegisterFlagCompletionFunc("id", c.comp.ProjectListCompletion))
	genericcli.Must(projectListCmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))

	projectApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
	Example project update:

	# cloudctl project describe project1 -o yaml > project1.yaml
	# vi project1.yaml
	## either via stdin
	# cat project1.yaml | cloudctl project apply -f -
	## or via file
	# cloudctl project apply -f project1.yaml
	`)

	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDescribeCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectApplyCmd)
	projectCmd.AddCommand(projectEditCmd)
	projectCmd.AddCommand(newMachineReservationsCmd(c))

	return projectCmd
}

func (c *config) projectCreate() error {
	tenant := viper.GetString("tenant")
	name := viper.GetString("name")
	desc := viper.GetString("description")
	labels := viper.GetStringSlice("label")
	as := viper.GetStringSlice("annotation")
	var (
		clusterQuota, machineQuota, ipQuota *models.V1Quota
	)
	if viper.IsSet("cluster-quota") {
		clusterQuota = &models.V1Quota{Quota: viper.GetInt32("cluster-quota")}
	}
	if viper.IsSet("machine-quota") {
		machineQuota = &models.V1Quota{Quota: viper.GetInt32("machine-quota")}
	}
	if viper.IsSet("ip-quota") {
		ipQuota = &models.V1Quota{Quota: viper.GetInt32("ip-quota")}
	}

	annotations, err := annotationsAsMap(as)
	if err != nil {
		return err
	}

	pcr := &models.V1ProjectCreateRequest{
		Name:        name,
		Description: desc,
		TenantID:    tenant,
		Quotas: &models.V1QuotaSet{
			Cluster: clusterQuota,
			Machine: machineQuota,
			IP:      ipQuota,
		},
		Meta: &models.V1Meta{
			Kind:        "Project",
			Apiversion:  "v1",
			Annotations: annotations,
			Labels:      labels,
		},
	}

	request := project.NewCreateProjectParams()
	request.SetBody(pcr)

	response, err := c.cloud.Project.CreateProject(request, nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(response.Payload)
}

func (c *config) projectDescribe(args []string) error {
	id, err := c.projectID("describe", args)
	if err != nil {
		return err
	}

	request := project.NewFindProjectParams()
	request.SetID(id)
	p, err := c.cloud.Project.FindProject(request, nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(p.Payload)
}

func (c *config) projectDelete(args []string) error {
	id, err := c.projectID("delete", args)
	if err != nil {
		return err
	}

	request := project.NewDeleteProjectParams().WithID(id)

	response, err := c.cloud.Project.DeleteProject(request, nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(response.Payload)
}

func (c *config) projectList() error {
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
		if err != nil {
			return err
		}

		return c.listPrinter.Print(response.Payload.Projects)
	}

	request := project.NewListProjectsParams()
	response, err := c.cloud.Project.ListProjects(request, nil)
	if err != nil {
		return err
	}
	return c.listPrinter.Print(response.Payload.Projects)
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

func (c *config) projectApply() error {
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
		return err
	}
	var response []*models.V1ProjectResponse
	for i, par := range pars {
		request := project.NewFindProjectParams()
		request.SetID(par.Meta.ID)
		p, err := c.cloud.Project.FindProject(request, nil)
		if err != nil {
			var r *project.FindProjectDefault
			if !errors.As(err, &r) {
				return err
			}
			if r.Code() != http.StatusNotFound {
				return err
			}
		}
		if p == nil || p.Payload == nil {
			params := project.NewCreateProjectParams()
			params.SetBody(&pars[i])
			resp, err := c.cloud.Project.CreateProject(params, nil)
			if err != nil {
				return err
			}
			response = append(response, resp.Payload)
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
				return err
			}
			response = append(response, resp.Payload)
			continue
		}
	}
	return c.listPrinter.Print(response)
}

func (c *config) projectEdit(args []string) error {
	id, err := c.projectID("edit", args)
	if err != nil {
		return err
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

	return helper.Edit(id, getFunc, updateFunc)
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

func annotationsAsMap(annotations []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, a := range annotations {
		parts := strings.Split(strings.TrimSpace(a), "=")
		if len(parts) != 2 {
			return result, fmt.Errorf("given annotation %s does not contain exactly one =", a)
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}
