package cmd

import (
	"testing"

	"github.com/fi-ts/cloud-go/api/client/volume"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloud-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"

	"github.com/stretchr/testify/mock"
)

var (
	volume1 = &models.V1VolumeResponse{
		ConnectedHosts: nil,
		NodeIPList: []string{
			"1.1.1.1",
			"2.2.2.2",
			"3.3.3.3",
		},
		PartitionID:        pointer.Pointer("partition-a"),
		PrimaryNodeUUID:    pointer.Pointer(""),
		ProjectID:          pointer.Pointer("project-a"),
		ProtectionState:    pointer.Pointer("Healthy"),
		QosPolicyName:      pointer.Pointer("no-limit-policy"),
		QosPolicyUUID:      pointer.Pointer("qos-id"),
		RebuildProgress:    pointer.Pointer("None"),
		ReplicaCount:       pointer.Pointer(int64(3)),
		Size:               pointer.Pointer(int64(10737418240)),
		SourceSnapshotUUID: nil,
		State:              pointer.Pointer("Available"),
		Statistics: &models.V1VolumeStatistics{
			CompressionRatio:    pointer.Pointer(float64(100)),
			LogicalUsedStorage:  pointer.Pointer(int64(10737418240)),
			PhysicalUsedStorage: pointer.Pointer(int64(10737418240)),
		},
		StorageClass: pointer.Pointer("partition-gold"),
		TenantID:     pointer.Pointer("fits"),
		VolumeHandle: new(string),
		VolumeID:     pointer.Pointer("volume-1"),
		VolumeName:   pointer.Pointer("volume 1"),
	}
	volume2 = &models.V1VolumeResponse{
		ConnectedHosts: nil,
		NodeIPList: []string{
			"1.1.1.1",
			"2.2.2.2",
			"3.3.3.3",
		},
		PartitionID:        pointer.Pointer("partition-a"),
		PrimaryNodeUUID:    pointer.Pointer(""),
		ProjectID:          pointer.Pointer("project-a"),
		ProtectionState:    pointer.Pointer("Healthy"),
		QosPolicyName:      pointer.Pointer("no-limit-policy"),
		QosPolicyUUID:      pointer.Pointer("qos-id"),
		RebuildProgress:    pointer.Pointer("None"),
		ReplicaCount:       pointer.Pointer(int64(2)),
		Size:               pointer.Pointer(int64(10737418240)),
		SourceSnapshotUUID: nil,
		State:              pointer.Pointer("Available"),
		Statistics: &models.V1VolumeStatistics{
			CompressionRatio:    pointer.Pointer(float64(100)),
			LogicalUsedStorage:  pointer.Pointer(int64(10737418240)),
			PhysicalUsedStorage: pointer.Pointer(int64(10737418240)),
		},
		StorageClass: pointer.Pointer("partition-gold"),
		TenantID:     pointer.Pointer("fits"),
		VolumeHandle: new(string),
		VolumeID:     pointer.Pointer("volume-2"),
		VolumeName:   pointer.Pointer("volume 2"),
	}

	snapshot1 = &models.V1SnapshotResponse{
		NodeIPList: []string{
			"1.1.1.1",
			"2.2.2.2",
			"3.3.3.3",
		},
		PartitionID:      pointer.Pointer("partition-a"),
		PrimaryNodeUUID:  pointer.Pointer(""),
		ProjectID:        pointer.Pointer("project-a"),
		ReplicaCount:     pointer.Pointer(int64(3)),
		Size:             pointer.Pointer(int64(10737418240)),
		State:            pointer.Pointer("Available"),
		TenantID:         pointer.Pointer("fits"),
		SnapshotID:       pointer.Pointer("snap-1"),
		SourceVolumeName: pointer.Pointer("volume 1"),
		SourceVolumeID:   pointer.Pointer("volume-1"),
	}
	snapshot2 = &models.V1SnapshotResponse{
		NodeIPList: []string{
			"1.1.1.1",
			"2.2.2.2",
			"3.3.3.3",
		},
		PartitionID:      pointer.Pointer("partition-a"),
		PrimaryNodeUUID:  pointer.Pointer(""),
		ProjectID:        pointer.Pointer("project-a"),
		ReplicaCount:     pointer.Pointer(int64(3)),
		Size:             pointer.Pointer(int64(10737418240)),
		State:            pointer.Pointer("Available"),
		TenantID:         pointer.Pointer("fits"),
		SnapshotID:       pointer.Pointer("snap-2"),
		SourceVolumeName: pointer.Pointer("volume 2"),
		SourceVolumeID:   pointer.Pointer("volume-2"),
	}
)

func Test_VolumeCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1VolumeResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1VolumeResponse) []string {
				return []string{"volume", "list"}
			},
			mocks: &client.CloudMockFns{
				Volume: func(mock *mock.Mock) {
					mock.On("FindVolumes", testcommon.MatchIgnoreContext(t, volume.NewFindVolumesParams().WithBody(&models.V1VolumeFindRequest{})), nil).Return(&volume.FindVolumesOK{
						Payload: []*models.V1VolumeResponse{
							volume2,
							volume1,
						},
					}, nil)
				},
			},
			want: []*models.V1VolumeResponse{
				volume1,
				volume2,
			},
			wantTable: pointer.Pointer(`
ID         TENANT   PROJECT     PARTITION     NAME       SIZE     USAGE    REPLICAS   STORAGECLASS
volume-1   fits     project-a   partition-a   volume 1   10 GiB   10 GiB   3          partition-gold
volume-2   fits     project-a   partition-a   volume 2   10 GiB   10 GiB   2          partition-gold
`),
			wantWideTable: pointer.Pointer(`
ID         TENANT   PROJECT     PARTITION     NAME       SIZE     USAGE    REPLICAS   STORAGECLASS     NODES
volume-1   fits     project-a   partition-a   volume 1   10 GiB   10 GiB   3          partition-gold
volume-2   fits     project-a   partition-a   volume 2   10 GiB   10 GiB   2          partition-gold
`),
			template: pointer.Pointer("{{ .VolumeID }} {{ .VolumeName }}"),
			wantTemplate: pointer.Pointer(`
volume-1 volume 1
volume-2 volume 2
`),
			wantMarkdown: pointer.Pointer(`
|    ID    | TENANT |  PROJECT  |  PARTITION  |   NAME   |  SIZE  | USAGE  | REPLICAS |  STORAGECLASS  |
|----------|--------|-----------|-------------|----------|--------|--------|----------|----------------|
| volume-1 | fits   | project-a | partition-a | volume 1 | 10 GiB | 10 GiB |        3 | partition-gold |
| volume-2 | fits   | project-a | partition-a | volume 2 | 10 GiB | 10 GiB |        2 | partition-gold |
`),
		},
		{
			name: "list with filters",
			cmd: func(want []*models.V1VolumeResponse) []string {
				args := []string{"volume", "list",
					"--id", *want[0].VolumeID,
					"--tenant", *want[0].TenantID,
					"--project", *want[0].ProjectID,
					"--partition", *want[0].PartitionID,
					"--only-unbound"}
				assertExhaustiveArgs(t, args, "sort-by")
				return args
			},
			mocks: &client.CloudMockFns{
				Volume: func(mock *mock.Mock) {
					mock.On("FindVolumes", testcommon.MatchIgnoreContext(t, volume.NewFindVolumesParams().WithBody(&models.V1VolumeFindRequest{
						PartitionID: pointer.Pointer("partition-a"),
						ProjectID:   pointer.Pointer("project-a"),
						TenantID:    pointer.Pointer("fits"),
						VolumeID:    pointer.Pointer("volume-1"),
					})), nil).Return(&volume.FindVolumesOK{
						Payload: []*models.V1VolumeResponse{
							volume1,
						},
					}, nil)
				},
			},
			want: []*models.V1VolumeResponse{
				volume1,
			},
			wantTable: pointer.Pointer(`
ID         TENANT   PROJECT     PARTITION     NAME       SIZE     USAGE    REPLICAS   STORAGECLASS
volume-1   fits     project-a   partition-a   volume 1   10 GiB   10 GiB   3          partition-gold
`),
			wantWideTable: pointer.Pointer(`
ID         TENANT   PROJECT     PARTITION     NAME       SIZE     USAGE    REPLICAS   STORAGECLASS     NODES
volume-1   fits     project-a   partition-a   volume 1   10 GiB   10 GiB   3          partition-gold
`),
			template: pointer.Pointer("{{ .VolumeID }} {{ .VolumeName }}"),
			wantTemplate: pointer.Pointer(`
volume-1 volume 1
`),
			wantMarkdown: pointer.Pointer(`
|    ID    | TENANT |  PROJECT  |  PARTITION  |   NAME   |  SIZE  | USAGE  | REPLICAS |  STORAGECLASS  |
|----------|--------|-----------|-------------|----------|--------|--------|----------|----------------|
| volume-1 | fits   | project-a | partition-a | volume 1 | 10 GiB | 10 GiB |        3 | partition-gold |
				`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_VolumeCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1VolumeResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1VolumeResponse) []string {
				return []string{"volume", "describe", *want.VolumeID}
			},
			mocks: &client.CloudMockFns{
				Volume: func(mock *mock.Mock) {
					mock.On("GetVolume", testcommon.MatchIgnoreContext(t, volume.NewGetVolumeParams().WithID(*volume1.VolumeID)), nil).Return(&volume.GetVolumeOK{
						Payload: volume1,
					}, nil)
				},
			},
			want: volume1,
			wantTable: pointer.Pointer(`
ID         TENANT   PROJECT     PARTITION     NAME       SIZE     USAGE    REPLICAS   STORAGECLASS
volume-1   fits     project-a   partition-a   volume 1   10 GiB   10 GiB   3          partition-gold
`),
			wantWideTable: pointer.Pointer(`
ID         TENANT   PROJECT     PARTITION     NAME       SIZE     USAGE    REPLICAS   STORAGECLASS     NODES
volume-1   fits     project-a   partition-a   volume 1   10 GiB   10 GiB   3          partition-gold
`),
			template: pointer.Pointer("{{ .VolumeID }} {{ .VolumeName }}"),
			wantTemplate: pointer.Pointer(`
volume-1 volume 1
`),
			wantMarkdown: pointer.Pointer(`
|    ID    | TENANT |  PROJECT  |  PARTITION  |   NAME   |  SIZE  | USAGE  | REPLICAS |  STORAGECLASS  |
|----------|--------|-----------|-------------|----------|--------|--------|----------|----------------|
| volume-1 | fits   | project-a | partition-a | volume 1 | 10 GiB | 10 GiB |        3 | partition-gold |
`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1VolumeResponse) []string {
				return []string{"volume", "rm", *want.VolumeID, "--yes-i-really-mean-it"}
			},
			mocks: &client.CloudMockFns{
				Volume: func(mock *mock.Mock) {
					mock.On("GetVolume", testcommon.MatchIgnoreContext(t, volume.NewGetVolumeParams().WithID(*volume1.VolumeID)), nil).Return(&volume.GetVolumeOK{
						Payload: volume1,
					}, nil)
					mock.On("DeleteVolume", testcommon.MatchIgnoreContext(t, volume.NewDeleteVolumeParams().WithID(*volume1.VolumeID)), nil).Return(&volume.DeleteVolumeOK{
						Payload: volume1,
					}, nil)
				},
			},
			want: volume1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_SnapshotCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1SnapshotResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1SnapshotResponse) []string {
				return []string{"volume", "snapshot", "list", "--project", *snapshot1.ProjectID}
			},
			mocks: &client.CloudMockFns{
				Volume: func(mock *mock.Mock) {
					mock.On("FindSnapshots", testcommon.MatchIgnoreContext(t, volume.NewFindSnapshotsParams().WithBody(&models.V1SnapshotFindRequest{
						ProjectID: snapshot1.ProjectID,
					})), nil).Return(&volume.FindSnapshotsOK{
						Payload: []*models.V1SnapshotResponse{
							snapshot2,
							snapshot1,
						},
					}, nil)
				},
			},
			want: []*models.V1SnapshotResponse{
				snapshot1,
				snapshot2,
			},
			wantTable: pointer.Pointer(`
ID       TENANT   PARTITION     NAME     SOURCEVOLUMEID   SOURCEVOLUMENAME   SIZE
snap-1   fits     partition-a   snap-1   volume-1         volume 1           10 GiB
snap-2   fits     partition-a   snap-2   volume-2         volume 2           10 GiB
`),
			wantWideTable: pointer.Pointer(`
ID       TENANT   PARTITION     NAME     SOURCEVOLUMEID   SOURCEVOLUMENAME   SIZE
snap-1   fits     partition-a   snap-1   volume-1         volume 1           10 GiB
snap-2   fits     partition-a   snap-2   volume-2         volume 2           10 GiB
`),
			template: pointer.Pointer("{{ .SnapshotID }}"),
			wantTemplate: pointer.Pointer(`
snap-1
snap-2
`),
			wantMarkdown: pointer.Pointer(`
|   ID   | TENANT |  PARTITION  |  NAME  | SOURCEVOLUMEID | SOURCEVOLUMENAME |  SIZE  |
|--------|--------|-------------|--------|----------------|------------------|--------|
| snap-1 | fits   | partition-a | snap-1 | volume-1       | volume 1         | 10 GiB |
| snap-2 | fits   | partition-a | snap-2 | volume-2       | volume 2         | 10 GiB |
`),
		},
		{
			name: "list with filters",
			cmd: func(want []*models.V1SnapshotResponse) []string {
				args := []string{"volume", "snapshot", "list",
					"--id", *want[0].SnapshotID,
					"--name", *want[0].SourceVolumeName,
					"--project", *want[0].ProjectID,
					"--partition", *want[0].PartitionID}
				assertExhaustiveArgs(t, args, "sort-by")
				return args
			},
			mocks: &client.CloudMockFns{
				Volume: func(mock *mock.Mock) {
					mock.On("FindSnapshots", testcommon.MatchIgnoreContext(t, volume.NewFindSnapshotsParams().WithBody(&models.V1SnapshotFindRequest{
						PartitionID: snapshot1.PartitionID,
						ProjectID:   snapshot1.ProjectID,
						Name:        snapshot1.SourceVolumeName,
						SnapshotID:  snapshot1.SnapshotID,
					})), nil).Return(&volume.FindSnapshotsOK{
						Payload: []*models.V1SnapshotResponse{
							snapshot1,
						},
					}, nil)
				},
			},
			want: []*models.V1SnapshotResponse{
				snapshot1,
			},
			wantTable: pointer.Pointer(`
ID       TENANT   PARTITION     NAME     SOURCEVOLUMEID   SOURCEVOLUMENAME   SIZE
snap-1   fits     partition-a   snap-1   volume-1         volume 1           10 GiB
`),
			wantWideTable: pointer.Pointer(`
ID       TENANT   PARTITION     NAME     SOURCEVOLUMEID   SOURCEVOLUMENAME   SIZE
snap-1   fits     partition-a   snap-1   volume-1         volume 1           10 GiB
`),
			template: pointer.Pointer("{{ .SnapshotID }}"),
			wantTemplate: pointer.Pointer(`
snap-1
`),
			wantMarkdown: pointer.Pointer(`
|   ID   | TENANT |  PARTITION  |  NAME  | SOURCEVOLUMEID | SOURCEVOLUMENAME |  SIZE  |
|--------|--------|-------------|--------|----------------|------------------|--------|
| snap-1 | fits   | partition-a | snap-1 | volume-1       | volume 1         | 10 GiB |
`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_SnapshotCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1SnapshotResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1SnapshotResponse) []string {
				return []string{"volume", "snapshot", "describe", *want.SnapshotID, "--project", *want.ProjectID}
			},
			mocks: &client.CloudMockFns{
				Volume: func(mock *mock.Mock) {
					mock.On("GetSnapshot", testcommon.MatchIgnoreContext(t, volume.NewGetSnapshotParams().WithID(*snapshot1.SnapshotID).WithProjectID(snapshot1.ProjectID)), nil).Return(&volume.GetSnapshotOK{
						Payload: snapshot1,
					}, nil)
				},
			},
			want: snapshot1,
			wantTable: pointer.Pointer(`
ID       TENANT   PARTITION     NAME     SOURCEVOLUMEID   SOURCEVOLUMENAME   SIZE
snap-1   fits     partition-a   snap-1   volume-1         volume 1           10 GiB
`),
			wantWideTable: pointer.Pointer(`
ID       TENANT   PARTITION     NAME     SOURCEVOLUMEID   SOURCEVOLUMENAME   SIZE
snap-1   fits     partition-a   snap-1   volume-1         volume 1           10 GiB
`),
			template: pointer.Pointer("{{ .SnapshotID }}"),
			wantTemplate: pointer.Pointer(`
snap-1
`),
			wantMarkdown: pointer.Pointer(`
|   ID   | TENANT |  PARTITION  |  NAME  | SOURCEVOLUMEID | SOURCEVOLUMENAME |  SIZE  |
|--------|--------|-------------|--------|----------------|------------------|--------|
| snap-1 | fits   | partition-a | snap-1 | volume-1       | volume 1         | 10 GiB |
`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1SnapshotResponse) []string {
				return []string{"volume", "snapshot", "rm", *want.SnapshotID, "--yes-i-really-mean-it", "--project", *want.ProjectID}
			},
			mocks: &client.CloudMockFns{
				Volume: func(mock *mock.Mock) {
					mock.On("DeleteSnapshot", testcommon.MatchIgnoreContext(t, volume.NewDeleteSnapshotParams().WithID(*snapshot1.SnapshotID).WithProjectID(snapshot1.ProjectID)), nil).Return(&volume.DeleteSnapshotOK{
						Payload: snapshot1,
					}, nil)
				},
			},
			want: snapshot1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
