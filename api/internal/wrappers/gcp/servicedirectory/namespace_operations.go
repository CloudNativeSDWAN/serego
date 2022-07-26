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

type sdNamespaceOperation struct {
	wrapper  *GoogleServiceDirectoryWrapper
	name     string
	pathName string
}

func (n *sdNamespaceOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Namespace, error) {
	if !opts.ForceRefresh {
		if ns := n.wrapper.getFromCache(n.pathName); ns != nil {
			return ns.(*coretypes.Namespace), nil
		}
	}

	ns, err := n.wrapper.client.GetNamespace(ctx, &pb.GetNamespaceRequest{
		Name: n.pathName,
	})
	if err != nil {
		return nil, err
	}

	namespace := toCoreNamespace(ns)
	n.wrapper.putOnCache(n.pathName, namespace)

	return namespace, nil
}

func (n *sdNamespaceOperation) Create(ctx context.Context, metadata map[string]string) (*coretypes.Namespace, error) {
	res, err := n.wrapper.client.CreateNamespace(ctx, &pb.CreateNamespaceRequest{
		Parent:      path.Dir(path.Dir(n.pathName)),
		NamespaceId: n.name,
		Namespace: &pb.Namespace{
			Name:   n.pathName,
			Labels: metadata,
		},
	})
	if err != nil {
		return nil, err
	}

	namespace := toCoreNamespace(res)
	n.wrapper.putOnCache(n.pathName, namespace)

	return namespace, nil
}

func (n *sdNamespaceOperation) Update(ctx context.Context, metadata map[string]string) (*coretypes.Namespace, error) {
	res, err := n.wrapper.client.UpdateNamespace(ctx, &pb.UpdateNamespaceRequest{
		Namespace: &pb.Namespace{
			Name:   n.pathName,
			Labels: metadata,
		},
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"labels"},
		},
	})
	if err != nil {
		return nil, err
	}

	namespace := toCoreNamespace(res)
	n.wrapper.putOnCache(n.pathName, namespace)

	return namespace, err
}

func (n *sdNamespaceOperation) Delete(ctx context.Context) error {
	n.wrapper.cache.Delete(n.pathName)
	return n.wrapper.client.DeleteNamespace(ctx, &pb.DeleteNamespaceRequest{
		Name: n.pathName,
	})
}

func (n *sdNamespaceOperation) List(opts *list.Options) ops.NamespaceLister {
	parentPathName := path.Dir(n.pathName)
	if n.name != "" {
		// one more time, so we also remove "/namespaces/"
		parentPathName = path.Dir(parentPathName)
	}

	if n.name != "" {
		if opts.NameFilters == nil {
			opts.NameFilters = &list.NameFilters{}
		}

		opts.NameFilters.In = append(opts.NameFilters.In, n.name)
	}

	return &ServiceDirectoryNamespaceIterator{
		wrapper:        n.wrapper,
		parentPathName: parentPathName,
		options:        opts,
	}
}

type ServiceDirectoryNamespaceIterator struct {
	wrapper        *GoogleServiceDirectoryWrapper
	parentPathName string
	options        *list.Options

	// Request is the actual request that will be sent to Service Directory.
	// Here it is exported so that it could mocked and tested.
	Request *pb.ListNamespacesRequest

	// Iterator is the interface that represents Service Directory's own
	// iterator. Here is used as interface so that it could be mocked
	// and tested.
	Iterator namespaceIteratorClient
}

func (ni *ServiceDirectoryNamespaceIterator) Next(ctx context.Context) (*coretypes.Namespace, ops.NamespaceOperation, error) {
	client := ni.wrapper.client

	if ni.Request == nil {
		req := &pb.ListNamespacesRequest{
			Parent:   ni.parentPathName,
			PageSize: ni.options.Results,
		}

		reqFilters := getRequestFilters(
			path.Join(ni.wrapper.pathName, pathNamespaces), ni.options)

		if reqFilters != "" {
			req.Filter = reqFilters
		}
		ni.Request = req
	}

	if ni.Iterator == nil {
		ni.Iterator = client.ListNamespaces(ctx, ni.Request)
	}

	var (
		ns       *coretypes.Namespace
		pathName string
	)
	for ns == nil {
		next, err := ni.Iterator.Next()
		if err != nil {
			return nil, nil, err
		}
		nsToFilter := toCoreNamespace(next)

		if passed, _ := ni.options.Filter(nsToFilter); passed {
			ns = nsToFilter
			pathName = next.Name
		}
	}

	ni.wrapper.putOnCache(pathName, ns)
	return ns, ni.wrapper.Namespace(path.Base(ns.Name)), nil
}

func (n *sdNamespaceOperation) Service(name string) ops.ServiceOperation {
	return &sdServiceOperation{
		wrapper:  n.wrapper,
		name:     name,
		pathName: path.Join(n.pathName, pathServices, name),
		parentOp: n,
	}
}

func toCoreNamespace(ns *pb.Namespace) *coretypes.Namespace {
	metadata := map[string]string{}
	if ns.Labels != nil {
		metadata = ns.Labels
	}

	return &coretypes.Namespace{
		Name:           path.Base(ns.Name),
		Metadata:       metadata,
		OriginalObject: ns,
	}
}
