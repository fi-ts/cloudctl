package cmd

import (
	"log"

	"git.f-i-ts.de/cloud-native/cloudctl/api/models"

	"git.f-i-ts.de/cloud-native/cloudctl/api/client/s3"
	"git.f-i-ts.de/cloud-native/cloudctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	s3Cmd = &cobra.Command{
		Use:   "s3",
		Short: "manage s3",
		Long:  "manges access to s3 storage located in different partitions",
	}
	s3DescribeCmd = &cobra.Command{
		Use:   "describe",
		Short: "describe an s3 user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return s3Describe()
		},
		PreRun: bindPFlags,
	}
	s3CreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create an s3 user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return s3Create()
		},
		PreRun: bindPFlags,
	}
	s3DeleteCmd = &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm", "delete"},
		Short:   "delete an s3 user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return s3Delete(args)
		},
		PreRun: bindPFlags,
	}
	s3ListCmd = &cobra.Command{
		Use:     "list",
		Short:   "list s3 users",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return s3List()
		},
		PreRun: bindPFlags,
	}
	s3PartitionListCmd = &cobra.Command{
		Use:     "partitions",
		Short:   "list s3 partitions",
		Aliases: []string{"partition"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return s3ListPartitions()
		},
		PreRun: bindPFlags,
	}
)

func init() {
	s3CreateCmd.Flags().StringP("name", "n", "", "name of s3 user. [required]")
	s3CreateCmd.Flags().StringP("partition", "p", "", "name of s3 partition. [required]")
	s3CreateCmd.Flags().StringP("tenant", "t", "", "create s3 for given tenant")
	err := s3CreateCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = s3CreateCmd.MarkFlagRequired("partition")
	if err != nil {
		log.Fatal(err.Error())
	}

	s3ListCmd.Flags().StringP("partition", "p", "", "name of s3 partition. [required]")
	err = s3ListCmd.MarkFlagRequired("partition")
	if err != nil {
		log.Fatal(err.Error())
	}

	s3DescribeCmd.Flags().StringP("name", "n", "", "name of s3 user. [required]")
	s3DescribeCmd.Flags().StringP("partition", "p", "", "name of s3 partition. [required]")
	s3DescribeCmd.Flags().StringP("tenant", "t", "", "create s3 for given tenant")
	err = s3DescribeCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = s3DescribeCmd.MarkFlagRequired("partition")
	if err != nil {
		log.Fatal(err.Error())
	}

	s3DeleteCmd.Flags().StringP("name", "n", "", "name of s3 user. [required]")
	s3DeleteCmd.Flags().StringP("partition", "p", "", "name of s3 partition. [required]")
	s3DeleteCmd.Flags().StringP("tenant", "t", "", "create s3 for given tenant")
	err = s3DeleteCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = s3DeleteCmd.MarkFlagRequired("partition")
	if err != nil {
		log.Fatal(err.Error())
	}

	s3Cmd.AddCommand(s3CreateCmd)
	s3Cmd.AddCommand(s3DescribeCmd)
	s3Cmd.AddCommand(s3DeleteCmd)
	s3Cmd.AddCommand(s3ListCmd)
	s3Cmd.AddCommand(s3PartitionListCmd)
}

func s3Describe() error {
	tenant := viper.GetString("tenant")
	name := viper.GetString("name")
	partition := viper.GetString("partition")

	p := &models.V1S3GetRequest{
		Name:      &name,
		Partition: &partition,
		Tenant:    &tenant,
	}

	request := s3.NewGets3Params()
	request.SetBody(p)

	response, err := cloud.S3.Gets3(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *s3.Gets3Default:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	return output.YAMLPrinter{}.Print(response.Payload)
}

func s3Create() error {
	tenant := viper.GetString("tenant")
	name := viper.GetString("name")
	partition := viper.GetString("partition")

	p := &models.V1S3CreateRequest{
		Name:      &name,
		Partition: &partition,
		Tenant:    &tenant,
	}

	request := s3.NewCreates3Params()
	request.SetBody(p)

	response, err := cloud.S3.Creates3(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *s3.Creates3Default:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	return output.YAMLPrinter{}.Print(response.Payload)
}

func s3Delete(args []string) error {
	tenant := viper.GetString("tenant")
	name := viper.GetString("name")
	partition := viper.GetString("partition")

	p := &models.V1S3DeleteRequest{
		Name:      &name,
		Partition: &partition,
		Tenant:    &tenant,
	}

	request := s3.NewDeletes3Params()
	request.SetBody(p)

	response, err := cloud.S3.Deletes3(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *s3.Deletes3Default:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}

	return output.YAMLPrinter{}.Print(response.Payload)
}

func s3List() error {
	partition := viper.GetString("partition")

	p := &models.V1S3ListRequest{
		Partition: &partition,
	}

	request := s3.NewLists3Params()
	request.SetBody(p)

	response, err := cloud.S3.Lists3(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *s3.Lists3Default:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	return printer.Print(response.Payload)
}

func s3ListPartitions() error {
	request := s3.NewLists3partitionsParams()

	response, err := cloud.S3.Lists3partitions(request, cloud.Auth)
	if err != nil {
		switch e := err.(type) {
		case *s3.Lists3partitionsDefault:
			return output.HTTPError(e.Payload)
		default:
			return output.UnconventionalError(err)
		}
	}
	return printer.Print(response.Payload)
}
