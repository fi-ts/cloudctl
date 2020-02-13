package cmd

import (
	"fmt"
	"log"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"

	"git.f-i-ts.de/cloud-native/cloudctl/api/client/project"
	"git.f-i-ts.de/cloud-native/cloudctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	projectCmd = &cobra.Command{
		Use:   "project",
		Short: "manage projects",
		Long:  "TODO",
	}
	projectCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectCreate()
		},
		PreRun: bindPFlags,
	}
	projectDeleteCmd = &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm", "delete"},
		Short:   "delete a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectDelete(args)
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
	err := projectCreateCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}

	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)
}

func projectCreate() error {
	name := viper.GetString("name")
	desc := viper.GetString("description")

	p := &models.V1Project{
		Name:        name,
		Description: desc,
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

	return printer.Print(response.Payload)
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

	return printer.Print(response.Payload)
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
	return printer.Print(response.Payload)
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
