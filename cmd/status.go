package cmd

import (
	"context"
	"fmt"
	"html"
	"net"
	"net/http"
	"time"

	"github.com/fi-ts/cloudctl/cmd/helper"
	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultStatusURL = "https://status.fits.cloud"

func newStatusCmd() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "open the status page in the browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			authContext, err := api.GetAuthContext(viper.GetString("kubeconfig"))
			if err != nil {
				return fmt.Errorf("no valid session found, please run `cloudctl login` first: %w", err)
			}

			statusURL := viper.GetString("status-url")

			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				return fmt.Errorf("unable to start local server: %w", err)
			}

			port := listener.Addr().(*net.TCPAddr).Port
			localURL := fmt.Sprintf("http://127.0.0.1:%d", port)

			srv := &http.Server{
				ReadHeaderTimeout: 5 * time.Second,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html><body>
<form id="f" method="POST" action="%s/auth/login">
<input type="hidden" name="token" value="%s">
</form>
<script>document.getElementById("f").submit();</script>
</body></html>`, html.EscapeString(statusURL), html.EscapeString(authContext.IDToken))
				}),
			}

			go func() {
				time.Sleep(3 * time.Second)
				_ = srv.Shutdown(context.Background())
			}()

			if err := helper.OpenBrowser(localURL); err != nil {
				_ = srv.Shutdown(context.Background())
				fmt.Println("Could not open browser automatically.")
				fmt.Printf("Please open %s manually (available for 3 seconds).\n", localURL)
			} else {
				fmt.Println("Opening status page in browser...")
			}

			_ = srv.Serve(listener)

			return nil
		},
	}

	statusCmd.Flags().String("status-url", defaultStatusURL, "URL of the status page")

	return statusCmd
}
