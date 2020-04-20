package cmd

import (
	"fmt"
	"io/ioutil"

	"git.f-i-ts.de/cloud-native/cloudctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	contextCmd = &cobra.Command{
		Use:     "context <name>",
		Aliases: []string{"ctx"},
		Short:   "manage cloudctl context",
		Long:    "context defines the backend to which cloudctl talks to.",
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
)

func contextSet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no context name given")
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
	ctxs.CurrentContext = defaultCtxName
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
	c, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read config, please create a config.yaml in either: /etc/cloudctl/, $HOME/.cloudctl/ or in the current directory, see cloudctl ctx -h for examples")
	}
	err = yaml.Unmarshal(c, &ctxs)
	return &ctxs, err
}

func writeContexts(ctxs *api.Contexts) error {
	cfgFile := viper.GetViper().ConfigFileUsed()
	fmt.Printf("update config:%s\n", cfgFile)
	c, err := yaml.Marshal(ctxs)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cfgFile, c, 0644)
}
