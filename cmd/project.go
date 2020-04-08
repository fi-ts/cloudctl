package cmd

import (
	"fmt"
	"log"
	"strings"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
	"gopkg.in/yaml.v3"

	"git.f-i-ts.de/cloud-native/cloudctl/api/client/project"
	"git.f-i-ts.de/cloud-native/cloudctl/cmd/helper"
	"git.f-i-ts.de/cloud-native/cloudctl/cmd/output"
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
	projectCreateCmd.Flags().String("name", "", "name of the cluster, max 10 characters. [required]")
	projectCreateCmd.Flags().String("description", "", "description of the cluster. [required]")
	projectCreateCmd.Flags().String("tenant", "", "create project for given tenant")
	projectCreateCmd.Flags().StringSlice("label", nil, "add initial label")
	projectCreateCmd.Flags().StringSlice("annotation", nil, "add initial annotation, must be in the form of key=value")
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

	p := &models.V1Project{
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
	pcr := models.V1ProjectCreateRequest{
		Project: p,
	}

	request := project.NewCreateProjectParams()
	request.SetBody(&pcr)

	response, err := cloud.Project.CreateProject(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *project.CreateProjectConflict:
			return output.HTTPError(e.Payload)
		case *project.CreateProjectDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	return printer.Print(response.Payload.Project)
}

func projectDescribe(args []string) error {
	id, err := projectID("describe", args)
	if err != nil {
		return err
	}

	request := project.NewFindProjectParams()
	request.SetID(id)
	p, err := cloud.Project.FindProject(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *project.FindProjectDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	return printer.Print(p.Payload.Project)
}

func projectDelete(args []string) error {
	id, err := projectID("delete", args)
	if err != nil {
		return err
	}

	request := project.NewDeleteProjectParams().WithID(id)

	response, err := cloud.Project.DeleteProject(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *project.DeleteProjectDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	return printer.Print(response.Payload.Project)
}

func projectList() error {
	request := project.NewListProjectsParams()
	response, err := cloud.Project.ListProjects(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *project.ListProjectsDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
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
	var pars []models.V1Project
	var par models.V1Project
	err := helper.ReadFrom(viper.GetString("file"), &par, func(data interface{}) {
		doc := data.(*models.V1Project)
		pars = append(pars, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		par = models.V1Project{}
	})
	if err != nil {
		return err
	}
	var response []*models.V1Project
	for _, par := range pars {
		request := project.NewFindProjectParams()
		request.SetID(par.Meta.ID)
		p, err := cloud.Project.FindProject(request, cloud.Auth)
		if err != nil {
			switch e := err.(type) {
			case *project.FindProjectDefault:
				return output.HTTPError(e.Payload)
			default:
				return output.UnconventionalError(err)
			}
		}
		if p.Payload.Project == nil {
			params := project.NewCreateProjectParams()
			params.SetBody(&models.V1ProjectCreateRequest{Project: &par})
			resp, err := cloud.Project.CreateProject(params, cloud.Auth)
			if err != nil {
				switch e := err.(type) {
				case *project.CreateProjectDefault:
					return output.HTTPError(e.Payload)
				case *project.CreateProjectConflict:
					return output.HTTPError(e.Payload)
				default:
					return output.UnconventionalError(err)
				}
			}
			response = append(response, resp.Payload.Project)
			continue
		}
		if p.Payload.Project.Meta != nil {
			params := project.NewUpdateProjectParams()
			params.SetBody(&models.V1ProjectUpdateRequest{Project: &par})
			resp, err := cloud.Project.UpdateProject(params, cloud.Auth)
			if err != nil {
				switch e := err.(type) {
				case *project.UpdateProjectPreconditionFailed:
					return output.HTTPError(e.Payload)
				default:
					return output.UnconventionalError(err)
				}
			}
			response = append(response, resp.Payload.Project)
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
		resp, err := cloud.Project.FindProject(request, cloud.Auth)
		if err != nil {
			return nil, fmt.Errorf("project describe error:%v", err)
		}
		content, err := yaml.Marshal(resp.Payload.Project)
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
		pup.Body = &models.V1ProjectUpdateRequest{Project: &purs[0]}
		uresp, err := cloud.Project.UpdateProject(pup, cloud.Auth)
		if err != nil {
			switch e := err.(type) {
			case *project.UpdateProjectPreconditionFailed:
				return output.HTTPError(e.Payload)
			default:
				return output.UnconventionalError(err)
			}
		}
		return printer.Print(uresp.Payload.Project)
	}

	return helper.Edit(id, getFunc, updateFunc)
}

func readProjectUpdateRequests(filename string) ([]models.V1Project, error) {
	var pcrs []models.V1Project
	var pcr models.V1Project
	err := helper.ReadFrom(filename, &pcr, func(data interface{}) {
		doc := data.(*models.V1Project)
		pcrs = append(pcrs, *doc)
	})
	if err != nil {
		return pcrs, err
	}
	if len(pcrs) != 1 {
		return pcrs, fmt.Errorf("project update error more or less than one project given:%d", len(pcrs))
	}
	return pcrs, nil
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
