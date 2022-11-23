package cmd

import (
	"fmt"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/sorters"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"

	"github.com/fi-ts/cloud-go/api/client/tenant"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type tenantCmd struct {
	*config
}

func newTenantCmd(c *config) *cobra.Command {
	w := tenantCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[any, *models.V1TenantUpdateRequest, *models.V1TenantResponse]{
		BinaryName: binaryName,
		GenericCLI: genericcli.NewGenericCLI[any, *models.V1TenantUpdateRequest, *models.V1TenantResponse](w).WithFS(c.fs),
		OnlyCmds: map[genericcli.DefaultCmd]bool{
			genericcli.ListCmd:     true,
			genericcli.DescribeCmd: true,
			genericcli.UpdateCmd:   true,
			genericcli.ApplyCmd:    true,
			genericcli.EditCmd:     true,
		},
		Singular:        "tenant",
		Plural:          "tenants",
		Description:     "manage tenants",
		Sorter:          sorters.TenantSorter(),
		ValidArgsFn:     c.comp.TenantListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("id", "", "show projects of given id")
			cmd.Flags().String("name", "", "show projects of given name")
			must(cmd.RegisterFlagCompletionFunc("id", c.comp.TenantListCompletion))
		},
	}

	return genericcli.NewCmds(cmdsConfig)
}

func (c tenantCmd) Get(id string) (*models.V1TenantResponse, error) {
	resp, err := c.client.Tenant.GetTenant(tenant.NewGetTenantParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c tenantCmd) List() ([]*models.V1TenantResponse, error) {
	resp, err := c.client.Tenant.FindTenants(tenant.NewFindTenantsParams().WithBody(&models.V1TenantFindRequest{
		ID:   viper.GetString("id"),
		Name: viper.GetString("name"),
	}), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c tenantCmd) Delete(_ string) (*models.V1TenantResponse, error) {
	return nil, fmt.Errorf("tenant entity does not support delete operation")
}

func (c tenantCmd) Create(rq any) (*models.V1TenantResponse, error) {
	return nil, genericcli.AlreadyExistsError()
}

func (c tenantCmd) Update(rq *models.V1TenantUpdateRequest) (*models.V1TenantResponse, error) {
	resp, err := c.Get(rq.Tenant.Meta.ID)
	if err != nil {
		return nil, err
	}

	// FIXME: should not be done by the client, see https://github.com/fi-ts/cloudctl/pull/26
	rq.Tenant.Meta.Version = resp.Meta.Version + 1

	updateResp, err := c.client.Tenant.UpdateTenant(tenant.NewUpdateTenantParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}

func (c tenantCmd) Convert(r *models.V1TenantResponse) (string, any, *models.V1TenantUpdateRequest, error) {
	if r.Meta == nil {
		return "", nil, nil, fmt.Errorf("meta is nil")
	}
	return r.Meta.ID, nil, tenantResponseToUpdate(r), nil
}

func tenantResponseToUpdate(r *models.V1TenantResponse) *models.V1TenantUpdateRequest {
	return &models.V1TenantUpdateRequest{
		Tenant: &models.V1Tenant{
			Meta: &models.V1Meta{
				Apiversion:  r.Meta.Apiversion,
				Kind:        r.Meta.Kind,
				ID:          r.Meta.ID,
				Annotations: r.Meta.Annotations,
				Labels:      r.Meta.Labels,
				Version:     r.Meta.Version,
			},
			Description: r.Description,
			Name:        r.Name,
			Quotas:      r.Quotas,
		},
	}
}
