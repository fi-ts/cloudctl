package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/sorters"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"

	"github.com/fi-ts/cloud-go/api/client/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type s3Cmd struct {
	*config
}

func newS3Cmd(c *config) *cobra.Command {
	w := s3Cmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1S3CreateRequest, *models.V1S3UpdateRequest, *models.V1S3CredentialsResponse]{
		BinaryName:      binaryName,
		GenericCLI:      genericcli.NewGenericCLI[*models.V1S3CreateRequest, *models.V1S3UpdateRequest, *models.V1S3CredentialsResponse](w).WithFS(c.fs),
		Singular:        "s3",
		Plural:          "s3",
		Description:     "manages s3 users to access s3 storage located in different partitions.",
		Sorter:          sorters.S3Sorter(),
		ValidArgsFn:     c.comp.S3ListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		CreateRequestFromCLI: func() (*models.V1S3CreateRequest, error) {
			p := &models.V1S3CreateRequest{
				ID:         pointer.Pointer(viper.GetString("id")),
				Partition:  pointer.Pointer(viper.GetString("partition")),
				Tenant:     pointer.Pointer(viper.GetString("tenant")),
				Project:    pointer.Pointer(viper.GetString("project")),
				Name:       pointer.Pointer(viper.GetString("name")),
				MaxBuckets: pointer.Pointer(viper.GetInt64("max-buckets")),
			}

			accessKey := viper.GetString("access-key")
			secretKey := viper.GetString("secret-key")
			if accessKey != "" {
				p.Key = &models.V1S3Key{
					AccessKey: &accessKey,
					SecretKey: &secretKey,
				}
			}

			return p, nil
		},
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("id", "i", "", "id of the s3 user [required]")
			cmd.Flags().StringP("partition", "p", "", "name of s3 partition to create the s3 user in [required]")
			cmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")
			cmd.Flags().StringP("tenant", "t", "", "create s3 for given tenant, defaults to logged in tenant")
			cmd.Flags().StringP("name", "n", "", "name of s3 user, only for display")
			cmd.Flags().Int64("max-buckets", 0, "maximum number of buckets for the s3 user")
			cmd.Flags().StringP("access-key", "", "", "specify the access key, otherwise will be generated")
			cmd.Flags().StringP("secret-key", "", "", "specify the secret key, otherwise will be generated")

			cmd.MarkFlagsMutuallyExclusive("file", "id")
			cmd.MarkFlagsRequiredTogether("id", "partition", "project")
			must(cmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("id", "", "id of s3 user.")
			cmd.Flags().String("project", "", "project of s3 user.")
			cmd.Flags().String("tenant", "", "tenant of s3 user.")
			cmd.Flags().String("partition", "", "name of s3 partition.")

			must(cmd.RegisterFlagCompletionFunc("id", c.comp.S3ListCompletion))
			must(cmd.RegisterFlagCompletionFunc("tenant", c.comp.TenantListCompletion))
			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			must(cmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
		},
		DescribeCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("partition", "p", "", "name of s3 partition where this user is in [required]")
			cmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")
			cmd.Flags().StringP("tenant", "t", "", "tenant of the s3 user, defaults to logged in tenant")
			cmd.Flags().StringP("for-client", "", "", "output suitable client configuration for either minio|s3cmd")

			must(cmd.MarkFlagRequired("partition"))
			must(cmd.MarkFlagRequired("project"))

			must(cmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
		},
		DeleteCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("partition", "p", "", "name of s3 partition where this user is in [required]")
			cmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")
			cmd.Flags().StringP("tenant", "t", "", "tenant of the s3 user, defaults to logged in tenant")
			cmd.Flags().Bool("force-delete", false, "forces s3 user deletion along with buckets and bucket objects even if those still exist (dangerous!)")

			must(cmd.MarkFlagRequired("partition"))
			must(cmd.MarkFlagRequired("project"))

			must(cmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
		},
		EditCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("partition", "p", "", "name of s3 partition where this user is in [required]")
			cmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")

			must(cmd.MarkFlagRequired("partition"))
			must(cmd.MarkFlagRequired("project"))

			must(cmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
		},
	}

	s3PartitionListCmd := &cobra.Command{
		Use:     "partitions",
		Short:   "list s3 partitions",
		Aliases: []string{"partition"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.listPartitions()
		},
	}
	s3AddKeyCmd := &cobra.Command{
		Use:   "add-key <id>",
		Short: "adds a key for an s3 user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.addKey(args)
		},
		ValidArgsFunction: c.comp.S3ListCompletion,
	}
	s3RemoveKeyCmd := &cobra.Command{
		Use:   "remove-key <id>",
		Short: "remove a key for an s3 user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.removeKey(args)
		},
		ValidArgsFunction: c.comp.S3ListCompletion,
	}
	s3ClientConfigCmd := &cobra.Command{
		Use:     "client-config <id>",
		Short:   "returns fitting configuration of an s3 user for given client",
		Aliases: []string{"partition"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.clientConfig(args)
		},
		ValidArgsFunction: c.comp.S3ListCompletion,
	}

	s3AddKeyCmd.Flags().StringP("partition", "p", "", "name of s3 partition where this user is in [required]")
	s3AddKeyCmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")
	s3AddKeyCmd.Flags().StringP("tenant", "t", "", "tenant of the s3 user, defaults to logged in tenant")
	s3AddKeyCmd.Flags().StringP("access-key", "", "", "specify the access key, otherwise will be generated")
	s3AddKeyCmd.Flags().StringP("secret-key", "", "", "specify the secret key, otherwise will be generated")
	must(s3AddKeyCmd.MarkFlagRequired("partition"))
	must(s3AddKeyCmd.MarkFlagRequired("project"))
	must(s3AddKeyCmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
	must(s3AddKeyCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))

	s3RemoveKeyCmd.Flags().StringP("partition", "p", "", "name of s3 partition where this user is in [required]")
	s3RemoveKeyCmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")
	s3RemoveKeyCmd.Flags().StringP("tenant", "t", "", "tenant of the s3 user, defaults to logged in tenant")
	s3RemoveKeyCmd.Flags().StringP("access-key", "", "", "specify the access key to delete the access / secret key pair")
	must(s3RemoveKeyCmd.MarkFlagRequired("partition"))
	must(s3RemoveKeyCmd.MarkFlagRequired("project"))
	must(s3RemoveKeyCmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
	must(s3RemoveKeyCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))

	s3ClientConfigCmd.Flags().StringP("partition", "p", "", "name of s3 partition where this user is in [required]")
	s3ClientConfigCmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")
	s3ClientConfigCmd.Flags().StringP("tenant", "t", "", "tenant of the s3 user, defaults to logged in tenant")
	s3ClientConfigCmd.Flags().StringP("for-client", "", "minio", "output suitable client configuration for either minio|s3cmd")

	must(s3ClientConfigCmd.MarkFlagRequired("partition"))
	must(s3ClientConfigCmd.MarkFlagRequired("project"))

	must(s3ClientConfigCmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
	must(s3ClientConfigCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(s3ClientConfigCmd.RegisterFlagCompletionFunc("for-client", cobra.FixedCompletions([]string{"minio", "s3cmd"}, cobra.ShellCompDirectiveDefault)))

	return genericcli.NewCmds(cmdsConfig, s3PartitionListCmd, s3AddKeyCmd, s3RemoveKeyCmd, s3ClientConfigCmd)
}

func (c s3Cmd) Get(id string) (*models.V1S3CredentialsResponse, error) {
	response, err := c.client.S3.Gets3(s3.NewGets3Params().WithBody(&models.V1S3GetRequest{
		ID:        &id,
		Partition: pointer.PointerOrNil(viper.GetString("partition")),
		Tenant:    pointer.PointerOrNil(viper.GetString("tenant")),
		Project:   pointer.PointerOrNil(viper.GetString("project")),
	}), nil)
	if err != nil {
		return nil, err
	}

	return response.Payload, nil
}

func (c s3Cmd) List() ([]*models.V1S3CredentialsResponse, error) {
	if viper.IsSet("id") {
		resp, err := c.Get(viper.GetString("id"))
		if err != nil {
			return nil, err
		}

		return pointer.WrapInSlice(resp), nil
	}

	response, err := c.client.S3.Lists3(s3.NewLists3Params().WithBody(&models.V1S3ListRequest{
		Partition: pointer.PointerOrNil(viper.GetString("partition")),
	}), nil)
	if err != nil {
		return nil, err
	}

	var result []*models.V1S3CredentialsResponse
	for _, resp := range response.Payload {
		resp := resp

		// TODO: move to filtering api
		if viper.IsSet("project") && viper.GetString("project") != *resp.Project {
			continue
		}
		if viper.IsSet("tenant") && viper.GetString("tenant") != *resp.Tenant {
			continue
		}

		result = append(result, s3ResponseToCredentialsResponse(resp))
	}

	return result, nil
}

func (c s3Cmd) Delete(id string) (*models.V1S3CredentialsResponse, error) {
	response, err := c.client.S3.Deletes3(s3.NewDeletes3Params().WithBody(&models.V1S3DeleteRequest{
		ID:        &id,
		Partition: pointer.PointerOrNil(viper.GetString("partition")),
		Tenant:    pointer.PointerOrNil(viper.GetString("tenant")),
		Project:   pointer.PointerOrNil(viper.GetString("project")),
		Force:     pointer.Pointer(viper.GetBool("force-delete")),
	}), nil)
	if err != nil {
		return nil, err
	}

	return s3ResponseToCredentialsResponse(response.Payload), nil
}

func (c s3Cmd) Create(rq *models.V1S3CreateRequest) (*models.V1S3CredentialsResponse, error) {
	response, err := c.client.S3.Creates3(s3.NewCreates3Params().WithBody(rq), nil)
	if err != nil {
		var r *s3.Creates3Default
		if errors.As(err, &r) && r.Payload.StatusCode == http.StatusConflict {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return response.Payload, nil
}

func (c s3Cmd) Update(rq *models.V1S3UpdateRequest) (*models.V1S3CredentialsResponse, error) {
	response, err := c.client.S3.Updates3(s3.NewUpdates3Params().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return response.Payload, nil
}

func (c s3Cmd) Convert(r *models.V1S3CredentialsResponse) (string, *models.V1S3CreateRequest, *models.V1S3UpdateRequest, error) {
	if r.ID == nil {
		return "", nil, nil, fmt.Errorf("id is nil")
	}
	return *r.ID, s3ResponseToCreate(r), s3ResponseToUpdate(r), nil
}

func s3ResponseToCredentialsResponse(r *models.V1S3Response) *models.V1S3CredentialsResponse {
	return &models.V1S3CredentialsResponse{
		Endpoint:  r.Endpoint,
		ID:        r.ID,
		Partition: r.Partition,
		Project:   r.Project,
		Tenant:    r.Tenant,
	}
}

func s3ResponseToCreate(r *models.V1S3CredentialsResponse) *models.V1S3CreateRequest {
	var key *models.V1S3Key
	if len(r.Keys) > 0 {
		key = &models.V1S3Key{
			AccessKey: r.Keys[0].AccessKey,
			SecretKey: r.Keys[0].SecretKey,
		}
	}

	return &models.V1S3CreateRequest{
		ID:         r.ID,
		MaxBuckets: r.MaxBuckets,
		Name:       r.Name,
		Partition:  r.Partition,
		Project:    r.Project,
		Tenant:     r.Tenant,
		Key:        key,
	}
}

func s3ResponseToUpdate(r *models.V1S3CredentialsResponse) *models.V1S3UpdateRequest {
	return &models.V1S3UpdateRequest{
		ID:        r.ID,
		Partition: r.Partition,
		Project:   r.Project,
		Tenant:    r.Tenant,
	}
}

func (c *s3Cmd) addKey(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	response, err := c.client.S3.Updates3(s3.NewUpdates3Params().WithBody(&models.V1S3UpdateRequest{
		ID:        &id,
		Partition: pointer.Pointer(viper.GetString("partition")),
		Tenant:    pointer.Pointer(viper.GetString("tenant")),
		Project:   pointer.Pointer(viper.GetString("project")),
		AddKeys: []*models.V1S3Key{
			{
				AccessKey: pointer.Pointer(viper.GetString("access-key")),
				SecretKey: pointer.Pointer(viper.GetString("secret-key")),
			},
		},
	}), nil)
	if err != nil {
		return err
	}

	return c.describePrinter.Print(response.Payload)
}

func (c *s3Cmd) removeKey(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	response, err := c.client.S3.Updates3(s3.NewUpdates3Params().WithBody(&models.V1S3UpdateRequest{
		ID:        &id,
		Partition: pointer.Pointer(viper.GetString("partition")),
		Tenant:    pointer.Pointer(viper.GetString("tenant")),
		Project:   pointer.Pointer(viper.GetString("project")),
		RemoveAccessKeys: []string{
			viper.GetString("access-key"),
		},
	}), nil)
	if err != nil {
		return err
	}

	return c.describePrinter.Print(response.Payload)
}

func (c *s3Cmd) listPartitions() error {
	response, err := c.client.S3.Lists3partitions(s3.NewLists3partitionsParams(), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(response.Payload)
}

func (c s3Cmd) clientConfig(args []string) error {
	var s3cmdTemplate = `cat << EOF > ${HOME}/.s3cfg
[default]
access_key = %s
host_base = %s
host_bucket = %s
secret_key = %s
EOF
`

	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	response, err := c.Get(id)
	if err != nil {
		return err
	}

	switch client := viper.GetString("for-client"); client {
	case "minio":
		fmt.Fprintf(c.config.out, "mc config host add %s %s %s %s\n", *response.ID, *response.Endpoint, *response.Keys[0].AccessKey, *response.Keys[0].SecretKey)
		return nil
	case "s3cmd":
		fmt.Fprintf(c.config.out, s3cmdTemplate, *response.Keys[0].AccessKey, *response.Endpoint, *response.Endpoint, *response.Keys[0].SecretKey)
		return nil
	default:
		return fmt.Errorf("unsupported s3 client configuration:%s", client)
	}
}
