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

package operations

import (
	"context"

	"github.com/CloudNativeSDWAN/serego/api/core/types"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
)

type ServiceRegistryWrapper interface {
	Namespace(string) NamespaceOperation
}

type NamespaceOperation interface {
	Get(ctx context.Context, opts *get.Options) (*types.Namespace, error)
	Create(ctx context.Context, metadata map[string]string) (*types.Namespace, error)
	Update(ctx context.Context, metadata map[string]string) (*types.Namespace, error)
	Delete(ctx context.Context) error
	List(opts *list.Options) NamespaceLister
	Service(string) ServiceOperation
}

type NamespaceLister interface {
	Next(context.Context) (*types.Namespace, NamespaceOperation, error)
}

type ServiceOperation interface {
	Get(ctx context.Context, opts *get.Options) (*types.Service, error)
	Create(ctx context.Context, metadata map[string]string) (*types.Service, error)
	Update(ctx context.Context, metadata map[string]string) (*types.Service, error)
	Delete(ctx context.Context) error
	List(opts *list.Options) ServiceLister
	Endpoint(string) EndpointOperation
}

type ServiceLister interface {
	Next(context.Context) (*types.Service, ServiceOperation, error)
}

type EndpointOperation interface {
	Get(ctx context.Context, opts *get.Options) (*types.Endpoint, error)
	Create(ctx context.Context, address string, port int32, metadata map[string]string) (*types.Endpoint, error)
	Update(ctx context.Context, address string, port int32, metadata map[string]string) (*types.Endpoint, error)
	Delete(ctx context.Context) error
	List(opts *list.Options) EndpointLister
}

type EndpointLister interface {
	Next(context.Context) (*types.Endpoint, EndpointOperation, error)
}
