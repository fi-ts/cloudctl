package cmd

import (
	"fmt"

	"github.com/fi-ts/cloud-go/api/models"
	"github.com/fi-ts/cloudctl/cmd/helper"
	"gopkg.in/yaml.v3"

	"github.com/fi-ts/cloud-go/api/client/tenant"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tenantCmd = &cobra.Command{
		Use:   "tenant",
		Short: "manage tenants",
	}
	tenantListCmd = &cobra.Command{
		Use:     "list",
		Short:   "lists tenants",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return tenantList(args)
		},
		PreRun: bindPFlags,
	}
	tenantDescribeCmd = &cobra.Command{
		Use:   "describe <tenantID>",
		Short: "describe a tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tenantDescribe(args)
		},
		PreRun: bindPFlags,
	}
	tenantEditCmd = &cobra.Command{
		Use:   "edit <tenantID>",
		Short: "edit a tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tenantEdit(args)
		},
		PreRun: bindPFlags,
	}
	tenantApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "create/update a tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tenantApply()
		},
		PreRun: bindPFlags,
	}
)

func init() {
	tenantCmd.AddCommand(tenantDescribeCmd)
	tenantCmd.AddCommand(tenantEditCmd)
	tenantCmd.AddCommand(tenantListCmd)
	tenantApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
	Example tenant update:

	# cloudctl tenant describe tenant1 -o yaml > tenant1.yaml
	# vi tenant1.yaml
	## either via stdin
	# cat tenant1.yaml | cloudctl tenant apply -f -
	## or via file
	# cloudctl tenant apply -f tenant1.yaml
	`)
	tenantCmd.AddCommand(tenantApplyCmd)

}

func tenantID(verb string, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("tenant %s requires tenantID as argument", verb)
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", fmt.Errorf("tenant %s requires exactly one tenantID as argument", verb)
}

func tenantDescribe(args []string) error {
	id, err := tenantID("edit", args)
	if err != nil {
		return err
	}
	request := tenant.NewGetTenantParams()
	request.SetID(id)
	resp, err := cloud.Tenant.GetTenant(request, nil)
	if err != nil {
		return fmt.Errorf("tenant describe error:%w", err)
	}
	return printer.Print(resp.Payload)
}

func tenantList(args []string) error {
	request := tenant.NewListTenantsParams()
	resp, err := cloud.Tenant.ListTenants(request, nil)
	if err != nil {
		return fmt.Errorf("tenant list error:%w", err)
	}
	return printer.Print(resp.Payload)
}

func tenantApply() error {
	var tars []models.V1Tenant
	var tar models.V1Tenant
	err := helper.ReadFrom(viper.GetString("file"), &tar, func(data interface{}) {
		doc := data.(*models.V1Tenant)
		tars = append(tars, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		tar = models.V1Tenant{}
	})
	if err != nil {
		return err
	}
	response := []*models.V1TenantResponse{}
	for i, tar := range tars {
		request := tenant.NewGetTenantParams()
		request.SetID(tar.Meta.ID)
		t, err := cloud.Tenant.GetTenant(request, nil)
		if err != nil {
			return err
		}
		if t.Payload == nil {
			return fmt.Errorf("Only tenant update is supported")
		}
		if t.Payload.Meta != nil {
			params := tenant.NewUpdateTenantParams()
			params.SetBody(&models.V1TenantUpdateRequest{Tenant: &tars[i]})
			resp, err := cloud.Tenant.UpdateTenant(params, nil)
			if err != nil {
				return err
			}
			response = append(response, resp.Payload)
			continue
		}
	}
	return printer.Print(response)
}

func tenantEdit(args []string) error {
	id, err := tenantID("edit", args)
	if err != nil {
		return err
	}

	getFunc := func(id string) ([]byte, error) {
		request := tenant.NewGetTenantParams()
		request.SetID(id)
		resp, err := cloud.Tenant.GetTenant(request, nil)
		if err != nil {
			return nil, fmt.Errorf("tenant describe error:%w", err)
		}
		content, err := yaml.Marshal(resp.Payload)
		if err != nil {
			return nil, err
		}
		return content, nil
	}
	updateFunc := func(filename string) error {
		purs, err := readtenantUpdateRequests(filename)
		if err != nil {
			return err
		}
		if len(purs) != 1 {
			return fmt.Errorf("tenant update error more or less than one tenant given:%d", len(purs))
		}
		pup := tenant.NewUpdateTenantParams()
		pup.Body = &models.V1TenantUpdateRequest{Tenant: &purs[0]}
		uresp, err := cloud.Tenant.UpdateTenant(pup, nil)
		if err != nil {
			return err
		}
		return printer.Print(uresp.Payload)
	}

	return helper.Edit(id, getFunc, updateFunc)
}

func readtenantUpdateRequests(filename string) ([]models.V1Tenant, error) {
	var pcrs []models.V1Tenant
	var pcr models.V1Tenant
	err := helper.ReadFrom(filename, &pcr, func(data interface{}) {
		doc := data.(*models.V1Tenant)
		pcrs = append(pcrs, *doc)
	})
	if err != nil {
		return pcrs, err
	}
	if len(pcrs) != 1 {
		return pcrs, fmt.Errorf("tenant update error more or less than one tenant given:%d", len(pcrs))
	}
	return pcrs, nil
}
