package cmd

import (
	"log"
	"os"

	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/fi-ts/cloud-go/api/client/s3"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion scripts",
	Long: `To load completion run

. <(cloudctl completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(cloudctl completion)
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := rootCmd.GenBashCompletion(os.Stdout)
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

var zshCompletionCmd = &cobra.Command{
	Use:   "zsh-completion",
	Short: "Generates Z shell completion scripts",
	Long: `To load completion run

. <(cloudctl zsh-completion)

To configure your Z shell (with oh-my-zshell framework) to load completions for each session run

echo -e '#compdef _cloudctl cloudctl\n. <(cloudctl zsh-completion)' > $ZSH/completions/_cloudctl
rm -f ~/.zcompdump*
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := rootCmd.GenZshCompletion(os.Stdout)
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

func contextListCompletion() ([]string, cobra.ShellCompDirective) {
	ctxs, err := getContexts()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for name := range ctxs.Contexts {
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveDefault
}

func clusterListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListClustersParams()
	response, err := cloud.Cluster.ListClusters(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, c := range response.Payload {
		names = append(names, *c.ID)
	}
	return names, cobra.ShellCompDirectiveDefault
}

func clusterMachineListCompletion(clusterID string) ([]string, cobra.ShellCompDirective) {
	findRequest := cluster.NewFindClusterParams()
	findRequest.SetID(clusterID)
	shoot, err := cloud.Cluster.FindCluster(findRequest, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var machines []string
	for _, m := range shoot.Payload.Machines {
		machines = append(machines, *m.ID)
	}
	return machines, cobra.ShellCompDirectiveDefault
}

func projectListCompletion() ([]string, cobra.ShellCompDirective) {
	request := project.NewListProjectsParams()
	response, err := cloud.Project.ListProjects(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range response.Payload.Projects {
		names = append(names, p.Meta.ID)
	}
	return names, cobra.ShellCompDirectiveDefault
}

func partitionListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return sc.Payload.Partitions, cobra.ShellCompDirectiveDefault
}

func networkListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return sc.Payload.Networks, cobra.ShellCompDirectiveDefault
}
func versionListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return sc.Payload.KubernetesVersions, cobra.ShellCompDirectiveDefault
}

func machineTypeListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return sc.Payload.MachineTypes, cobra.ShellCompDirectiveDefault
}

func machineImageListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, t := range sc.Payload.MachineImages {
		name := *t.Name + "-" + *t.Version
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveDefault
}

func firewallTypeListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return sc.Payload.FirewallTypes, cobra.ShellCompDirectiveDefault
}

func firewallImageListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return sc.Payload.FirewallImages, cobra.ShellCompDirectiveDefault
}

func s3ListPartitionsCompletion() ([]string, cobra.ShellCompDirective) {
	request := s3.NewLists3partitionsParams()
	response, err := cloud.S3.Lists3partitions(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range response.Payload {
		names = append(names, *p.ID)
	}
	return names, cobra.ShellCompDirectiveDefault
}
