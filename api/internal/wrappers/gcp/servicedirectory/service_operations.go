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

type sdServiceOperation struct {
	wrapper  *GoogleServiceDirectoryWrapper
	parentOp *sdNamespaceOperation
	name     string
	pathName string
}

func (s *sdServiceOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Service, error) {
	if !opts.ForceRefresh {
		if ns := s.wrapper.getFromCache(s.pathName); ns != nil {
			return ns.(*coretypes.Service), nil
		}
	}

	serv, err := s.wrapper.client.GetService(ctx, &pb.GetServiceRequest{
		Name: s.pathName,
	})
	if err != nil {
		return nil, err
	}

	service := toCoreService(serv)
	s.wrapper.putOnCache(s.pathName, service)

	return service, nil
}

func (s *sdServiceOperation) Create(ctx context.Context, metadata map[string]string) (*coretypes.Service, error) {
	res, err := s.wrapper.client.CreateService(ctx, &pb.CreateServiceRequest{
		Parent:    s.parentOp.pathName,
		ServiceId: s.name,
		Service: &pb.Service{
			Name:        s.pathName,
			Annotations: metadata,
		},
	})
	if err != nil {
		return nil, err
	}

	service := toCoreService(res)
	s.wrapper.putOnCache(s.pathName, service)

	return service, nil
}

func (s *sdServiceOperation) Update(ctx context.Context, metadata map[string]string) (*coretypes.Service, error) {
	res, err := s.wrapper.client.UpdateService(ctx, &pb.UpdateServiceRequest{
		Service: &pb.Service{
			Name:        s.pathName,
			Annotations: metadata,
		},
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"annotations"},
		},
	})
	if err != nil {
		return nil, err
	}

	service := toCoreService(res)
	s.wrapper.putOnCache(s.pathName, service)

	return service, nil
}

func (s *sdServiceOperation) Delete(ctx context.Context) error {
	s.wrapper.cache.Delete(s.pathName)

	return s.wrapper.client.DeleteService(ctx, &pb.DeleteServiceRequest{
		Name: s.pathName,
	})
}

func (s *sdServiceOperation) List(opts *list.Options) ops.ServiceLister {
	if s.name != "" {
		if opts.NameFilters == nil {
			opts.NameFilters = &list.NameFilters{}
		}

		opts.NameFilters.In = append(opts.NameFilters.In, s.name)
	}

	return &ServiceDirectoryServiceIterator{
		wrapper:  s.wrapper,
		parentOp: s.parentOp,
		options:  opts,
	}
}

type ServiceDirectoryServiceIterator struct {
	wrapper  *GoogleServiceDirectoryWrapper
	parentOp *sdNamespaceOperation
	options  *list.Options

	// Request is the actual request that will be sent to Service Directory.
	// Here it is exported so that it could mocked and tested.
	Request *pb.ListServicesRequest

	// Iterator is the interface that represents Service Directory's own
	// iterator. Here is used as interface so that it could be mocked
	// and tested.
	Iterator serviceIteratorClient
}

func (s *ServiceDirectoryServiceIterator) Next(ctx context.Context) (*coretypes.Service, ops.ServiceOperation, error) {
	client := s.wrapper.client

	if s.Request == nil {
		req := &pb.ListServicesRequest{
			PageSize: s.options.Results,
			Parent:   s.parentOp.pathName,
		}

		reqFilters := getRequestFilters(
			path.Join(s.parentOp.pathName, pathServices), s.options)

		if reqFilters != "" {
			req.Filter = reqFilters
		}
		s.Request = req
	}

	if s.Iterator == nil {
		s.Iterator = client.ListServices(ctx, s.Request)
	}

	var (
		serv     *coretypes.Service
		pathName string
	)

	for serv == nil {
		next, err := s.Iterator.Next()
		if err != nil {
			return nil, nil, err
		}
		servToFilter := toCoreService(next)

		if passed, _ := s.options.Filter(servToFilter); passed {
			serv = servToFilter
			pathName = next.Name
		}
	}

	s.wrapper.putOnCache(pathName, serv)
	return serv, s.parentOp.Service(path.Base(serv.Name)), nil
}

func (s *sdServiceOperation) Endpoint(name string) ops.EndpointOperation {
	return &sdEndpointOperation{
		wrapper:  s.wrapper,
		pathName: path.Join(s.pathName, pathEndpoints, name),
		name:     name,
		parentOp: s,
	}
}

func toCoreService(serv *pb.Service) *coretypes.Service {
	metadata := map[string]string{}
	if serv.Annotations != nil {
		metadata = serv.Annotations
	}

	return &coretypes.Service{
		Name: path.Base(serv.Name),
		Namespace: func() string {
			return path.Base(path.Dir(path.Dir(serv.Name)))
		}(),
		Metadata:       metadata,
		OriginalObject: serv,
	}
}
