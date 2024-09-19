package cmd

import (
	"testing"

	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/fi-ts/cloud-go/api/models"
	testclient "github.com/fi-ts/cloud-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/stretchr/testify/mock"
)

var (
	machineReservation1 = &models.V1MachineReservationResponse{
		Amount:      pointer.Pointer(int32(3)),
		Description: "for firewalls",
		Labels: map[string]string{
			"size.metal-stack.io/reserved-at": "2024-09-19T08:57:40Z",
			"size.metal-stack.io/reserved-by": "fits",
		},
		Partitionids: []string{"partition-a"},
		Projectid:    pointer.Pointer("project-a"),
		Sizeid:       pointer.Pointer("size-a"),
		Tenant:       pointer.Pointer("fits"),
	}
	machineReservation2 = &models.V1MachineReservationResponse{
		Amount:      pointer.Pointer(int32(3)),
		Description: "for machines",
		Labels: map[string]string{
			"size.metal-stack.io/reserved-by": "fits",
		},
		Partitionids: []string{"partition-a", "partition-b"},
		Projectid:    pointer.Pointer("project-b"),
		Sizeid:       pointer.Pointer("size-b"),
		Tenant:       pointer.Pointer("fits"),
	}
)

func Test_ProjectMachineReservationsCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1MachineReservationResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1MachineReservationResponse) []string {
				return []string{"project", "machine-reservation", "list"}
			},
			mocks: &testclient.CloudMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("ListMachineReservations", testcommon.MatchIgnoreContext(t, project.NewListMachineReservationsParams().WithBody(&models.V1MachineReservationFindRequest{})), nil).Return(&project.ListMachineReservationsOK{
						Payload: []*models.V1MachineReservationResponse{
							machineReservation2,
							machineReservation1,
						},
					}, nil)
				},
			},
			want: []*models.V1MachineReservationResponse{
				machineReservation1,
				machineReservation2,
			},
			wantTable: pointer.Pointer(`
TENANT   PROJECT     SIZE     AMOUNT   PARTITIONS                DESCRIPTION
fits     project-a   size-a   3        partition-a               for firewalls
fits     project-b   size-b   3        partition-a,partition-b   for machines
`),
			wantWideTable: pointer.Pointer(`
TENANT   PROJECT     SIZE     AMOUNT   PARTITIONS                DESCRIPTION     LABELS
fits     project-a   size-a   3        partition-a               for firewalls   for firewalls   size.metal-stack.io/reserved-at=2024-09-19T08:57:40Z
                                                                                                    size.metal-stack.io/reserved-by=fits
fits     project-b   size-b   3        partition-a,partition-b   for machines    for machines    size.metal-stack.io/reserved-by=fits
`),
			template: pointer.Pointer("{{ .sizeid }} {{ .projectid }}"),
			wantTemplate: pointer.Pointer(`
size-a project-a
size-b project-b
`),
			wantMarkdown: pointer.Pointer(`
| TENANT |  PROJECT  |  SIZE  | AMOUNT |       PARTITIONS        |  DESCRIPTION  |
|--------|-----------|--------|--------|-------------------------|---------------|
| fits   | project-a | size-a |      3 | partition-a             | for firewalls |
| fits   | project-b | size-b |      3 | partition-a,partition-b | for machines  |
`),
		},
		// 		{
		// 			name: "list with filters",
		// 			cmd: func(want []*models.V1MachineReservationResponse) []string {
		// 				args := []string{"project", "list", "--name", "project-1", "--tenant", "metal-stack", "--id", want[0].Meta.ID}
		// 				assertExhaustiveArgs(t, args, "sort-by")
		// 				return args
		// 			},
		// 			mocks: &client.MetalMockFns{
		// 				Project: func(mock *mock.Mock) {
		// 					mock.On("FindProjects", testcommon.MatchIgnoreContext(t, project.NewFindProjectsParams().WithBody(&models.V1ProjectFindRequest{
		// 						Name:     "project-1",
		// 						TenantID: "metal-stack",
		// 						ID:       "1",
		// 					})), nil).Return(&project.FindProjectsOK{
		// 						Payload: []*models.V1ProjectResponse{
		// 							project1,
		// 						},
		// 					}, nil)
		// 				},
		// 			},
		// 			want: []*models.V1ProjectResponse{
		// 				project1,
		// 			},
		// 			wantTable: pointer.Pointer(`
		// UID   TENANT        NAME        DESCRIPTION   LABELS   ANNOTATIONS
		// 1     metal-stack   project-1   project 1     c        a=b
		// `),
		// 			wantWideTable: pointer.Pointer(`
		// UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
		// 1     metal-stack   project-1   project 1     1/3/2                          c        a=b
		// `),
		// 			template: pointer.Pointer("{{ .meta.id }} {{ .name }}"),
		// 			wantTemplate: pointer.Pointer(`
		// 1 project-1
		// `),
		// 			wantMarkdown: pointer.Pointer(`
		// | UID |   TENANT    |   NAME    | DESCRIPTION | LABELS | ANNOTATIONS |
		// |-----|-------------|-----------|-------------|--------|-------------|
		// |   1 | metal-stack | project-1 | project 1   | c      | a=b         |
		// `),
		// 		},
		// 		{
		// 			name: "apply",
		// 			cmd: func(want []*models.V1ProjectResponse) []string {
		// 				return appendFromFileCommonArgs("project", "apply")
		// 			},
		// 			fsMocks: func(fs afero.Fs, want []*models.V1ProjectResponse) {
		// 				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
		// 			},
		// 			mocks: &client.MetalMockFns{
		// 				Project: func(mock *mock.Mock) {
		// 					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(projectResponseToCreate(project1))), nil).Return(nil, &project.CreateProjectConflict{}).Once()
		// 					mock.On("FindProject", testcommon.MatchIgnoreContext(t, project.NewFindProjectParams().WithID(project1.Meta.ID)), nil).Return(&project.FindProjectOK{
		// 						Payload: project1,
		// 					}, nil)
		// 					mock.On("UpdateProject", testcommon.MatchIgnoreContext(t, project.NewUpdateProjectParams().WithBody(projectResponseToUpdate(project1))), nil).Return(&project.UpdateProjectOK{
		// 						Payload: project1,
		// 					}, nil)
		// 					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(projectResponseToCreate(project2))), nil).Return(&project.CreateProjectCreated{
		// 						Payload: project2,
		// 					}, nil)
		// 				},
		// 			},
		// 			want: []*models.V1ProjectResponse{
		// 				project1,
		// 				project2,
		// 			},
		// 		},
		// 		{
		// 			name: "create from file",
		// 			cmd: func(want []*models.V1ProjectResponse) []string {
		// 				return appendFromFileCommonArgs("project", "create")
		// 			},
		// 			fsMocks: func(fs afero.Fs, want []*models.V1ProjectResponse) {
		// 				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
		// 			},
		// 			mocks: &client.MetalMockFns{
		// 				Project: func(mock *mock.Mock) {
		// 					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(projectResponseToCreate(project1))), nil).Return(&project.CreateProjectCreated{
		// 						Payload: project1,
		// 					}, nil)
		// 				},
		// 			},
		// 			want: []*models.V1ProjectResponse{
		// 				project1,
		// 			},
		// 		},
		// 		{
		// 			name: "update from file",
		// 			cmd: func(want []*models.V1ProjectResponse) []string {
		// 				return appendFromFileCommonArgs("project", "update")
		// 			},
		// 			fsMocks: func(fs afero.Fs, want []*models.V1ProjectResponse) {
		// 				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
		// 			},
		// 			mocks: &client.MetalMockFns{
		// 				Project: func(mock *mock.Mock) {
		// 					mock.On("FindProject", testcommon.MatchIgnoreContext(t, project.NewFindProjectParams().WithID(project1.Meta.ID)), nil).Return(&project.FindProjectOK{
		// 						Payload: project1,
		// 					}, nil)
		// 					mock.On("UpdateProject", testcommon.MatchIgnoreContext(t, project.NewUpdateProjectParams().WithBody(projectResponseToUpdate(project1))), nil).Return(&project.UpdateProjectOK{
		// 						Payload: project1,
		// 					}, nil)
		// 				},
		// 			},
		// 			want: []*models.V1ProjectResponse{
		// 				project1,
		// 			},
		// 		},
		// 		{
		// 			name: "delete from file",
		// 			cmd: func(want []*models.V1ProjectResponse) []string {
		// 				return appendFromFileCommonArgs("project", "delete")
		// 			},
		// 			fsMocks: func(fs afero.Fs, want []*models.V1ProjectResponse) {
		// 				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
		// 			},
		// 			mocks: &client.MetalMockFns{
		// 				Project: func(mock *mock.Mock) {
		// 					mock.On("DeleteProject", testcommon.MatchIgnoreContext(t, project.NewDeleteProjectParams().WithID(project1.Meta.ID)), nil).Return(&project.DeleteProjectOK{
		// 						Payload: project1,
		// 					}, nil)
		// 				},
		// 			},
		// 			want: []*models.V1ProjectResponse{
		// 				project1,
		// 			},
		// 		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

// func Test_ProjectCmd_SingleResult(t *testing.T) {
// 	tests := []*test[*models.V1ProjectResponse]{
// 		{
// 			name: "describe",
// 			cmd: func(want *models.V1ProjectResponse) []string {
// 				return []string{"project", "describe", want.Meta.ID}
// 			},
// 			mocks: &client.MetalMockFns{
// 				Project: func(mock *mock.Mock) {
// 					mock.On("FindProject", testcommon.MatchIgnoreContext(t, project.NewFindProjectParams().WithID(project1.Meta.ID)), nil).Return(&project.FindProjectOK{
// 						Payload: project1,
// 					}, nil)
// 				},
// 			},
// 			want: project1,
// 			wantTable: pointer.Pointer(`
// UID   TENANT        NAME        DESCRIPTION   LABELS   ANNOTATIONS
// 1     metal-stack   project-1   project 1     c        a=b
// `),
// 			wantWideTable: pointer.Pointer(`
// UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
// 1     metal-stack   project-1   project 1     1/3/2                          c        a=b
// `),
// 			template: pointer.Pointer("{{ .meta.id }} {{ .name }}"),
// 			wantTemplate: pointer.Pointer(`
// 1 project-1
// `),
// 			wantMarkdown: pointer.Pointer(`
// | UID |   TENANT    |   NAME    | DESCRIPTION | LABELS | ANNOTATIONS |
// |-----|-------------|-----------|-------------|--------|-------------|
// |   1 | metal-stack | project-1 | project 1   | c      | a=b         |
// `),
// 		},
// 		{
// 			name: "delete",
// 			cmd: func(want *models.V1ProjectResponse) []string {
// 				return []string{"project", "rm", want.Meta.ID}
// 			},
// 			mocks: &client.MetalMockFns{
// 				Project: func(mock *mock.Mock) {
// 					mock.On("DeleteProject", testcommon.MatchIgnoreContext(t, project.NewDeleteProjectParams().WithID(project1.Meta.ID)), nil).Return(&project.DeleteProjectOK{
// 						Payload: project1,
// 					}, nil)
// 				},
// 			},
// 			want: project1,
// 		},
// 		{
// 			name: "create",
// 			cmd: func(want *models.V1ProjectResponse) []string {
// 				args := []string{"project", "create",
// 					"--name", want.Name,
// 					"--description", want.Description,
// 					"--tenant", want.TenantID,
// 					"--label", strings.Join(want.Meta.Labels, ","),
// 					"--annotation", strings.Join(genericcli.MapToLabels(want.Meta.Annotations), ","),
// 					"--cluster-quota", strconv.FormatInt(int64(want.Quotas.Cluster.Quota), 10),
// 					"--machine-quota", strconv.FormatInt(int64(want.Quotas.Machine.Quota), 10),
// 					"--ip-quota", strconv.FormatInt(int64(want.Quotas.IP.Quota), 10),
// 				}
// 				assertExhaustiveArgs(t, args, commonExcludedFileArgs()...)
// 				return args
// 			},
// 			mocks: &client.MetalMockFns{
// 				Project: func(mock *mock.Mock) {
// 					p := project1
// 					p.Meta.ID = ""
// 					p.Meta.Version = 0
// 					p.Quotas.Cluster.Used = 0
// 					p.Quotas.IP.Used = 0
// 					p.Quotas.Machine.Used = 0
// 					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(projectResponseToCreate(p))), nil).Return(&project.CreateProjectCreated{
// 						Payload: project1,
// 					}, nil)
// 				},
// 			},
// 			want: project1,
// 		},
// 	}
// 	for _, tt := range tests {
// 		tt.testCmd(t)
// 	}
// }
