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

package servicedirectory

import (
	"context"

	sd "cloud.google.com/go/servicedirectory/apiv1"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	pb "google.golang.org/genproto/googleapis/cloud/servicedirectory/v1"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
)

type regClient interface {
	Close() error
	CreateEndpoint(ctx context.Context, req *pb.CreateEndpointRequest, opts ...gax.CallOption) (*pb.Endpoint, error)
	CreateNamespace(ctx context.Context, req *pb.CreateNamespaceRequest, opts ...gax.CallOption) (*pb.Namespace, error)
	CreateService(ctx context.Context, req *pb.CreateServiceRequest, opts ...gax.CallOption) (*pb.Service, error)
	DeleteEndpoint(ctx context.Context, req *pb.DeleteEndpointRequest, opts ...gax.CallOption) error
	DeleteNamespace(ctx context.Context, req *pb.DeleteNamespaceRequest, opts ...gax.CallOption) error
	DeleteService(ctx context.Context, req *pb.DeleteServiceRequest, opts ...gax.CallOption) error
	GetEndpoint(ctx context.Context, req *pb.GetEndpointRequest, opts ...gax.CallOption) (*pb.Endpoint, error)
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
	GetNamespace(ctx context.Context, req *pb.GetNamespaceRequest, opts ...gax.CallOption) (*pb.Namespace, error)
	GetService(ctx context.Context, req *pb.GetServiceRequest, opts ...gax.CallOption) (*pb.Service, error)
	ListEndpoints(ctx context.Context, req *pb.ListEndpointsRequest, opts ...gax.CallOption) *sd.EndpointIterator
	ListNamespaces(ctx context.Context, req *pb.ListNamespacesRequest, opts ...gax.CallOption) *sd.NamespaceIterator
	ListServices(ctx context.Context, req *pb.ListServicesRequest, opts ...gax.CallOption) *sd.ServiceIterator
	SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error)
	TestIamPermissions(ctx context.Context, req *iampb.TestIamPermissionsRequest, opts ...gax.CallOption) (*iampb.TestIamPermissionsResponse, error)
	UpdateEndpoint(ctx context.Context, req *pb.UpdateEndpointRequest, opts ...gax.CallOption) (*pb.Endpoint, error)
	UpdateNamespace(ctx context.Context, req *pb.UpdateNamespaceRequest, opts ...gax.CallOption) (*pb.Namespace, error)
	UpdateService(ctx context.Context, req *pb.UpdateServiceRequest, opts ...gax.CallOption) (*pb.Service, error)
}

type namespaceIteratorClient interface {
	Next() (*pb.Namespace, error)
	PageInfo() *iterator.PageInfo
}

type serviceIteratorClient interface {
	Next() (*pb.Service, error)
	PageInfo() *iterator.PageInfo
}

type endpointIteratorClient interface {
	Next() (*pb.Endpoint, error)
	PageInfo() *iterator.PageInfo
}
