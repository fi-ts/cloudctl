package cmd

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/fi-ts/cloud-go/api/client/s3"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloud-go/test/client"
	"github.com/metal-stack/metal-lib/httperrors"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	s31Response = &models.V1S3Response{
		Endpoint:  pointer.Pointer("https://endpoint-a"),
		ID:        pointer.Pointer("s3-1"),
		Partition: pointer.Pointer("partition-a"),
		Project:   pointer.Pointer("project-a"),
		Tenant:    pointer.Pointer("fits"),
	}
	s31 = &models.V1S3CredentialsResponse{
		Endpoint:  s31Response.Endpoint,
		ID:        s31Response.ID,
		Partition: s31Response.Partition,
		Project:   s31Response.Project,
		Tenant:    s31Response.Tenant,
		Keys: []*models.V1S3Key{
			{
				AccessKey: pointer.Pointer("access-key-a"),
				SecretKey: pointer.Pointer("secret-key-a"),
			},
		},
		MaxBuckets: pointer.Pointer(int64(100)),
		Name:       pointer.Pointer("s3 1"),
	}
	s32Response = &models.V1S3Response{
		Endpoint:  pointer.Pointer("https://endpoint-b"),
		ID:        pointer.Pointer("s3-2"),
		Partition: pointer.Pointer("partition-a"),
		Project:   pointer.Pointer("project-a"),
		Tenant:    pointer.Pointer("fits"),
	}
	s32 = &models.V1S3CredentialsResponse{
		Endpoint:  s32Response.Endpoint,
		ID:        s32Response.ID,
		Partition: s32Response.Partition,
		Project:   s32Response.Project,
		Tenant:    s32Response.Tenant,
		Keys: []*models.V1S3Key{
			{
				AccessKey: pointer.Pointer("access-key-b"),
				SecretKey: pointer.Pointer("secret-key-b"),
			},
		},
		MaxBuckets: pointer.Pointer(int64(200)),
		Name:       pointer.Pointer("s3 2"),
	}
)

func Test_S3Cmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1S3CredentialsResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1S3CredentialsResponse) []string {
				return []string{"s3", "list"}
			},
			mocks: &client.CloudMockFns{
				S3: func(mock *mock.Mock) {
					mock.On("Lists3", testcommon.MatchIgnoreContext(t, s3.NewLists3Params().WithBody(&models.V1S3ListRequest{})), nil).Return(&s3.Lists3OK{
						Payload: []*models.V1S3Response{
							s32Response,
							s31Response,
						},
					}, nil)
				},
			},
			want: []*models.V1S3CredentialsResponse{
				s3ResponseToCredentialsResponse(s31Response),
				s3ResponseToCredentialsResponse(s32Response),
			},
			wantTable: pointer.Pointer(`
ID     TENANT   PROJECT     PARTITION     ENDPOINT
s3-1   fits     project-a   partition-a   https://endpoint-a
s3-2   fits     project-a   partition-a   https://endpoint-b
`),
			wantWideTable: pointer.Pointer(`
ID     TENANT   PROJECT     PARTITION     ENDPOINT
s3-1   fits     project-a   partition-a   https://endpoint-a
s3-2   fits     project-a   partition-a   https://endpoint-b
`),
			template: pointer.Pointer("{{ .id }}"),
			wantTemplate: pointer.Pointer(`
s3-1
s3-2
`),
			wantMarkdown: pointer.Pointer(`
|  ID  | TENANT |  PROJECT  |  PARTITION  |      ENDPOINT      |
|------|--------|-----------|-------------|--------------------|
| s3-1 | fits   | project-a | partition-a | https://endpoint-a |
| s3-2 | fits   | project-a | partition-a | https://endpoint-b |
`),
		},
		{
			name: "list with filters",
			cmd: func(want []*models.V1S3CredentialsResponse) []string {
				args := []string{"s3", "list", "--project", *want[0].Project, "--partition", *want[0].Partition, "--tenant", *want[0].Tenant}
				assertExhaustiveArgs(t, args, "sort-by", "id")
				return args
			},
			mocks: &client.CloudMockFns{
				S3: func(mock *mock.Mock) {
					mock.On("Lists3", testcommon.MatchIgnoreContext(t, s3.NewLists3Params().WithBody(&models.V1S3ListRequest{
						Partition: s31.Partition,
					})), nil).Return(&s3.Lists3OK{
						Payload: []*models.V1S3Response{
							s32Response,
							s31Response,
						},
					}, nil)
				},
			},
			want: []*models.V1S3CredentialsResponse{
				s3ResponseToCredentialsResponse(s31Response),
				s3ResponseToCredentialsResponse(s32Response),
			},
			wantTable: pointer.Pointer(`
ID     TENANT   PROJECT     PARTITION     ENDPOINT
s3-1   fits     project-a   partition-a   https://endpoint-a
s3-2   fits     project-a   partition-a   https://endpoint-b
`),
			wantWideTable: pointer.Pointer(`
ID     TENANT   PROJECT     PARTITION     ENDPOINT
s3-1   fits     project-a   partition-a   https://endpoint-a
s3-2   fits     project-a   partition-a   https://endpoint-b
`),
			template: pointer.Pointer("{{ .id }}"),
			wantTemplate: pointer.Pointer(`
s3-1
s3-2
`),
			wantMarkdown: pointer.Pointer(`
|  ID  | TENANT |  PROJECT  |  PARTITION  |      ENDPOINT      |
|------|--------|-----------|-------------|--------------------|
| s3-1 | fits   | project-a | partition-a | https://endpoint-a |
| s3-2 | fits   | project-a | partition-a | https://endpoint-b |
`),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1S3CredentialsResponse) []string {
				return []string{"s3", "apply", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1S3CredentialsResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.CloudMockFns{
				S3: func(mock *mock.Mock) {
					mock.On("Creates3", testcommon.MatchIgnoreContext(t, s3.NewCreates3Params().WithBody(s3ResponseToCreate(s31))), nil).Return(nil, &s3.Creates3Default{Payload: httperrors.Conflict(fmt.Errorf("already exists"))}).Once()
					mock.On("Updates3", testcommon.MatchIgnoreContext(t, s3.NewUpdates3Params().WithBody(s3ResponseToUpdate(s31))), nil).Return(&s3.Updates3OK{
						Payload: s31,
					}, nil)
					mock.On("Creates3", testcommon.MatchIgnoreContext(t, s3.NewCreates3Params().WithBody(s3ResponseToCreate(s32))), nil).Return(&s3.Creates3OK{
						Payload: s32,
					}, nil)
				},
			},
			want: []*models.V1S3CredentialsResponse{
				s31,
				s32,
			},
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1S3CredentialsResponse) []string {
				return []string{"s3", "create", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1S3CredentialsResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.CloudMockFns{
				S3: func(mock *mock.Mock) {
					mock.On("Creates3", testcommon.MatchIgnoreContext(t, s3.NewCreates3Params().WithBody(s3ResponseToCreate(s31))), nil).Return(&s3.Creates3OK{
						Payload: s31,
					}, nil)
				},
			},
			want: []*models.V1S3CredentialsResponse{
				s31,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1S3CredentialsResponse) []string {
				return []string{"s3", "update", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1S3CredentialsResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.CloudMockFns{
				S3: func(mock *mock.Mock) {
					mock.On("Updates3", testcommon.MatchIgnoreContext(t, s3.NewUpdates3Params().WithBody(s3ResponseToUpdate(s31))), nil).Return(&s3.Updates3OK{
						Payload: s31,
					}, nil)
				},
			},
			want: []*models.V1S3CredentialsResponse{
				s31,
			},
		},
		{
			name: "delete from file",
			cmd: func(want []*models.V1S3CredentialsResponse) []string {
				return []string{"s3", "delete", "-f", "/file.yaml", "--project", *want[0].Project, "--partition", *want[0].Partition}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1S3CredentialsResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.CloudMockFns{
				S3: func(mock *mock.Mock) {
					mock.On("Deletes3", testcommon.MatchIgnoreContext(t, s3.NewDeletes3Params().WithBody(&models.V1S3DeleteRequest{
						ID:        s31.ID,
						Project:   s31.Project,
						Partition: s31.Partition,
						Force:     pointer.Pointer(false),
					})), nil).Return(&s3.Deletes3OK{
						Payload: s31Response,
					}, nil)
				},
			},
			want: []*models.V1S3CredentialsResponse{
				s3ResponseToCredentialsResponse(s31Response),
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_S3Cmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1S3CredentialsResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1S3CredentialsResponse) []string {
				return []string{"s3", "describe", *want.ID, "--project", *want.Project, "--partition", *want.Partition}
			},
			mocks: &client.CloudMockFns{
				S3: func(mock *mock.Mock) {
					mock.On("Gets3", testcommon.MatchIgnoreContext(t, s3.NewGets3Params().WithBody(&models.V1S3GetRequest{
						ID:        s31.ID,
						Project:   s31.Project,
						Partition: s31.Partition,
					})), nil).Return(&s3.Gets3OK{
						Payload: s31,
					}, nil)
				},
			},
			want: s31,
			wantTable: pointer.Pointer(`
ID     TENANT   PROJECT     PARTITION     ENDPOINT
s3-1   fits     project-a   partition-a   https://endpoint-a
`),
			wantWideTable: pointer.Pointer(`
ID     TENANT   PROJECT     PARTITION     ENDPOINT
s3-1   fits     project-a   partition-a   https://endpoint-a
`),
			template: pointer.Pointer("{{ .id }}"),
			wantTemplate: pointer.Pointer(`
s3-1
`),
			wantMarkdown: pointer.Pointer(`
|  ID  | TENANT |  PROJECT  |  PARTITION  |      ENDPOINT      |
|------|--------|-----------|-------------|--------------------|
| s3-1 | fits   | project-a | partition-a | https://endpoint-a |
`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1S3CredentialsResponse) []string {
				return []string{"s3", "rm", *want.ID, "--project", *want.Project, "--partition", *want.Partition}
			},
			mocks: &client.CloudMockFns{
				S3: func(mock *mock.Mock) {
					mock.On("Deletes3", testcommon.MatchIgnoreContext(t, s3.NewDeletes3Params().WithBody(&models.V1S3DeleteRequest{
						ID:        s31.ID,
						Project:   s31.Project,
						Partition: s31.Partition,
						Force:     pointer.Pointer(false),
					})), nil).Return(&s3.Deletes3OK{
						Payload: s31Response,
					}, nil)
				},
			},
			want: s3ResponseToCredentialsResponse(s31Response),
		},
		{
			name: "create",
			cmd: func(want *models.V1S3CredentialsResponse) []string {
				args := []string{"s3", "create",
					"--id", *want.ID,
					"--partition", *want.Partition,
					"--project", *want.Project,
					"--tenant", *want.Tenant,
					"--name", *want.Name,
					"--max-buckets", strconv.FormatInt(int64(*want.MaxBuckets), 10),
					"--access-key", *want.Keys[0].AccessKey,
					"--secret-key", *want.Keys[0].SecretKey,
				}
				assertExhaustiveArgs(t, args, "file", "bulk-output")
				return args
			},
			mocks: &client.CloudMockFns{
				S3: func(mock *mock.Mock) {
					mock.On("Creates3", testcommon.MatchIgnoreContext(t, s3.NewCreates3Params().WithBody(s3ResponseToCreate(s31))), nil).Return(&s3.Creates3OK{
						Payload: s31,
					}, nil)
				},
			},
			want: s31,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
