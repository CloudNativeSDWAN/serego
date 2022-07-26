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

package fake

import (
	"context"

	coretypes "github.com/CloudNativeSDWAN/serego/api/core/types"
	ops "github.com/CloudNativeSDWAN/serego/api/internal/operations"
	"github.com/CloudNativeSDWAN/serego/api/options/get"
	"github.com/CloudNativeSDWAN/serego/api/options/list"
)

type EndpointOperation struct {
	Name_   string
	Get_    func(context.Context, *get.Options) (*coretypes.Endpoint, error)
	Create_ func(context.Context, string, int32, map[string]string) (*coretypes.Endpoint, error)
	Update_ func(context.Context, string, int32, map[string]string) (*coretypes.Endpoint, error)
	Delete_ func(context.Context) error
	List_   func(*list.Options) ops.EndpointLister
}

func (e *EndpointOperation) Get(ctx context.Context, opts *get.Options) (*coretypes.Endpoint, error) {
	return e.Get_(ctx, opts)
}

func (e *EndpointOperation) Create(ctx context.Context, address string, port int32, metadata map[string]string) (*coretypes.Endpoint, error) {
	return e.Create_(ctx, address, port, metadata)
}

func (e *EndpointOperation) Update(ctx context.Context, address string, port int32, metadata map[string]string) (*coretypes.Endpoint, error) {
	return e.Update_(ctx, address, port, metadata)
}

func (e *EndpointOperation) Delete(ctx context.Context) error {
	return e.Delete_(ctx)
}

func (e *EndpointOperation) List(opts *list.Options) ops.EndpointLister {
	return e.List_(opts)
}

type FakeEndpointIterator struct {
	Next_ func(ctx context.Context) (*coretypes.Endpoint, ops.EndpointOperation, error)
}

func (e *FakeEndpointIterator) Next(ctx context.Context) (*coretypes.Endpoint, ops.EndpointOperation, error) {
	return e.Next_(ctx)
}
