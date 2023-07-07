package cmd

import (
	"testing"

	"github.com/dcorbe/termui-dpc/widgets"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func Test_kubernetesVersions_toNodes(t *testing.T) {
	tests := []struct {
		name        string
		previous    []*widgets.TreeNode
		clusters    []*models.V1ClusterResponse
		want        []*widgets.TreeNode
		wantChanged bool
	}{
		{
			name: "initial fill",
			clusters: []*models.V1ClusterResponse{
				{
					ID:        pointer.Pointer("cluster-a-id"),
					Name:      pointer.Pointer("cluster-a"),
					ProjectID: pointer.Pointer("project-a"),
					Tenant:    pointer.Pointer("tenant-a"),
					Kubernetes: &models.V1Kubernetes{
						Version: pointer.Pointer("1.24.3"),
					},
				},
			},
			want: []*widgets.TreeNode{
				{
					Value: nodeValue("1.24 (1)"),
					Nodes: []*widgets.TreeNode{
						{
							Value: nodeValue("1.24.3 (1)"),
							Nodes: []*widgets.TreeNode{
								{
									Value: nodeValue("tenant-a"),
									Nodes: []*widgets.TreeNode{
										{
											Value: nodeValue("project-a"),
											Nodes: []*widgets.TreeNode{
												{
													Value: nodeValue("cluster-a (cluster-a-id)"),
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
			previous: []*widgets.TreeNode{
				{
					Value: nodeValue("1.24 (2)"),
					Nodes: []*widgets.TreeNode{
						{
							Value: nodeValue("1.24.3 (2)"),
							Nodes: []*widgets.TreeNode{
								{
									Value: nodeValue("tenant-a"),
									Nodes: []*widgets.TreeNode{
										{
											Value: nodeValue("project-a"),
											Nodes: []*widgets.TreeNode{
												{
													Value: nodeValue("cluster-a (cluster-a-id)"),
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
					ID:        pointer.Pointer("cluster-a-id"),
					Name:      pointer.Pointer("cluster-a"),
					ProjectID: pointer.Pointer("project-a"),
					Tenant:    pointer.Pointer("tenant-a"),
					Kubernetes: &models.V1Kubernetes{
						Version: pointer.Pointer("1.24.3"),
					},
				},
				{
					ID:        pointer.Pointer("cluster-a-id"),
					Name:      pointer.Pointer("cluster-a"),
					ProjectID: pointer.Pointer("project-a"),
					Tenant:    pointer.Pointer("tenant-a"),
					Kubernetes: &models.V1Kubernetes{
						Version: pointer.Pointer("1.24.3"),
					},
				},
			},
			want: []*widgets.TreeNode{
				{
					Value: nodeValue("1.24 (2)"),
					Nodes: []*widgets.TreeNode{
						{
							Value: nodeValue("1.24.3 (2)"),
							Nodes: []*widgets.TreeNode{
								{
									Value: nodeValue("tenant-a"),
									Nodes: []*widgets.TreeNode{
										{
											Value: nodeValue("project-a"),
											Nodes: []*widgets.TreeNode{
												{
													Value: nodeValue("cluster-a (cluster-a-id)"),
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
			if diff := cmp.Diff(got, tt.want, cmp.AllowUnexported(widgets.TreeNode{})); diff != "" {
				t.Errorf("kubernetesVersions.toNodes() diff = %s", diff)
			}
			if got1 != tt.wantChanged {
				t.Errorf("kubernetesVersions.toNodes() gotChanged = %v, want %v", got1, tt.wantChanged)
			}
		})
	}
}
