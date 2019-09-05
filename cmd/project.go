package cmd

import (
	metalgo "github.com/metal-pod/metal-go"
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
	projectCreateCmd.Flags().StringP("name", "", "", "name of the cluster, max 10 characters. [required]")
	projectCreateCmd.Flags().StringP("description", "", "", "description of the cluster. [required]")

	projectCreateCmd.MarkFlagRequired("name")

	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectListCmd)
}

func projectCreate() error {
	name := viper.GetString("name")
	desc := viper.GetString("description")
	tenant := "TODO from oidc token"

	pcr := metalgo.ProjectCreateRequest{
		Name:        name,
		Description: desc,
		Tenant:      tenant,
	}
	response, err := metal.ProjectCreate(pcr)
	if err != nil {
		return err
	}

	return printer.Print(response.Project)
}
func projectList() error {
	response, err := metal.ProjectList()
	if err != nil {
		return err
	}
	return printer.Print(response.Project)
}
