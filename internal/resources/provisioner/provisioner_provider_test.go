/*
Copyright © 2023 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
*/

package provisioner

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/authctx"
	"testing"
)

func initTestProvider(t *testing.T) *schema.Provider {
	testAccProvider := &schema.Provider{
		Schema: authctx.ProviderAuthSchema(),
		ResourcesMap: map[string]*schema.Resource{
			resourceName: ResourceProvisioner(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			resourceName: DataSourceProvisioner(),
		},
		ConfigureContextFunc: authctx.ProviderConfigureContext,
	}
	if err := testAccProvider.InternalValidate(); err != nil {
		require.NoError(t, err)
	}

	return testAccProvider
}
