package completion

import (
	"sort"

	"github.com/fi-ts/cloud-go/api/client/database"
	"github.com/spf13/cobra"
)

func (c *Completion) PostgresListPartitionsCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := database.NewGetPostgresPartitionsParams()
	response, err := c.cloud.Database.GetPostgresPartitions(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for name := range response.Payload {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) PostgresListVersionsCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := database.NewGetPostgresVersionsParams()
	response, err := c.cloud.Database.GetPostgresVersions(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, v := range response.Payload {
		names = append(names, v.Version)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}
func (c *Completion) PostgresListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := database.NewListPostgresParams()
	response, err := c.cloud.Database.ListPostgres(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range response.Payload {
		names = append(names, *p.ID+"\t"+p.Description)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}
