package cmd

import (
	"log"
	"os"

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
