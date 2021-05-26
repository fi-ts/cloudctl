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

const (
	serverPatchFlagPipesToAdd = "pipes-to-add"
	clientCreateFlagPipes     = "pipes"
	clientCreateFlagServer    = "server"
	clientPatchFlagPipes      = "pipes"
)

var (
	gwCmd = &cobra.Command{
		Use:     "gateway",
		Aliases: []string{"gw"},
		Short:   "Manage gateways",
		Long:    "Manage gateways which enable access to services in another cluster",
	}
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Manage servers",
	}
	serverCreateCmd = &cobra.Command{
		Use:    "create <projectUID>--<name>",
		Long:   "Create a server",
		Short:  "Create a server",
		RunE:   serverCreate,
		PreRun: bindPFlags,
	}
	serverDescribeCmd = &cobra.Command{
		Use:   "describe <projectUID>--<name>",
		Long:  "Describe a server",
		Short: "Describe a server",
		RunE:  serverDescribe,
	}
	serverListCmd = &cobra.Command{
		Use:   "list <projectUID>",
		Long:  "List servers under a specific project-UID",
		Short: "List servers under a specific project-UID",
		RunE:  serverList,
	}
	serverPatchCmd = &cobra.Command{
		Use:    "patch <projectUID>--<name>",
		Long:   "Patch a server",
		Short:  "Patch a server",
		RunE:   serverPatch,
		PreRun: bindPFlags,
	}
	serverDeleteCmd = &cobra.Command{
		Use:   "delete <projectUID>--<name>",
		Long:  "Delete a server",
		Short: "Delete a server",
		RunE:  serverDelete,
	}
	// clientCmd = &cobra.Command{
	// 	Use:   "client",
	// 	Short: "manages gateway clients",
	// 	Long:  "Manage gateway clients which constitute half of gateway pairs",
	// }
	// gwClientAddPipesCmd = &cobra.Command{
	// 	Use:   "add-pipes",
	// 	Short: "add pipes to a gateway",
	// 	Long:  "add pipes to a gateway",
	// 	RunE: func(cmd *cobra.Command, args []string) error {
	// 		return gatewayAddPipes()
	// 	},
	// 	PreRun: bindPFlags,
	// }
	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Manage clients",
	}
	clientCreateCmd = &cobra.Command{
		Use:    "create <projectUID>--<name>",
		Long:   "Create a client",
		Short:  "Create a client",
		RunE:   clientCreate,
		PreRun: bindPFlags,
	}
	clientDescribeCmd = &cobra.Command{
		Use:    "describe <projectUID>--<name>",
		Short:  "Describe a gateway",
		Long:   "Describe a gateway",
		RunE:   clientDescribe,
		PreRun: bindPFlags,
	}
	clientListCmd = &cobra.Command{
		Use:   "list <projectUID>",
		Long:  "List clients under a specific project-UID",
		Short: "List clients under a specific project-UID",
		RunE:  clientList,
	}
	clientPatchCmd = &cobra.Command{
		Use:    "patch <projectUID>--<name>",
		Long:   "Patch a client",
		Short:  "Patch a client",
		RunE:   clientPatch,
		PreRun: bindPFlags,
	}
	clientDeleteCmd = &cobra.Command{
		Use:   "delete <projectUID>--<name>",
		Long:  "Delete a client",
		Short: "Delete a client",
		RunE:  clientDelete,
	}
)

func init() {
	gwCmd.AddCommand(serverCmd)
	gwCmd.AddCommand(clientCmd)

	serverCmd.AddCommand(serverCreateCmd)
	serverCmd.AddCommand(serverDescribeCmd)
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverPatchCmd)
	serverCmd.AddCommand(serverDeleteCmd)

	clientCmd.AddCommand(clientCreateCmd)
	clientCmd.AddCommand(clientDescribeCmd)
	clientCmd.AddCommand(clientListCmd)
	clientCmd.AddCommand(clientPatchCmd)
	clientCmd.AddCommand(clientDeleteCmd)

	serverPatchCmd.Flags().String(serverPatchFlagPipesToAdd, "", "New pipes to add to the gateway server. `Pipes` is a comma-separated-list of pipes (e.g. PIPE_1,PIPE_2). Each pipe has format SVC_NAME_1:CLIENT_POD_PORT_1:REMOTE_SVC_ENDPOINT_1.")
	if err := serverPatchCmd.MarkFlagRequired(serverPatchFlagPipesToAdd); err != nil {
		log.Fatal(err)
	}

	// client create
	clientCreateCmd.Flags().String(clientCreateFlagServer, "", "UID of the peer server of the client")
	if err := clientCreateCmd.MarkFlagRequired(clientCreateFlagServer); err != nil {
		log.Fatal(err)
	}
	clientCreateCmd.Flags().String(clientCreateFlagPipes, "", "Comma-separated list of pipe names chosen from the server's `pipes` spec, e.g. `PIPE_NAME_1,PIPE_NAME_2`")
	if err := clientCreateCmd.MarkFlagRequired(clientCreateFlagPipes); err != nil {
		log.Fatal(err)
	}

	// client patch
	clientPatchCmd.Flags().String(clientPatchFlagPipes, "", "Comma-separated list of the new pipe names to add, which are chosen from the server's `pipes` spec, e.g. `NEW_PIPE_NAME_1,NEW_PIPE_NAME_2`")
	if err := clientPatchCmd.MarkFlagRequired(clientPatchFlagPipes); err != nil {
		log.Fatal(err)
	}
}

func clientCreate(cmd *cobra.Command, args []string) error {
	projectUID, name, err := validateCommandArgsAndParseInstanceID(args)
	if err != nil {
		return err
	}

	serverProjectUID, serverName, err := parseInstanceID(viper.GetString(clientCreateFlagServer))
	if err != nil {
		return err
	}

	parsedPipeNames, err := parseCommaSeparatedString(viper.GetString(clientCreateFlagPipes))
	if err != nil {
		return err
	}

	// parsedPipes, err := parseFlagPipes()
	// if err != nil {
	// 	return err
	// }
	params := gateway.NewPostClientParams()
	params.SetProjectuid(projectUID)
	params.SetBody(&models.V1GatewayClientPostRequest{
		ProjectUID:       &projectUID,
		Name:             &name,
		ServerProjectUID: serverProjectUID,
		ServerName:       serverName,
		// Pipes:            parsedPipes,
		PipeNames: parsedPipeNames,
	})

	resp, err := cloud.Gateway.PostClient(params, nil)
	if err != nil {
		return fmt.Errorf("post gateway client: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func clientDelete(cmd *cobra.Command, args []string) error {
	projectUID, name, err := validateCommandArgsAndParseInstanceID(args)
	if err != nil {
		return err
	}

	params := gateway.NewDeleteClientParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)
	resp, err := cloud.Gateway.DeleteClient(params, nil)
	if err != nil {
		return fmt.Errorf("deleting gateway client %s/%s: %w", projectUID, name, err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func clientDescribe(cmd *cobra.Command, args []string) error {
	projectUID, name, err := validateCommandArgsAndParseInstanceID(args)
	if err != nil {
		return err
	}
	params := gateway.NewGetClientParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)

	// params.SetBody(&models.V1GatewayDescribeRequest{
	// 	Name:       ptr(viper.GetString("name")),
	// 	ProjectUID: ptr(viper.GetString("project")),
	// })

	resp, err := cloud.Gateway.GetClient(params, nil)
	if err != nil {
		return fmt.Errorf("fetching a gateway client: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func clientList(cmd *cobra.Command, args []string) error {
	params := gateway.NewListClientsParams()
	params.SetProjectuid(args[0])
	resp, err := cloud.Gateway.ListClients(params, nil)
	if err != nil {
		return fmt.Errorf("failed to list gateway clients: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func clientPatch(cmd *cobra.Command, args []string) error {
	projectUID, name, err := validateCommandArgsAndParseInstanceID(args)
	if err != nil {
		return err
	}
	params := gateway.NewPatchClientParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)

	pipes, err := parseCommaSeparatedString(viper.GetString(clientPatchFlagPipes))
	if err != nil {
		return fmt.Errorf("parsing the flag `pipes`: %w", err)
	}
	req := &models.V1GatewayClientPatchRequest{
		ProjectUID: &projectUID,
		Name:       &name,
		PipeNames:  pipes,
	}
	params.SetBody(req)

	resp, err := cloud.Gateway.PatchClient(params, nil)
	if err != nil {
		return fmt.Errorf("patching the gateway server: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func serverCreate(cmd *cobra.Command, args []string) error {
	projectUID, name, err := validateCommandArgsAndParseInstanceID(args)
	if err != nil {
		return err
	}
	params := gateway.NewPostServerParams()
	params.SetProjectuid(projectUID)

	req := &models.V1GatewayServerPostRequest{
		ProjectUID: &projectUID,
		Name:       &name,
		// Peers: []*models.V1PeerSpec{{
		// 	Name:      ptr(""),
		// 	PublicKey: ptr(""),
		// }},
	}
	// if viper.GetString("pipes") != "" {
	// 	pipes, err := parseFlagPipes()
	// 	if err != nil {
	// 		return fmt.Errorf("failed to parse flag `pipes`: %w", err)
	// 	}
	// 	req.Pipes = pipes
	// }
	params.SetBody(req)

	resp, err := cloud.Gateway.PostServer(params, nil)
	if err != nil {
		return fmt.Errorf("failed to create gateway: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func serverDelete(cmd *cobra.Command, args []string) error {
	projectUID, name, err := validateCommandArgsAndParseInstanceID(args)
	if err != nil {
		return err
	}

	params := gateway.NewDeleteServerParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)
	resp, err := cloud.Gateway.DeleteServer(params, nil)
	if err != nil {
		return fmt.Errorf("failed to delete gateway server %s/%s: %w", projectUID, name, err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func serverDescribe(cmd *cobra.Command, args []string) error {
	projectUID, name, err := validateCommandArgsAndParseInstanceID(args)
	if err != nil {
		return err
	}
	params := gateway.NewGetServerParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)
	resp, err := cloud.Gateway.GetServer(params, nil)
	if err != nil {
		return fmt.Errorf("failed to describe a gateway: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func serverList(cmd *cobra.Command, args []string) error {
	params := gateway.NewListServersParams()
	params.SetProjectuid(args[0])
	resp, err := cloud.Gateway.ListServers(params, nil)
	if err != nil {
		return fmt.Errorf("failed to list gateway servers: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func serverPatch(cmd *cobra.Command, args []string) error {
	projectUID, name, err := validateCommandArgsAndParseInstanceID(args)
	if err != nil {
		return err
	}
	params := gateway.NewPatchServerParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)

	parsedPipes, err := parsePipeSpecs(viper.GetString(serverPatchFlagPipesToAdd))
	if err != nil {
		return fmt.Errorf("failed to parse flag `pipes`: %w", err)
	}
	req := &models.V1GatewayServerPatchRequest{
		ProjectUID: ptr(projectUID),
		Name:       ptr(name),
		Pipes:      parsedPipes,
	}
	params.SetBody(req)

	resp, err := cloud.Gateway.PatchServer(params, nil)
	if err != nil {
		return fmt.Errorf("failed to create gateway server: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

// other helpers

func parseCommaSeparatedString(s string) ([]string, error) {
	ss := strings.Split(s, ",")
	if len(ss) == 0 {
		return nil, fmt.Errorf("failed to parse %s", s)
	}
	return ss, nil
}

func parseInstanceID(id string) (string, string, error) {
	ss := strings.Split(id, "--")
	if len(ss) != 2 {
		return "", "", errors.New("`%s` should have the format <project UID>--<instance name>")
	}
	return ss[0], ss[1], nil
}

func parsePipeSpec(unparsed string) (*models.V1PipeSpec, error) {
	ss := strings.Split(unparsed, ":")
	if len(ss) < 3 {
		return nil, fmt.Errorf("`pipe` %s incomplete: it should be a colon-separated-list `SVC_NAME_1:CLIENT_POD_PORT_1:REMOTE_SVC_ENDPOINT_1`", unparsed)
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

func parsePipeSpecs(specs string) ([]*models.V1PipeSpec, error) {
	ss := strings.Split(specs, ",")
	pipes := []*models.V1PipeSpec{}
	for i := range ss {
		pipe, err := parsePipeSpec(ss[i])
		if err != nil {
			return nil, fmt.Errorf("parsing pipe spec %s: %w", ss[i], err)
		}
		pipes = append(pipes, pipe)
	}
	return pipes, nil
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

func validateCommandArgsAndParseInstanceID(args []string) (string, string, error) {
	if len(args) != 1 {
		return "", "", fmt.Errorf("command args `%s` should only contain one string", args)
	}
	return parseInstanceID(args[0])
}
