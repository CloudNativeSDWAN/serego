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
	"path"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	pb "google.golang.org/genproto/googleapis/cloud/servicedirectory/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

type sdEndpointOperation struct {
	wrapper  *GoogleServiceDirectoryWrapper
	parentOp *sdServiceOperation
	pathName string
	name     string
}

func (e *sdEndpointOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Endpoint, error) {
	if !opts.ForceRefresh {
		if ep := e.wrapper.getFromCache(e.pathName); ep != nil {
			return ep.(*coretypes.Endpoint), nil
		}
	}

	ep, err := e.wrapper.client.GetEndpoint(ctx, &pb.GetEndpointRequest{
		Name: e.pathName,
	})
	if err != nil {
		return nil, err
	}

	endpoint := toCoreEndpoint(ep)
	e.wrapper.putOnCache(e.pathName, endpoint)

	return endpoint, nil
}

func (e *sdEndpointOperation) Create(ctx context.Context, address string, port int32, metadata map[string]string) (*coretypes.Endpoint, error) {
	res, err := e.wrapper.client.CreateEndpoint(ctx, &pb.CreateEndpointRequest{
		Parent:     e.parentOp.pathName,
		EndpointId: e.name,
		Endpoint: &pb.Endpoint{
			Name:        e.pathName,
			Annotations: metadata,
			Address:     address,
			Port:        port,
		},
	})
	if err != nil {
		return nil, err
	}

	endpoint := toCoreEndpoint(res)
	e.wrapper.putOnCache(e.pathName, endpoint)

	return endpoint, nil
}

func (e *sdEndpointOperation) Update(ctx context.Context, address string, port int32, metadata map[string]string) (*coretypes.Endpoint, error) {
	res, err := e.wrapper.client.UpdateEndpoint(ctx, &pb.UpdateEndpointRequest{
		Endpoint: &pb.Endpoint{
			Name:        e.pathName,
			Annotations: metadata,
			Address:     address,
			Port:        port,
		},
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"annotations", "address", "port"},
		},
	})
	if err != nil {
		return nil, err
	}

	endpoint := toCoreEndpoint(res)
	e.wrapper.putOnCache(e.pathName, endpoint)

	return endpoint, nil
}

func (e *sdEndpointOperation) Delete(ctx context.Context) error {
	e.wrapper.cache.Delete(e.pathName)

	return e.wrapper.client.DeleteEndpoint(ctx, &pb.DeleteEndpointRequest{
		Name: e.pathName,
	})
}

func (e *sdEndpointOperation) List(opts *list.Options) ops.EndpointLister {
	if e.name != "" {
		if opts.NameFilters == nil {
			opts.NameFilters = &list.NameFilters{}
		}

		opts.NameFilters.In = append(opts.NameFilters.In, e.name)
	}

	return &ServiceDirectoryEndpointIterator{
		wrapper:  e.wrapper,
		parentOp: e.parentOp,
		options:  opts,
	}
}

type ServiceDirectoryEndpointIterator struct {
	wrapper  *GoogleServiceDirectoryWrapper
	parentOp *sdServiceOperation
	options  *list.Options

	// Request is the actual request that will be sent to Service Directory.
	// Here it is exported so that it could mocked and tested.
	Request *pb.ListEndpointsRequest

	// Iterator is the interface that represents Service Directory's own
	// iterator. Here is used as interface so that it could be mocked
	// and tested.
	Iterator endpointIteratorClient
}

func (e *ServiceDirectoryEndpointIterator) Next(ctx context.Context) (*coretypes.Endpoint, ops.EndpointOperation, error) {
	client := e.wrapper.client

	if e.Request == nil {
		req := &pb.ListEndpointsRequest{
			PageSize: e.options.Results,
			Parent:   e.parentOp.pathName,
		}

		reqFilters := getRequestFilters(
			path.Join(e.parentOp.pathName, pathEndpoints), e.options)

		if reqFilters != "" {
			req.Filter = reqFilters
		}
		e.Request = req
	}

	if e.Iterator == nil {
		e.Iterator = client.ListEndpoints(ctx, e.Request)
	}

	var (
		endp     *coretypes.Endpoint
		pathName string
	)
	for endp == nil {
		next, err := e.Iterator.Next()
		if err != nil {
			return nil, nil, err
		}
		endpToFilter := toCoreEndpoint(next)

		if passed, _ := e.options.Filter(endpToFilter); passed {
			endp = endpToFilter
			pathName = next.Name
		}
	}

	e.wrapper.putOnCache(pathName, endp)
	return endp, e.parentOp.Endpoint(path.Base(endp.Name)), nil
}

func toCoreEndpoint(endp *pb.Endpoint) *coretypes.Endpoint {
	metadata := map[string]string{}
	if endp.Annotations != nil {
		metadata = endp.Annotations
	}

	nsName, servName := "", ""
	{
		serv := path.Dir(path.Dir(endp.Name))
		servName = path.Base(serv)

		ns := path.Dir(path.Dir(serv))
		nsName = path.Base(ns)
	}

	return &coretypes.Endpoint{
		Name:           path.Base(endp.Name),
		Namespace:      nsName,
		Service:        servName,
		Address:        endp.Address,
		Port:           endp.Port,
		Metadata:       metadata,
		OriginalObject: endp,
	}
}
