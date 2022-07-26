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

// ServiceOperation contains data and code that will be used to perform
// an operation on a given service or on any service - e.g. Get or
// Register - on the chosen service registry.
//
// You should not use this directly: look at
// http://localhost:6060/pkg/github.com/CloudNativeSDWAN/serego/api/core/#NamespaceOperation.Service
type ServiceOperation struct {
	root     *ServiceRegistry
	name     string
	pathName string
	parent   *NamespaceOperation
	op       ops.ServiceOperation
}

// Service denotes the intention to start an operation on a given service
// and thus acts as a factory for a service operation: it takes care of
// initializing, setting up values and checking cache, among other stuff.
func (n *NamespaceOperation) Service(name string) *ServiceOperation {
	return &ServiceOperation{
		name:     name,
		pathName: path.Join(n.pathName, pathServices, name),
		parent:   n,
		op:       n.op.Service(name),
		root:     n.root,
	}
}

// Get tries to retrieve the service.
//
// Unless cache is disabled, it will first check its cache and later the
// service registry if not found there.
func (s *ServiceOperation) Get(ctx context.Context, opts ...get.Option) (*types.Service, error) {
	if err := s.checkNames(); err != nil {
		return nil, err
	}

	getOptions := &get.Options{}
	for _, opt := range opts {
		if err := opt(getOptions); err != nil {
			return nil, err
		}
	}

	serv, err := s.op.Get(ctx, getOptions)
	if err != nil {
		return nil, err
	}

	return serv, nil
}

// Register will insert - or update, if already exists - the service with the
// provided options on the service registry.
func (s *ServiceOperation) Register(ctx context.Context, opts ...register.Option) error {
	if err := s.checkNames(); err != nil {
		return err
	}

	regOpts := &register.Options{}
	for _, opt := range opts {
		if err := opt(regOpts); err != nil {
			return err
		}
	}

	serv, err := s.Get(ctx)
	registerMode, newMetadata, err := prepareRegisterOperation(regOpts, serv, err)
	if err != nil {
		return err
	}

	if registerMode == register.CreateMode {
		_, err = s.op.Create(ctx, newMetadata)
	} else {
		servToCreate := &types.Service{Name: s.name, Namespace: s.parent.name, Metadata: newMetadata}
		if serv.DeepEqualTo(servToCreate) {
			return nil
		}

		_, err = s.op.Update(ctx, newMetadata)
	}

	return err
}

// Deregister removes the service from the service registry and from the
// service operation's internal cache.
//
// By default, it will not return an error if the service does not exist.
// Please read the Deregister operations section to learn more.
func (s *ServiceOperation) Deregister(ctx context.Context, opts ...deregister.Option) error {
	if err := s.checkNames(); err != nil {
		return err
	}

	derOpts := &deregister.Options{}
	for _, opt := range opts {
		if err := opt(derOpts); err != nil {
			return err
		}
	}

	if err := s.op.Delete(ctx); err != nil {
		if srerr.IsNotFound(err) && !derOpts.FailNotExists {
			return nil
		}

		return err
	}

	return nil
}

func (s *ServiceOperation) checkNames() error {
	if s.name == "" {
		return srerr.EmptyServiceName
	}

	return s.parent.checkName()
}

// List returns an iterator that will get a list of namespaces according to the
// options provided.
//
// After you called this function you can iterate through all results with its
// Next function.
func (s *ServiceOperation) List(opts ...list.Option) *ServicesIterator {
	var (
		err      error
		listOpts = &list.Options{Results: list.DefaultListResultsNumber}
	)

	if s.root == nil {
		// If this is initiliazed as &ServiceOperation{} instead of going
		// through a namespace operation then just return it as is: .Next()
		// will handle it.
		return &ServicesIterator{}
	}

	servIterator := &ServicesIterator{
		root:   s.root,
		parent: s.parent,
	}

	// As of now, you *need* to list services on a specific namespace and can't
	// list them on any namespace, so we check that the parent operation
	// (namespace) does have a name.
	if nameErr := s.parent.checkName(); nameErr != nil {
		servIterator.err = nameErr
		return servIterator
	}

	for _, opt := range opts {
		if err = opt(listOpts); err != nil {
			servIterator.err = err
			return servIterator
		}
	}

	servIterator.iterator = s.op.List(listOpts)
	return servIterator
}

// ServicesIterator is in charge of retrieving the list of services, running
// filters and automatically getting the next page of results.
//
// Note that you cannot use this structure directly but you must initialize one
// through the List() function, for example:
// 	iterator := myServiceRegistry.Namespace("sales").Service(core.Any).List()
type ServicesIterator struct {
	root     *ServiceRegistry
	iterator ops.ServiceLister
	parent   *NamespaceOperation
	err      error
}

// Next retrieves the next item from the results that were pulled from the
// service registry according to the options you gave in the List function.
//
// Only results that passed all the filters are returned.
func (si *ServicesIterator) Next(ctx context.Context) (*types.Service, *ServiceOperation, error) {
	if si.root == nil {
		return nil, nil, srerr.UninitializedOperation
	}

	if si.err != nil {
		return nil, nil, si.err
	}

	serv, op, err := si.iterator.Next(ctx)
	if err != nil {
		return nil, nil, err
	}

	// The actual operation -- i.e. the one for Cloud Map or Service Directory
	// may have stored some things to speed up search, i.e. IDs.
	// So we overwrite the operation by using the one returned by Next().
	servOp := si.parent.Service(serv.Name)
	servOp.op = op

	return serv, servOp, nil
}
