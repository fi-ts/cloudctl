package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/fi-ts/cloud-go/api/models"
	"gopkg.in/yaml.v3"

	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	projectCmd = &cobra.Command{
		Use:   "project",
		Short: "manage projects",
		Long:  "a project organizes cloud resources regarding tenancy, quotas, billing and authentication",
	}
	projectCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectCreate()
		},
		PreRun: bindPFlags,
	}
	projectDescribeCmd = &cobra.Command{
		Use:   "describe <projectID>",
		Short: "describe a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectDescribe(args)
		},
		PreRun: bindPFlags,
	}
	projectDeleteCmd = &cobra.Command{
		Use:     "remove <projectID>",
		Aliases: []string{"rm", "delete"},
		Short:   "delete a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectDelete(args)
		},
		PreRun: bindPFlags,
	}
	projectApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "create/update a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectApply()
		},
		PreRun: bindPFlags,
	}
	projectEditCmd = &cobra.Command{
		Use:   "edit <projectID>",
		Short: "edit a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectEdit(args)
		},
		PreRun: bindPFlags,
	}
	projectListCmd = &cobra.Command{
		Use:     "list",
		Short:   "list projects",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectList()
		},
		PreRun: bindPFlags,
	}
)

func init() {
	projectCreateCmd.Flags().String("name", "", "name of the project, max 10 characters. [required]")
	projectCreateCmd.Flags().String("description", "", "description of the project. [required]")
	projectCreateCmd.Flags().String("tenant", "", "create project for given tenant")
	projectCreateCmd.Flags().StringSlice("label", nil, "add initial label, can be given multiple times to add multiple labels, e.g. --label=foo --label=bar")
	projectCreateCmd.Flags().StringSlice("annotation", nil, "add initial annotation, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar")
	projectCreateCmd.Flags().Int32("cluster-quota", 0, "cluster quota")
	projectCreateCmd.Flags().Int32("machine-quota", 0, "machine quota")
	projectCreateCmd.Flags().Int32("ip-quota", 0, "ip quota")
	err := projectCreateCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}

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
}

func projectCreate() error {
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

	response, err := cloud.Project.CreateProject(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}

func projectDescribe(args []string) error {
	id, err := projectID("describe", args)
	if err != nil {
		return err
	}

	request := project.NewFindProjectParams()
	request.SetID(id)
	p, err := cloud.Project.FindProject(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(p.Payload)
}

func projectDelete(args []string) error {
	id, err := projectID("delete", args)
	if err != nil {
		return err
	}

	request := project.NewDeleteProjectParams().WithID(id)

	response, err := cloud.Project.DeleteProject(request, nil)
	if err != nil {
		return err
	}

	return printer.Print(response.Payload)
}

func projectList() error {
	request := project.NewListProjectsParams()
	response, err := cloud.Project.ListProjects(request, nil)
	if err != nil {
		return err
	}
	return printer.Print(response.Payload.Projects)
}

func projectID(verb string, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("project %s requires projectID as argument", verb)
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", fmt.Errorf("project %s requires exactly one projectID as argument", verb)
}

func projectApply() error {
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
	for _, par := range pars {
		request := project.NewFindProjectParams()
		request.SetID(par.Meta.ID)
		p, err := cloud.Project.FindProject(request, nil)
		if err != nil {
			return err
		}
		if p.Payload == nil {
			params := project.NewCreateProjectParams()
			params.SetBody(&par)
			resp, err := cloud.Project.CreateProject(params, nil)
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
			resp, err := cloud.Project.UpdateProject(params, nil)
			if err != nil {
				return err
			}
			response = append(response, resp.Payload)
			continue
		}
	}
	return printer.Print(response)
}

func projectEdit(args []string) error {
	id, err := projectID("edit", args)
	if err != nil {
		return err
	}

	getFunc := func(id string) ([]byte, error) {
		request := project.NewFindProjectParams()
		request.SetID(id)
		resp, err := cloud.Project.FindProject(request, nil)
		if err != nil {
			return nil, fmt.Errorf("project describe error:%v", err)
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
		uresp, err := cloud.Project.UpdateProject(pup, nil)
		if err != nil {
			return err
		}
		return printer.Print(uresp.Payload)
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
