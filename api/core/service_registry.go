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
	"time"

	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
)

const (
	// Any is just a way to signal the intention on performing an operation
	// on any namespace/service/endpoint, and is provided just to make your
	// code more readable.
	Any string = ""
	// DefaultCacheExpirationTime is the maximum time an object can stay on
	// cache before being considered stale and thus removed.
	//
	// This can be overridden through options.
	DefaultCacheExpirationTime time.Duration = 5 * time.Minute
	// DefaultCacheCleanUpTime is the frequency that the cache will use to
	// purge expired elements.
	DefaultCacheCleanUpTime time.Duration = 10 * time.Minute

	pathNamespaces string = "namespaces"
	pathServices   string = "services"
	pathEndpoints  string = "endpoints"
)

// ServiceRegistry is the "root" object and represents a generic service
// registry, abstracting the real one that you want to use and allowing you to
// perform operations using the same API and functions regardless of the actual
// service registry you chose.
//
// In order to work, it must be initialized through wrappers and cannot be used
// directly: use the NewWrapper functions to do so.
type ServiceRegistry struct {
	wrapper ops.ServiceRegistryWrapper
}
