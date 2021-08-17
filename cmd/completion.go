package cmd

import (
	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/client/database"
	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/fi-ts/cloud-go/api/client/s3"
	"github.com/spf13/cobra"
)

func contextListCompletion() ([]string, cobra.ShellCompDirective) {
	ctxs, err := getContexts()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for name := range ctxs.Contexts {
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func clusterListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListClustersParams()
	response, err := cloud.Cluster.ListClusters(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, c := range response.Payload {
		names = append(names, *c.ID+"\t"+*c.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func clusterMachineListCompletion(clusterIDs []string, includeMachines bool) ([]string, cobra.ShellCompDirective) {
	if len(clusterIDs) != 1 {
		return []string{"no clusterid given"}, cobra.ShellCompDirectiveNoFileComp
	}
	clusterID := clusterIDs[0]
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(clusterID)
	shoot, err := cloud.Cluster.FindCluster(findRequest, nil)
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

func projectListCompletion() ([]string, cobra.ShellCompDirective) {
	request := project.NewListProjectsParams()
	response, err := cloud.Project.ListProjects(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range response.Payload.Projects {
		names = append(names, p.Meta.ID+"\t"+p.TenantID+"/"+p.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func partitionListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return sc.Payload.Partitions, cobra.ShellCompDirectiveNoFileComp
}

func networkListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return sc.Payload.Networks, cobra.ShellCompDirectiveNoFileComp
}
func versionListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return sc.Payload.KubernetesVersions, cobra.ShellCompDirectiveNoFileComp
}

func machineTypeListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return sc.Payload.MachineTypes, cobra.ShellCompDirectiveNoFileComp
}

func machineImageListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, t := range sc.Payload.MachineImages {
		name := *t.Name + "-" + *t.Version
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func firewallTypeListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return sc.Payload.FirewallTypes, cobra.ShellCompDirectiveNoFileComp
}

func firewallImageListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return sc.Payload.FirewallImages, cobra.ShellCompDirectiveNoFileComp
}

func firewallControllerVersionListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, nil)
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
	return fwcvs, cobra.ShellCompDirectiveNoFileComp
}

func s3ListPartitionsCompletion() ([]string, cobra.ShellCompDirective) {
	request := s3.NewLists3partitionsParams()
	response, err := cloud.S3.Lists3partitions(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range response.Payload {
		names = append(names, *p.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func postgresListPartitionsCompletion() ([]string, cobra.ShellCompDirective) {
	request := database.NewGetPostgresPartitionsParams()
	response, err := cloud.Database.GetPostgresPartitions(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for name := range response.Payload {
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func postgresListVersionsCompletion() ([]string, cobra.ShellCompDirective) {
	request := database.NewGetPostgresVersionsParams()
	response, err := cloud.Database.GetPostgresVersions(request, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, v := range response.Payload {
		names = append(names, v.Version)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
