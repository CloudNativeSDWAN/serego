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

package wrapper

import (
	"time"

	srerr "github.com/CloudNativeSDWAN/serego/api/errors"
)

const (
	DefaultCacheExpirationTime time.Duration = 5 * time.Minute
	DefaultCacheCleanUpTime    time.Duration = 10 * time.Minute
)

// Options to fine tune the behavior of the Service Registry API.
type Options struct {
	// CacheExpirationTime defines the time after which an element will be
	// considered expired by the cache.
	CacheExpirationTime time.Duration
	// CacheCleanUpTime defines the frequency with the cache will delete
	// expired elements.
	CacheCleanUpTime time.Duration
	// Region where to register all resources in the service registry.
	//
	// This is *required* for Google Service Directory, and ignored by all
	// other service registries.
	Region string
	// Project ID for the service registry.
	//
	// This is *required* for Google Service Directory, and ignored by all
	// other service registries.
	ProjectID string
}

type Option func(*Options) error

// WithNoCache instructs the API to never use cache, but rather always
// perform calls to the service registry.
//
// Note that this may cause latencies and slow performance in some situations.
//
// For example:
// 	sd, err := core.NewServiceRegistryFromMyProvider(myClient, wrapper.WithNoCache())
func WithNoCache() Option {
	return func(wo *Options) error {
		wo.CacheExpirationTime = 0
		return nil
	}
}

// WithCacheExpirationTime instructs the API to use a custom cache expiration
// time instead of the default one, which is 5 minutes.
//
// Keep in mind that a long time could potentially cause the API to retain
// inaccurate or outdated values in case other systems/applications are
// updating the service registry as well.
//
// Do not use this function if you want to disable cache entirely, but rather
// use WithNoCache.
//
// For example:
// 	sd, err := core.NewServiceRegistryFromMyProvider(myClient, wrapper.WithCacheExpirationTime(10*time.Minute))
func WithCacheExpirationTime(expTime time.Duration) Option {
	return func(wo *Options) error {
		if expTime <= 0 {
			return srerr.InvalidCacheExpirationTime
		}

		wo.CacheExpirationTime = expTime
		return nil
	}
}

// WithRegion defines the region where to register all resources in the
// service registry.
//
// This is *required* for Google Service Directory, and ignored by all
// other service registries.
//
// For example:
// 	sd, err := core.NewServiceRegistryFromServiceDirectory(
// 		myClient,
// 		wrapper.WithRegion("us-east1"),
// 		wrapper.WithProjectID("my-project-id"),
// 	)
func WithRegion(region string) Option {
	return func(o *Options) error {
		o.Region = region
		return nil
	}
}

// WithProjectID defines the project ID for Google Service Directory.
//
// This is *required* for Google Service Directory, and ignored by all
// other service registries.
//
// For example:
// 	sd, err := core.NewServiceRegistryFromServiceDirectory(
// 		myClient,
// 		wrapper.WithRegion("us-east1"),
// 		wrapper.WithProjectID("my-project-id"),
// 	)
func WithProjectID(ID string) Option {
	return func(o *Options) error {
		o.ProjectID = ID
		return nil
	}
}
