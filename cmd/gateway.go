package cmd

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/fi-ts/cloud-go/api/client/gateway"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/utils/pointer"
)

var (
	gwCmd = &cobra.Command{
		Use:     "gateway",
		Aliases: []string{"gw"},
		Short:   "manages gateways",
		Long:    "Manage gateways which enable access to services in another cluster",
	}
	gwClientCmd = &cobra.Command{
		Use:   "client",
		Short: "manages gateway clients",
		Long:  "Manage gateway clients which constitute half of gateway pairs",
	}
	gwClientAddPipesCmd = &cobra.Command{
		Use:   "add-pipes",
		Short: "add pipes to a gateway",
		Long:  "add pipes to a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return gatewayAddPipes()
		},
		PreRun: bindPFlags,
	}
	gwClientCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "creates a gateway",
		Long:  "Create a new gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return gatewayCreate()
		},
		PreRun: bindPFlags,
	}
	gwClientDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "deletes a gateway",
		Long:  "Delete a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return gatewayDelete()
		},
		PreRun: bindPFlags,
	}
	gwClientDescribeCmd = &cobra.Command{
		Use:   "describe",
		Short: "describe a gateway",
		Long:  "Describe a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return gatewayDescribe()
		},
		PreRun: bindPFlags,
	}
)

func init() {
	gwCmd.AddCommand(gwClientCmd)

	gwClientCreateCmd.Flags().String("name", "", "Name of the gateway")
	gwClientCreateCmd.Flags().String("clients", "", "Name of the gateway clients")
	gwClientCreateCmd.Flags().String("project", "", "Project-UID which the gateway belongs to")
	gwClientCreateCmd.Flags().String("pipes", "", "Comma-separated-list of pipes (e.g. PIPE_1,PIPE_2). Pipe has format SVC_NAME_1:CLIENT_POD_PORT_1:REMOTE_SVC_ENDPOINT_1.")
	if err := gwClientCreateCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
	if err := gwClientCreateCmd.MarkFlagRequired("project"); err != nil {
		log.Fatal(err)
	}
	if err := gwClientCreateCmd.MarkFlagRequired("pipes"); err != nil {
		log.Fatal(err)
	}
	gwClientCmd.AddCommand(gwClientCreateCmd)

	gwClientAddPipesCmd.Flags().String("name", "", "Name of the gateway")
	gwClientAddPipesCmd.Flags().String("clients", "", "Name of the gateway clients")
	gwClientAddPipesCmd.Flags().String("project", "", "Project-UID which the gateway belongs to")
	gwClientAddPipesCmd.Flags().String("pipes", "", "Comma-separated-list of pipes (e.g. PIPE_1,PIPE_2). Pipe has format SVC_NAME_1:CLIENT_POD_PORT_1:REMOTE_SVC_ENDPOINT_1.")
	if err := gwClientAddPipesCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
	if err := gwClientAddPipesCmd.MarkFlagRequired("project"); err != nil {
		log.Fatal(err)
	}
	if err := gwClientAddPipesCmd.MarkFlagRequired("pipes"); err != nil {
		log.Fatal(err)
	}
	gwClientCmd.AddCommand(gwClientAddPipesCmd)

	gwClientDeleteCmd.Flags().String("name", "", "Name of the gateway")
	gwClientDeleteCmd.Flags().String("clients", "", "Name of the gateway clients")
	gwClientDeleteCmd.Flags().String("project", "", "Project-UID which the gateway belongs to")
	if err := gwClientDeleteCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
	if err := gwClientDeleteCmd.MarkFlagRequired("project"); err != nil {
		log.Fatal(err)
	}
	gwClientCmd.AddCommand(gwClientDeleteCmd)

	gwClientDescribeCmd.Flags().String("name", "", "Name of the gateway")
	gwClientDescribeCmd.Flags().String("project", "default", "Project-UID which the gateway belongs to")
	if err := gwClientDescribeCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
	gwClientCmd.AddCommand(gwClientDescribeCmd)
}

func gatewayCreate() error {
	params := gateway.NewCreateParams()

	parsed, err := parseFlagPipes()
	if err != nil {
		return fmt.Errorf("failed to parse flag `pipes`: %w", err)
	}

	params.SetBody(&models.V1GatewayCreateRequest{
		Name:  ptr(viper.GetString("name")),
		Pipes: parsed,
		Peers: []*models.V1PeerSpec{{
			Name:      ptr(""),
			PublicKey: ptr(""),
		}},
		ProjectUID: ptr(viper.GetString("project")),
		Type:       ptr("client"),
	})

	resp, err := cloud.Gateway.Create(params, nil)
	if err != nil {
		return fmt.Errorf("failed to create gateway: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func gatewayDelete() error {
	params := gateway.NewDeleteParams()
	params.SetBody(&models.V1GatewayDeleteRequest{
		ProjectUID: ptr(viper.GetString("project")),
		Name:       ptr(viper.GetString("name")),
	})
	resp, err := cloud.Gateway.Delete(params, nil)
	if err != nil {
		return fmt.Errorf("failed to delete gateway: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func gatewayAddPipes() error {
	params := gateway.NewAddPipesParams()

	parsed, err := parseFlagPipes()
	if err != nil {
		return fmt.Errorf("failed to parse flag `pipes`: %w", err)
	}

	params.SetBody(&models.V1GatewayAddPipesRequest{
		Name:  ptr(viper.GetString("name")),
		Pipes: parsed,
		Peers: []*models.V1PeerSpec{{
			Name:      ptr(""),
			PublicKey: ptr(""),
		}},
		ProjectUID: ptr(viper.GetString("project")),
		Type:       ptr("client"),
	})

	resp, err := cloud.Gateway.AddPipes(params, nil)
	if err != nil {
		return fmt.Errorf("failed to add pipes to the gateway: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func gatewayDescribe() error {
	params := gateway.NewDescribeParams()

	params.SetBody(&models.V1GatewayDescribeRequest{
		Name:       ptr(viper.GetString("name")),
		ProjectUID: ptr(viper.GetString("project")),
	})

	resp, err := cloud.Gateway.Describe(params, nil)
	if err != nil {
		return fmt.Errorf("failed to describe a gateway: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func parseFlagPipes() ([]*models.V1PipeSpec, error) {
	ss := strings.Split(viper.GetString("pipes"), ",")
	pipes := []*models.V1PipeSpec{}
	for i := range ss {
		pipe, err := parsePipe(ss[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse flag `pipes`: %w", err)
		}
		pipes = append(pipes, pipe)
	}
	return pipes, nil
}

func parsePipe(unparsed string) (*models.V1PipeSpec, error) {
	ss := strings.Split(unparsed, ":")
	if len(ss) < 3 {
		return nil, errors.New("pipe incomplete: it should be a colon-separated-list `SVC_NAME_1:CLIENT_POD_PORT_1:REMOTE_SVC_ENDPOINT_1`")
	}

	pipe := &models.V1PipeSpec{}
	pipe.Name = ptr(ss[0])
	port, err := u16StrToI64Ptr(ss[1])
	if err != nil {
		return nil, fmt.Errorf("failed to convert `%s` to pointer to int64: %w", ss[2], err)
	}
	pipe.Port = port
	pipe.Remote = ptr(strings.TrimPrefix(unparsed, ss[0]+":"+ss[1]+":"))
	return pipe, nil
}

func ptr(s string) *string {
	return pointer.StringPtr(s)
}

// Convert an uint16 as string to a pointer to int64
func u16StrToI64Ptr(s string) (*int64, error) {
	u16AsU64, err := strconv.ParseUint(s, 10, 16) // uint16 in gateway k8s api
	if err != nil {
		return nil, fmt.Errorf("failed to convert the port in pipe %s to uint16: %w", s, err)
	}

	return pointer.Int64Ptr(int64(u16AsU64)), nil
}
