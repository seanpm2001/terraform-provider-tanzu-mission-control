/*
Copyright © 2023 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
*/

package provisioner

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/authctx"
	clienterrors "github.com/vmware/terraform-provider-tanzu-mission-control/internal/client/errors"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/helper"
	provisioner "github.com/vmware/terraform-provider-tanzu-mission-control/internal/models/provisioner"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/resources/common"
)

func DataSourceProvisioner() *schema.Resource {
	return &schema.Resource{
		ReadContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			return dataSourceProvisionerListRead(helper.GetContextWithCaller(ctx, helper.DataRead), d, m)
		},
		Schema: provisionerListSchema,
	}
}

var provisionerListSchema = map[string]*schema.Schema{
	nameKey: {
		Type:        schema.TypeString,
		Description: "Name of the provisioner",
		Optional:    true,
	},
	managementClusterNameKey: {
		Type:        schema.TypeString,
		Description: "Name of the management cluster",
		Required:    true,
		ForceNew:    true,
	},
	orgIDKey: {
		Type:        schema.TypeString,
		Description: "ID of the organization",
		Optional:    true,
	},
	common.MetaKey: common.Meta,
}

func dataSourceProvisionerListRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	config := m.(authctx.TanzuContext)

	managementClusterName, ok := d.Get(managementClusterNameKey).(string)
	if !ok {
		return diag.Errorf("unable to read management cluster name")
	}

	provisionerName, _ := d.Get(nameKey).(string)
	if provisionerName != "" {
		return dataSourceProvisionerRead(ctx, d, m)
	}

	fn := &provisioner.VmwareTanzuManageV1alpha1ManagementclusterProvisionerFullName{
		ManagementClusterName: managementClusterName,
	}

	resp, err := config.TMCConnection.ProvisionerResourceService.ProvisionerResourceServiceList(fn)
	if err != nil {
		if clienterrors.IsNotFoundError(err) && !helper.IsDataRead(ctx) {
			_ = schema.RemoveFromState(d, m)
			return
		}
	}

	for i := range resp.Provisioners {
		d.SetId(resp.Provisioners[i].Meta.UID)

		if err := d.Set(common.MetaKey, common.FlattenMeta(resp.Provisioners[i].Meta)); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
