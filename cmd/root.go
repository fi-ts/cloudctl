package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	output "git.f-i-ts.de/cloud-native/cloudctl/cmd/output"
	g "git.f-i-ts.de/cloud-native/cloudctl/pkg/gardener"
	m "git.f-i-ts.de/cloud-native/cloudctl/pkg/metal"
	"git.f-i-ts.de/cloud-native/metallib/auth"
	"github.com/metal-pod/v"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cfgFileType = "yaml"
	programName = "cloudctl"
)

var (
	kubeconfig string
	gardener   *g.Gardener
	metal      *m.Metal
	printer    output.Printer
	// will bind all viper flags to subcommands and
	// prevent overwrite of identical flag names from other commands
	// see https://github.com/spf13/viper/issues/233#issuecomment-386791444
	bindPFlags = func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	}

	rootCmd = &cobra.Command{
		Use:     programName,
		Aliases: []string{"m"},
		Short:   "a cli to manage cloud entities.",
		Long:    "with cloudctl you can manage kubernetes cluster, view networks et.al.",
		Version: v.V.String(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initPrinter()
		},
		SilenceUsage: true,
	}
)

// Execute is the entrypoint of the metal-go application
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		if viper.GetBool("debug") {
			st := errors.WithStack(err)
			fmt.Printf("%+v", st)
		}
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringP("url", "u", "http://127.0.0.1:8080", "api server address. Can be specified with CLOUDCTL_URL environment variable.")
	rootCmd.PersistentFlags().String("apitoken", "", "api token to authenticate. Can be specified with CLOUDCTL_APITOKEN environment variable.")
	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to the kube-config to use for authentication and authorization. Is updated by login.")
	rootCmd.PersistentFlags().StringP("output-format", "o", "table", "output format (table|wide|markdown|json|yaml|template), wide is a table with more columns.")
	rootCmd.AddCommand(clusterCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(whoamiCmd)

	err := viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("error setup root cmd:%v", err)
	}
}

func initConfig() {
	viper.SetEnvPrefix(strings.ToUpper(programName))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetConfigType(cfgFileType)
	cfgFile := viper.GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("config file path set explicitly, but unreadable:%v", err)
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(fmt.Sprintf("/etc/%s", programName))
		viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", programName))
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			usedCfg := viper.ConfigFileUsed()
			if usedCfg != "" {
				log.Fatalf("config %s file unreadable:%v", usedCfg, err)
			}
		}
	}

	kubeconfig = viper.GetString("kubeconfig")

	driverURL := viper.GetString("url")
	apiToken := viper.GetString("apitoken")
	hmacKey := viper.GetString("hmac")

	// if there is no api token explicitly specified we try to pull it out of
	// the kubeconfig context
	if apiToken == "" {
		kubeconfig := viper.GetString("kubeconfig")
		authContext, err := auth.CurrentAuthContext(kubeconfig)
		// if there is an error, no kubeconfig exists for us ... this is not really an error
		// if metalctl is used in scripting with an hmac-key
		if err == nil {
			apiToken = authContext.IDToken
		}
	}

	var err error
	metal, err = m.New(driverURL, apiToken, hmacKey)
	if err != nil {
		log.Fatalf("error setup root cmd:%v", err)
	}
}

func initPrinter() {
	var err error
	printer, err = output.NewPrinter(
		viper.GetString("output-format"),
		viper.GetString("order"),
		viper.GetString("template"),
		viper.GetBool("no-headers"),
	)
	if err != nil {
		log.Fatalf("unable to initialize printer:%v", err)
	}
}
