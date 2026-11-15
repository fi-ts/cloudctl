package cmd

import (
	"testing"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/google/go-cmp/cmp"
)

func Test_kubernetesVersions_toNodes(t *testing.T) {
	tests := []struct {
		name        string
		previous    []*versionNode
		clusters    []*models.V1ClusterResponse
		want        []*versionNode
		wantChanged bool
	}{
		{
			name: "initial fill",
			clusters: []*models.V1ClusterResponse{
				{
					ID:        new("cluster-a-id"),
					Name:      new("cluster-a"),
					ProjectID: new("project-a"),
					Tenant:    new("tenant-a"),
					Kubernetes: &models.V1Kubernetes{
						Version: new("1.24.3"),
					},
				},
			},
			want: []*versionNode{
				{
					value: "1.24 (1)",
					kids: []*versionNode{
						{
							value: "1.24.3 (1)",
							kids: []*versionNode{
								{
									value: "tenant-a",
									kids: []*versionNode{
										{
											value: "project-a",
											kids: []*versionNode{
												{
													value: "cluster-a (cluster-a-id)",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantChanged: true,
		},
		{
			name: "keep expanded child",
			previous: []*versionNode{
				{
					value:    "1.24 (2)",
					expanded: true,
					kids: []*versionNode{
						{
							value: "1.24.3 (2)",
							kids: []*versionNode{
								{
									value: "tenant-a",
									kids: []*versionNode{
										{
											value: "project-a",
											kids: []*versionNode{
												{
													value: "cluster-a (cluster-a-id)",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			clusters: []*models.V1ClusterResponse{
				{
					ID:        new("cluster-a-id"),
					Name:      new("cluster-a"),
					ProjectID: new("project-a"),
					Tenant:    new("tenant-a"),
					Kubernetes: &models.V1Kubernetes{
						Version: new("1.24.3"),
					},
				},
				{
					ID:        new("cluster-a-id"),
					Name:      new("cluster-a"),
					ProjectID: new("project-a"),
					Tenant:    new("tenant-a"),
					Kubernetes: &models.V1Kubernetes{
						Version: new("1.24.3"),
					},
				},
			},
			want: []*versionNode{
				{
					value:    "1.24 (2)",
					expanded: true,
					kids: []*versionNode{
						{
							value: "1.24.3 (2)",
							kids: []*versionNode{
								{
									value: "tenant-a",
									kids: []*versionNode{
										{
											value: "project-a",
											kids: []*versionNode{
												{
													value: "cluster-a (cluster-a-id)",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantChanged: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := kubernetesVersions{
				previous: tt.previous,
			}
			got, got1 := k.update(tt.clusters)
			if diff := cmp.Diff(got, tt.want, cmp.AllowUnexported(versionNode{})); diff != "" {
				t.Errorf("kubernetesVersions.toNodes() diff = %s", diff)
			}
			if got1 != tt.wantChanged {
				t.Errorf("kubernetesVersions.toNodes() gotChanged = %v, want %v", got1, tt.wantChanged)
			}
		})
	}
}
