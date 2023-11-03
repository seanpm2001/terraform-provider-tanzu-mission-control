/*
Copyright © 2023 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
*/

package inspectionsclient

import (
	"net/url"

	"github.com/pkg/errors"

	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/client/transport"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/helper"
	inspectionsmodel "github.com/vmware/terraform-provider-tanzu-mission-control/internal/models/inspections"
)

const (
	// URL Paths.
	clustersAPIVersionAndGroupPath = "v1alpha1/clusters"
	inspectionsPath                = "inspection/scans"

	// Query Params.
	managementClusterNameParam = "searchScope.managementClusterName"
	provisionerNameParam       = "searchScope.provisionerName"
	inspectionNameParam        = "searchScope.name"
)

// New creates a new inspections resource service API client.
func New(transport *transport.Client) ClientService {
	return &Client{Client: transport}
}

/*
Client for inspections resource service API.
*/
type Client struct {
	*transport.Client
}

// ClientService is the interface for Client methods.
type ClientService interface {
	InspectionsResourceServiceList(fn *inspectionsmodel.VmwareTanzuManageV1alpha1ClusterInspectionScanFullName) (*inspectionsmodel.VmwareTanzuManageV1alpha1ClusterInspectionScanListData, error)
}

/*
InspectionsResourceServiceList lists inspections.
*/
func (c *Client) InspectionsResourceServiceList(fn *inspectionsmodel.VmwareTanzuManageV1alpha1ClusterInspectionScanFullName) (*inspectionsmodel.VmwareTanzuManageV1alpha1ClusterInspectionScanListData, error) {
	resp := &inspectionsmodel.VmwareTanzuManageV1alpha1ClusterInspectionScanListData{}

	if fn.ManagementClusterName == "" || fn.ProvisionerName == "" || fn.ClusterName == "" {
		return nil, errors.New("Management Cluster Name, Provisioner Name and Cluster Name must be provided.")
	}

	requestURL := helper.ConstructRequestURL(clustersAPIVersionAndGroupPath, fn.ClusterName, inspectionsPath)
	queryParams := url.Values{}

	queryParams.Add(managementClusterNameParam, fn.ManagementClusterName)
	queryParams.Add(provisionerNameParam, fn.ProvisionerName)

	if fn.Name != "" {
		queryParams.Add(inspectionNameParam, fn.ProvisionerName)
	}

	requestURL = requestURL.AppendQueryParams(queryParams)

	err := c.Get(requestURL.String(), resp)

	return resp, err
}
