package cmd

import (
	"log"
	"os"

	"git.f-i-ts.de/cloud-native/cloudctl/api/client/cluster"
	"git.f-i-ts.de/cloud-native/cloudctl/api/client/project"
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
	var names []string
	for p := range sc.Payload.PartitionConstraints {
		names = append(names, p)
	}
	return names, cobra.ShellCompDirectiveDefault
}

func networkListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, pc := range sc.Payload.PartitionConstraints {
		names = append(names, pc.Networks...)
	}
	return names, cobra.ShellCompDirectiveDefault
}
func versionListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, v := range sc.Payload.KubernetesVersions {
		names = append(names, v)
	}
	return names, cobra.ShellCompDirectiveDefault
}

func machineTypeListCompletion() ([]string, cobra.ShellCompDirective) {
	request := cluster.NewListConstraintsParams()
	sc, err := cloud.Cluster.ListConstraints(request, cloud.Auth)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, t := range sc.Payload.MachineTypes {
		names = append(names, t)
	}
	return names, cobra.ShellCompDirectiveDefault
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
