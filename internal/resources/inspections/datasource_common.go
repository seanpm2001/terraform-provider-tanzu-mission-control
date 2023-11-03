/*
Copyright © 2023 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
*/

package inspections

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/authctx"
	inspectionsmodel "github.com/vmware/terraform-provider-tanzu-mission-control/internal/models/inspections"
)

func listInspections(data *schema.ResourceData, m interface{}) (resp *inspectionsmodel.VmwareTanzuManageV1alpha1ClusterInspectionScanListData, err error) {
	config := m.(authctx.TanzuContext)
	model, err := tfInspectionModelConverter.ConvertTFSchemaToAPIModel(data, []string{NameKey, ClusterNameKey, ManagementClusterNameKey, ProvisionerNameKey})
	inspectionFullName := model.FullName

	if err != nil {
		return nil, errors.Wrap(err, "Converting schema failed.")
	}

	resp, err = config.TMCConnection.InspectionsResourceService.InspectionsResourceServiceList(inspectionFullName)

	return resp, err
}
