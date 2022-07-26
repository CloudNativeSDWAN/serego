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

// NamespaceOperation contains data and code that will be used to perform
// an operation on a given namespace or on any namespace - e.g. Get or
// Register - on the chosen service registry.
//
// You should not use this directly but rather initialized via the Namespace
// function from the ServiceRegistry struct:
// http://localhost:6060/pkg/github.com/CloudNativeSDWAN/serego/api/core/#ServiceRegistry.Namespace
type NamespaceOperation struct {
	name     string
	pathName string
	op       ops.NamespaceOperation
	root     *ServiceRegistry
}

// Namespace denotes the intention to start an operation on a given namespace
// and thus acts as a factory for a namespace operation: it takes care of
// initializing, setting up values and checking cache, among other things.
func (s *ServiceRegistry) Namespace(name string) *NamespaceOperation {
	return &NamespaceOperation{
		name:     name,
		pathName: path.Join(pathNamespaces, name),
		op:       s.wrapper.Namespace(name),
		root:     s,
	}
}

// Get tries to retrieve the namespace.
//
// Unless cache is disabled, it will first check its cache and later the
// service registry if not found there.
func (n *NamespaceOperation) Get(ctx context.Context, opts ...get.Option) (*types.Namespace, error) {
	if err := n.checkName(); err != nil {
		return nil, err
	}

	getOptions := &get.Options{}
	for _, opt := range opts {
		if err := opt(getOptions); err != nil {
			return nil, err
		}
	}

	ns, err := n.op.Get(ctx, getOptions)
	if err != nil {
		return nil, err
	}

	return ns, nil
}

// Register will insert - or update, if already exists - the namespace with the
// provided options on the service registry.
func (n *NamespaceOperation) Register(ctx context.Context, opts ...register.Option) error {
	if err := n.checkName(); err != nil {
		return err
	}

	regOpts := &register.Options{}
	for _, opt := range opts {
		if err := opt(regOpts); err != nil {
			return err
		}
	}

	ns, err := n.Get(ctx)
	registerMode, newMetadata, err := prepareRegisterOperation(regOpts, ns, err)
	if err != nil {
		return err
	}

	if registerMode == register.CreateMode {
		_, err = n.op.Create(ctx, newMetadata)
	} else {
		nsToUpdate := &types.Namespace{Name: n.name, Metadata: newMetadata}
		if ns.DeepEqualTo(nsToUpdate) {
			// Avoid update if nothing is changed.
			// Note that if cache is enabled, the previous .Get() operation
			// already cached the result, so we can safely return here.
			return nil
		}

		_, err = n.op.Update(ctx, newMetadata)
	}

	return err
}

// Deregister removes the namespace from the service registry and from the
// namespace operation's internal cache.
//
// By default, it will not return an error if the namespace does not exist.
// Please read the Deregister operations section to learn more.
func (n *NamespaceOperation) Deregister(ctx context.Context, opts ...deregister.Option) error {
	if err := n.checkName(); err != nil {
		return err
	}

	derOpts := &deregister.Options{}
	for _, opt := range opts {
		if err := opt(derOpts); err != nil {
			return err
		}
	}

	if err := n.op.Delete(ctx); err != nil {
		if srerr.IsNotFound(err) && !derOpts.FailNotExists {
			return nil
		}

		return err
	}

	return nil
}

// List returns an iterator that will get a list of namespaces according to the
// options provided.
//
// After you called this function you can iterate through all results with its
// Next function.
func (n *NamespaceOperation) List(opts ...list.Option) *NamespacesIterator {
	var (
		err      error
		listOpts = &list.Options{Results: list.DefaultListResultsNumber}
	)

	if n.root == nil {
		// If this is initiliazed as &NamespaceOperation{} instead of going
		// through the wrapper then just return it as is: .Next() will handle
		// it.
		return &NamespacesIterator{}
	}

	for _, opt := range opts {
		if err = opt(listOpts); err != nil {
			break
		}
	}

	var iterator ops.NamespaceLister
	if err == nil {
		iterator = n.op.List(listOpts)
	}

	return &NamespacesIterator{
		root:     n.root,
		iterator: iterator,
		err:      err,
	}
}

func (n *NamespaceOperation) checkName() error {
	if n.name == "" {
		return srerr.EmptyNamespaceName
	}

	return nil
}

// NamespacesIterator is in charge of retrieving the list of namespaces,
// running filters and automatically getting the next page of results.
//
// Note that you cannot use this structure directly but you must initialize one
// through the List() function, for example:
// 	iterator := myServiceRegistry.Namespace(core.Any).List()
type NamespacesIterator struct {
	root     *ServiceRegistry
	iterator ops.NamespaceLister
	err      error
}

// Next retrieves the next item from the results that were pulled from the
// service registry according to the options you gave in the List function.
//
// Only results that passed all the filters are returned.
func (ni *NamespacesIterator) Next(ctx context.Context) (*types.Namespace, *NamespaceOperation, error) {
	if ni.root == nil {
		return nil, nil, srerr.UninitializedOperation
	}

	if ni.err != nil {
		return nil, nil, ni.err
	}

	ns, op, err := ni.iterator.Next(ctx)
	if err != nil {
		return nil, nil, err
	}

	// The actual operation -- i.e. the one for Cloud Map or Service Directory
	// may have stored some things to speed up search, i.e. IDs.
	// So we overwrite the operation by using the one returned by Next().
	nsOp := ni.root.Namespace(ns.Name)
	nsOp.op = op

	return ns, nsOp, nil
}
