package cmd

import (
	"git.f-i-ts.de/cloud-native/cloudctl/api/client/project"
	output "git.f-i-ts.de/cloud-native/cloudctl/cmd/output"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
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
	projectCreateCmd.MarkFlagRequired("name")

	projectListCmd.Flags().String("tenant", "", "show projects of given tenant")
	projectListCmd.Flags().Bool("all", false, "show all projects")

	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectListCmd)
}

func projectCreate() error {
	name := viper.GetString("name")
	desc := viper.GetString("description")

	pcr := models.ModelsV1ProjectCreateRequest{
		Name:        name,
		Description: desc,
	}
	request := project.NewCreateProjectParams()
	request.SetBody(&pcr)

	response, err := cloud.Project.CreateProject(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *project.CreateProjectConflict:
			output.PrintHTTPError(e.Payload)
		case *project.CreateProjectDefault:
			output.PrintHTTPError(e.Payload)
		default:
			output.PrintUnconventionalError(err)
		}
	}

	return printer.Print(response.Payload)
}
func projectList() error {
	tenant := viper.GetString("tenant")
	all := viper.GetBool("all")
	var pfr *models.V1ProjectFindRequest
	if tenant != "" {
		pfr = &models.V1ProjectFindRequest{
			Tenant: &tenant,
		}
	}
	if all {
		pfr = &models.V1ProjectFindRequest{
			All: &all,
		}
	}
	if pfr != nil {
		fpp := project.NewFindProjectsParams()
		fpp.SetBody(pfr)
		response, err := cloud.Project.FindProjects(fpp, cloud.Auth)
		if err != nil {
			switch e := err.(type) {
			case *project.ListProjectsDefault:
				output.PrintHTTPError(e.Payload)
			default:
				output.PrintUnconventionalError(err)
			}
		}
		return printer.Print(response.Payload)
	}

	request := project.NewListProjectsParams()
	response, err := cloud.Project.ListProjects(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *project.ListProjectsDefault:
			output.PrintHTTPError(e.Payload)
		default:
			output.PrintUnconventionalError(err)
		}
	}
	return printer.Print(response.Payload)
}
