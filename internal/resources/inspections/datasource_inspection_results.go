/*
Copyright © 2023 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
*/

package inspections

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func DataSourceInspectionResults() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInspectionResultsRead,
		Schema:      inspectionResultsDataSourceSchema,
	}
}

func dataSourceInspectionResultsRead(ctx context.Context, data *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	resp, err := listInspections(data, m)

	switch {
	case err != nil:
		return diag.FromErr(errors.Wrapf(err, "Couldn't read inspection results."))
	case resp.Scans == nil:
		data.SetId("NO_DATA")
	default:
		err = tfInspectionModelConverter.FillTFSchema(resp.Scans[0], data)

		if err != nil {
			return diag.FromErr(err)
		}

		inspectionFullName := resp.Scans[0].FullName

		var idKeys = []string{
			inspectionFullName.ManagementClusterName,
			inspectionFullName.ProvisionerName,
			inspectionFullName.ClusterName,
			inspectionFullName.Name,
		}

		data.SetId(strings.Join(idKeys, "/"))
	}

	return diags
}
