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
	"path"
	"reflect"

	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/options/wrapper"
	"github.com/patrickmn/go-cache"
)

const (
	pathProjects   string = "projects"
	pathLocations  string = "locations"
	pathNamespaces string = "namespaces"
	pathServices   string = "services"
	pathEndpoints  string = "endpoints"
)

type GoogleServiceDirectoryWrapper struct {
	// client *servicedirectory.RegistrationClient
	// We're using regClient instead of *servicedirectory.RegistrationClient
	// for testing purposes
	client   regClient
	pathName string
	cache    *cache.Cache
}

func NewServiceDirectoryWrapper(client regClient, wopts *wrapper.Options) (*GoogleServiceDirectoryWrapper, error) {
	if reflect.ValueOf(client).IsNil() {
		return nil, srerr.NoClientProvided
	}
	if wopts.ProjectID == "" {
		return nil, srerr.NoProjectIDSet
	}
	if wopts.Region == "" {
		return nil, srerr.NoLocationSet
	}

	return &GoogleServiceDirectoryWrapper{
		client:   client,
		pathName: path.Join(pathProjects, wopts.ProjectID, pathLocations, wopts.Region),
		cache: func() *cache.Cache {
			if wopts.CacheExpirationTime == 0 {
				return nil
			}

			return cache.New(wopts.CacheExpirationTime, wrapper.DefaultCacheCleanUpTime)
		}(),
	}, nil
}

func (g *GoogleServiceDirectoryWrapper) putOnCache(pathName string, object interface{}) {
	if g.cache != nil {
		g.cache.SetDefault(pathName, object)
	}
}

func (g *GoogleServiceDirectoryWrapper) getFromCache(pathName string) interface{} {
	if g.cache == nil {
		return nil
	}

	object, found := g.cache.Get(pathName)
	if !found {
		return nil
	}

	return object
}

func (g *GoogleServiceDirectoryWrapper) Namespace(name string) ops.NamespaceOperation {
	return &sdNamespaceOperation{
		name:     name,
		pathName: path.Join(g.pathName, pathNamespaces, name),
		wrapper:  g,
	}
}
