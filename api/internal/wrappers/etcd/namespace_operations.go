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

package etcd

import (
	"context"
	"fmt"
	"path"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

type etcdNamespaceOperation struct {
	name     string
	pathName string
	wrapper  *EtcdWrapper
	kv       clientv3.KV
}

func (n *etcdNamespaceOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Namespace, error) {
	if !opts.ForceRefresh {
		if ns := n.wrapper.getFromCache(n.pathName); ns != nil {
			return ns.(*coretypes.Namespace), nil
		}
	}

	keyValue, err := getOne(ctx, n.kv, n.name)
	if err != nil {
		return nil, err
	}

	var ns coretypes.Namespace
	if err := yaml.Unmarshal(keyValue.Value, &ns); err != nil {
		return nil, fmt.Errorf("error while trying to decode the resource: %w", err)
	}
	ns.OriginalObject = keyValue
	n.wrapper.putOnCache(n.pathName, &ns)

	return &ns, nil
}

func (n *etcdNamespaceOperation) Create(ctx context.Context, metadata map[string]string) (*coretypes.Namespace, error) {
	if metadata == nil {
		metadata = map[string]string{}
	}

	nsBytes, _ := yaml.Marshal(&coretypes.Namespace{
		Name:     n.name,
		Metadata: metadata,
	})
	if _, err := n.kv.Put(ctx, prependSlash(n.name), string(nsBytes)); err != nil {
		return nil, err
	}

	return n.Get(ctx, &get.Options{})
}

func (n *etcdNamespaceOperation) Update(ctx context.Context, metadata map[string]string) (*coretypes.Namespace, error) {
	return n.Create(ctx, metadata)
}

func (n *etcdNamespaceOperation) Delete(ctx context.Context) error {
	defer n.wrapper.cache.Delete(n.pathName)

	// TODO: as of now, we delete all children of this. In future we will
	// return an error if resource is not empty and an option to override
	// it anyways.
	if _, err := n.kv.Delete(ctx, prependSlash(n.name), clientv3.WithPrefix()); err != nil {
		return err
	}

	return nil
}

func (n *etcdNamespaceOperation) List(opts *list.Options) ops.NamespaceLister {
	if n.name != "" {
		if opts == nil {
			opts = &list.Options{}
		}

		// Add the name as a filter
		if opts.NameFilters == nil {
			opts.NameFilters = &list.NameFilters{}
		}

		opts.NameFilters.In = append(opts.NameFilters.In, n.name)
	}

	if opts.Results == 0 {
		opts.Results = int32(list.DefaultListResultsNumber)
	}

	return &EtcdNamespacesIterator{
		wrapper: n.wrapper,
		kv:      n.kv,
		options: opts,
		hasMore: true,
		lastKey: "/",
	}
}

type EtcdNamespacesIterator struct {
	wrapper   *EtcdWrapper
	kv        clientv3.KV
	currIndex int
	hasMore   bool
	lastKey   string
	keyValues []*mvccpb.KeyValue
	options   *list.Options
}

func (ni *EtcdNamespacesIterator) Next(ctx context.Context) (*coretypes.Namespace, ops.NamespaceOperation, error) {
	for i := ni.currIndex; i < len(ni.keyValues); i++ {
		currKeyValue := ni.keyValues[i]

		if path.Dir(string(currKeyValue.Key)) != "/" ||
			string(currKeyValue.Key) == ni.lastKey {
			continue
		}

		var ns coretypes.Namespace
		if err := yaml.Unmarshal(currKeyValue.Value, &ns); err != nil {
			// This is not a valid namespace
			continue
		}
		ns.OriginalObject = currKeyValue

		if passed, _ := ni.options.Filter(&ns); passed {
			newOp := ni.wrapper.Namespace(ns.Name).(*etcdNamespaceOperation)
			ni.lastKey = prependSlash(ns.Name)
			ni.currIndex = i + 1
			ni.wrapper.putOnCache(newOp.pathName, &ns)

			return &ns, newOp, nil
		}
	}

	if ni.hasMore {
		values, err := getList(ctx, ni.kv, ni.lastKey, ni.options.Results)
		if err != nil {
			ni.hasMore = false
			return nil, nil, fmt.Errorf("could not get next results: %w", err)
		}

		if len(values) < int(ni.options.Results) {
			ni.hasMore = false
		}

		if len(values) > 0 && ni.lastKey == string(values[0].Key) {
			// Skip the first one in this case
			values = values[1:]
		}

		ni.keyValues = append(ni.keyValues, values...)

		return ni.Next(ctx)
	}

	return nil, nil, srerr.IteratorDone
}

func (n *etcdNamespaceOperation) Service(name string) ops.ServiceOperation {
	return &etcdServiceOperation{
		name:     name,
		pathName: path.Join(n.pathName, pathServices, name),
		parentOp: n,
		wrapper:  n.wrapper,
		kv:       NewKV(n.kv, path.Join("/", n.name, pathServices)),
	}
}
