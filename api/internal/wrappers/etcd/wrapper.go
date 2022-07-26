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
	"path"

	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
	"github.com/patrickmn/go-cache"
	clientv3 "go.etcd.io/etcd/client/v3"
	etcdns "go.etcd.io/etcd/client/v3/namespace"
)

const (
	pathNamespaces string = "/namespaces"
	pathServices   string = "/services"
	pathEndpoints  string = "/endpoints"
)

type NewKVFunc func(kv clientv3.KV, prefix string) clientv3.KV

var NewKV NewKVFunc

func init() {
	NewKV = etcdns.NewKV
}

type EtcdWrapper struct {
	client *clientv3.Client
	cache  *cache.Cache
}

func NewEtcdWrapper(client *clientv3.Client, wopts *wrapper.Options) (*EtcdWrapper, error) {
	if client == nil {
		return nil, srerr.NoClientProvided
	}

	return &EtcdWrapper{
		client: client,
		cache: func() *cache.Cache {
			if wopts.CacheExpirationTime == 0 {
				return nil
			}

			return cache.New(wopts.CacheExpirationTime, wrapper.DefaultCacheCleanUpTime)
		}(),
	}, nil
}

func (c *EtcdWrapper) putOnCache(pathName string, object interface{}) {
	if c.cache != nil {
		c.cache.SetDefault(pathName, object)
	}
}

func (c *EtcdWrapper) getFromCache(pathName string) interface{} {
	if c.cache == nil {
		return nil
	}

	object, found := c.cache.Get(pathName)
	if !found {
		return nil
	}

	return object
}

func (c *EtcdWrapper) Namespace(name string) ops.NamespaceOperation {
	return &etcdNamespaceOperation{
		name:     name,
		pathName: path.Join(pathNamespaces, name),
		wrapper:  c,
		kv:       NewKV(c.client.KV, pathNamespaces),
	}
}
