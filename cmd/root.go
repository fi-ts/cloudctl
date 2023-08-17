package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"

	cloudgo "github.com/fi-ts/cloud-go"
	"github.com/fi-ts/cloud-go/api/client"
	"github.com/fi-ts/cloudctl/cmd/completion"
	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
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
)

func newRootCmd() *cobra.Command {
	name := "cloudctl"
	rootCmd := &cobra.Command{
		Use:          name,
		Short:        "a cli to manage cloud entities.",
		Long:         "with cloudctl you can manage kubernetes cluster, view networks et.al.",
		SilenceUsage: true,
	}

	rootCmd.PersistentFlags().StringP("url", "u", "", "api server address. Can be specified with CLOUDCTL_URL environment variable.")
	rootCmd.PersistentFlags().String("apitoken", "", "api token to authenticate. Can be specified with CLOUDCTL_APITOKEN environment variable.")
	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to the kube-config to use for authentication and authorization. Is updated by login. Uses default path if not specified.")
	rootCmd.PersistentFlags().StringP("order", "", "", "order by (comma separated) column(s)")
	rootCmd.PersistentFlags().BoolP("no-headers", "", false, "ommit headers in tables")
	rootCmd.PersistentFlags().BoolP("debug", "", false, "enable debug")
	rootCmd.PersistentFlags().StringP("output-format", "o", "table", "output format (table|wide|markdown|json|yaml|template), wide is a table with more columns.")
	rootCmd.PersistentFlags().StringP("template", "", "", `output template for template output-format, go template format.
	For property names inspect the output of -o json for reference.
	Example for clusters:

	cloudctl cluster ls -o template --template "{{ .ID }} {{ .Name }}"

	`)
	rootCmd.PersistentFlags().BoolP("yes-i-really-mean-it", "", false, "skips security prompts (which can be dangerous to set blindly because actions can lead to data loss or additional costs)")

	must(viper.BindPFlags(rootCmd.Flags()))
	must(viper.BindPFlags(rootCmd.PersistentFlags()))

	cfg := getConfig(name)

	rootCmd.AddCommand(newAuditCmd(cfg))
	rootCmd.AddCommand(newClusterCmd(cfg))
	rootCmd.AddCommand(newDashboardCmd(cfg))
	rootCmd.AddCommand(newUpdateCmd(name))
	rootCmd.AddCommand(newLoginCmd(cfg))
	rootCmd.AddCommand(newLogoutCmd(cfg))
	rootCmd.AddCommand(newWhoamiCmd())
	rootCmd.AddCommand(newProjectCmd(cfg))
	rootCmd.AddCommand(newTenantCmd(cfg))
	rootCmd.AddCommand(newContextCmd(cfg))
	rootCmd.AddCommand(newS3Cmd(cfg))
	rootCmd.AddCommand(newVersionCmd(cfg))
	rootCmd.AddCommand(newVolumeCmd(cfg))
	rootCmd.AddCommand(newPostgresCmd(cfg))
	rootCmd.AddCommand(newIPCmd(cfg))
	rootCmd.AddCommand(newBillingCmd(cfg))
	rootCmd.AddCommand(newHealthCmd(cfg))

	return rootCmd
}

// Execute is the entrypoint of the cloudctl application
func Execute() {
	cmd := newRootCmd()
	err := cmd.Execute()
	if err != nil {
		if viper.GetBool("debug") {
			panic(err)
		}
		os.Exit(1)
	}
}

type config struct {
	name        string
	cloud       *client.CloudAPI
	comp        *completion.Completion
	consoleHost string
	log         *slog.Logger
}

func getConfig(name string) *config {
	viper.SetEnvPrefix(strings.ToUpper(name))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetConfigType("yaml")
	cfgFile := viper.GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("config file path set explicitly, but unreadable:%v", err)
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(fmt.Sprintf("/etc/%s", name))
		h, err := os.UserHomeDir()
		if err != nil {
			log.Printf("unable to figure out user home directory, skipping config lookup path: %v", err)
		} else {
			viper.AddConfigPath(fmt.Sprintf(h+"/.%s", name))
		}
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			usedCfg := viper.ConfigFileUsed()
			if usedCfg != "" {
				log.Fatalf("config %s file unreadable:%v", usedCfg, err)
			}
		}
	}

	ctx := api.MustDefaultContext()

	opts := &slog.HandlerOptions{}
	if viper.GetBool("debug") {
		opts.Level = slog.LevelDebug
	}

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
		authContext, err := api.GetAuthContext(viper.GetString("kubeconfig"))
		// if there is an error, no kubeconfig exists for us ... this is not really an error
		// if cloudctl is used in scripting with an hmac-key
		if err == nil {
			apiToken = authContext.IDToken
		}
	}

	cloud, err := cloudgo.NewClient(driverURL, apiToken, hmac)
	if err != nil {
		log.Fatalf("error initializing cloud-api client: %v", err)
	}

	comp := completion.NewCompletion(cloud)

	parsedURL, err := url.Parse(driverURL)
	if err != nil {
		log.Fatalf("could not parse driver url: %v", err)
	}
	consoleHost := parsedURL.Host

	return &config{
		name:        name,
		cloud:       cloud,
		comp:        comp,
		consoleHost: consoleHost,
		log:         slog.New(slog.NewJSONHandler(os.Stdout, opts)),
	}
}
