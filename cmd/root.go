package cmd

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"

	cloudgo "github.com/fi-ts/cloud-go"
	"github.com/fi-ts/cloud-go/api/client"
	"github.com/fi-ts/cloudctl/cmd/completion"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	binaryName = "cloudctl"
)

type config struct {
	fs          afero.Fs
	out         io.Writer
	cloud       *client.CloudAPI
	comp        *completion.Completion
	consoleHost string
	log         *slog.Logger
}

func newRootCmd(cfg *config) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          binaryName,
		Short:        "a cli to manage cloud entities.",
		Long:         "with cloudctl you can manage kubernetes cluster, view networks et.al.",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.SetFs(cfg.fs)
			genericcli.Must(viper.BindPFlags(cmd.Flags()))
			genericcli.Must(viper.BindPFlags(cmd.PersistentFlags()))
			// we cannot instantiate the config earlier because
			// cobra flags do not work so early in the game
			genericcli.Must(initConfigWithViperCtx(cfg))

			return nil
		},
	}

	rootCmd.PersistentFlags().StringP("url", "u", "", "api server address. Can be specified with CLOUDCTL_URL environment variable.")
	rootCmd.PersistentFlags().String("apitoken", "", "api token to authenticate. Can be specified with CLOUDCTL_APITOKEN environment variable.")
	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to the kube-config to use for authentication and authorization. Is updated by login. Uses default path if not specified.")
	rootCmd.PersistentFlags().StringP("order", "", "", "order by (comma separated) column(s)")
	rootCmd.PersistentFlags().BoolP("no-headers", "", false, "omit headers in tables")
	rootCmd.PersistentFlags().BoolP("debug", "", false, "enable debug")
	rootCmd.PersistentFlags().Bool("force-color", false, "force colored output even without tty")
	rootCmd.PersistentFlags().StringP("output-format", "o", "table", "output format (table|wide|markdown|json|yaml|template), wide is a table with more columns.")
	rootCmd.PersistentFlags().StringP("template", "", "", `output template for template output-format, go template format.
	For property names inspect the output of -o json for reference.
	Example for clusters:

	cloudctl cluster ls -o template --template "{{ .ID }} {{ .Name }}"

	`)
	rootCmd.PersistentFlags().BoolP("yes-i-really-mean-it", "", false, "skips security prompts (which can be dangerous to set blindly because actions can lead to data loss or additional costs)")

	rootCmd.AddCommand(newAuditCmd(cfg))
	rootCmd.AddCommand(newClusterCmd(cfg))
	rootCmd.AddCommand(newDashboardCmd(cfg))
	rootCmd.AddCommand(newUpdateCmd(cfg, binaryName))
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
	// the config will be provided with more values on cobra init
	// cobra flags do not work so early in the game
	c := &config{
		fs:   afero.NewOsFs(),
		out:  os.Stdout,
		comp: &completion.Completion{},
	}

	cmd := newRootCmd(c)
	err := cmd.Execute()
	if err != nil {
		if viper.GetBool("debug") {
			panic(err)
		}
		os.Exit(1)
	}
}

func initConfigWithViperCtx(cfg *config) error {
	viper.SetEnvPrefix(strings.ToUpper(binaryName))
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
		viper.AddConfigPath(fmt.Sprintf("/etc/%s", binaryName))
		h, err := os.UserHomeDir()
		if err != nil {
			log.Printf("unable to figure out user home directory, skipping config lookup path: %v", err)
		} else {
			viper.AddConfigPath(fmt.Sprintf(h+"/.%s", binaryName))
		}
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			usedCfg := viper.ConfigFileUsed()
			if usedCfg != "" {
				log.Fatalf("config %s file unreadable:%v", usedCfg, err)
			}
		}
	}

	if viper.IsSet("kubeconfig") {
		kubeconfigPath, err := helper.ExpandHomeDir(viper.GetString("kubeconfig"))
		if err != nil {
			return fmt.Errorf("unable to get kubeconfig path: %w", err)
		}

		viper.Set("kubeconfig", kubeconfigPath)
	}

	ctx := api.MustDefaultContext()

	opts := &slog.HandlerOptions{}
	if viper.GetBool("debug") {
		opts.Level = slog.LevelDebug
	}
	cfg.log = slog.New(slog.NewJSONHandler(os.Stdout, opts))

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

	cfg.cloud = cloud
	cfg.comp.SetClient(cloud)

	parsedURL, err := url.Parse(driverURL)
	if err != nil {
		log.Fatalf("could not parse driver url: %v", err)
	}
	cfg.consoleHost = parsedURL.Host

	return nil
}
