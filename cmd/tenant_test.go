package cmd

import (
	"testing"

	"github.com/fi-ts/cloud-go/api/client/tenant"
	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloud-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	tenant1 = &models.V1TenantResponse{
		Meta: &models.V1Meta{
			Kind:       "Tenant",
			Apiversion: "v1",
			ID:         "1",
			Annotations: map[string]string{
				"a": "b",
			},
			Labels:  []string{"c"},
			Version: 1,
		},
		Description: "tenant 1",
		Name:        "tenant-1",
		Quotas: &models.V1QuotaSet{
			Cluster: &models.V1Quota{
				Quota: 1,
				Used:  1,
			},
			IP: &models.V1Quota{
				Quota: 2,
				Used:  2,
			},
			Machine: &models.V1Quota{
				Quota: 3,
				Used:  3,
			},
		},
	}
	tenant2 = &models.V1TenantResponse{
		Meta: &models.V1Meta{
			Kind:       "Tenant",
			Apiversion: "v1",
			ID:         "2",
			Annotations: map[string]string{
				"a": "b",
			},
			Labels:  []string{"c"},
			Version: 1,
		},
		Description: "tenant 2",
		Name:        "tenant-2",
		Quotas: &models.V1QuotaSet{
			Cluster: &models.V1Quota{},
			IP:      &models.V1Quota{},
			Machine: &models.V1Quota{},
		},
	}
)

func Test_TenantCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1TenantResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1TenantResponse) []string {
				return []string{"tenant", "list"}
			},
			mocks: &client.CloudMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("FindTenants", testcommon.MatchIgnoreContext(t, tenant.NewFindTenantsParams().WithBody(&models.V1TenantFindRequest{})), nil).Return(&tenant.FindTenantsOK{
						Payload: []*models.V1TenantResponse{
							tenant2,
							tenant1,
						},
					}, nil)
				},
			},
			want: []*models.V1TenantResponse{
				tenant1,
				tenant2,
			},
			wantTable: pointer.Pointer(`
ID   NAME       DESCRIPTION   LABELS   ANNOTATIONS
1    tenant-1   tenant 1      c        a=b
2    tenant-2   tenant 2      c        a=b
`),
			wantWideTable: pointer.Pointer(`
ID   NAME       DESCRIPTION   LABELS   ANNOTATIONS
1    tenant-1   tenant 1      c        a=b
2    tenant-2   tenant 2      c        a=b
`),
			template: pointer.Pointer("{{ .meta.id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 tenant-1
2 tenant-2
`),
			wantMarkdown: pointer.Pointer(`
| ID |   NAME   | DESCRIPTION | LABELS | ANNOTATIONS |
|----|----------|-------------|--------|-------------|
|  1 | tenant-1 | tenant 1    | c      | a=b         |
|  2 | tenant-2 | tenant 2    | c      | a=b         |
`),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1TenantResponse) []string {
				return []string{"tenant", "apply", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1TenantResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.CloudMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("GetTenant", testcommon.MatchIgnoreContext(t, tenant.NewGetTenantParams().WithID(tenant1.Meta.ID)), nil).Return(&tenant.GetTenantOK{
						Payload: &models.V1TenantResponse{
							Meta: &models.V1Meta{
								Version: 0,
							},
						},
					}, nil)
					mock.On("UpdateTenant", testcommon.MatchIgnoreContext(t, tenant.NewUpdateTenantParams().WithBody(tenantResponseToUpdate(tenant1))), nil).Return(&tenant.UpdateTenantOK{
						Payload: tenant1,
					}, nil)
					mock.On("GetTenant", testcommon.MatchIgnoreContext(t, tenant.NewGetTenantParams().WithID(tenant2.Meta.ID)), nil).Return(&tenant.GetTenantOK{
						Payload: &models.V1TenantResponse{
							Meta: &models.V1Meta{
								Version: 0,
							},
						},
					}, nil)
					mock.On("UpdateTenant", testcommon.MatchIgnoreContext(t, tenant.NewUpdateTenantParams().WithBody(tenantResponseToUpdate(tenant2))), nil).Return(&tenant.UpdateTenantOK{
						Payload: tenant2,
					}, nil)
				},
			},
			want: []*models.V1TenantResponse{
				tenant1,
				tenant2,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1TenantResponse) []string {
				return []string{"tenant", "update", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1TenantResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.CloudMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("GetTenant", testcommon.MatchIgnoreContext(t, tenant.NewGetTenantParams().WithID(tenant1.Meta.ID)), nil).Return(&tenant.GetTenantOK{
						Payload: &models.V1TenantResponse{
							Meta: &models.V1Meta{
								Version: 0,
							},
						},
					}, nil)
					p := tenant1
					p.Meta.Version = 1
					mock.On("UpdateTenant", testcommon.MatchIgnoreContext(t, tenant.NewUpdateTenantParams().WithBody(tenantResponseToUpdate(p))), nil).Return(&tenant.UpdateTenantOK{
						Payload: p,
					}, nil)
				},
			},
			want: []*models.V1TenantResponse{
				tenant1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_TenantCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1TenantResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1TenantResponse) []string {
				return []string{"tenant", "describe", want.Meta.ID}
			},
			mocks: &client.CloudMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("GetTenant", testcommon.MatchIgnoreContext(t, tenant.NewGetTenantParams().WithID(tenant1.Meta.ID)), nil).Return(&tenant.GetTenantOK{
						Payload: tenant1,
					}, nil)
				},
			},
			want: tenant1,
			wantTable: pointer.Pointer(`
ID   NAME       DESCRIPTION   LABELS   ANNOTATIONS
1    tenant-1   tenant 1      c        a=b
`),
			wantWideTable: pointer.Pointer(`
ID   NAME       DESCRIPTION   LABELS   ANNOTATIONS
1    tenant-1   tenant 1      c        a=b
`),
			template: pointer.Pointer("{{ .meta.id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 tenant-1
`),
			wantMarkdown: pointer.Pointer(`
| ID |   NAME   | DESCRIPTION | LABELS | ANNOTATIONS |
|----|----------|-------------|--------|-------------|
|  1 | tenant-1 | tenant 1    | c      | a=b         |
`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
