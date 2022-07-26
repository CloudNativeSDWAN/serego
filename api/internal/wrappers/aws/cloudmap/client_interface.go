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

package cloudmap

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
)

type cloudMapClientIface interface {
	CreateHttpNamespace(ctx context.Context, params *servicediscovery.CreateHttpNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.CreateHttpNamespaceOutput, error)
	CreatePrivateDnsNamespace(ctx context.Context, params *servicediscovery.CreatePrivateDnsNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.CreatePrivateDnsNamespaceOutput, error)
	CreatePublicDnsNamespace(ctx context.Context, params *servicediscovery.CreatePublicDnsNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.CreatePublicDnsNamespaceOutput, error)
	CreateService(ctx context.Context, params *servicediscovery.CreateServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.CreateServiceOutput, error)
	DeleteNamespace(ctx context.Context, params *servicediscovery.DeleteNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DeleteNamespaceOutput, error)
	DeleteService(ctx context.Context, params *servicediscovery.DeleteServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DeleteServiceOutput, error)
	DeregisterInstance(ctx context.Context, params *servicediscovery.DeregisterInstanceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DeregisterInstanceOutput, error)
	DiscoverInstances(ctx context.Context, params *servicediscovery.DiscoverInstancesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.DiscoverInstancesOutput, error)
	GetInstance(ctx context.Context, params *servicediscovery.GetInstanceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetInstanceOutput, error)
	GetInstancesHealthStatus(ctx context.Context, params *servicediscovery.GetInstancesHealthStatusInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetInstancesHealthStatusOutput, error)
	GetNamespace(ctx context.Context, params *servicediscovery.GetNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetNamespaceOutput, error)
	GetOperation(ctx context.Context, params *servicediscovery.GetOperationInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetOperationOutput, error)
	GetService(ctx context.Context, params *servicediscovery.GetServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.GetServiceOutput, error)
	ListInstances(ctx context.Context, params *servicediscovery.ListInstancesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListInstancesOutput, error)
	ListNamespaces(ctx context.Context, params *servicediscovery.ListNamespacesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListNamespacesOutput, error)
	ListOperations(ctx context.Context, params *servicediscovery.ListOperationsInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListOperationsOutput, error)
	ListServices(ctx context.Context, params *servicediscovery.ListServicesInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListServicesOutput, error)
	ListTagsForResource(ctx context.Context, params *servicediscovery.ListTagsForResourceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.ListTagsForResourceOutput, error)
	RegisterInstance(ctx context.Context, params *servicediscovery.RegisterInstanceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.RegisterInstanceOutput, error)
	TagResource(ctx context.Context, params *servicediscovery.TagResourceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.TagResourceOutput, error)
	UntagResource(ctx context.Context, params *servicediscovery.UntagResourceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UntagResourceOutput, error)
	UpdateHttpNamespace(ctx context.Context, params *servicediscovery.UpdateHttpNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UpdateHttpNamespaceOutput, error)
	UpdateInstanceCustomHealthStatus(ctx context.Context, params *servicediscovery.UpdateInstanceCustomHealthStatusInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UpdateInstanceCustomHealthStatusOutput, error)
	UpdatePrivateDnsNamespace(ctx context.Context, params *servicediscovery.UpdatePrivateDnsNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UpdatePrivateDnsNamespaceOutput, error)
	UpdatePublicDnsNamespace(ctx context.Context, params *servicediscovery.UpdatePublicDnsNamespaceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UpdatePublicDnsNamespaceOutput, error)
	UpdateService(ctx context.Context, params *servicediscovery.UpdateServiceInput, optFns ...func(*servicediscovery.Options)) (*servicediscovery.UpdateServiceOutput, error)
}
