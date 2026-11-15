package cmd

import (
	"strings"
	"testing"

	"charm.land/bubbletea/v2"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/google/go-cmp/cmp"
)

func Test_dashboardTabs_names(t *testing.T) {
	tabs := dashboardTabs(nil)

	want := []string{"ClusterHealth", "ClusterVersions", "Volumes", "Clusters"}
	if len(tabs) != len(want) {
		t.Fatalf("dashboardTabs() returned %d tabs, want %d", len(tabs), len(want))
	}
	for i, name := range want {
		if tabs[i].Name() != name {
			t.Errorf("dashboardTabs()[%d].Name() = %q, want %q", i, tabs[i].Name(), name)
		}
	}
}

func Test_dashboardClusterListPane_Update(t *testing.T) {
	newPane := func() *dashboardClusterListPane {
		d := &dashboardClusterListPane{cache: nil}
		d.clusters = []clusterListItem{
			{id: "id-a", name: "cluster-a"},
			{id: "id-b", name: "cluster-b"},
			{id: "id-c", name: "cluster-c"},
		}
		return d
	}

	t.Run("navigation", func(t *testing.T) {
		d := newPane()

		d.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		if d.selected != 1 {
			t.Errorf("after down selected = %d, want 1", d.selected)
		}
		d.Update(tea.KeyPressMsg{Code: tea.KeyUp})
		if d.selected != 0 {
			t.Errorf("after up selected = %d, want 0", d.selected)
		}
		d.Update(tea.KeyPressMsg{Code: tea.KeyUp})
		if d.selected != 0 {
			t.Errorf("up past top selected = %d, want 0", d.selected)
		}
		d.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
		if d.selected != 2 {
			t.Errorf("after end selected = %d, want 2", d.selected)
		}
		d.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		if d.selected != 2 {
			t.Errorf("down past bottom selected = %d, want 2", d.selected)
		}
		d.Update(tea.KeyPressMsg{Code: tea.KeyHome})
		if d.selected != 0 {
			t.Errorf("after home selected = %d, want 0", d.selected)
		}
	})

	t.Run("enter selects cluster", func(t *testing.T) {
		d := newPane()
		d.selected = 1

		cmd := d.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		if cmd == nil {
			t.Fatal("enter returned nil command")
		}

		msg, ok := cmd().(dashboardSelectClusterMsg)
		if !ok {
			t.Fatalf("enter returned %T, want dashboardSelectClusterMsg", cmd())
		}
		if msg.id != "id-b" {
			t.Errorf("enter selected id = %q, want %q", msg.id, "id-b")
		}
	})

	t.Run("enter on empty list", func(t *testing.T) {
		d := &dashboardClusterListPane{}
		if cmd := d.Update(tea.KeyPressMsg{Code: tea.KeyEnter}); cmd != nil {
			t.Errorf("enter on empty list returned non-nil command")
		}
	})
}

func Test_dashboardClusterListPane_View(t *testing.T) {
	d := &dashboardClusterListPane{}
	for _, name := range []string{"cluster-a", "cluster-b", "cluster-c"} {
		d.clusters = append(d.clusters, clusterListItem{id: "id-" + name, name: name, status: "Succeeded"})
	}

	out := d.View(120, 10)
	for _, want := range []string{"cluster-a", "cluster-b", "cluster-c", "NAME", "STATUS"} {
		if !strings.Contains(out, want) {
			t.Errorf("View() does not contain %q", want)
		}
	}
}

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
