package completion

import (
	"sort"

	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/spf13/cobra"
)

func (c *Completion) MachineReservationListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		resp, err := c.cloud.Project.ListMachineReservations(project.NewListMachineReservationsParams().WithBody(&models.V1MachineReservationFindRequest{}), nil)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var projects []string

		for _, rv := range resp.Payload {
			if rv.Projectid == nil {
				continue
			}

			projects = append(projects, *rv.Projectid)
		}

		sort.Strings(projects)

		return projects, cobra.ShellCompDirectiveNoFileComp
	}

	p := args[0]

	resp, err := c.cloud.Project.ListMachineReservations(project.NewListMachineReservationsParams().WithBody(&models.V1MachineReservationFindRequest{
		Projectid: &p,
	}), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var sizes []string

	for _, rv := range resp.Payload {
		if rv.Projectid == nil || rv.Sizeid == nil {
			continue
		}

		if *rv.Projectid != p {
			continue
		}

		sizes = append(sizes, *rv.Sizeid)
	}

	sort.Strings(sizes)

	return sizes, cobra.ShellCompDirectiveNoFileComp
}
