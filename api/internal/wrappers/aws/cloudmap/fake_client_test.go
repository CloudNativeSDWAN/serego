// Copyright (c) 2022 Cisco Systems, Inc. and its affiliates
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package cloudmap_test

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
)

type fakeCloudMapClient struct {
	_GetNamespace        func(ctx context.Context, params *servicediscovery.GetNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetNamespaceOutput, error)
	_CreateHttpNamespace func(ctx context.Context, params *servicediscovery.CreateHttpNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.CreateHttpNamespaceOutput, error)
	_GetOperation        func(ctx context.Context, params *servicediscovery.GetOperationInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetOperationOutput, error)
	_ListNamespaces      func(ctx context.Context, params *servicediscovery.ListNamespacesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListNamespacesOutput, error)
	_ListTagsForResource func(ctx context.Context, params *servicediscovery.ListTagsForResourceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListTagsForResourceOutput, error)
	_TagResource         func(ctx context.Context, params *servicediscovery.TagResourceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.TagResourceOutput, error)
	_UntagResource       func(ctx context.Context, params *servicediscovery.UntagResourceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UntagResourceOutput, error)
	_DeleteNamespace     func(ctx context.Context, params *servicediscovery.DeleteNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DeleteNamespaceOutput, error)
	_GetService          func(ctx context.Context, params *servicediscovery.GetServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetServiceOutput, error)
	_ListServices        func(ctx context.Context, params *servicediscovery.ListServicesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListServicesOutput, error)
	_CreateService       func(ctx context.Context, params *servicediscovery.CreateServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.CreateServiceOutput, error)
	_UpdateService       func(ctx context.Context, params *servicediscovery.UpdateServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UpdateServiceOutput, error)
	_DeleteService       func(ctx context.Context, params *servicediscovery.DeleteServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DeleteServiceOutput, error)
	_ListInstances       func(ctx context.Context, params *servicediscovery.ListInstancesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListInstancesOutput, error)
	_DiscoverInstances   func(ctx context.Context, params *servicediscovery.DiscoverInstancesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DiscoverInstancesOutput, error)
	_RegisterInstance    func(ctx context.Context, params *servicediscovery.RegisterInstanceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.RegisterInstanceOutput, error)
	_DeregisterInstance  func(ctx context.Context, params *servicediscovery.DeregisterInstanceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DeregisterInstanceOutput, error)
	_GetInstance         func(ctx context.Context, params *servicediscovery.GetInstanceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetInstanceOutput, error)
}

func (f *fakeCloudMapClient) CreateHttpNamespace(ctx context.Context, params *servicediscovery.CreateHttpNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.CreateHttpNamespaceOutput, error) {
	return f._CreateHttpNamespace(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) GetNamespace(ctx context.Context, params *servicediscovery.GetNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetNamespaceOutput, error) {
	return f._GetNamespace(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) GetOperation(ctx context.Context, params *servicediscovery.GetOperationInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetOperationOutput, error) {
	return f._GetOperation(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) ListNamespaces(ctx context.Context, params *servicediscovery.ListNamespacesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListNamespacesOutput, error) {
	return f._ListNamespaces(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) ListTagsForResource(ctx context.Context, params *servicediscovery.ListTagsForResourceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListTagsForResourceOutput, error) {
	return f._ListTagsForResource(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) TagResource(ctx context.Context, params *servicediscovery.TagResourceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.TagResourceOutput, error) {
	return f._TagResource(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) UntagResource(ctx context.Context, params *servicediscovery.UntagResourceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UntagResourceOutput, error) {
	return f._UntagResource(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) DeleteNamespace(ctx context.Context, params *servicediscovery.DeleteNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DeleteNamespaceOutput, error) {
	return f._DeleteNamespace(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) GetService(ctx context.Context, params *servicediscovery.GetServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetServiceOutput, error) {
	return f._GetService(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) ListServices(ctx context.Context, params *servicediscovery.ListServicesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListServicesOutput, error) {
	return f._ListServices(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) CreateService(ctx context.Context, params *servicediscovery.CreateServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.CreateServiceOutput, error) {
	return f._CreateService(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) UpdateService(ctx context.Context, params *servicediscovery.UpdateServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UpdateServiceOutput, error) {
	return f._UpdateService(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) DeleteService(ctx context.Context, params *servicediscovery.DeleteServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DeleteServiceOutput, error) {
	return f._DeleteService(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) ListInstances(ctx context.Context, params *servicediscovery.ListInstancesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListInstancesOutput, error) {
	return f._ListInstances(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) DiscoverInstances(ctx context.Context, params *servicediscovery.DiscoverInstancesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DiscoverInstancesOutput, error) {
	return f._DiscoverInstances(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) RegisterInstance(ctx context.Context, params *servicediscovery.RegisterInstanceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.RegisterInstanceOutput, error) {
	return f._RegisterInstance(ctx, params, optFns...)
}

func (f *fakeCloudMapClient) DeregisterInstance(ctx context.Context, params *servicediscovery.DeregisterInstanceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DeregisterInstanceOutput, error) {
	return f._DeregisterInstance(ctx, params, optFns...)
}

// We don't use the following ones.
func (f *fakeCloudMapClient) CreatePrivateDnsNamespace(ctx context.Context, params *servicediscovery.CreatePrivateDnsNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.CreatePrivateDnsNamespaceOutput, error) {
	return nil, nil
}

func (f *fakeCloudMapClient) CreatePublicDnsNamespace(ctx context.Context, params *servicediscovery.CreatePublicDnsNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.CreatePublicDnsNamespaceOutput, error) {
	return nil, nil
}

func (f *fakeCloudMapClient) GetInstance(ctx context.Context, params *servicediscovery.GetInstanceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetInstanceOutput, error) {
	return f._GetInstance(ctx, params)
}

func (f *fakeCloudMapClient) GetInstancesHealthStatus(ctx context.Context, params *servicediscovery.GetInstancesHealthStatusInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetInstancesHealthStatusOutput, error) {
	return nil, nil
}

func (f *fakeCloudMapClient) ListOperations(ctx context.Context, params *servicediscovery.ListOperationsInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListOperationsOutput, error) {
	return nil, nil
}

func (f *fakeCloudMapClient) UpdateHttpNamespace(ctx context.Context, params *servicediscovery.UpdateHttpNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UpdateHttpNamespaceOutput, error) {
	return nil, nil
}

func (f *fakeCloudMapClient) UpdateInstanceCustomHealthStatus(ctx context.Context, params *servicediscovery.UpdateInstanceCustomHealthStatusInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UpdateInstanceCustomHealthStatusOutput, error) {
	return nil, nil
}

func (f *fakeCloudMapClient) UpdatePrivateDnsNamespace(ctx context.Context, params *servicediscovery.UpdatePrivateDnsNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UpdatePrivateDnsNamespaceOutput, error) {
	return nil, nil
}

func (f *fakeCloudMapClient) UpdatePublicDnsNamespace(ctx context.Context, params *servicediscovery.UpdatePublicDnsNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UpdatePublicDnsNamespaceOutput, error) {
	return nil, nil
}
