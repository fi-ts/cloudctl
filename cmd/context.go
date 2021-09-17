package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	contextCmd = &cobra.Command{
		Use:     "context <name>",
		Aliases: []string{"ctx"},
		Short:   "manage cloudctl context",
		Long:    "context defines the backend to which cloudctl talks to. You can switch back and forth with \"-\"",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return contextListCompletion()
		},
		Example: `
~/.cloudctl/config.yaml
---
current: prod
contexts:
  prod:
    url: https://api.metal-stack.io/cloud
    issuer_url: https://dex.metal-stack.io/dex
    client_id: metal_client
    client_secret: 456
  dev:
    url: https://api.metal-stack.dev/cloud
    issuer_url: https://dex.metal-stack.dev/dex
    client_id: metal_client
    client_secret: 123
...
`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) == 1 {
				return contextSet(args)
			}
			if len(args) == 0 {
				return contextList()
			}
			return nil
		},
		PreRun: bindPFlags,
	}

	defaultCtx = api.Context{
		ApiURL:    "http://localhost:8080/cloud",
		IssuerURL: "http://localhost:8080/",
	}
	contextShortCmd = &cobra.Command{
		Use:   "short",
		Short: "only show the default context name",
		RunE: func(cmd *cobra.Command, args []string) error {
			return contextShort()
		},
		PreRun: bindPFlags,
	}
)

func init() {
	contextCmd.AddCommand(contextShortCmd)
}

func contextShort() error {
	ctxs, err := getContexts()
	if err != nil {
		return err
	}
	fmt.Println(ctxs.CurrentContext)
	return nil
}

func contextSet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no context name given")
	}
	if args[0] == "-" {
		return previous()
	}
	ctxs, err := getContexts()
	if err != nil {
		return err
	}
	defaultCtxName := args[0]
	_, ok := ctxs.Contexts[defaultCtxName]
	if !ok {
		return fmt.Errorf("context %s not found", defaultCtxName)
	}
	if defaultCtxName == ctxs.CurrentContext {
		fmt.Printf("%s context \"%s\" already active\n", color.GreenString("✔"), color.GreenString(ctxs.CurrentContext))
		return nil
	}
	ctxs.PreviousContext = ctxs.CurrentContext
	ctxs.CurrentContext = defaultCtxName
	return writeContexts(ctxs)
}

func previous() error {
	ctxs, err := getContexts()
	if err != nil {
		return err
	}
	prev := ctxs.PreviousContext
	if prev == "" {
		return fmt.Errorf("no previous context found")
	}
	curr := ctxs.CurrentContext
	ctxs.PreviousContext = curr
	ctxs.CurrentContext = prev
	return writeContexts(ctxs)
}

func contextList() error {
	ctxs, err := getContexts()
	if err != nil {
		return err
	}
	return printer.Print(ctxs)
}

func mustDefaultContext() api.Context {
	ctxs, err := getContexts()
	if err != nil {
		return defaultCtx
	}
	ctx, ok := ctxs.Contexts[ctxs.CurrentContext]
	if !ok {
		return defaultCtx
	}
	return ctx
}

func getContexts() (*api.Contexts, error) {
	var ctxs api.Contexts
	cfgFile := viper.GetViper().ConfigFileUsed()
	c, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read config, please create a config.yaml in either: /etc/cloudctl/, $HOME/.cloudctl/ or in the current directory, see cloudctl ctx -h for examples")
	}
	err = yaml.Unmarshal(c, &ctxs)
	return &ctxs, err
}

func writeContexts(ctxs *api.Contexts) error {
	c, err := yaml.Marshal(ctxs)
	if err != nil {
		return err
	}
	cfgFile := viper.GetViper().ConfigFileUsed()
	err = os.WriteFile(cfgFile, c, 0600)
	if err != nil {
		return err
	}
	fmt.Printf("%s switched context to \"%s\"\n", color.GreenString("✔"), color.GreenString(ctxs.CurrentContext))
	return nil
}
