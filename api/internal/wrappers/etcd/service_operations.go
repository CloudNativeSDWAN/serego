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
	"bytes"
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

type etcdServiceOperation struct {
	wrapper  *EtcdWrapper
	parentOp *etcdNamespaceOperation
	name     string
	pathName string
	kv       clientv3.KV
}

func (s *etcdServiceOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Service, error) {
	if _, err := s.parentOp.Get(ctx, &get.Options{}); err != nil {
		if srerr.IsNotFound(err) {
			// Auto-correct any misconfiguration.
			s.Delete(ctx)
			return nil, srerr.NamespaceNotFound
		}

		return nil, fmt.Errorf(`error while getting parent namespace "%s" before getting service: %w`, s.parentOp.name, err)
	}

	if !opts.ForceRefresh {
		if serv := s.wrapper.getFromCache(s.pathName); serv != nil {
			return serv.(*coretypes.Service), nil
		}
	}

	keyValue, err := getOne(ctx, s.kv, s.name)
	if err != nil {
		return nil, err
	}

	var serv coretypes.Service
	if err := yaml.
		NewDecoder(bytes.NewReader(keyValue.Value)).
		Decode(&serv); err != nil {
		return nil, fmt.Errorf("error while trying to decode the resource: %w", err)
	}
	serv.OriginalObject = keyValue
	s.wrapper.putOnCache(s.pathName, &serv)

	return &serv, nil
}

func (s *etcdServiceOperation) Create(ctx context.Context, metadata map[string]string) (*coretypes.Service, error) {
	// Does the namespace exist, though?
	if _, err := s.parentOp.Get(ctx, &get.Options{}); err != nil {
		return nil, fmt.Errorf(`error while getting parent namespace "%s" before creating service: %w`, s.parentOp.name, err)
	}

	if metadata == nil {
		metadata = map[string]string{}
	}

	nsBytes, _ := yaml.Marshal(&coretypes.Service{
		Name:      s.name,
		Namespace: s.parentOp.name,
		Metadata:  metadata,
	})
	if _, err := s.kv.Put(ctx, prependSlash(s.name), string(nsBytes)); err != nil {
		return nil, err
	}

	return s.Get(ctx, &get.Options{})
}

func (s *etcdServiceOperation) Update(ctx context.Context, metadata map[string]string) (*coretypes.Service, error) {
	return s.Create(ctx, metadata)
}

func (s *etcdServiceOperation) Delete(ctx context.Context) error {
	defer s.wrapper.cache.Delete(s.pathName)

	// TODO: as of now, we delete all children of this. In future we will
	// return an error if resource is not empty and an option to override
	// it anyways.
	if _, err := s.kv.Delete(ctx, prependSlash(s.name), clientv3.WithPrefix()); err != nil {
		return err
	}

	return nil
}

func (s *etcdServiceOperation) List(opts *list.Options) ops.ServiceLister {
	if s.name != "" {
		if opts == nil {
			opts = &list.Options{}
		}

		// Add the name as a filter
		if opts.NameFilters == nil {
			opts.NameFilters = &list.NameFilters{}
		}

		opts.NameFilters.In = append(opts.NameFilters.In, s.name)
	}

	if opts.Results == 0 {
		opts.Results = int32(list.DefaultListResultsNumber)
	}

	return &EtcdServicesIterator{
		wrapper: s.wrapper,
		kv:      s.kv,
		options: opts,
		hasMore: true,
		lastKey: "/",
		nsName:  s.parentOp.name,
	}
}

type EtcdServicesIterator struct {
	wrapper   *EtcdWrapper
	kv        clientv3.KV
	currIndex int
	hasMore   bool
	lastKey   string
	keyValues []*mvccpb.KeyValue
	options   *list.Options
	nsName    string
}

func (si *EtcdServicesIterator) Next(ctx context.Context) (*coretypes.Service, ops.ServiceOperation, error) {
	if si.nsName == "" {
		return nil, nil, fmt.Errorf("cannot get next resource: %w", srerr.EmptyNamespaceName)
	}

	for i := si.currIndex; i < len(si.keyValues); i++ {
		currKeyValue := si.keyValues[i]

		if path.Dir(string(currKeyValue.Key)) != "/" ||
			string(currKeyValue.Key) == si.lastKey {
			continue
		}

		var serv coretypes.Service
		if err := yaml.Unmarshal(currKeyValue.Value, &serv); err != nil {
			// This is not a valid service
			continue
		}
		serv.OriginalObject = currKeyValue

		if passed, _ := si.options.Filter(&serv); passed {
			newOp := si.wrapper.Namespace(serv.Namespace).
				Service(serv.Name).(*etcdServiceOperation)
			si.lastKey = prependSlash(serv.Name)
			si.currIndex = i + 1
			si.wrapper.putOnCache(newOp.pathName, &serv)

			return &serv, newOp, nil
		}
	}

	if si.hasMore {
		values, err := getList(ctx, si.kv, si.lastKey, si.options.Results)
		if err != nil {
			si.hasMore = false
			return nil, nil, fmt.Errorf("could not get next results: %w", err)
		}

		if len(values) < int(si.options.Results) {
			si.hasMore = false
		}

		if len(values) > 0 && si.lastKey == string(values[0].Key) {
			// Skip the first one in this case
			values = values[1:]
		}

		si.keyValues = append(si.keyValues, values...)

		return si.Next(ctx)
	}

	return nil, nil, srerr.IteratorDone
}

func (s *etcdServiceOperation) Endpoint(name string) ops.EndpointOperation {
	return &etcdEndpointOperation{
		name:     name,
		pathName: path.Join(s.pathName, pathEndpoints, name),
		wrapper:  s.wrapper,
		parentOp: s,
		kv:       NewKV(s.kv, path.Join("/", s.name, pathEndpoints)),
	}
}
