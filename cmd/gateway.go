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
	flagProject                    = "project"
	flagAllProjects                = "all-projects"
	clientCreateFlagPipes          = "pipes"
	clientCreateFlagServer         = "server"
	clientUpdateFlagPipesToAdd     = "pipes-to-add"
	serverCreateFlagLoadBalancerIP = "ip"
	serverCreateFlagPipes          = "pipes"
	serverUpdateFlagPipesToAdd     = "pipes-to-add"
)

var (
	gwCmd = &cobra.Command{
		Aliases: []string{"gw"},
		Use:     "gateway",
		Short:   "Manage gateways",
		Long:    "Manage gateways which enable access to services in another cluster",
		PreRun:  bindPFlags,
	}
	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Manage clients",
	}
	clientCreateCmd = &cobra.Command{
		Use:    "create --project <projectUID> <name>",
		Long:   "Create a client",
		Short:  "Create a client",
		RunE:   clientCreate,
		PreRun: bindPFlags,
	}
	clientDeleteCmd = &cobra.Command{
		Use:    "delete --project <projectUID> <name>",
		Long:   "Delete a client",
		Short:  "Delete a client",
		RunE:   clientDelete,
		PreRun: bindPFlags,
	}
	clientDescribeCmd = &cobra.Command{
		Use:    "describe --project <projectUID> <name>",
		Short:  "Describe a gateway",
		Long:   "Describe a gateway",
		RunE:   clientDescribe,
		PreRun: bindPFlags,
	}
	clientListCmd = &cobra.Command{
		Use:    "list --project <projectUID>",
		Long:   "List clients under a specific project-UID",
		Short:  "List clients under a specific project-UID",
		RunE:   clientList,
		PreRun: bindPFlags,
	}
	clientUpdateCmd = &cobra.Command{
		Use:    "update --project <projectUID> <name>",
		Long:   "Update a client",
		Short:  "Update a client",
		RunE:   clientUpdate,
		PreRun: bindPFlags,
	}
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Manage servers",
	}
	serverCreateCmd = &cobra.Command{
		Use:    "create --project <projectUID> <name>",
		Long:   "Create a server",
		Short:  "Create a server",
		RunE:   serverCreate,
		PreRun: bindPFlags,
	}
	serverDescribeCmd = &cobra.Command{
		Use:    "describe --project <projectUID> <name>",
		Long:   "Describe a server",
		Short:  "Describe a server",
		RunE:   serverDescribe,
		PreRun: bindPFlags,
	}
	serverDeleteCmd = &cobra.Command{
		Use:    "delete --project <projectUID> <name>",
		Long:   "Delete a server",
		Short:  "Delete a server",
		RunE:   serverDelete,
		PreRun: bindPFlags,
	}
	serverListCmd = &cobra.Command{
		Aliases: []string{"ls"},
		Use:     "list --project <projectUID>",
		Long:    "List servers under a specific project-UID",
		Short:   "List servers under a specific project-UID",
		RunE:    serverList,
		PreRun:  bindPFlags,
	}
	serverUpdateCmd = &cobra.Command{
		Use:    "update --project <projectUID> <name>",
		Long:   "Update a server",
		Short:  "Update a server",
		RunE:   serverUpdate,
		PreRun: bindPFlags,
	}
	singleClientCmds = []*cobra.Command{clientCreateCmd, clientDescribeCmd, clientUpdateCmd, clientDeleteCmd}
	singleServerCmds = []*cobra.Command{serverCreateCmd, serverDescribeCmd, serverUpdateCmd, serverDeleteCmd}
	listCmds         = []*cobra.Command{clientListCmd, serverListCmd}
)

func init() {
	gwCmd.AddCommand(serverCmd, clientCmd)
	clientCmd.AddCommand(append(singleClientCmds, clientListCmd)...)
	serverCmd.AddCommand(append(singleServerCmds, serverListCmd)...)

	// all single instance commands
	must(defineRequiredFlagProject(append(singleClientCmds, singleServerCmds...)))

	// multiple instance commands
	for i := range listCmds {
		defineFlagProjectMIC(listCmds[i])
	}

	clientCreateCmd.Flags().String(clientCreateFlagServer, "", "UID of the peer server of the client")
	must(clientCreateCmd.MarkFlagRequired(clientCreateFlagServer))

	clientCreateCmd.Flags().StringSlice(clientCreateFlagPipes, nil, "Pipe names chosen from the server's `pipes` spec, e.g. `PIPE_NAME_1,PIPE_NAME_2`")
	must(clientCreateCmd.MarkFlagRequired(clientCreateFlagPipes))

	clientUpdateCmd.Flags().StringSlice(clientUpdateFlagPipesToAdd, nil, "Comma-separated list of the new pipe names to add, which are chosen from the server's `pipes` spec, e.g. `NEW_PIPE_NAME_1,NEW_PIPE_NAME_2`")
	must(clientUpdateCmd.MarkFlagRequired(clientUpdateFlagPipesToAdd))

	serverCreateCmd.Flags().String(serverCreateFlagLoadBalancerIP, "", "IP of the load balancer of the gateway server.")
	must(serverCreateCmd.MarkFlagRequired(serverCreateFlagLoadBalancerIP))

	serverCreateCmd.Flags().StringSlice(serverCreateFlagPipes, nil, "Pipes of the gateway server, e.g. PIPE_1,PIPE_2. Each pipe has format SVC_NAME:CLIENT_POD_PORT:REMOTE_SVC_ENDPOINT.")

	serverUpdateCmd.Flags().StringSlice(serverUpdateFlagPipesToAdd, nil, "New pipes to add to the gateway server, e.g. PIPE_1,PIPE_2. Each pipe has format SVC_NAME:CLIENT_POD_PORT:REMOTE_SVC_ENDPOINT.")
	must(serverUpdateCmd.MarkFlagRequired(serverUpdateFlagPipesToAdd))
}

func clientCreate(cmd *cobra.Command, args []string) error {
	projectUID, name, err := projectUIDAndInstanceName(args)
	if err != nil {
		return err
	}

	serverProjectUID, serverName, err := parseInstanceID(viper.GetString(clientCreateFlagServer))
	if err != nil {
		return err
	}

	params := gateway.NewClientPostParams()
	params.SetProjectuid(projectUID)
	params.SetBody(&models.V1GatewayClientPostRequest{
		ProjectUID:       &projectUID,
		Name:             &name,
		ServerProjectUID: serverProjectUID,
		ServerName:       serverName,
		PipeNames:        viper.GetStringSlice(clientCreateFlagPipes),
	})

	resp, err := cloud.Gateway.ClientPost(params, nil)
	if err != nil {
		return fmt.Errorf("post gateway client: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func clientDelete(cmd *cobra.Command, args []string) error {
	projectUID, name, err := projectUIDAndInstanceName(args)
	if err != nil {
		return err
	}

	params := gateway.NewClientDeleteParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)
	resp, err := cloud.Gateway.ClientDelete(params, nil)
	if err != nil {
		return fmt.Errorf("deleting gateway client %s/%s: %w", projectUID, name, err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func clientDescribe(cmd *cobra.Command, args []string) error {
	projectUID, name, err := projectUIDAndInstanceName(args)
	if err != nil {
		return err
	}
	params := gateway.NewClientGetParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)

	resp, err := cloud.Gateway.ClientGet(params, nil)
	if err != nil {
		return fmt.Errorf("fetching a gateway client: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func clientList(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("no argument allowed")
	}

	project := viper.GetString(flagProject)
	if viper.GetBool(flagAllProjects) {
		if project != "" {
			return errors.New("only one of two flags allowed")
		}
		resp, err := cloud.Gateway.ClientListAll(gateway.NewClientListAllParams(), nil)
		if err != nil {
			return fmt.Errorf("listing gateway clients: %w", err)
		}
		return output.YAMLPrinter{}.Print(resp.Payload)
	}

	if project == "" {
		return errors.New("project missing")
	}
	params := gateway.NewClientListParams()
	params.SetProjectuid(viper.GetString(flagProject))
	resp, err := cloud.Gateway.ClientList(params, nil)
	if err != nil {
		return fmt.Errorf("failed to list gateway clients: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func clientUpdate(cmd *cobra.Command, args []string) error {
	projectUID, name, err := projectUIDAndInstanceName(args)
	if err != nil {
		return err
	}
	params := gateway.NewClientPatchParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)

	req := &models.V1GatewayClientPatchRequest{
		ProjectUID: &projectUID,
		Name:       &name,
		PipeNames:  viper.GetStringSlice(clientUpdateFlagPipesToAdd),
	}
	params.SetBody(req)

	resp, err := cloud.Gateway.ClientPatch(params, nil)
	if err != nil {
		return fmt.Errorf("patching the gateway server: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func serverCreate(cmd *cobra.Command, args []string) error {
	projectUID, name, err := projectUIDAndInstanceName(args)
	if err != nil {
		return err
	}
	params := gateway.NewServerPostParams()
	params.SetProjectuid(projectUID)

	req := &models.V1GatewayServerPostRequest{
		ProjectUID:     &projectUID,
		Name:           &name,
		LoadBalancerIP: viper.GetString(serverCreateFlagLoadBalancerIP),
	}
	if pipes := viper.GetStringSlice(serverCreateFlagPipes); len(pipes) > 0 {
		pipes, err := parsePipeSpecs(pipes)
		if err != nil {
			return fmt.Errorf("failed to parse flag `pipes`: %w", err)
		}
		req.Pipes = pipes
	}
	params.SetBody(req)

	resp, err := cloud.Gateway.ServerPost(params, nil)
	if err != nil {
		return fmt.Errorf("failed to create gateway: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func serverDelete(cmd *cobra.Command, args []string) error {
	projectUID, name, err := projectUIDAndInstanceName(args)
	if err != nil {
		return err
	}

	params := gateway.NewServerDeleteParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)
	resp, err := cloud.Gateway.ServerDelete(params, nil)
	if err != nil {
		return fmt.Errorf("failed to delete gateway server %s/%s: %w", projectUID, name, err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func serverDescribe(cmd *cobra.Command, args []string) error {
	projectUID, name, err := projectUIDAndInstanceName(args)
	if err != nil {
		return err
	}
	params := gateway.NewServerGetParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)
	resp, err := cloud.Gateway.ServerGet(params, nil)
	if err != nil {
		return fmt.Errorf("failed to describe a gateway: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func serverList(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("no argument allowed")
	}

	project := viper.GetString(flagProject)
	if viper.GetBool(flagAllProjects) {
		if project != "" {
			return errors.New("only one of two flags allowed")
		}
		resp, err := cloud.Gateway.ServerListAll(gateway.NewServerListAllParams(), nil)
		if err != nil {
			return fmt.Errorf("listing gateway servers: %w", err)
		}
		return output.YAMLPrinter{}.Print(resp.Payload)
	}

	if project == "" {
		return errors.New("project missing")
	}
	params := gateway.NewServerListParams()
	params.SetProjectuid(project)
	resp, err := cloud.Gateway.ServerList(params, nil)
	if err != nil {
		return fmt.Errorf("listing gateway servers: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

func serverUpdate(cmd *cobra.Command, args []string) error {
	projectUID, name, err := projectUIDAndInstanceName(args)
	if err != nil {
		return err
	}
	params := gateway.NewServerPatchParams()
	params.SetProjectuid(projectUID)
	params.SetName(name)

	parsedPipes, err := parsePipeSpecs(viper.GetStringSlice(serverUpdateFlagPipesToAdd))
	if err != nil {
		return fmt.Errorf("failed to parse flag `pipes`: %w", err)
	}
	req := &models.V1GatewayServerPatchRequest{
		ProjectUID: ptr(projectUID),
		Name:       ptr(name),
		Pipes:      parsedPipes,
	}
	params.SetBody(req)

	resp, err := cloud.Gateway.ServerPatch(params, nil)
	if err != nil {
		return fmt.Errorf("failed to create gateway server: %w", err)
	}
	return output.YAMLPrinter{}.Print(resp.Payload)
}

/* other than handlers */

func defineFlagProject(c *cobra.Command) {
	c.Flags().String(flagProject, "", "Project UID of gateway instance")
}

// MIC := multiple-instance-command
func defineFlagProjectMIC(c *cobra.Command) {
	// `--project` or `--all-projects`
	defineFlagProject(c)
	c.Flags().BoolP(flagAllProjects, "A", false, "Instances of all projects")
}

func defineRequiredFlagProject(cc []*cobra.Command) error {
	for i := range cc {
		defineFlagProject(cc[i])
		if err := cc[i].MarkFlagRequired(flagProject); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
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

func parsePipeSpecs(specs []string) ([]*models.V1PipeSpec, error) {
	pipes := []*models.V1PipeSpec{}
	for i := range specs {
		pipe, err := parsePipeSpec(specs[i])
		if err != nil {
			return nil, fmt.Errorf("parsing pipe spec %s: %w", specs[i], err)
		}
		pipes = append(pipes, pipe)
	}
	return pipes, nil
}

func projectUIDAndInstanceName(args []string) (string, string, error) {
	project := viper.GetString(flagProject)
	if project == "" {
		return "", "", errors.New("project UID missing")
	}
	if len(args) != 1 {
		return "", "", errors.New("There should be one and only one argument.")
	}
	return project, args[0], nil
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
