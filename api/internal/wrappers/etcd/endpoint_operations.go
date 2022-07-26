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

type etcdEndpointOperation struct {
	wrapper  *EtcdWrapper
	parentOp *etcdServiceOperation
	name     string
	pathName string
	kv       clientv3.KV
}

func (e *etcdEndpointOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Endpoint, error) {
	if _, err := e.parentOp.Get(ctx, &get.Options{}); err != nil {
		if srerr.IsNotFound(err) {
			// Auto-correct any misconfiguration.
			e.Delete(ctx)
			return nil, srerr.ServiceNotFound
		}

		return nil, fmt.Errorf(`error while getting parent service "%s" before getting endpoint: %w`, e.parentOp.name, err)
	}

	if !opts.ForceRefresh {
		if endp := e.wrapper.getFromCache(e.pathName); endp != nil {
			return endp.(*coretypes.Endpoint), nil
		}
	}

	keyValue, err := getOne(ctx, e.kv, e.name)
	if err != nil {
		return nil, err
	}

	var endp coretypes.Endpoint
	if err := yaml.Unmarshal(keyValue.Value, &endp); err != nil {
		return nil, fmt.Errorf("error while trying to decode the resource: %w", err)
	}
	endp.OriginalObject = keyValue
	e.wrapper.putOnCache(e.pathName, &endp)

	return &endp, nil
}

func (e *etcdEndpointOperation) Create(ctx context.Context, address string, port int32, metadata map[string]string) (*coretypes.Endpoint, error) {
	// Do the parents exist, though?
	if _, err := e.parentOp.Get(ctx, &get.Options{}); err != nil {
		return nil, fmt.Errorf(`error while getting parent service "%s" before creating endpoint: %w`, e.parentOp.name, err)
	}

	if metadata == nil {
		metadata = map[string]string{}
	}

	endpBytes, _ := yaml.Marshal(&coretypes.Endpoint{
		Name:      e.name,
		Namespace: e.parentOp.parentOp.name,
		Service:   e.parentOp.name,
		Address:   address,
		Port:      port,
		Metadata:  metadata,
	})
	if _, err := e.kv.Put(ctx, prependSlash(e.name), string(endpBytes)); err != nil {
		return nil, err
	}

	return e.Get(ctx, &get.Options{})
}

func (e *etcdEndpointOperation) Update(ctx context.Context, address string, port int32, metadata map[string]string) (*coretypes.Endpoint, error) {
	return e.Create(ctx, address, port, metadata)
}

func (e *etcdEndpointOperation) Delete(ctx context.Context) error {
	defer e.wrapper.cache.Delete(e.pathName)

	// Note that WithPrefix() doesn't really matter here, as endpoints are at
	// the bottom of the hierarchy.
	if _, err := e.kv.Delete(ctx, prependSlash(e.name), clientv3.WithPrefix()); err != nil {
		return err
	}

	return nil
}

func (e *etcdEndpointOperation) List(opts *list.Options) ops.EndpointLister {
	if e.name != "" {
		if opts == nil {
			opts = &list.Options{}
		}

		// Add the name as a filter
		if opts.NameFilters == nil {
			opts.NameFilters = &list.NameFilters{}
		}

		opts.NameFilters.In = append(opts.NameFilters.In, e.name)
	}

	if opts.Results == 0 {
		opts.Results = int32(list.DefaultListResultsNumber)
	}

	return &EtcdEndpointsIterator{
		wrapper:  e.wrapper,
		kv:       e.kv,
		options:  opts,
		hasMore:  true,
		lastKey:  "/",
		servName: e.parentOp.name,
		nsName:   e.parentOp.parentOp.name,
	}
}

type EtcdEndpointsIterator struct {
	wrapper   *EtcdWrapper
	kv        clientv3.KV
	currIndex int
	hasMore   bool
	lastKey   string
	keyValues []*mvccpb.KeyValue
	options   *list.Options
	nsName    string
	servName  string
}

func (ei *EtcdEndpointsIterator) Next(ctx context.Context) (*coretypes.Endpoint, ops.EndpointOperation, error) {
	if ei.nsName == "" {
		return nil, nil, srerr.EmptyNamespaceName
	}
	if ei.servName == "" {
		return nil, nil, srerr.EmptyServiceName
	}

	for i := ei.currIndex; i < len(ei.keyValues); i++ {
		currKeyValue := ei.keyValues[i]

		if path.Dir(string(currKeyValue.Key)) != "/" ||
			string(currKeyValue.Key) == ei.lastKey {
			continue
		}

		var endp coretypes.Endpoint
		if err := yaml.Unmarshal(currKeyValue.Value, &endp); err != nil {
			// This is not a valid endpoint
			continue
		}

		if passed, _ := ei.options.Filter(&endp); passed {
			newOp := ei.wrapper.Namespace(endp.Namespace).
				Service(endp.Service).Endpoint(endp.Name).(*etcdEndpointOperation)
			ei.lastKey = prependSlash(endp.Name)
			ei.currIndex = i + 1
			endp.OriginalObject = currKeyValue
			ei.wrapper.putOnCache(newOp.pathName, &endp)

			return &endp, newOp, nil
		}
	}

	if ei.hasMore {
		values, err := getList(ctx, ei.kv, ei.lastKey, ei.options.Results)
		if err != nil {
			ei.hasMore = false
			return nil, nil, fmt.Errorf("could not get next results: %w", err)
		}

		if len(values) < int(ei.options.Results) {
			ei.hasMore = false
		}

		if len(values) > 0 && ei.lastKey == string(values[0].Key) {
			// Skip the first one in this case
			values = values[1:]
		}

		ei.keyValues = append(ei.keyValues, values...)

		return ei.Next(ctx)
	}

	return nil, nil, srerr.IteratorDone
}
