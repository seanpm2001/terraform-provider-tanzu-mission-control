/*
Copyright © 2023 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
*/

package provisioner

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/authctx"
	clienterrors "github.com/vmware/terraform-provider-tanzu-mission-control/internal/client/errors"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/helper"
	provisioner "github.com/vmware/terraform-provider-tanzu-mission-control/internal/models/provisioner"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/resources/common"
)

type (
	contextMethodKey struct{}
)

func ResourceProvisioner() *schema.Resource {
	return &schema.Resource{
		Schema:        provisionerSchema,
		ReadContext:   dataSourceProvisionerRead,
		CreateContext: resourceProvisionerCreate,
		UpdateContext: resourceProvisionerInPlaceUpdate,
		DeleteContext: resourceProvisionerDelete,
		Description:   "Tanzu Mission Control Provisioner Resource",
	}
}

var provisionerSchema = map[string]*schema.Schema{
	nameKey: {
		Type:        schema.TypeString,
		Description: "Name of the provisioner",
		Required:    true,
		ForceNew:    true,
	},
	managementClusterNameKey: {
		Type:        schema.TypeString,
		Description: "Name of the management cluster. Edit operation such as create, update and delete is not supported for TKGm & TKGs management cluster provisioners.",
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

func resourceProvisionerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	config, ok := m.(authctx.TanzuContext)

	if !ok {
		log.Println("[ERROR] error while retrieving Tanzu auth config")
		return diag.Errorf("error while retrieving Tanzu auth config")
	}

	provisionerName, ok := d.Get(nameKey).(string)
	if !ok {
		return diag.Errorf("unable to read provisioner name")
	}

	managementClusterName, ok := d.Get(managementClusterNameKey).(string)
	if !ok {
		return diag.Errorf("unable to read management cluster name")
	}

	provisionerRequest := &provisioner.VmwareTanzuManageV1alpha1ManagementclusterProvisionerCreateProvisionerRequest{
		Provisioner: &provisioner.VmwareTanzuManageV1alpha1ManagementclusterProvisionerProvisioner{
			FullName: &provisioner.VmwareTanzuManageV1alpha1ManagementclusterProvisionerFullName{
				ManagementClusterName: managementClusterName,
				Name:                  provisionerName,
			},
			Meta: common.ConstructMeta(d),
		},
	}

	provisionerResponse, err := config.TMCConnection.ProvisionerResourceService.ProvisionerResourceServiceCreate(provisionerRequest)

	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "Unable to create Tanzu Mission Control provisioner entry, name : %s", provisionerName))
	}

	d.SetId(provisionerResponse.Provisioner.Meta.UID)

	return append(diags, dataSourceProvisionerRead(context.WithValue(ctx, contextMethodKey{}, helper.CreateState), d, m)...)
}

func resourceProvisionerInPlaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	config := m.(authctx.TanzuContext)

	updateRequired := common.HasMetaChanged(d)

	if !updateRequired {
		return diags
	}

	provisionerName, ok := d.Get(nameKey).(string)
	if !ok {
		return diag.Errorf("unable to read provisioner name")
	}

	managementClusterName, ok := d.Get(managementClusterNameKey).(string)
	if !ok {
		return diag.Errorf("unable to read management cluster name")
	}

	fn := &provisioner.VmwareTanzuManageV1alpha1ManagementclusterProvisionerFullName{
		ManagementClusterName: managementClusterName,
		Name:                  provisionerName,
	}

	getResp, err := config.TMCConnection.ProvisionerResourceService.ProvisionerResourceServiceGet(fn)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to get Tanzu Mission Control provisioner entry, name: %s", provisionerName))
	}

	if updateRequired {
		meta := common.ConstructMeta(d)

		getResp.Provisioner.Meta.Labels = meta.Labels
		getResp.Provisioner.Meta.Description = meta.Description
	}

	_, err = config.TMCConnection.ProvisionerResourceService.ProvisionerResourceServiceUpdate(
		&provisioner.VmwareTanzuManageV1alpha1ManagementclusterProvisionerCreateProvisionerRequest{
			Provisioner: getResp.Provisioner,
		},
	)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "Unable to update Tanzu Mission Control provisioner entry, name : %s", provisionerName))
	}

	return dataSourceProvisionerRead(ctx, d, m)
}

func dataSourceProvisionerRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	config := m.(authctx.TanzuContext)

	provisionerName, ok := d.Get(nameKey).(string)
	if !ok {
		return diag.Errorf("unable to read provisioner name")
	}

	managementClusterName, ok := d.Get(managementClusterNameKey).(string)
	if !ok {
		return diag.Errorf("unable to read management cluster name")
	}

	fn := &provisioner.VmwareTanzuManageV1alpha1ManagementclusterProvisionerFullName{
		Name:                  provisionerName,
		ManagementClusterName: managementClusterName,
	}

	resp, err := config.TMCConnection.ProvisionerResourceService.ProvisionerResourceServiceGet(fn)
	if err != nil {
		if clienterrors.IsNotFoundError(err) && !helper.IsDataRead(ctx) {
			_ = schema.RemoveFromState(d, m)
			return
		}
	}

	d.SetId(resp.Provisioner.Meta.UID)

	if err := d.Set(common.MetaKey, common.FlattenMeta(resp.Provisioner.Meta)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceProvisionerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	config := m.(authctx.TanzuContext)

	provisionerName, ok := d.Get(nameKey).(string)
	if !ok {
		return diag.Errorf("unable to read provisioner name")
	}

	managementClusterName, ok := d.Get(managementClusterNameKey).(string)
	if !ok {
		return diag.Errorf("unable to read management cluster name")
	}

	fn := &provisioner.VmwareTanzuManageV1alpha1ManagementclusterProvisionerFullName{
		ManagementClusterName: managementClusterName,
		Name:                  provisionerName,
	}

	err := config.TMCConnection.ProvisionerResourceService.ProvisionerResourceServiceDelete(fn)
	if err != nil && !clienterrors.IsNotFoundError(err) {
		return diag.FromErr(errors.Wrapf(err, "Unable to delete Tanzu Mission Control provisioner entry, name : %s", provisionerName))
	}

	_ = schema.RemoveFromState(d, m)

	return diags
}
