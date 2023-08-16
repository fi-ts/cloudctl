package cmd

import (
	"fmt"
	"time"

	"github.com/fi-ts/cloud-go/api/client/audit"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newAuditCmd(c *config) *cobra.Command {
	auditCmd := &cobra.Command{
		Use:   "audit",
		Short: "show audit traces of the api. feature must be enabled on server-side.",
		Long:  "show audit traces of the api. feature must be enabled on server-side.",
	}
	auditListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list audit traces",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.auditList()
		},
		PreRun: bindPFlags,
	}
	auditDescribeCmd := &cobra.Command{
		Use:   "describe <rqid>",
		Short: "describe an audit trace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.auditDescribe(args)
		},
		PreRun: bindPFlags,
	}

	auditDescribeCmd.Flags().String("phase", "response", "phase of the audit trace. One of [request, response, single, error, opened, closed]")

	auditListCmd.Flags().StringP("query", "q", "", "filters audit trace body payloads for the given text.")

	auditListCmd.Flags().String("from", "1h", "start of range of the audit traces. e.g. 1h, 10m, 2006-01-02 15:04:05")
	auditListCmd.Flags().String("to", "", "end of range of the audit traces. e.g. 1h, 10m, 2006-01-02 15:04:05")

	auditListCmd.Flags().String("component", "", "component of the audit trace.")
	auditListCmd.Flags().String("request-id", "", "request id of the audit trace.")
	auditListCmd.Flags().String("type", "", "type of the audit trace. One of [http, grpc, event].")

	auditListCmd.Flags().String("user", "", "user of the audit trace.")
	auditListCmd.Flags().String("tenant", "", "tenant of the audit trace.")

	auditListCmd.Flags().String("detail", "", "detail of the audit trace. An HTTP method, unary or stream")
	auditListCmd.Flags().String("phase", "", "phase of the audit trace. One of [request, response, single, error, opened, closed]")

	auditListCmd.Flags().String("path", "", "api path of the audit trace.")
	auditListCmd.Flags().String("forwarded-for", "", "forwarded for of the audit trace.")
	auditListCmd.Flags().String("remote-addr", "", "remote address of the audit trace.")

	auditListCmd.Flags().String("error", "", "error of the audit trace.")
	auditListCmd.Flags().Int32("status-code", 0, "HTTP status code of the audit trace.")

	auditListCmd.Flags().Int64("limit", 100, "limit the number of audit traces.")

	auditCmd.AddCommand(auditDescribeCmd)
	auditCmd.AddCommand(auditListCmd)

	return auditCmd
}

func (c *config) auditList() error {
	fromDateTime, err := eventuallyRelativeDateTime(viper.GetString("from"))
	if err != nil {
		return err
	}
	toDateTime, err := eventuallyRelativeDateTime(viper.GetString("to"))
	if err != nil {
		return err
	}
	resp, err := c.cloud.Audit.FindAuditTraces(audit.NewFindAuditTracesParams().WithBody(&models.V1AuditFindRequest{
		Body:         viper.GetString("query"),
		From:         fromDateTime,
		To:           toDateTime,
		Component:    viper.GetString("component"),
		Rqid:         viper.GetString("request-id"),
		Type:         viper.GetString("type"),
		User:         viper.GetString("user"),
		Tenant:       viper.GetString("tenant"),
		Detail:       viper.GetString("detail"),
		Phase:        viper.GetString("phase"),
		Path:         viper.GetString("path"),
		ForwardedFor: viper.GetString("forwarded-for"),
		RemoteAddr:   viper.GetString("remote-addr"),
		Error:        viper.GetString("error"),
		StatusCode:   viper.GetInt32("status-code"),
		Limit:        viper.GetInt64("limit"),
	}), nil)
	if err != nil {
		return err
	}

	return output.New().Print(resp.Payload)
}

func (c *config) auditDescribe(args []string) error {
	id, err := c.auditID("describe", args)
	if err != nil {
		return err
	}

	traces, err := c.cloud.Audit.FindAuditTraces(audit.NewFindAuditTracesParams().WithBody(&models.V1AuditFindRequest{
		Rqid:  id,
		Phase: viper.GetString("phase"),
	}), nil)
	if err != nil {
		return err
	}
	if len(traces.Payload) == 0 {
		return fmt.Errorf("no audit trace found with request id %s", id)
	}

	return output.New().Print(traces.Payload[0])
}

func eventuallyRelativeDateTime(s string) (strfmt.DateTime, error) {
	if s == "" {
		return strfmt.DateTime{}, nil
	}
	duration, err := strfmt.ParseDuration(s)
	if err == nil {
		return strfmt.DateTime(time.Now().Add(-duration)), nil
	}
	return strfmt.ParseDateTime(s)
}

func (c *config) auditID(verb string, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("audit %s requires projectID as argument", verb)
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", fmt.Errorf("audit %s requires exactly one projectID as argument", verb)
}
