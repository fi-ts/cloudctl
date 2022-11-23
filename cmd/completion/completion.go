package completion

import (
	"sort"

	"github.com/fi-ts/cloud-go/api/client"
	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/spf13/cobra"
)

var (
	ClusterPurposes = []string{"production", "development", "evaluation", "infrastructure"}
)

type Completion struct {
	cloud *client.CloudAPI
}

func NewCompletion(cloud *client.CloudAPI) *Completion {
	return &Completion{
		cloud: cloud,
	}
}

func (c *Completion) SetClient(cloud *client.CloudAPI) {
	c.cloud = cloud
}

func OutputFormatListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"table", "wide", "markdown", "json", "yaml", "template"}, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) ContextListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctxs, err := api.GetContexts()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for name := range ctxs.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) ClusterListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListClustersParams()
	response, err := c.cloud.Cluster.ListClusters(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, c := range response.Payload {
		names = append(names, *c.ID+"\t"+*c.Name)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) ClusterNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListClustersParams()
	response, err := c.cloud.Cluster.ListClusters(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, c := range response.Payload {
		names = append(names, *c.Name+"\t"+*c.ID)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) ClusterMachineListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return c.clusterMachineListCompletion(args, true)
}

func (c *Completion) ClusterFirewallListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return c.clusterMachineListCompletion(args, false)
}

func (c *Completion) ClusterPurposeListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return ClusterPurposes, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) clusterMachineListCompletion(clusterIDs []string, includeMachines bool) ([]string, cobra.ShellCompDirective) {
	if len(clusterIDs) != 1 {
		return []string{"no clusterid given"}, cobra.ShellCompDirectiveNoFileComp
	}
	clusterID := clusterIDs[0]
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(clusterID)
	shoot, err := c.cloud.Cluster.FindCluster(findRequest, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var machines []string
	for _, m := range shoot.Payload.Firewalls {
		machines = append(machines, *m.ID+"\t"+*m.Allocation.Hostname)
	}
	if includeMachines {
		for _, m := range shoot.Payload.Machines {
			machines = append(machines, *m.ID+"\t"+*m.Allocation.Hostname)
		}
	}
	return machines, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) PartitionListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := c.cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	sort.Strings(sc.Payload.Partitions)
	return sc.Payload.Partitions, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) SeedListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := c.cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var names []string
	for _, seedNames := range sc.Payload.Seeds {
		names = append(names, seedNames...)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) NetworkListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := c.cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	sort.Strings(sc.Payload.Networks)
	return sc.Payload.Networks, cobra.ShellCompDirectiveNoFileComp
}
func (c *Completion) VersionListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := c.cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	sort.Strings(sc.Payload.KubernetesVersions)
	return sc.Payload.KubernetesVersions, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) MachineTypeListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := c.cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	sort.Strings(sc.Payload.MachineTypes)
	return sc.Payload.MachineTypes, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) MachineImageListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := c.cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, t := range sc.Payload.MachineImages {
		name := *t.Name + "-" + *t.Version
		names = append(names, name)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) FirewallTypeListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := c.cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	sort.Strings(sc.Payload.FirewallTypes)
	return sc.Payload.FirewallTypes, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) FirewallImageListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := c.cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	sort.Strings(sc.Payload.FirewallImages)
	return sc.Payload.FirewallImages, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) FirewallControllerVersionListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := c.cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	fwcvs := []string{"auto"}
	for _, v := range sc.Payload.FirewallControllerVersions {
		if v.Version == nil {
			continue
		}
		fwcvs = append(fwcvs, *v.Version)
	}
	sort.Strings(fwcvs)
	return fwcvs, cobra.ShellCompDirectiveNoFileComp
}
