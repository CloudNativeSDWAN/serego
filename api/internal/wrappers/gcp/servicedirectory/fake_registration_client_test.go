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

package servicedirectory_test

import (
	"context"

	sd "cloud.google.com/go/servicedirectory/apiv1"
	"github.com/googleapis/gax-go/v2"
	pb "google.golang.org/genproto/googleapis/cloud/servicedirectory/v1"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
)

type fakeRegistrationClient struct {
	_createNamespace func(context.Context, *pb.CreateNamespaceRequest, ...gax.CallOption) (*pb.Namespace, error)
	_createService   func(context.Context, *pb.CreateServiceRequest, ...gax.CallOption) (*pb.Service, error)
	_createEndpoint  func(context.Context, *pb.CreateEndpointRequest, ...gax.CallOption) (*pb.Endpoint, error)
	_getNamespace    func(context.Context, *pb.GetNamespaceRequest, ...gax.CallOption) (*pb.Namespace, error)
	_getService      func(context.Context, *pb.GetServiceRequest, ...gax.CallOption) (*pb.Service, error)
	_getEndpoint     func(context.Context, *pb.GetEndpointRequest, ...gax.CallOption) (*pb.Endpoint, error)
	_updateNamespace func(context.Context, *pb.UpdateNamespaceRequest, ...gax.CallOption) (*pb.Namespace, error)
	_updateService   func(context.Context, *pb.UpdateServiceRequest, ...gax.CallOption) (*pb.Service, error)
	_updateEndpoint  func(context.Context, *pb.UpdateEndpointRequest, ...gax.CallOption) (*pb.Endpoint, error)
	_deleteNamespace func(context.Context, *pb.DeleteNamespaceRequest, ...gax.CallOption) error
	_deleteService   func(context.Context, *pb.DeleteServiceRequest, ...gax.CallOption) error
	_deleteEndpoint  func(context.Context, *pb.DeleteEndpointRequest, ...gax.CallOption) error
	_listNamespaces  func(context.Context, *pb.ListNamespacesRequest, ...gax.CallOption) *sd.NamespaceIterator
	_listServices    func(context.Context, *pb.ListServicesRequest, ...gax.CallOption) *sd.ServiceIterator
	_listEndpoints   func(context.Context, *pb.ListEndpointsRequest, ...gax.CallOption) *sd.EndpointIterator
}

func (f *fakeRegistrationClient) Close() error {
	return nil
}

func (f *fakeRegistrationClient) CreateNamespace(ctx context.Context, req *pb.CreateNamespaceRequest, _ ...gax.CallOption) (*pb.Namespace, error) {
	return f._createNamespace(ctx, req)
}

func (f *fakeRegistrationClient) CreateService(ctx context.Context, req *pb.CreateServiceRequest, _ ...gax.CallOption) (*pb.Service, error) {
	return f._createService(ctx, req)
}

func (f *fakeRegistrationClient) CreateEndpoint(ctx context.Context, req *pb.CreateEndpointRequest, _ ...gax.CallOption) (*pb.Endpoint, error) {
	return f._createEndpoint(ctx, req)
}

func (f *fakeRegistrationClient) GetNamespace(ctx context.Context, req *pb.GetNamespaceRequest, _ ...gax.CallOption) (*pb.Namespace, error) {
	return f._getNamespace(ctx, req)
}

func (f *fakeRegistrationClient) GetService(ctx context.Context, req *pb.GetServiceRequest, _ ...gax.CallOption) (*pb.Service, error) {
	return f._getService(ctx, req)
}

func (f *fakeRegistrationClient) GetEndpoint(ctx context.Context, req *pb.GetEndpointRequest, _ ...gax.CallOption) (*pb.Endpoint, error) {
	return f._getEndpoint(ctx, req)
}

func (f *fakeRegistrationClient) UpdateNamespace(ctx context.Context, req *pb.UpdateNamespaceRequest, _ ...gax.CallOption) (*pb.Namespace, error) {
	return f._updateNamespace(ctx, req)
}

func (f *fakeRegistrationClient) UpdateService(ctx context.Context, req *pb.UpdateServiceRequest, _ ...gax.CallOption) (*pb.Service, error) {
	return f._updateService(ctx, req)
}

func (f *fakeRegistrationClient) UpdateEndpoint(ctx context.Context, req *pb.UpdateEndpointRequest, _ ...gax.CallOption) (*pb.Endpoint, error) {
	return f._updateEndpoint(ctx, req)
}

func (f *fakeRegistrationClient) DeleteNamespace(ctx context.Context, req *pb.DeleteNamespaceRequest, _ ...gax.CallOption) error {
	return f._deleteNamespace(ctx, req)
}

func (f *fakeRegistrationClient) DeleteService(ctx context.Context, req *pb.DeleteServiceRequest, _ ...gax.CallOption) error {
	return f._deleteService(ctx, req)
}

func (f *fakeRegistrationClient) DeleteEndpoint(ctx context.Context, req *pb.DeleteEndpointRequest, _ ...gax.CallOption) error {
	return f._deleteEndpoint(ctx, req)
}

func (f *fakeRegistrationClient) ListNamespaces(ctx context.Context, req *pb.ListNamespacesRequest, _ ...gax.CallOption) *sd.NamespaceIterator {
	return f._listNamespaces(ctx, req)
}

func (f *fakeRegistrationClient) ListServices(ctx context.Context, req *pb.ListServicesRequest, _ ...gax.CallOption) *sd.ServiceIterator {
	return f._listServices(ctx, req)
}

func (f *fakeRegistrationClient) ListEndpoints(ctx context.Context, req *pb.ListEndpointsRequest, _ ...gax.CallOption) *sd.EndpointIterator {
	return f._listEndpoints(ctx, req)
}

func (f *fakeRegistrationClient) GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, _ ...gax.CallOption) (*iampb.Policy, error) {
	return nil, nil
}

func (f *fakeRegistrationClient) SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest, _ ...gax.CallOption) (*iampb.Policy, error) {
	return nil, nil
}

func (f *fakeRegistrationClient) TestIamPermissions(ctx context.Context, req *iampb.TestIamPermissionsRequest, _ ...gax.CallOption) (*iampb.TestIamPermissionsResponse, error) {
	return nil, nil
}
