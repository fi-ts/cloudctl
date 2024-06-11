package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"

	"github.com/fi-ts/cloud-go/api/client/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newS3Cmd(c *config) *cobra.Command {
	s3Cmd := &cobra.Command{
		Use:   "s3",
		Short: "manage s3",
		Long:  "manages access to s3 storage located in different partitions",
	}
	s3DescribeCmd := &cobra.Command{
		Use:   "describe",
		Short: "describe an s3 user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.s3Describe()
		},
	}
	s3CreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create an s3 user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.s3Create()
		},
	}
	s3DeleteCmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"destroy", "rm", "remove"},
		Short:   "delete an s3 user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.s3Delete()
		},
	}
	s3ListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list s3 users",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.s3List()
		},
	}
	s3PartitionListCmd := &cobra.Command{
		Use:     "partitions",
		Short:   "list s3 partitions",
		Aliases: []string{"partition"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.s3ListPartitions()
		},
	}
	s3AddKeyCmd := &cobra.Command{
		Use:   "add-key",
		Short: "adds a key for an s3 user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.s3AddKey()
		},
	}
	s3RemoveKeyCmd := &cobra.Command{
		Use:   "remove-key",
		Short: "remove a key for an s3 user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.s3RemoveKey()
		},
	}

	s3CreateCmd.Flags().StringP("id", "i", "", "id of the s3 user [required]")
	s3CreateCmd.Flags().StringP("partition", "p", "", "name of s3 partition to create the s3 user in [required]")
	s3CreateCmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")
	s3CreateCmd.Flags().StringP("tenant", "t", "", "create s3 for given tenant, defaults to logged in tenant")
	s3CreateCmd.Flags().StringP("name", "n", "", "name of s3 user, only for display")
	s3CreateCmd.Flags().Int64("max-buckets", 0, "maximum number of buckets for the s3 user")
	s3CreateCmd.Flags().StringP("access-key", "", "", "specify the access key, otherwise will be generated")
	s3CreateCmd.Flags().StringP("secret-key", "", "", "specify the secret key, otherwise will be generated")
	genericcli.Must(s3CreateCmd.MarkFlagRequired("id"))
	genericcli.Must(s3CreateCmd.MarkFlagRequired("partition"))
	genericcli.Must(s3CreateCmd.MarkFlagRequired("project"))
	genericcli.Must(s3CreateCmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
	genericcli.Must(s3CreateCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))

	s3ListCmd.Flags().StringP("partition", "p", "", "name of s3 partition.")
	s3ListCmd.Flags().String("project", "", "id of the project that the s3 user belongs to")
	genericcli.Must(s3ListCmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
	genericcli.Must(s3ListCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))

	s3DescribeCmd.Flags().StringP("id", "i", "", "id of the s3 user [required]")
	s3DescribeCmd.Flags().StringP("partition", "p", "", "name of s3 partition where this user is in [required]")
	s3DescribeCmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")
	s3DescribeCmd.Flags().StringP("tenant", "t", "", "tenant of the s3 user, defaults to logged in tenant")
	s3DescribeCmd.Flags().StringP("for-client", "", "", "output suitable client configuration for either minio|s3cmd")
	genericcli.Must(s3DescribeCmd.MarkFlagRequired("id"))
	genericcli.Must(s3DescribeCmd.MarkFlagRequired("partition"))
	genericcli.Must(s3DescribeCmd.MarkFlagRequired("project"))
	genericcli.Must(s3DescribeCmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
	genericcli.Must(s3DescribeCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))

	s3DeleteCmd.Flags().StringP("id", "i", "", "id of the s3 user [required]")
	s3DeleteCmd.Flags().StringP("partition", "p", "", "name of s3 partition where this user is in [required]")
	s3DeleteCmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")
	s3DeleteCmd.Flags().StringP("tenant", "t", "", "tenant of the s3 user, defaults to logged in tenant")
	s3DeleteCmd.Flags().Bool("force", false, "forces s3 user deletion along with buckets and bucket objects even if those still exist (dangerous!)")
	genericcli.Must(s3DeleteCmd.MarkFlagRequired("id"))
	genericcli.Must(s3DeleteCmd.MarkFlagRequired("partition"))
	genericcli.Must(s3DeleteCmd.MarkFlagRequired("project"))
	genericcli.Must(s3DeleteCmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
	genericcli.Must(s3DeleteCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))

	s3AddKeyCmd.Flags().StringP("id", "i", "", "id of the s3 user [required]")
	s3AddKeyCmd.Flags().StringP("partition", "p", "", "name of s3 partition where this user is in [required]")
	s3AddKeyCmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")
	s3AddKeyCmd.Flags().StringP("tenant", "t", "", "tenant of the s3 user, defaults to logged in tenant")
	s3AddKeyCmd.Flags().StringP("access-key", "", "", "specify the access key, otherwise will be generated")
	s3AddKeyCmd.Flags().StringP("secret-key", "", "", "specify the secret key, otherwise will be generated")
	genericcli.Must(s3AddKeyCmd.MarkFlagRequired("id"))
	genericcli.Must(s3AddKeyCmd.MarkFlagRequired("partition"))
	genericcli.Must(s3AddKeyCmd.MarkFlagRequired("project"))
	genericcli.Must(s3AddKeyCmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
	genericcli.Must(s3AddKeyCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))

	s3RemoveKeyCmd.Flags().StringP("id", "i", "", "id of the s3 user [required]")
	s3RemoveKeyCmd.Flags().StringP("partition", "p", "", "name of s3 partition where this user is in [required]")
	s3RemoveKeyCmd.Flags().String("project", "", "id of the project that the s3 user belongs to [required]")
	s3RemoveKeyCmd.Flags().StringP("tenant", "t", "", "tenant of the s3 user, defaults to logged in tenant")
	s3RemoveKeyCmd.Flags().StringP("access-key", "", "", "specify the access key to delete the access / secret key pair")
	genericcli.Must(s3RemoveKeyCmd.MarkFlagRequired("id"))
	genericcli.Must(s3RemoveKeyCmd.MarkFlagRequired("partition"))
	genericcli.Must(s3RemoveKeyCmd.MarkFlagRequired("project"))
	genericcli.Must(s3RemoveKeyCmd.RegisterFlagCompletionFunc("partition", c.comp.S3ListPartitionsCompletion))
	genericcli.Must(s3RemoveKeyCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))

	s3Cmd.AddCommand(s3CreateCmd)
	s3Cmd.AddCommand(s3DescribeCmd)
	s3Cmd.AddCommand(s3DeleteCmd)
	s3Cmd.AddCommand(s3ListCmd)
	s3Cmd.AddCommand(s3PartitionListCmd)
	s3Cmd.AddCommand(s3AddKeyCmd)
	s3Cmd.AddCommand(s3RemoveKeyCmd)
	return s3Cmd
}

var s3cmdTemplate = `cat << EOF > ${HOME}/.s3cfg
[default]
access_key = %s
host_base = %s
host_bucket = %s
secret_key = %s
EOF
`

func (c *config) s3Describe() error {
	tenant := viper.GetString("tenant")
	id := viper.GetString("id")
	partition := viper.GetString("partition")
	project := viper.GetString("project")
	client := viper.GetString("for-client")

	p := &models.V1S3GetRequest{
		ID:        &id,
		Partition: &partition,
		Tenant:    &tenant,
		Project:   &project,
	}

	request := s3.NewGets3Params()
	request.SetBody(p)

	response, err := c.cloud.S3.Gets3(request, nil)
	if err != nil {
		return err
	}
	cfg := response.Payload
	switch client {
	case "":
	case "minio":
		fmt.Printf("mc config host add %s %s %s %s\n", *cfg.ID, *cfg.Endpoint, *cfg.Keys[0].AccessKey, *cfg.Keys[0].SecretKey)
		return nil
	case "s3cmd":
		fmt.Printf(s3cmdTemplate, *cfg.Keys[0].AccessKey, *cfg.Endpoint, *cfg.Endpoint, *cfg.Keys[0].SecretKey)
		return nil
	default:
		return fmt.Errorf("unsupported s3 client configuration:%s", client)
	}
	return c.describePrinter.Print(response.Payload)
}

func (c *config) s3Create() error {
	tenant := viper.GetString("tenant")
	id := viper.GetString("id")
	partition := viper.GetString("partition")
	project := viper.GetString("project")
	name := viper.GetString("name")
	maxBuckets := viper.GetInt64("max-buckets")
	accessKey := viper.GetString("access-key")
	secretKey := viper.GetString("secret-key")

	p := &models.V1S3CreateRequest{
		ID:         &id,
		Partition:  &partition,
		Tenant:     &tenant,
		Project:    &project,
		Name:       &name,
		MaxBuckets: &maxBuckets,
	}

	if accessKey != "" {
		p.Key = &models.V1S3Key{
			AccessKey: &accessKey,
			SecretKey: &secretKey,
		}
	}

	request := s3.NewCreates3Params()
	request.SetBody(p)

	response, err := c.cloud.S3.Creates3(request, nil)
	if err != nil {
		return err
	}

	return c.describePrinter.Print(response.Payload)
}

func (c *config) s3Delete() error {
	tenant := viper.GetString("tenant")
	id := viper.GetString("id")
	partition := viper.GetString("partition")
	project := viper.GetString("project")
	force := viper.GetBool("force")

	p := &models.V1S3DeleteRequest{
		ID:        &id,
		Partition: &partition,
		Tenant:    &tenant,
		Project:   &project,
		Force:     &force,
	}

	request := s3.NewDeletes3Params()
	request.SetBody(p)

	response, err := c.cloud.S3.Deletes3(request, nil)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("the server took too long to respond. for s3 users with a lot of data the deletion may take longer than configured request timeout. please check again if the deletion was carried out successfully at a later point in time.")
		}
		return err
	}

	return c.describePrinter.Print(response.Payload)
}

func (c *config) s3AddKey() error {
	tenant := viper.GetString("tenant")
	id := viper.GetString("id")
	partition := viper.GetString("partition")
	project := viper.GetString("project")
	accessKey := viper.GetString("access-key")
	secretKey := viper.GetString("secret-key")

	p := &models.V1S3UpdateRequest{
		ID:        &id,
		Partition: &partition,
		Tenant:    &tenant,
		Project:   &project,
		AddKeys: []*models.V1S3Key{
			{
				AccessKey: &accessKey,
				SecretKey: &secretKey,
			},
		},
	}

	request := s3.NewUpdates3Params()
	request.SetBody(p)

	response, err := c.cloud.S3.Updates3(request, nil)
	if err != nil {
		return err
	}

	return c.describePrinter.Print(response.Payload)
}

func (c *config) s3RemoveKey() error {
	tenant := viper.GetString("tenant")
	id := viper.GetString("id")
	partition := viper.GetString("partition")
	project := viper.GetString("project")
	accessKey := viper.GetString("access-key")

	p := &models.V1S3UpdateRequest{
		ID:        &id,
		Partition: &partition,
		Tenant:    &tenant,
		Project:   &project,
		RemoveAccessKeys: []string{
			accessKey,
		},
	}

	request := s3.NewUpdates3Params()
	request.SetBody(p)

	response, err := c.cloud.S3.Updates3(request, nil)
	if err != nil {
		return err
	}

	return c.describePrinter.Print(response.Payload)
}

func (c *config) s3List() error {
	partition := viper.GetString("partition")
	project := viper.GetString("project")

	p := &models.V1S3ListRequest{
		Partition: &partition,
	}

	request := s3.NewLists3Params()
	request.SetBody(p)

	response, err := c.cloud.S3.Lists3(request, nil)
	if err != nil {
		return err
	}

	if project == "" {
		return c.listPrinter.Print(response.Payload)
	}

	var result []*models.V1S3Response
	for _, s3 := range response.Payload {
		if *s3.Project == project {
			result = append(result, s3)
		}
	}
	return c.listPrinter.Print(result)
}

func (c *config) s3ListPartitions() error {
	request := s3.NewLists3partitionsParams()

	response, err := c.cloud.S3.Lists3partitions(request, nil)
	if err != nil {
		return err
	}
	return c.listPrinter.Print(response.Payload)
}
