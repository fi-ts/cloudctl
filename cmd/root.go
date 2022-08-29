package cmd

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	cloudgo "github.com/fi-ts/cloud-go"
	"github.com/fi-ts/cloud-go/api/client"
	"github.com/fi-ts/cloudctl/cmd/completion"
	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const binaryName = "cloudctl"

type config struct {
	name            string
	fs              afero.Fs
	out             io.Writer
	client          *client.CloudAPI
	comp            *completion.Completion
	consoleHost     string
	log             *zap.SugaredLogger
	describePrinter printers.Printer
	listPrinter     printers.Printer
}

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

func newRootCmd(c *config) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          binaryName,
		Short:        "a cli to manage cloud entities.",
		Long:         "with cloudctl you can manage kubernetes cluster, view networks et.al.",
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			must(viper.BindPFlags(cmd.Flags()))
			must(viper.BindPFlags(cmd.PersistentFlags()))
			// we cannot instantiate the config earlier because
			// cobra flags do not work so early in the game
			must(initConfigWithViperCtx(c))
		},
	}

	markdownCmd := &cobra.Command{
		Use:   "markdown",
		Short: "create markdown documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doc.GenMarkdownTree(rootCmd, "./docs")
		},
		DisableAutoGenTag: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			recursiveAutoGenDisable(rootCmd)
		},
	}

	rootCmd.PersistentFlags().StringP("api-url", "", "", "api server address. Can be specified with CLOUDCTL_API_URL environment variable.")
	rootCmd.PersistentFlags().String("api-token", "", "api token to authenticate. Can be specified with CLOUDCTL_API_TOKEN environment variable.")
	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to the kube-config to use for authentication and authorization. Is updated by login. Uses default path if not specified.")

	rootCmd.PersistentFlags().StringP("order", "", "", "order by (comma separated) column(s)")

	rootCmd.PersistentFlags().StringP("output-format", "o", "table", "output format (table|wide|markdown|json|yaml|template), wide is a table with more columns.")
	rootCmd.PersistentFlags().StringP("template", "", "", `output template for template output-format, go template format.
	For property names inspect the output of -o json for reference.
	Example for clusters:

	cloudctl cluster ls -o template --template "{{ .ID }} {{ .Name }}"

	`)
	rootCmd.PersistentFlags().BoolP("no-headers", "", false, "omit headers in tables")

	rootCmd.PersistentFlags().BoolP("yes-i-really-mean-it", "", false, "skips security prompts (which can be dangerous to set blindly because actions can lead to data loss or additional costs)")
	rootCmd.PersistentFlags().Bool("debug", false, "debug output")
	rootCmd.PersistentFlags().Bool("force-color", false, "force colored output even without tty")

	must(rootCmd.RegisterFlagCompletionFunc("output-format", completion.OutputFormatListCompletion))

	rootCmd.AddCommand(newClusterCmd(c))
	rootCmd.AddCommand(newDashboardCmd(c))
	rootCmd.AddCommand(newUpdateCmd())
	rootCmd.AddCommand(newLoginCmd(c))
	rootCmd.AddCommand(newLogoutCmd(c))
	rootCmd.AddCommand(markdownCmd)
	rootCmd.AddCommand(newWhoamiCmd())
	rootCmd.AddCommand(newProjectCmd(c))
	rootCmd.AddCommand(newTenantCmd(c))
	rootCmd.AddCommand(newContextCmd(c))
	rootCmd.AddCommand(newS3Cmd(c))
	rootCmd.AddCommand(newVersionCmd(c))
	rootCmd.AddCommand(newVolumeCmd(c))
	rootCmd.AddCommand(newPostgresCmd(c))
	rootCmd.AddCommand(newIPCmd(c))
	rootCmd.AddCommand(newBillingCmd(c))
	rootCmd.AddCommand(newHealthCmd(c))

	cobra.OnInitialize(func() {
		must(readConfigFile())
	})

	return rootCmd
}

func readConfigFile() error {
	viper.SetEnvPrefix(strings.ToUpper(binaryName))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetConfigType("yaml")
	cfgFile := viper.GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("config file path set explicitly, but unreadable: %w", err)
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(fmt.Sprintf("/etc/%s", binaryName))

		h, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to figure out user home directory, skipping config lookup path: %w", err)
		} else {
			viper.AddConfigPath(fmt.Sprintf(h+"/.%s", binaryName))
		}

		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			usedCfg := viper.ConfigFileUsed()
			if usedCfg != "" {
				return fmt.Errorf("config %s file unreadable: %w", usedCfg, err)
			}
		}
	}

	return nil
}

func initConfigWithViperCtx(c *config) error {
	ctx := api.MustDefaultContext()

	c.listPrinter = newPrinterFromCLI(c.out)
	c.describePrinter = defaultToYAMLPrinter(c.out)

	if c.log == nil {
		logger, err := newLogger()
		if err != nil {
			return fmt.Errorf("error creating logger: %w", err)
		}
		c.log = logger
	}

	if c.client != nil {
		return nil
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

	parsedURL, err := url.Parse(driverURL)
	if err != nil {
		log.Fatalf("could not parse driver url: %v", err)
	}

	c.comp.SetClient(cloud)
	c.consoleHost = parsedURL.Host
	c.client = cloud

	return nil
}

func newLogger() (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	if viper.GetBool("debug") {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return l.Sugar(), nil
}

func recursiveAutoGenDisable(cmd *cobra.Command) {
	cmd.DisableAutoGenTag = true
	for _, child := range cmd.Commands() {
		recursiveAutoGenDisable(child)
	}
}
