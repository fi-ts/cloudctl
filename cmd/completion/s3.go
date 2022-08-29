package completion

import (
	"sort"

	"github.com/fi-ts/cloud-go/api/client/s3"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/spf13/cobra"
)

func (c *Completion) S3ListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var partition *string
	if p, _ := cmd.Flags().GetString("partition"); p != "" {
		partition = pointer.Pointer(p)
	}

	request := s3.NewLists3Params().WithBody(&models.V1S3ListRequest{Partition: partition})
	response, err := c.cloud.S3.Lists3(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range response.Payload {
		// TODO: move to filtering api
		if project, _ := cmd.Flags().GetString("project"); project != *p.Project {
			continue
		}
		names = append(names, *p.ID)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) S3ListPartitionsCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := s3.NewLists3partitionsParams()
	response, err := c.cloud.S3.Lists3partitions(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range response.Payload {
		names = append(names, *p.ID)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}
