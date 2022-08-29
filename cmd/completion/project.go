package completion

import (
	"sort"

	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/spf13/cobra"
)

func (c *Completion) ProjectListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := project.NewListProjectsParams()
	response, err := c.cloud.Project.ListProjects(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range response.Payload.Projects {
		names = append(names, p.Meta.ID+"\t"+p.TenantID+"/"+p.Name)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}
