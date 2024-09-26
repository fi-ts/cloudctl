package cmd

import (
	"strconv"
	"strings"
	"testing"

	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/fi-ts/cloud-go/api/models"
	testclient "github.com/fi-ts/cloud-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	machineReservation1 = &models.V1MachineReservationResponse{
		ID:          pointer.Pointer("1"),
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
		ID:          pointer.Pointer("2"),
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
ID   TENANT   PROJECT     SIZE     AMOUNT   PARTITIONS                DESCRIPTION
1    fits     project-a   size-a   3        partition-a               for firewalls
2    fits     project-b   size-b   3        partition-a,partition-b   for machines
`),
			wantWideTable: pointer.Pointer(`
ID   TENANT   PROJECT     SIZE     AMOUNT   PARTITIONS                DESCRIPTION     LABELS
1    fits     project-a   size-a   3        partition-a               for firewalls   for firewalls   size.metal-stack.io/reserved-at=2024-09-19T08:57:40Z
                                                                                                      size.metal-stack.io/reserved-by=fits
2    fits     project-b   size-b   3        partition-a,partition-b   for machines    for machines    size.metal-stack.io/reserved-by=fits
`),
			template: pointer.Pointer("{{ .sizeid }} {{ .projectid }}"),
			wantTemplate: pointer.Pointer(`
size-a project-a
size-b project-b
`),
			wantMarkdown: pointer.Pointer(`
| ID | TENANT |  PROJECT  |  SIZE  | AMOUNT |       PARTITIONS        |  DESCRIPTION  |
|----|--------|-----------|--------|--------|-------------------------|---------------|
|  1 | fits   | project-a | size-a |      3 | partition-a             | for firewalls |
|  2 | fits   | project-b | size-b |      3 | partition-a,partition-b | for machines  |
`),
		},
		{
			name: "list with filters",
			cmd: func(want []*models.V1MachineReservationResponse) []string {
				args := []string{"project", "machine-reservation", "list", "--tenant", *want[0].Tenant, "--project", *want[0].Projectid, "--size", *want[0].Sizeid, "--id", *want[0].ID}
				assertExhaustiveArgs(t, args, "sort-by")
				return args
			},
			mocks: &testclient.CloudMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("ListMachineReservations", testcommon.MatchIgnoreContext(t, project.NewListMachineReservationsParams().WithBody(&models.V1MachineReservationFindRequest{
						Projectid: pointer.Pointer("project-a"),
						Sizeid:    pointer.Pointer("size-a"),
						Tenant:    pointer.Pointer("fits"),
						ID:        pointer.Pointer("1"),
					})), nil).Return(&project.ListMachineReservationsOK{
						Payload: []*models.V1MachineReservationResponse{
							machineReservation1,
						},
					}, nil)
				},
			},
			want: []*models.V1MachineReservationResponse{
				machineReservation1,
			},
		},
		{
			name: "apply",
			cmd: func(want []*models.V1MachineReservationResponse) []string {
				return appendFromFileCommonArgs("project", "machine-reservation", "apply")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1MachineReservationResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &testclient.CloudMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("CreateMachineReservation", testcommon.MatchIgnoreContext(t, project.NewCreateMachineReservationParams().
						WithBody(toMachineReservationCreateRequest(machineReservation1)).WithForce(pointer.Pointer(false))), nil).
						Return(nil, &project.CreateMachineReservationConflict{}).Once()
					mock.On("UpdateMachineReservation", testcommon.MatchIgnoreContext(t, project.NewUpdateMachineReservationParams().
						WithBody(toMachineReservationUpdateRequest(machineReservation1)).WithForce(pointer.Pointer(false))), nil).
						Return(&project.UpdateMachineReservationOK{Payload: machineReservation1}, nil)

					mock.On("CreateMachineReservation", testcommon.MatchIgnoreContext(t, project.NewCreateMachineReservationParams().
						WithBody(toMachineReservationCreateRequest(machineReservation2)).WithForce(pointer.Pointer(false))), nil).
						Return(&project.CreateMachineReservationCreated{Payload: machineReservation2}, nil)
				},
			},
			want: []*models.V1MachineReservationResponse{
				machineReservation1,
				machineReservation2,
			},
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1MachineReservationResponse) []string {
				return appendFromFileCommonArgs("project", "machine-reservation", "create")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1MachineReservationResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &testclient.CloudMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("CreateMachineReservation", testcommon.MatchIgnoreContext(t, project.NewCreateMachineReservationParams().
						WithBody(toMachineReservationCreateRequest(machineReservation1)).WithForce(pointer.Pointer(false))), nil).
						Return(&project.CreateMachineReservationCreated{Payload: machineReservation1}, nil)
				},
			},
			want: []*models.V1MachineReservationResponse{
				machineReservation1,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1MachineReservationResponse) []string {
				return appendFromFileCommonArgs("project", "machine-reservation", "update")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1MachineReservationResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &testclient.CloudMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("UpdateMachineReservation", testcommon.MatchIgnoreContext(t, project.NewUpdateMachineReservationParams().
						WithBody(toMachineReservationUpdateRequest(machineReservation1)).WithForce(pointer.Pointer(false))), nil).
						Return(&project.UpdateMachineReservationOK{Payload: machineReservation1}, nil)
				},
			},
			want: []*models.V1MachineReservationResponse{
				machineReservation1,
			},
		},
		{
			name: "delete from file",
			cmd: func(want []*models.V1MachineReservationResponse) []string {
				return appendFromFileCommonArgs("project", "machine-reservation", "delete")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1MachineReservationResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &testclient.CloudMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("DeleteMachineReservation", testcommon.MatchIgnoreContext(t, project.NewDeleteMachineReservationParams().WithID(*machineReservation1.ID)), nil).
						Return(&project.DeleteMachineReservationOK{Payload: machineReservation1}, nil)
				},
			},
			want: []*models.V1MachineReservationResponse{
				machineReservation1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_ProjectMachineReservationsCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1MachineReservationResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1MachineReservationResponse) []string {
				return []string{"project", "machine-reservation", "describe", *want.ID}
			},
			mocks: &testclient.CloudMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("GetMachineReservation", testcommon.MatchIgnoreContext(t, project.NewGetMachineReservationParams().WithID(*machineReservation1.ID)), nil).Return(&project.GetMachineReservationOK{
						Payload: machineReservation1,
					}, nil)
				},
			},
			want: machineReservation1,
			wantTable: pointer.Pointer(`
ID   TENANT   PROJECT     SIZE     AMOUNT   PARTITIONS    DESCRIPTION
1    fits     project-a   size-a   3        partition-a   for firewalls
`),
			wantWideTable: pointer.Pointer(`
ID   TENANT   PROJECT     SIZE     AMOUNT   PARTITIONS    DESCRIPTION     LABELS
1    fits     project-a   size-a   3        partition-a   for firewalls   for firewalls   size.metal-stack.io/reserved-at=2024-09-19T08:57:40Z
                                                                                          size.metal-stack.io/reserved-by=fits
`),
			template: pointer.Pointer("{{ .sizeid }} {{ .projectid }}"),
			wantTemplate: pointer.Pointer(`
size-a project-a
`),
			wantMarkdown: pointer.Pointer(`
| ID | TENANT |  PROJECT  |  SIZE  | AMOUNT | PARTITIONS  |  DESCRIPTION  |
|----|--------|-----------|--------|--------|-------------|---------------|
|  1 | fits   | project-a | size-a |      3 | partition-a | for firewalls |
`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1MachineReservationResponse) []string {
				return []string{"project", "machine-reservation", "rm", *want.ID}
			},
			mocks: &testclient.CloudMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("DeleteMachineReservation", testcommon.MatchIgnoreContext(t, project.NewDeleteMachineReservationParams().WithID(*machineReservation1.ID)), nil).
						Return(&project.DeleteMachineReservationOK{Payload: machineReservation1}, nil)
				},
			},
			want: machineReservation1,
		},
		{
			name: "create",
			cmd: func(want *models.V1MachineReservationResponse) []string {
				args := []string{"project", "machine-reservation", "create",
					"--amount", strconv.Itoa(int(*want.Amount)), //nolint:gosec
					"--description", want.Description,
					"--project", *want.Projectid,
					"--force",
					"--partitions", strings.Join(want.Partitionids, ","),
					"--size", *want.Sizeid,
				}

				assertExhaustiveArgs(t, args, commonExcludedFileArgs()...)
				return args
			},
			mocks: &testclient.CloudMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("CreateMachineReservation", testcommon.MatchIgnoreContext(t, project.NewCreateMachineReservationParams().
						WithBody(toMachineReservationCreateRequest(machineReservation1)).WithForce(pointer.Pointer(true))), nil).
						Return(&project.CreateMachineReservationCreated{Payload: machineReservation1}, nil)
				},
			},
			want: machineReservation1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
