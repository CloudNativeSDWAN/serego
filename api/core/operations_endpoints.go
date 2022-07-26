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

package core

import (
	"context"
	"path"

	"github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/options/deregister"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"github.com/CloudNativeSDWAN/serego/api/options/register"
)

// EndpointOperation contains data and code that will be used to perform
// an operation on a given endpoint or on any endpoint - e.g. Get or
// Register - on the chosen service registry.
//
// You should not use this directly: look at
// http://localhost:6060/pkg/github.com/CloudNativeSDWAN/serego/api/core/#ServiceOperation.Endpoint
type EndpointOperation struct {
	name     string
	pathName string
	op       ops.EndpointOperation
	parent   *ServiceOperation
	root     *ServiceRegistry
}

// Endpoint denotes the intention to start an operation on a given endpoint
// and thus acts as a factory for an endpoint operation: it takes care of
// initializing, setting up values and checking cache, among other things.
func (s *ServiceOperation) Endpoint(name string) *EndpointOperation {
	return &EndpointOperation{
		name:     name,
		pathName: path.Join(s.pathName, pathEndpoints, name),
		parent:   s,
		op:       s.op.Endpoint(name),
		root:     s.root,
	}
}

// Get tries to retrieve the endpoint.
//
// Unless cache is disabled, it will first check its cache and later the
// service registry if not found there.
func (e *EndpointOperation) Get(ctx context.Context, opts ...get.Option) (*types.Endpoint, error) {
	if err := e.checkNames(); err != nil {
		return nil, err
	}

	getOptions := &get.Options{}
	for _, opt := range opts {
		if err := opt(getOptions); err != nil {
			return nil, err
		}
	}

	ep, err := e.op.Get(ctx, getOptions)
	if err != nil {
		return nil, err
	}

	return ep, nil
}

// Register will insert - or update, if already exists - the endpoint with the
// provided options on the service registry.
//
// You can register an endpoint without explicitly providing a name for it
// as long as you also have the GenerateName option enabled, in which case
// a name will be generated for this endpoint starting from its parent service
// name.
func (e *EndpointOperation) Register(ctx context.Context, opts ...register.Option) error {
	if e.root == nil {
		return srerr.UninitializedOperation
	}

	if err := e.parent.checkNames(); err != nil {
		return err
	}

	regOpts := &register.Options{}
	for _, opt := range opts {
		if err := opt(regOpts); err != nil {
			return err
		}
	}

	if e.name == "" {
		if !regOpts.GenerateName ||
			(regOpts.GenerateName && regOpts.RegisterMode == register.UpdateMode) {
			return srerr.EmptyEndpointName
		}

		name := generateRandomName(e.parent.name)
		e.pathName = path.Join(e.pathName, name)
		e.name = name
		e.op = e.parent.op.Endpoint(name)
	}

	ep, err := e.Get(ctx)
	registerMode, newMetadata, err := prepareRegisterOperation(regOpts, ep, err)
	if err != nil {
		return err
	}

	// Reset some values.
	if regOpts.Address == nil {
		regOpts.Address = func() *string {
			var address string
			if ep != nil {
				address = ep.Address
			}

			return &address
		}()
	}

	if regOpts.Port == nil {
		regOpts.Port = func() *int32 {
			var port int32
			if ep != nil {
				port = ep.Port
			}

			return &port
		}()
	}

	if registerMode == register.CreateMode {
		_, err = e.op.Create(ctx, *regOpts.Address, *regOpts.Port, newMetadata)
	} else {
		epToUpdate := &types.Endpoint{
			Name:      e.name,
			Service:   e.parent.name,
			Namespace: e.parent.parent.name,
			Address:   *regOpts.Address,
			Port:      *regOpts.Port,
			Metadata:  newMetadata,
		}
		if ep.DeepEqualTo(epToUpdate) {
			return nil
		}

		_, err = e.op.Update(ctx, *regOpts.Address, *regOpts.Port, newMetadata)
	}

	return err
}

// Deregister removes the endpoint from the service registry and from the
// endpoint operation's internal cache.
//
// By default, it will not return an error if the endpoint does not exist.
// Please read the Deregister operations section to learn more.
func (e *EndpointOperation) Deregister(ctx context.Context, opts ...deregister.Option) error {
	if err := e.checkNames(); err != nil {
		return err
	}

	derOpts := &deregister.Options{}
	for _, opt := range opts {
		if err := opt(derOpts); err != nil {
			return err
		}
	}

	if err := e.op.Delete(ctx); err != nil {
		if srerr.IsNotFound(err) && !derOpts.FailNotExists {
			return nil
		}

		return err
	}

	return nil
}

func (e *EndpointOperation) checkNames() error {
	if e.name == "" {
		return srerr.EmptyEndpointName
	}

	return e.parent.checkNames()
}

// List returns an iterator that will get a list of endpoints according to the
// options provided.
//
// After you called this function you can iterate through all results with its
// Next function.
func (e *EndpointOperation) List(opts ...list.Option) *EndpointsIterator {
	var (
		err      error
		listOpts = &list.Options{Results: list.DefaultListResultsNumber}
	)

	if e.root == nil {
		// If this is initiliazed as &EndpointOperation{} instead of going
		// through a service operation then just return it as is: .Next()
		// will handle it.
		return &EndpointsIterator{}
	}

	endpIterator := &EndpointsIterator{
		root:   e.root,
		parent: e.parent,
	}

	// As with services, you *need* to specify a service where to list
	// endpoints from, so we check that parent operation (service) does have
	// a name. This will bubble up till the namespace operation and make
	// the same check.
	if namesErr := e.parent.checkNames(); namesErr != nil {
		endpIterator.err = namesErr
		return endpIterator
	}

	for _, opt := range opts {
		if err = opt(listOpts); err != nil {
			endpIterator.err = err
			return endpIterator
		}
	}

	endpIterator.iterator = e.op.List(listOpts)
	return endpIterator
}

// EndpointsIterator is in charge of retrieving the list of endpoints, running
// filters and automatically getting the next page of results.
//
// Note that you must not use this directly but rather initialize one
// through the List() function, for example:
// 	iterator := wrapper.Namespace("sales").Service("support").Endpoint("").List()
type EndpointsIterator struct {
	root     *ServiceRegistry
	iterator ops.EndpointLister
	parent   *ServiceOperation
	err      error
}

// Next retrieves the next item from the results that were pulled from the
// service registry according to the options you gave in the List function.
//
// Only results that passed all the filters are returned.
func (ei *EndpointsIterator) Next(ctx context.Context) (*types.Endpoint, *EndpointOperation, error) {
	if ei.root == nil {
		return nil, nil, srerr.UninitializedOperation
	}

	if ei.err != nil {
		return nil, nil, ei.err
	}

	ep, op, err := ei.iterator.Next(ctx)
	if err != nil {
		return nil, nil, err
	}

	// The actual operation -- i.e. the one for Cloud Map or Service Directory
	// may have stored some things to speed up search, i.e. IDs.
	// So we overwrite the operation by using the one returned by Next().
	epOp := ei.parent.Endpoint(ep.Name)
	epOp.op = op

	return ep, epOp, nil
}
