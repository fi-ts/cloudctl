package cmd

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	cloudgo "github.com/fi-ts/cloud-go"
	"github.com/fi-ts/cloud-go/api/client"
	output "github.com/fi-ts/cloudctl/cmd/output"
	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cfgFileType = "yaml"
	programName = "cloudctl"
)

var (
	ctx         api.Context
	cloud       *client.CloudAPI
	consoleHost string
	printer     output.Printer
	// will bind all viper flags to subcommands and
	// prevent overwrite of identical flag names from other commands
	// see https://github.com/spf13/viper/issues/233#issuecomment-386791444
	bindPFlags = func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			fmt.Printf("error during setup:%v", err)
			os.Exit(1)
		}
	}

	rootCmd = &cobra.Command{
		Use:   programName,
		Short: "a cli to manage cloud entities.",
		Long:  "with cloudctl you can manage kubernetes cluster, view networks et.al.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initPrinter()
		},
		SilenceUsage: true,
	}
)

// Execute is the entrypoint of the cloudctl application
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
	rootCmd.PersistentFlags().StringP("url", "u", "", "api server address. Can be specified with CLOUDCTL_URL environment variable.")
	rootCmd.PersistentFlags().String("apitoken", "", "api token to authenticate. Can be specified with CLOUDCTL_APITOKEN environment variable.")
	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to the kube-config to use for authentication and authorization. Is updated by login.")
	rootCmd.PersistentFlags().StringP("order", "", "", "order by (comma separated) column(s)")
	rootCmd.PersistentFlags().StringP("output-format", "o", "table", "output format (table|wide|markdown|json|yaml|template), wide is a table with more columns.")
	rootCmd.PersistentFlags().StringP("template", "", "", `output template for template output-format, go template format.
	For property names inspect the output of -o json or -o yaml for reference.
	Example for clusters:

	cloudctl cluster ls -o template --template "{{ .metadata.uid }}"

	`)
	rootCmd.PersistentFlags().BoolP("yes-i-really-mean-it", "", false, "skips security prompts (which can be dangerous to set blindly because actions can lead to data loss or additional costs)")

	rootCmd.AddCommand(clusterCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(tenantCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(s3Cmd)
	rootCmd.AddCommand(versionCmd)

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
		h, err := os.UserHomeDir()
		if err != nil {
			log.Printf("unable to figure out user home directory, skipping config lookup path: %v", err)
		} else {
			viper.AddConfigPath(fmt.Sprintf(h+"/.%s", programName))
		}
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			usedCfg := viper.ConfigFileUsed()
			if usedCfg != "" {
				log.Fatalf("config %s file unreadable:%v", usedCfg, err)
			}
		}
	}

	ctx = mustDefaultContext()
	driverURL := viper.GetString("url")
	if driverURL == "" && ctx.ApiURL != "" {
		driverURL = ctx.ApiURL
	}
	hmac := viper.GetString("hmac")
	if hmac == "" && ctx.HMAC != nil {
		hmac = *ctx.HMAC
	}
	apiToken := viper.GetString("apitoken")

	// if there is no api token explicitly specified we try to pull it out of
	// the kubeconfig context
	if apiToken == "" {
		authContext, err := getAuthContext(viper.GetString("kubeConfig"))
		// if there is an error, no kubeconfig exists for us ... this is not really an error
		// if cloudctl is used in scripting with an hmac-key
		if err == nil {
			apiToken = authContext.IDToken
		}
	}

	var err error
	cloud, err = cloudgo.NewClient(driverURL, apiToken, hmac)
	if err != nil {
		log.Fatalf("error initializing cloud-api client: %v", err)
	}

	parsedURL, err := url.Parse(driverURL)
	if err != nil {
		log.Fatalf("could not parse driver url: %v", err)
	}
	consoleHost = parsedURL.Host
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
