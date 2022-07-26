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

package cloudmap

import (
	"path"
	"time"

	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
	"github.com/patrickmn/go-cache"
)

const (
	pathNamespaces string = "namespaces"
	pathServices   string = "services"
	pathEndpoints  string = "endpoints"

	pathID  string = "id"
	pathARN string = "arn"
)

type AwsCloudMapWrapper struct {
	client cloudMapClientIface
	cache  *cache.Cache
}

func NewCloudMapWrapper(client cloudMapClientIface, wopts *wrapper.Options) (*AwsCloudMapWrapper, error) {
	if client == nil {
		return nil, srerr.NoClientProvided
	}

	return &AwsCloudMapWrapper{
		client: client,
		cache: func() *cache.Cache {
			if wopts.CacheExpirationTime == 0 {
				return cache.New(time.Nanosecond, wrapper.DefaultCacheCleanUpTime)
			}

			return cache.New(wopts.CacheExpirationTime, wrapper.DefaultCacheCleanUpTime)
		}(),
	}, nil
}

func (c *AwsCloudMapWrapper) putOnCache(pathName string, object interface{}) {
	if c.cache != nil {
		c.cache.SetDefault(pathName, object)
	}
}

func (c *AwsCloudMapWrapper) getFromCache(pathName string) interface{} {
	if c.cache == nil {
		return nil
	}

	object, found := c.cache.Get(pathName)
	if !found {
		return nil
	}

	return object
}

func (c *AwsCloudMapWrapper) Namespace(name string) ops.NamespaceOperation {
	var (
		pathName string
	)

	if name != "" {
		pathName = path.Join(pathNamespaces, name)
	}

	return &cmNamespaceOperation{
		name:     name,
		pathName: pathName,
		wrapper:  c,
	}
}
