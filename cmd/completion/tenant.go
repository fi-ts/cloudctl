package completion

import (
	"sort"

	"github.com/fi-ts/cloud-go/api/client/tenant"
	"github.com/spf13/cobra"
)

func (c *Completion) TenantListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := tenant.NewListTenantsParams()
	ts, err := c.cloud.Tenant.ListTenants(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var names []string
	for _, t := range ts.Payload {
		name := t.Meta.ID + "\t" + t.Name
		names = append(names, name)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}
